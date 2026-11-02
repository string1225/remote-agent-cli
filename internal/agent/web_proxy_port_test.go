package agent

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSelectWebProxyAddressSkipsOccupiedPort(t *testing.T) {
	occupied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "another application", http.StatusBadRequest)
	}))
	defer occupied.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	address, err := SelectWebProxyAddress(ctx, Config{ID: "fixture"}, occupied.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if address == occupied.Listener.Addr().String() {
		t.Fatal("selected another application's occupied port")
	}
	listener, err := ListenWebProxy(address)
	if err != nil {
		t.Fatalf("selected port cannot be bound: %v", err)
	}
	defer listener.Close()
	host, _, _ := net.SplitHostPort(address)
	if host != "127.0.0.1" {
		t.Fatalf("loopback IP changed: %s", address)
	}
}

func TestSelectWebProxyAddressReusesOnlyThisClient(t *testing.T) {
	cfg := Config{ID: "fixture", Server: "https://relay.example.test", Credential: "mock-secret"}
	address, _, _ := startTestWebProxy(t, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	got, err := SelectWebProxyAddress(ctx, cfg, address)
	if err != nil || got != address {
		t.Fatalf("healthy same-client listener moved: %q, %v", got, err)
	}
	cfg.ID = "different-client"
	got, err = SelectWebProxyAddress(ctx, cfg, address)
	if err != nil || got == address {
		t.Fatalf("another client's listener reused: %q, %v", got, err)
	}
}

func TestSelectWebProxyAddressKeepsAvailablePort(t *testing.T) {
	listener, err := ListenWebProxy("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	got, err := SelectWebProxyAddress(context.Background(), Config{ID: "fixture"}, address)
	if err != nil || got != address {
		t.Fatalf("available requested port changed: %q, %v", got, err)
	}
}

func TestSelectWebProxyAddressRejectsUnsafeAddressAndCancellation(t *testing.T) {
	for _, address := range []string{"0.0.0.0:17891", "localhost:17891", "127.0.0.1:65536", "[::]:17891"} {
		if got, err := SelectWebProxyAddress(context.Background(), Config{}, address); err == nil {
			t.Errorf("unsafe address %s accepted as %s", address, got)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SelectWebProxyAddress(ctx, Config{}, "127.0.0.1:17891"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled selection: %v", err)
	}
}
