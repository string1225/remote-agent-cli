package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/webproxy"
)

const DefaultWebProxyListen = "127.0.0.1:17891"

type webProxy struct {
	config Config
	pool   webproxy.Pool
}

// ListenWebProxy never accepts LAN/wildcard listeners or DNS hostnames.
// Port zero is supported for isolated callers and tests.
func ListenWebProxy(address string) (net.Listener, error) {
	if err := validateWebProxyListen(address); err != nil {
		return nil, err
	}
	return net.Listen("tcp", address)
}

func validateWebProxyListen(address string) error {
	host, port, err := net.SplitHostPort(address)
	ip, ipErr := netip.ParseAddr(host)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || ipErr != nil || !ip.IsLoopback() || ip.Zone() != "" || portErr != nil || portNumber < 0 || portNumber > 65535 {
		return errors.New("web proxy must listen on a literal loopback IP and valid port")
	}
	return nil
}

// ServeWebProxy owns listener until cancellation. It never rewrites global
// Git/system proxy configuration, and credentials only travel to the bound site.
func ServeWebProxy(ctx context.Context, cfg Config, listener net.Listener) error {
	if err := ValidateServer(cfg.Server); err != nil {
		listener.Close()
		return err
	}
	if cfg.ID == "" || cfg.Credential == "" {
		listener.Close()
		return errors.New("web proxy requires an enrolled client")
	}
	p := &webProxy{config: cfg}
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

func (p *webProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && !r.URL.IsAbs() && r.URL.Path == WebProxyHealthPath {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"clientId": p.config.ID, "version": buildinfo.Current()})
		return
	}
	if r.Method != http.MethodConnect {
		p.forwardHTTP(w, r)
		return
	}
	target, err := webproxy.Address(r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	host, _, _ := net.SplitHostPort(target)
	if webproxy.RelayHost(host) {
		if _, err := webproxy.Target(target); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	ctx, release, ok := p.pool.Open(r.Context(), p.config.ID)
	if !ok {
		http.Error(w, "web proxy connection limit reached", http.StatusTooManyRequests)
		return
	}
	defer release()
	upstream, err := p.dial(ctx, target)
	if err != nil {
		log.Printf("web proxy connection failed target=%s: %v", target, err)
		http.Error(w, "web proxy destination unavailable", http.StatusBadGateway)
		return
	}
	defer upstream.Close()
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
	if _, err := buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil || buffered.Flush() != nil {
		return
	}
	_ = conn.SetWriteDeadline(time.Time{})
	webproxy.Stream(ctx, conn, buffered.Reader, upstream)
}

// The route is chosen before dialing. Relay failures must never dial the
// selected host directly. Ordinary destinations use the local network as-is.
func (p *webProxy) dial(ctx context.Context, authority string) (net.Conn, error) {
	target, err := webproxy.Address(authority)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(target)
	if !webproxy.RelayHost(host) {
		dialer := net.Dialer{Timeout: 15 * time.Second}
		return dialer.DialContext(ctx, "tcp", target)
	}
	if _, err := webproxy.Target(target); err != nil {
		return nil, err
	}
	u, _ := protocol.ParseServerURL(p.config.Server)
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path += "/api/agent/web-proxy"
	query := u.Query()
	query.Set("id", p.config.ID)
	query.Set("target", target)
	u.RawQuery = query.Encode()
	dialer := websocket.Dialer{HandshakeTimeout: 20 * time.Second, WriteBufferSize: webproxy.ChunkSize}
	ws, response, err := dialer.DialContext(ctx, u.String(), http.Header{"Authorization": {"Bearer " + p.config.Credential}})
	if err != nil {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		if response != nil {
			return nil, fmt.Errorf("web relay returned HTTP %d; check server DNS and relay logs", response.StatusCode)
		}
		return nil, errors.New("web relay unavailable; check the client binding and server version")
	}
	// net.Pipe gives HTTP transports and CONNECT the same streaming connection
	// interface while the existing WebSocket heartbeat and cancellation run.
	relay, local := net.Pipe()
	go webproxy.Tunnel(ctx, ws, relay, relay, nil)
	return local, nil
}

func (a *Agent) startWebProxy(ctx context.Context) func() {
	a.mu.Lock()
	cfg := a.Config
	a.mu.Unlock()
	managerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		var stopProxy func()
		active := ""
		nextAttempt := time.Time{}
		defer func() {
			if stopProxy != nil {
				stopProxy()
			}
		}()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			current := cfg
			address, settingErr := WebProxyAddress(cfg.Path)
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
				active = ""
				if address != "" {
					if time.Now().Before(nextAttempt) {
						select {
						case <-managerCtx.Done():
							return
						case <-ticker.C:
							continue
						}
					}
					listener, err := ListenWebProxy(address)
					if err != nil {
						nextAttempt = time.Now().Add(3 * time.Second)
						log.Printf("web proxy unavailable at %s: %v", address, err)
					} else {
						active = address
						nextAttempt = time.Time{}
						proxyCtx, stop := context.WithCancel(managerCtx)
						proxyDone := make(chan struct{})
						go func() {
							defer close(proxyDone)
							if err := ServeWebProxy(proxyCtx, current, listener); err != nil {
								log.Printf("web proxy stopped: %v", err)
							}
						}()
						stopProxy = func() { stop(); <-proxyDone }
						log.Printf("web proxy listening on http://%s via the bound relay", listener.Addr())
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

// WebProxyAddress returns the configured listener, or empty when disabled.
// A separate settings file keeps stale project/configuration saves from
// re-enabling or disabling the proxy, and contains no device credentials.
func WebProxyAddress(configPath string) (string, error) {
	if configPath == "" {
		return "", nil
	}
	data, err := os.ReadFile(configPath + ".web-proxy.json")
	if os.IsNotExist(err) {
		data, err = os.ReadFile(configPath + ".github-proxy.json")
	}
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
		if err := validateWebProxyListen(settings.Listen); err != nil {
			return "", err
		}
	}
	return settings.Listen, nil
}

// SetWebProxy updates only the independent proxy settings, preserving an
// exact backup and any unknown settings. Empty disables the service.
func SetWebProxy(configPath, address string) (string, error) {
	if address != "" {
		if err := validateWebProxyListen(address); err != nil {
			return "", err
		}
	}
	if _, err := Load(configPath); err != nil {
		return "", err
	}
	path := configPath + ".web-proxy.json"
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
	newFile := os.IsNotExist(err)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if newFile {
		before, err = os.ReadFile(configPath + ".github-proxy.json")
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
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
	if !newFile && bytes.Equal(bytes.TrimSpace(before), after) {
		return "", nil
	}
	backup := ""
	if len(before) > 0 {
		backup = path + ".backup-" + time.Now().UTC().Format("20060102T150405.000000000")
		if err := writeWebProxyFile(backup, before); err != nil {
			return "", err
		}
	}
	latest, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return backup, err
	}
	expected := before
	if newFile {
		expected = nil
	}
	if !bytes.Equal(expected, latest) {
		return backup, errors.New("proxy configuration changed; setting was not saved")
	}
	if err := writeWebProxyFile(path, after); err != nil {
		return backup, err
	}
	return backup, nil
}

// These settings contain no secrets. Inherit the config directory's access so
// the Windows Service can read user-written toggles; credential-file protection
// deliberately removes that inherited service grant and must not be used here.
func writeWebProxyFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".web-proxy-*")
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
