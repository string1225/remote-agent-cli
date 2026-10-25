package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/store"
	"github.com/string1225/remote-agent-cli/internal/webproxy"
)

func webProxyFixture(t *testing.T) (*Server, string, string) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "proxy.db"))
	if err != nil {
		t.Fatal(err)
	}
	s, _, err := New(db, Config{PublicURL: "http://localhost/agents"}, fstest.MapFS{})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(); db.Close() })
	if err := db.CreateUser(store.User{ID: "owner", Username: "owner"}); err != nil {
		t.Fatal(err)
	}
	_, invite, err := db.Enroll("owner", "proxy fixture", "linux")
	if err != nil {
		t.Fatal(err)
	}
	a, credential, err := db.Redeem(invite, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	return s, a.ID, credential
}

func TestWebProxyRelayAuthenticatesBeforeDialAndRestrictsTargets(t *testing.T) {
	s, id, credential := webProxyFixture(t)
	var calls atomic.Int32
	dial := func(context.Context, string) (net.Conn, error) {
		calls.Add(1)
		return nil, errors.New("fixture dial failure")
	}
	for _, item := range []struct {
		target, token, origin string
		upgrade               bool
		status                int
	}{
		{"github.com:443", "", "", true, 401},
		{"github.com:443", "wrong", "", true, 401},
		{"github.com:443", credential, "https://evil.test", true, 403},
		{"example.com:443", credential, "", true, 403},
		{"github.com.evil.test:443", credential, "", true, 403},
		{"169.254.169.254:443", credential, "", true, 403},
		{"github.com:22", credential, "", true, 403},
		{"github.com:443", credential, "", false, 426},
		{"github.com:443", credential, "", true, 502},
		{"chatgpt.com:443", credential, "", true, 502},
		{"ws.chatgpt.com:443", credential, "", true, 502},
		{"github.com:80", credential, "", true, 502},
		{"chatgpt.com.evil.test:443", credential, "", true, 403},
	} {
		r := httptest.NewRequest("GET", "/api/agent/web-proxy?id="+id+"&target="+url.QueryEscape(item.target), nil)
		r.Header.Set("Authorization", "Bearer "+item.token)
		r.Header.Set("Origin", item.origin)
		if item.upgrade {
			r.Header.Set("Upgrade", "websocket")
			r.Header.Set("Connection", "Upgrade")
		}
		w := httptest.NewRecorder()
		s.serveWebProxy(w, r, dial)
		if w.Code != item.status || strings.Contains(w.Body.String(), credential) {
			t.Fatalf("%s: got %d: %s", item.target, w.Code, w.Body.String())
		}
	}
	if calls.Load() != 4 {
		t.Fatal("denied request dialed network", calls.Load())
	}
	if err := s.Store.Revoke("owner", id); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/api/agent/web-proxy?id="+id+"&target=github.com:443", nil)
	r.Header.Set("Authorization", "Bearer "+credential)
	w := httptest.NewRecorder()
	s.serveWebProxy(w, r, dial)
	if w.Code != 401 || calls.Load() != 4 {
		t.Fatal("revoked device admitted")
	}
}

func TestWebProxyRelayBinaryStreamingAndShutdown(t *testing.T) {
	s, id, credential := webProxyFixture(t)
	echoDone := make(chan struct{})
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.serveWebProxy(w, r, func(ctx context.Context, target string) (net.Conn, error) {
			if target != "raw.githubusercontent.com:443" {
				t.Error("wrong destination", target)
			}
			one, two := net.Pipe()
			go func() { defer close(echoDone); defer two.Close(); io.Copy(two, two) }()
			return one, nil
		})
	}))
	defer relay.Close()
	u := "ws" + strings.TrimPrefix(relay.URL, "http") + "?id=" + id + "&target=raw.githubusercontent.com:443"
	ws, response, err := websocket.DefaultDialer.Dial(u, http.Header{"Authorization": {"Bearer " + credential}})
	if err != nil {
		if response != nil {
			t.Log(response.Status)
		}
		t.Fatal(err)
	}
	defer ws.Close()
	payload := bytes.Repeat([]byte("binary fixture"), 2000)
	if err := ws.WriteMessage(websocket.BinaryMessage, payload); err != nil {
		t.Fatal(err)
	}
	ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	var got []byte
	for len(got) < len(payload) {
		_, part, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, part...)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("relay changed stream", err)
	}
	s.Close()
	if _, _, err := ws.ReadMessage(); err == nil {
		t.Fatal("shutdown left proxy socket open")
	}
	select {
	case <-echoDone:
	case <-time.After(3 * time.Second):
		t.Fatal("upstream leaked after shutdown")
	}
}

func TestWebProxyRelayLimitsPendingDialsAndRechecksCredentials(t *testing.T) {
	s, id, credential := webProxyFixture(t)
	for range webproxy.MaxPerDevice {
		_, release, ok := s.webProxyPool.Open(context.Background(), id)
		if !ok {
			t.Fatal("pool setup failed")
		}
		defer release()
	}
	r := httptest.NewRequest("GET", "/api/agent/web-proxy?id="+id+"&target=github.com:443", nil)
	r.Header.Set("Authorization", "Bearer "+credential)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	w := httptest.NewRecorder()
	s.serveWebProxy(w, r, func(context.Context, string) (net.Conn, error) {
		t.Error("limit allowed dial")
		return nil, errors.New("unexpected")
	})
	if w.Code != 429 {
		t.Fatal("limit returned", w.Code)
	}

	s2, id2, credential2 := webProxyFixture(t)
	r = httptest.NewRequest("GET", "/api/agent/web-proxy?id="+id2+"&target=github.com:443", nil)
	r.Header.Set("Authorization", "Bearer "+credential2)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	w = httptest.NewRecorder()
	s2.serveWebProxy(w, r, func(context.Context, string) (net.Conn, error) {
		if err := s2.Store.Revoke("owner", id2); err != nil {
			t.Fatal(err)
		}
		one, two := net.Pipe()
		two.Close()
		return one, nil
	})
	if w.Code != 401 {
		t.Fatal("revocation during dial was ignored", w.Code)
	}
}
