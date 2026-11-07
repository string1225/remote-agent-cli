package agent

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/webproxy"
)

func TestWebProxyDirectTrafficExceedsRelayConnectionQuota(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "direct") }))
	defer upstream.Close()
	address, _, _ := startTestWebProxy(t, Config{ID: "fixture", Credential: "mock-secret", Server: "http://127.0.0.1:1"})
	connect := func() (net.Conn, int) {
		t.Helper()
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		conn.SetDeadline(time.Now().Add(3 * time.Second))
		fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", upstream.Listener.Addr(), upstream.Listener.Addr())
		res, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
		if err != nil {
			conn.Close()
			t.Fatal(err)
		}
		return conn, res.StatusCode
	}
	for i := 0; i < webproxy.MaxPerDevice; i++ {
		conn, status := connect()
		defer conn.Close()
		if status != http.StatusOK {
			t.Fatalf("connection %d: HTTP %d", i, status)
		}
	}
	proxyURL, _ := url.Parse("http://" + address)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	res, err := client.Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("ordinary HTTP blocked by relay quota: HTTP %d", res.StatusCode)
	}
	conn, status := connect()
	defer conn.Close()
	if status != http.StatusOK {
		t.Errorf("ordinary CONNECT blocked by relay quota: HTTP %d", status)
	}
}

func TestWebProxyRelaySaturationDoesNotBlockDirectHTTP(t *testing.T) {
	cfg := Config{ID: "fixture", Credential: "mock-secret", Server: "http://127.0.0.1:1"}
	p := &webProxy{config: cfg}
	for i := 0; i < webproxy.MaxPerDevice; i++ {
		_, release, ok := p.pool.Open(context.Background(), cfg.ID)
		if !ok {
			t.Fatal("fixture quota could not be filled")
		}
		defer release()
	}
	server := httptest.NewServer(p)
	defer server.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "direct") }))
	defer upstream.Close()
	proxyURL, _ := url.Parse(server.URL)
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	res, err := client.Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("relay saturation blocked an unrelated destination: HTTP %d", res.StatusCode)
	}
	res, err = client.Get("http://github.com/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("relay quota was bypassed: HTTP %d", res.StatusCode)
	}
}
