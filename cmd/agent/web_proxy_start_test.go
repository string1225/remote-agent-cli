package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func proxyStartFixture(t *testing.T, stealFirst bool) agent.Config {
	t.Helper()
	cfg := agent.Config{ID: "port-fixture", Credential: "mock-secret", Server: "https://relay.example.test", Path: filepath.Join(t.TempDir(), "config.json")}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		var servers []*http.Server
		defer func() {
			for _, s := range servers {
				s.Close()
			}
		}()
		seen := ""
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			address, err := agent.WebProxyAddress(cfg.Path)
			if err != nil || address == "" || address == seen {
				continue
			}
			listener, err := agent.ListenWebProxy(address)
			if err != nil {
				t.Error(err)
				return
			}
			seen = address
			id := cfg.ID
			if stealFirst {
				id = "another-application"
				stealFirst = false
			}
			server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(map[string]string{"clientId": id})
			})}
			servers = append(servers, server)
			go server.Serve(listener)
		}
	}()
	t.Cleanup(func() { cancel(); <-done })
	return cfg
}

func TestConfigureWebProxyListenerSwitchesAndReusesConfiguredPort(t *testing.T) {
	cfg := proxyStartFixture(t, false)
	before, _ := os.ReadFile(cfg.Path)
	occupied := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) }))
	defer occupied.Close()
	var output bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	address, err := configureWebProxyListener(ctx, cfg, occupied.Listener.Addr().String(), &output)
	if err != nil {
		t.Fatal(err)
	}
	if address == occupied.Listener.Addr().String() || !strings.Contains(output.String(), "自动切换") || !strings.Contains(output.String(), address) {
		t.Fatalf("conflict was not explained and switched: %s / %s", address, &output)
	}
	saved, err := agent.WebProxyAddress(cfg.Path)
	if err != nil || saved != address {
		t.Fatalf("selected port not saved: %q / %v", saved, err)
	}
	reused, err := configureWebProxyListener(ctx, cfg, "", io.Discard)
	if err != nil || reused != address {
		t.Fatalf("plain on did not reuse configured port: %q / %v", reused, err)
	}
	after, _ := os.ReadFile(cfg.Path)
	if !bytes.Equal(before, after) {
		t.Fatal("device configuration changed")
	}
}

func TestConfigureWebProxyListenerRecoversFromCompetingBind(t *testing.T) {
	cfg := proxyStartFixture(t, true)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var output bytes.Buffer
	address, err := configureWebProxyListener(ctx, cfg, "127.0.0.1:0", &output)
	if err != nil {
		t.Fatalf("competing bind did not recover: %v / %s", err, &output)
	}
	if err := agent.CheckWebProxy(ctx, cfg, address); err != nil {
		t.Fatal(err)
	}
	if strings.Count(output.String(), "自动切换") < 2 {
		t.Fatalf("competing bind was not reported: %s", &output)
	}
}

func TestConfigureWebProxyListenerDoesNotRotateWhenServiceIsStopped(t *testing.T) {
	cfg := agent.Config{ID: "stopped-fixture", Credential: "mock-secret", Server: "https://relay.example.test", Path: filepath.Join(t.TempDir(), "config.json")}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	var output bytes.Buffer
	if _, err := configureWebProxyListener(ctx, cfg, address, &output); err == nil {
		t.Fatal("stopped service reported ready")
	}
	if strings.Contains(output.String(), "自动切换") {
		t.Fatalf("stopped service incorrectly treated as a conflict: %s", &output)
	}
}
