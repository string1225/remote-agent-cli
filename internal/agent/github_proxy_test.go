package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/githubproxy"
)

func startTestGitHubProxy(t *testing.T, cfg Config) (string, context.CancelFunc, <-chan error) {
	t.Helper()
	listener, err := ListenGitHubProxy("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- ServeGitHubProxy(ctx, cfg, listener) }()
	t.Cleanup(cancel)
	return listener.Addr().String(), cancel, done
}

func TestGitHubProxyHTTPSStreamsThroughBoundRelay(t *testing.T) {
	payload := bytes.Repeat([]byte("GitHub release fixture\n"), 40000)
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "release-assets.githubusercontent.com" || r.Header.Get("Authorization") != "Bearer github-user-token" || r.Header.Get("Proxy-Authorization") != "" {
			t.Error("target request or GitHub credentials changed")
		}
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil || !bytes.Equal(body, payload) {
				t.Error("upload changed")
			}
		}
		w.Write(payload)
	}))
	defer upstream.Close()
	var calls atomic.Int32
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/agents/api/agent/github-proxy" || r.URL.Query().Get("id") != "device-fixture" || r.URL.Query().Get("target") != "release-assets.githubusercontent.com:443" || r.Header.Get("Authorization") != "Bearer device-only-secret" || r.Header.Get("Proxy-Authorization") != "" {
			t.Error("wrong relay routing or leaked proxy headers")
		}
		u, _ := url.Parse(upstream.URL)
		conn, err := net.DialTimeout("tcp", u.Host, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			conn.Close()
			return
		}
		githubproxy.Tunnel(r.Context(), ws, conn, conn, nil)
	}))
	defer relay.Close()
	address, cancel, done := startTestGitHubProxy(t, Config{Server: relay.URL + "/agents", ID: "device-fixture", Credential: "device-only-secret"})
	proxyURL, _ := url.Parse("http://" + address)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: upstream.Client().Transport.(*http.Transport).TLSClientConfig.Clone()}
	// The fixture certificate is issued for example.com; keep verification on and
	// use that trusted name while requesting the GitHub host through CONNECT.
	transport.TLSClientConfig.ServerName = "example.com"
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		var upload io.Reader
		if method == http.MethodPost {
			upload = bytes.NewReader(payload)
		}
		req, _ := http.NewRequest(method, "https://release-assets.githubusercontent.com/download", upload)
		req.Header.Set("Authorization", "Bearer github-user-token")
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || !bytes.Equal(body, payload) {
			t.Fatal("streamed response changed", err, len(body))
		}
	}
	if calls.Load() == 0 {
		t.Fatal("relay was bypassed")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("proxy did not stop")
	}
}

func TestGitHubProxyPreservesPipelinedCONNECTBytesAndClosesTunnels(t *testing.T) {
	relayDone := make(chan struct{})
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(relayDone)
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		for {
			kind, b, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if err := ws.WriteMessage(kind, b); err != nil {
				return
			}
		}
	}))
	defer relay.Close()
	address, cancel, _ := startTestGitHubProxy(t, Config{Server: relay.URL, ID: "fixture", Credential: "test-secret"})
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	io.WriteString(conn, "CONNECT github.com:443 HTTP/1.1\r\nHost: github.com:443\r\n\r\npipelined TLS bytes")
	reader := bufio.NewReader(conn)
	res, err := http.ReadResponse(reader, &http.Request{Method: http.MethodConnect})
	if err != nil || res.StatusCode != 200 {
		t.Fatal("CONNECT failed", err)
	}
	b := make([]byte, len("pipelined TLS bytes"))
	if _, err := io.ReadFull(reader, b); err != nil || string(b) != "pipelined TLS bytes" {
		t.Fatal("lost buffered CONNECT data", err, string(b))
	}
	cancel()
	if _, err := reader.ReadByte(); err == nil {
		t.Fatal("active tunnel survived shutdown")
	}
	select {
	case <-relayDone:
	case <-time.After(3 * time.Second):
		t.Fatal("relay connection leaked")
	}
}

func TestGitHubProxyRejectsForbiddenTargetsAndFailsClosed(t *testing.T) {
	var calls atomic.Int32
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "fixture server unavailable", 503)
	}))
	defer relay.Close()
	address, _, _ := startTestGitHubProxy(t, Config{Server: relay.URL, ID: "fixture", Credential: "test-secret"})
	for _, item := range []struct {
		method, target, origin string
		status                 int
	}{
		{"CONNECT", "example.com:443", "", 403},
		{"CONNECT", "github.com.evil.test:443", "", 403},
		{"CONNECT", "127.0.0.1:443", "", 403},
		{"CONNECT", "github.com:22", "", 403},
		{"CONNECT", "github.com:443", "https://evil.test", 403},
		{"GET", "http://github.com/", "", 405},
		{"CONNECT", "github.com:443", "", 502},
	} {
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		host := item.target
		if item.method == "GET" {
			host = "github.com"
		}
		io.WriteString(conn, item.method+" "+item.target+" HTTP/1.1\r\nHost: "+host+"\r\nOrigin: "+item.origin+"\r\n\r\n")
		res, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: item.method})
		if err != nil {
			conn.Close()
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		conn.Close()
		if res.StatusCode != item.status || strings.Contains(string(body), "test-secret") {
			t.Fatalf("%s: got %d: %s", item.target, res.StatusCode, body)
		}
	}
	if calls.Load() != 1 {
		t.Fatal("forbidden request reached relay", calls.Load())
	}
}

func TestGitHubProxyListenerRejectsNetworkExposure(t *testing.T) {
	for _, address := range []string{":17891", "0.0.0.0:17891", "[::]:17891", "192.168.1.10:17891", "localhost:17891", "127.0.0.1:-1", "127.0.0.1:65536", "[::1%zone]:17891"} {
		listener, err := ListenGitHubProxy(address)
		if err == nil {
			listener.Close()
			t.Errorf("accepted unsafe listener %s", address)
		}
	}
}

func TestGitHubProxySwitchPreservesConfigurationAndAppliesWithoutRestart(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Server: "https://relay.example.test/agents", ID: "switch-fixture", Credential: "temporary-secret", Path: filepath.Join(root, "config.json")}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(cfg.Path)
	var document map[string]json.RawMessage
	json.Unmarshal(before, &document)
	document["futureSetting"] = json.RawMessage(`{"preserve":true}`)
	before, _ = json.MarshalIndent(document, "", "  ")
	if err := os.WriteFile(cfg.Path, before, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(cfg.Path)
	initialAddress, settingErr := GitHubProxyAddress(cfg.Path)
	if err != nil || settingErr != nil || initialAddress != "" {
		t.Fatal("old config did not default off", err)
	}
	listener, err := ListenGitHubProxy("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	a := &Agent{Config: loaded}
	stop := a.startGitHubProxy(context.Background())
	defer stop()
	_, err = SetGitHubProxy(cfg.Path, address)
	if err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(cfg.Path)
	if !bytes.Equal(before, original) {
		t.Fatal("original configuration was not preserved")
	}
	var opened net.Conn
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		opened, err = net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("proxy was not enabled without restart", err)
	}
	defer opened.Close()
	onSettings, _ := os.ReadFile(cfg.Path + ".github-proxy.json")
	backup, err := SetGitHubProxy(cfg.Path, "")
	if err != nil {
		t.Fatal(err)
	}
	backedUp, _ := os.ReadFile(backup)
	if !bytes.Equal(onSettings, backedUp) {
		t.Fatal("proxy settings were not backed up")
	}
	deadline = time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if dialErr != nil {
			break
		}
		conn.Close()
		time.Sleep(20 * time.Millisecond)
	}
	if conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond); err == nil {
		conn.Close()
		t.Fatal("proxy still listening after off")
	}
	current, _ := os.ReadFile(cfg.Path)
	json.Unmarshal(current, &document)
	if string(document["futureSetting"]) != `{"preserve":true}` {
		var preserved map[string]bool
		if json.Unmarshal(document["futureSetting"], &preserved) != nil || !preserved["preserve"] {
			t.Fatal("unknown configuration was lost")
		}
	}
	loaded, err = Load(cfg.Path)
	finalAddress, settingErr := GitHubProxyAddress(cfg.Path)
	if err != nil || loaded.ID != cfg.ID || loaded.Credential != cfg.Credential || settingErr != nil || finalAddress != "" {
		t.Fatal("switch changed binding or failed to disable", err)
	}
	if _, err := SetGitHubProxy(cfg.Path, "0.0.0.0:17891"); err == nil {
		t.Fatal("unsafe configuration accepted")
	}
	after, _ := os.ReadFile(cfg.Path)
	if !bytes.Equal(current, after) {
		t.Fatal("failed setting changed configuration")
	}
}
