package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/githubproxy"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const DefaultGitHubProxyListen = "127.0.0.1:17891"

type gitHubProxy struct {
	config Config
	pool   githubproxy.Pool
}

// ListenGitHubProxy never accepts LAN/wildcard listeners or DNS hostnames.
// Port zero is supported for isolated callers and tests.
func ListenGitHubProxy(address string) (net.Listener, error) {
	if err := validateGitHubProxyListen(address); err != nil {
		return nil, err
	}
	return net.Listen("tcp", address)
}

func validateGitHubProxyListen(address string) error {
	host, port, err := net.SplitHostPort(address)
	ip, ipErr := netip.ParseAddr(host)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || ipErr != nil || !ip.IsLoopback() || ip.Zone() != "" || portErr != nil || portNumber < 0 || portNumber > 65535 {
		return errors.New("GitHub proxy must listen on a literal loopback IP and valid port")
	}
	return nil
}

// ServeGitHubProxy owns listener until cancellation. It never rewrites global
// Git/system proxy configuration, and credentials only travel to the bound site.
func ServeGitHubProxy(ctx context.Context, cfg Config, listener net.Listener) error {
	if err := ValidateServer(cfg.Server); err != nil {
		listener.Close()
		return err
	}
	if cfg.ID == "" || cfg.Credential == "" {
		listener.Close()
		return errors.New("GitHub proxy requires an enrolled client")
	}
	p := &gitHubProxy{config: cfg}
	server := &http.Server{
		Handler:           p,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 * 1024,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	stop := context.AfterFunc(ctx, func() { p.pool.Close(); server.Close() })
	defer stop()
	defer p.pool.Close()
	err := server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (p *gitHubProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		w.Header().Set("Allow", http.MethodConnect)
		http.Error(w, "Use this HTTP proxy for HTTPS GitHub URLs (CONNECT).", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Origin") != "" || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		http.Error(w, "browser origin denied", http.StatusForbidden)
		return
	}
	target, err := githubproxy.Target(r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	ctx, release, ok := p.pool.Open(r.Context(), p.config.ID)
	if !ok {
		http.Error(w, "GitHub proxy connection limit reached", http.StatusTooManyRequests)
		return
	}
	defer release()
	u, _ := protocol.ParseServerURL(p.config.Server)
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path += "/api/agent/github-proxy"
	query := u.Query()
	query.Set("id", p.config.ID)
	query.Set("target", target)
	u.RawQuery = query.Encode()
	dialer := websocket.Dialer{HandshakeTimeout: 20 * time.Second, WriteBufferSize: githubproxy.ChunkSize}
	ws, response, err := dialer.DialContext(ctx, u.String(), http.Header{"Authorization": {"Bearer " + p.config.Credential}})
	if err != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		http.Error(w, "GitHub relay unavailable; check the client binding and server version", http.StatusBadGateway)
		return
	}
	defer ws.Close()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "CONNECT unavailable", http.StatusInternalServerError)
		return
	}
	conn, buffered, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Time{})
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.WriteString(buffered, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil || buffered.Flush() != nil {
		return
	}
	githubproxy.Tunnel(ctx, ws, conn, buffered.Reader, nil)
}

func (a *Agent) startGitHubProxy(ctx context.Context) func() {
	a.mu.Lock()
	cfg := a.Config
	a.mu.Unlock()
	managerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		var stopProxy func()
		active := ""
		defer func() {
			if stopProxy != nil {
				stopProxy()
			}
		}()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			current := cfg
			address, settingErr := GitHubProxyAddress(cfg.Path)
			if cfg.Path != "" {
				loaded, err := Load(cfg.Path)
				if err != nil || loaded.ID != cfg.ID || loaded.Credential != cfg.Credential || loaded.Server != cfg.Server {
					address = ""
				} else {
					current = loaded
				}
			}
			if settingErr != nil {
				address = ""
			}
			if address != active {
				if stopProxy != nil {
					stopProxy()
					stopProxy = nil
				}
				active = address
				if address != "" {
					listener, err := ListenGitHubProxy(address)
					if err != nil {
						log.Printf("GitHub proxy unavailable at %s: %v", address, err)
					} else {
						proxyCtx, stop := context.WithCancel(managerCtx)
						proxyDone := make(chan struct{})
						go func() {
							defer close(proxyDone)
							if err := ServeGitHubProxy(proxyCtx, current, listener); err != nil {
								log.Printf("GitHub proxy stopped: %v", err)
							}
						}()
						stopProxy = func() { stop(); <-proxyDone }
						log.Printf("GitHub proxy listening on http://%s via the bound relay", listener.Addr())
					}
				}
			}
			select {
			case <-managerCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { cancel(); <-done }
}

// GitHubProxyAddress returns the configured listener, or empty when disabled.
// A separate settings file keeps stale project/configuration saves from
// re-enabling or disabling the proxy, and contains no device credentials.
func GitHubProxyAddress(configPath string) (string, error) {
	if configPath == "" {
		return "", nil
	}
	data, err := os.ReadFile(configPath + ".github-proxy.json")
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var settings struct {
		Listen string `json:"listen"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", err
	}
	if settings.Listen != "" {
		if err := validateGitHubProxyListen(settings.Listen); err != nil {
			return "", err
		}
	}
	return settings.Listen, nil
}

// SetGitHubProxy updates only the independent proxy settings, preserving an
// exact backup and any unknown settings. Empty disables the service.
func SetGitHubProxy(configPath, address string) (string, error) {
	if address != "" {
		if err := validateGitHubProxyListen(address); err != nil {
			return "", err
		}
	}
	if _, err := Load(configPath); err != nil {
		return "", err
	}
	path := configPath + ".github-proxy.json"
	unlock, err := lockFile(path + ".write.lock")
	if err != nil {
		return "", errors.New("device configuration is being updated; retry shortly")
	}
	defer unlock()
	if info, err := os.Lstat(path); err != nil && !os.IsNotExist(err) {
		return "", err
	} else if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("proxy configuration must not be a symbolic link")
	}
	before, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	document := map[string]json.RawMessage{}
	if len(before) > 0 {
		if err := json.Unmarshal(before, &document); err != nil {
			return "", err
		}
	}
	if document == nil {
		return "", errors.New("invalid proxy settings")
	}
	if address == "" {
		delete(document, "listen")
	} else {
		document["listen"], _ = json.Marshal(address)
	}
	after, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", err
	}
	if bytes.Equal(bytes.TrimSpace(before), after) {
		return "", nil
	}
	backup := ""
	if len(before) > 0 {
		backup = path + ".backup-" + time.Now().UTC().Format("20060102T150405.000000000")
		if err := writeGitHubProxyFile(backup, before); err != nil {
			return "", err
		}
	}
	latest, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return backup, err
	}
	if !bytes.Equal(before, latest) {
		return backup, errors.New("proxy configuration changed; setting was not saved")
	}
	if err := writeGitHubProxyFile(path, after); err != nil {
		return backup, err
	}
	return backup, nil
}

// These settings contain no secrets. Inherit the config directory's access so
// the Windows Service can read user-written toggles; credential-file protection
// deliberately removes that inherited service grant and must not be used here.
func writeGitHubProxyFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".github-proxy-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
