package agent

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebProxyReadinessChecksIdentityAndTunnelErrors(t *testing.T) {
	cfg := Config{Server: "https://relay.example.test", ID: "fixture", Credential: "mock-secret"}
	address, _, _ := startTestWebProxy(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := WaitWebProxy(ctx, cfg, address); err != nil {
		t.Fatal(err)
	}
	wrong := cfg
	wrong.ID = "another-client"
	if err := CheckWebProxy(ctx, wrong, address); err == nil {
		t.Fatal("accepted another client's listener")
	}
	failure := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "fixture relay failure", 502) }))
	defer failure.Close()
	if err := ProbeWebProxyTunnel(ctx, failure.Listener.Addr().String(), "github.com:443"); err == nil {
		t.Fatal("failed CONNECT reported ready")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	unused := listener.Addr().String()
	listener.Close()
	if err := CheckWebProxy(ctx, cfg, unused); err == nil {
		t.Fatal("dead listener reported ready")
	}
}
