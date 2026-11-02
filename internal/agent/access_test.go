package agent

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCredentialStorageFailureDoesNotTakeClientOffline(t *testing.T) {
	connected := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		select {
		case connected <- struct{}{}:
		default:
		}
		for {
			if _, _, err = conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "access"), []byte("fixture collision"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &Agent{Config: Config{Path: filepath.Join(dir, "config.json"), Server: server.URL, ID: "fixture", Credential: "fixture-only"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	select {
	case <-connected:
	case err := <-done:
		t.Fatalf("credential initialization stopped client: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("client did not connect")
	}
	a.accessOnce.Do(func() {})
	if a.accessErr == nil {
		t.Fatal("credential failure was hidden")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("client did not stop")
	}
}
