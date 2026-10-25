package agent

import (
	"bufio"
	"bytes"
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
	"github.com/string1225/remote-agent-cli/internal/webproxy"
)

func TestWebProxyPassesOtherHTTPAndHTTPSDirectly(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	var relayCalls atomic.Int32
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		relayCalls.Add(1)
		http.Error(w, "direct destinations must not reach the relay", 503)
	}))
	defer relay.Close()
	payload := bytes.Repeat([]byte("direct post fixture\n"), 20000)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer target-secret" || r.Header.Get("Proxy-Authorization") != "" || r.Header.Get("X-Proxy-Hop") != "" || r.Header.Get("Origin") != "https://browser.example.test" {
			t.Error("proxy changed origin headers or leaked hop-by-hop credentials")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || !bytes.Equal(body, payload) {
			t.Error("direct request body changed", err)
		}
		w.Header().Set("Connection", "X-Response-Hop")
		w.Header().Set("X-Response-Hop", "remove")
		w.Header().Set("X-End-To-End", "preserve")
		w.WriteHeader(http.StatusCreated)
		w.Write(payload)
	})
	plain := httptest.NewServer(handler)
	defer plain.Close()
	secure := httptest.NewTLSServer(handler)
	defer secure.Close()
	address, _, _ := startTestWebProxy(t, Config{Server: relay.URL, ID: "fixture", Credential: "device-only-secret"})
	proxyURL, _ := url.Parse("http://" + address)
	for _, upstream := range []*httptest.Server{plain, secure} {
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL), ProxyConnectHeader: http.Header{"Proxy-Authorization": {"Basic proxy-only"}}}
		if upstream == secure {
			transport.TLSClientConfig = secure.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
		}
		client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
		req, _ := http.NewRequest(http.MethodPost, upstream.URL+"/upload?key=value", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer target-secret")
		req.Header.Set("Origin", "https://browser.example.test")
		if upstream == plain {
			req.Header.Set("Proxy-Authorization", "Basic proxy-only")
			req.Header.Set("Connection", "X-Proxy-Hop")
			req.Header.Set("X-Proxy-Hop", "remove")
		}
		response, err := client.Do(req)
		if err != nil {
			transport.CloseIdleConnections()
			t.Fatal(err)
		}
		got, err := io.ReadAll(response.Body)
		response.Body.Close()
		transport.CloseIdleConnections()
		if err != nil || response.StatusCode != http.StatusCreated || !bytes.Equal(got, payload) || response.Header.Get("X-End-To-End") != "preserve" {
			t.Fatal("direct response changed", err, response.Status)
		}
		// HTTPS bytes stay opaque, so hop-by-hop response stripping applies to
		// ordinary HTTP forwarding only.
		if upstream == plain && response.Header.Get("X-Response-Hop") != "" {
			t.Fatal("forwarded hop-by-hop response header")
		}
	}
	if relayCalls.Load() != 0 {
		t.Fatal("direct request was relayed", relayCalls.Load())
	}
}

func TestWebProxyRelaysChatGPTHTTPAfterDirectRedirect(t *testing.T) {
	var relayCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "auth.chatgpt.com" || r.URL.Path != "/finish" || r.Header.Get("Proxy-Authorization") != "" {
			t.Error("relayed HTTP target changed")
		}
		io.WriteString(w, "relayed chatgpt fixture")
	}))
	defer upstream.Close()
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		relayCalls.Add(1)
		if r.URL.Path != "/agents/api/agent/web-proxy" || r.URL.Query().Get("target") != "auth.chatgpt.com:80" || r.Header.Get("Authorization") != "Bearer device-secret" {
			t.Error("wrong ChatGPT relay route")
		}
		upstreamURL, _ := url.Parse(upstream.URL)
		conn, err := net.DialTimeout("tcp", upstreamURL.Host, time.Second)
		if err != nil {
			t.Error(err)
			return
		}
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			conn.Close()
			return
		}
		webproxy.Tunnel(r.Context(), ws, conn, conn, nil)
	}))
	defer relay.Close()
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://auth.chatgpt.com/finish", http.StatusFound)
	}))
	defer direct.Close()
	address, _, _ := startTestWebProxy(t, Config{Server: relay.URL + "/agents", ID: "fixture", Credential: "device-secret"})
	proxyURL, _ := url.Parse("http://" + address)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	response, err := client.Get(direct.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || string(body) != "relayed chatgpt fixture" || relayCalls.Load() != 1 {
		t.Fatal("redirect was not routed by destination", err, string(body), relayCalls.Load())
	}
}

func TestWebProxyDirectCONNECTPreservesPipeliningAndStops(t *testing.T) {
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	echoDone := make(chan struct{})
	go func() {
		defer close(echoDone)
		conn, err := upstream.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.Copy(conn, conn)
	}()
	// No reachable relay exists. Ordinary destinations must remain usable.
	address, cancel, _ := startTestWebProxy(t, Config{Server: "http://127.0.0.1:1", ID: "fixture", Credential: "device-secret"})
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	io.WriteString(conn, "CONNECT "+upstream.Addr().String()+" HTTP/1.1\r\nHost: "+upstream.Addr().String()+"\r\n\r\ndirect buffered bytes")
	reader := bufio.NewReader(conn)
	response, err := http.ReadResponse(reader, &http.Request{Method: http.MethodConnect})
	if err != nil || response.StatusCode != 200 {
		t.Fatal("direct CONNECT failed", err)
	}
	got := make([]byte, len("direct buffered bytes"))
	if _, err := io.ReadFull(reader, got); err != nil || string(got) != "direct buffered bytes" {
		t.Fatal("direct CONNECT lost data", err)
	}
	cancel()
	if _, err := reader.ReadByte(); err == nil {
		t.Fatal("direct CONNECT survived off")
	}
	select {
	case <-echoDone:
	case <-time.After(3 * time.Second):
		t.Fatal("direct connection leaked")
	}
}

func TestWebProxyReadsLegacySettingsAndWritesIndependentOverride(t *testing.T) {
	cfg := Config{Server: "https://relay.example.test", ID: "fixture", Credential: "device-secret", Path: filepath.Join(t.TempDir(), "config.json")}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	legacyPath := cfg.Path + ".github-proxy.json"
	legacy := []byte(`{"listen":"127.0.0.1:17891","futureFlag":{"keep":true}}`)
	if err := os.WriteFile(legacyPath, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	if address, err := WebProxyAddress(cfg.Path); err != nil || address != DefaultWebProxyListen {
		t.Fatal("legacy setting not retained", address, err)
	}
	backup, err := SetWebProxy(cfg.Path, "")
	if err != nil {
		t.Fatal(err)
	}
	if address, err := WebProxyAddress(cfg.Path); err != nil || address != "" {
		t.Fatal("new off did not override legacy on", address, err)
	}
	original, _ := os.ReadFile(legacyPath)
	backedUp, _ := os.ReadFile(backup)
	if !bytes.Equal(original, legacy) || !bytes.Equal(backedUp, legacy) {
		t.Fatal("legacy settings or recovery backup changed")
	}
	settings, _ := os.ReadFile(cfg.Path + ".web-proxy.json")
	if !strings.Contains(string(settings), "futureFlag") {
		t.Fatal("unknown settings lost")
	}
	// A stale ordinary configuration save must not revert the separate switch.
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	if address, err := WebProxyAddress(cfg.Path); err != nil || address != "" {
		t.Fatal("configuration save reverted proxy switch", address, err)
	}
}

func TestWebProxyPassesHTTPWebSocketUpgradeAndStops(t *testing.T) {
	upstreamDone := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(upstreamDone)
		ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		ws.WriteMessage(websocket.BinaryMessage, []byte("upgrade fixture"))
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer upstream.Close()
	address, cancel, _ := startTestWebProxy(t, Config{Server: "http://127.0.0.1:1", ID: "fixture", Credential: "device-secret"})
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	u, _ := url.Parse(upstream.URL)
	io.WriteString(conn, "GET "+upstream.URL+"/echo HTTP/1.1\r\nHost: "+u.Host+"\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n")
	reader := bufio.NewReader(conn)
	response, err := http.ReadResponse(reader, &http.Request{Method: http.MethodGet})
	if err != nil || response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatal("WebSocket upgrade did not pass through", err)
	}
	frame := make([]byte, 2+len("upgrade fixture"))
	if _, err := io.ReadFull(reader, frame); err != nil || frame[0] != 0x82 || frame[1] != byte(len("upgrade fixture")) || string(frame[2:]) != "upgrade fixture" {
		t.Fatal("WebSocket bytes changed", err)
	}
	cancel()
	select {
	case <-upstreamDone:
	case <-time.After(3 * time.Second):
		t.Fatal("WebSocket survived proxy shutdown")
	}
}
