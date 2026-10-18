package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/server"
	"github.com/string1225/remote-agent-cli/internal/store"
	"github.com/string1225/remote-agent-cli/web"
)

func gatewayFixture(t *testing.T, upstream http.Handler) (*fixture, string) {
	t.Helper()
	up := httptest.NewServer(upstream)
	t.Cleanup(up.Close)
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	s, handler, err := server.New(db, server.Config{PublicURL: "http://localhost:8080/agents", GatewayURL: up.URL, GatewayToken: "private-service-key"}, web.Files)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(handler)
	s.Config.PublicURL = ts.URL + "/agents"
	t.Cleanup(func() { s.Close(); ts.Close(); db.Close() })
	if err := db.CreateUser(store.User{ID: "owner", Username: "owner"}); err != nil {
		t.Fatal(err)
	}
	_, token, err := db.NewCodexToken("owner", "test", "")
	if err != nil {
		t.Fatal(err)
	}
	return &fixture{t, db, s, ts.URL + "/agents"}, token
}

func modelRequest(t *testing.T, base, token, method, path, body string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(method, base+"/codex/v1/"+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	return r
}

func modelDo(t *testing.T, r *http.Request, status int) []byte {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != status {
		t.Fatalf("%s %s: %d want %d: %s", r.Method, r.URL, res.StatusCode, status, b)
	}
	if status >= 400 {
		var v struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(b, &v) != nil || v.Error.Message == "" || v.Error.Type == "" {
			t.Fatalf("invalid model API error: %s", b)
		}
	}
	return b
}

func TestGatewayReplacesCredentialsAndStripsRoutingHeaders(t *testing.T) {
	var calls atomic.Int32
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/responses" || r.Header.Get("Authorization") != "Bearer private-service-key" {
			t.Error("incorrect upstream route or credential", r.URL.Path, r.Header.Get("Authorization"))
		}
		for _, header := range []string{"Cookie", "Proxy-Authorization", "OpenAI-Organization", "OpenAI-Project", "ChatGPT-Account-ID", "X-Account-ID", "X-Api-Key"} {
			if r.Header.Get(header) != "" {
				t.Errorf("forwarded %s", header)
			}
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"model":"gpt-6.1-sol","input":"hello"}` {
			t.Errorf("body changed: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "secret=value")
		io.WriteString(w, `{"id":"ok"}`)
	}))
	r := modelRequest(t, f.url, token, "POST", "responses", `{"model":"gpt-6.1-sol","input":"hello"}`)
	for _, h := range []string{"Cookie", "Proxy-Authorization", "OpenAI-Organization", "OpenAI-Project", "ChatGPT-Account-ID", "X-Account-ID", "X-Api-Key"} {
		r.Header.Set(h, "sensitive")
	}
	modelDo(t, r, 200)
	if calls.Load() != 1 {
		t.Fatal(calls.Load())
	}
}

func TestGatewayRejectsUnapprovedRoutesOriginsAndCredentials(t *testing.T) {
	var calls atomic.Int32
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); io.WriteString(w, `{}`) }))
	for _, tc := range []struct {
		method, path string
		status       int
	}{{"POST", "chat/completions", 404}, {"GET", "responses", 405}, {"DELETE", "models", 405}, {"GET", "models?api_key=other", 400}, {"GET", "models?openai_api_key=other", 400}, {"GET", "models?account_id=other", 400}, {"POST", "responses?access_token=other", 400}} {
		t.Run(tc.method+tc.path, func(t *testing.T) { modelDo(t, modelRequest(t, f.url, token, tc.method, tc.path, `{}`), tc.status) })
	}
	r := modelRequest(t, f.url, token, "GET", "models", "")
	r.Header.Set("Origin", f.url)
	modelDo(t, r, 403)
	modelDo(t, modelRequest(t, f.url, "wrong", "GET", "models", ""), 401)
	list, _ := f.db.CodexTokens("owner")
	if err := f.db.RevokeCodexToken("owner", list[0].ID, false); err != nil {
		t.Fatal(err)
	}
	modelDo(t, modelRequest(t, f.url, token, "GET", "models", ""), 401)
	if calls.Load() != 0 {
		t.Fatal("blocked request reached upstream")
	}
}

func TestGatewayStreamsBeforeUpstreamCompletesAndBoundsConcurrency(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: first\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	client := &http.Client{Timeout: 5 * time.Second}
	responses := []*http.Response{}
	for range 2 {
		res, err := client.Do(modelRequest(t, f.url, token, "POST", "responses", `{"stream":true}`))
		if err != nil {
			t.Fatal("stream did not flush", err)
		}
		responses = append(responses, res)
		defer res.Body.Close()
		line, err := bufio.NewReader(res.Body).ReadString('\n')
		if err != nil || line != "data: first\n" {
			t.Fatal("missing streaming event", line, err)
		}
	}
	modelDo(t, modelRequest(t, f.url, token, "POST", "responses", `{}`), 429)
}

func TestGatewayBlocksRedirectsAndSanitizesUpstreamErrors(t *testing.T) {
	var redirected atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected.Add(1) }))
	defer destination.Close()
	for _, status := range []int{302, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", destination.URL)
				w.WriteHeader(status)
				io.WriteString(w, "private-service-key")
			}))
			body := modelDo(t, modelRequest(t, f.url, token, "GET", "models", ""), 502)
			if bytes.Contains(body, []byte("private-service-key")) {
				t.Fatal("upstream secret leaked")
			}
		})
	}
	if redirected.Load() != 0 {
		t.Fatal("gateway followed redirect")
	}
}

func TestGatewayRejectsOversizedBody(t *testing.T) {
	var calls atomic.Int32
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	modelDo(t, modelRequest(t, f.url, token, "POST", "responses", strings.Repeat("x", 32*1024*1024+1)), 413)
	if calls.Load() != 0 {
		t.Fatal("oversized request reached upstream")
	}
}

func TestGatewayCancelsUpstreamWhenClientClosesStream(t *testing.T) {
	canceled := make(chan struct{})
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: ready\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(canceled)
	}))
	req := modelRequest(t, f.url, token, "POST", "responses", `{"stream":true}`)
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	res.Body.Close()
	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("client cancellation did not reach upstream")
	}
}

func TestGatewayConvertsTransportFailureToModelError(t *testing.T) {
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Close()
	}))
	modelDo(t, modelRequest(t, f.url, token, "GET", "models", ""), 502)
}

func TestDisablingAccountCancelsItsInferenceAndPreservesOtherStreams(t *testing.T) {
	aliceCanceled, bobCanceled := make(chan struct{}), make(chan struct{})
	f, aliceToken := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			io.WriteString(w, `{}`)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: ready\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		if bytes.Contains(body, []byte("alice")) {
			close(aliceCanceled)
		} else {
			close(bobCanceled)
		}
	}))
	admin := f.admin("administrator")
	bob := f.user("bob")
	_ = bob
	bobUser, _ := f.db.User("bob")
	_, bobToken, err := f.db.NewCodexToken(bobUser.ID, "bob", "")
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	aliceResponse, err := client.Do(modelRequest(t, f.url, aliceToken, "POST", "responses", `{"input":"alice","stream":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer aliceResponse.Body.Close()
	bobResponse, err := client.Do(modelRequest(t, f.url, bobToken, "POST", "responses", `{"input":"bob","stream":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer bobResponse.Body.Close()
	f.request(admin, "PATCH", "/api/admin/users/owner", map[string]bool{"disabled": true}, 200)
	select {
	case <-aliceCanceled:
	case <-time.After(time.Second):
		t.Fatal("disabled user's active inference survived")
	}
	select {
	case <-bobCanceled:
		t.Fatal("disabling another account canceled Bob's stream")
	default:
	}
	// Cancellation interrupts the client transport, so a chunked SSE response
	// may end without its final chunk; waiting for the client timeout must fail.
	if _, err := io.ReadAll(aliceResponse.Body); err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("disabled stream failed to end", err)
	}
	// The canceled request releases its concurrency slot while Bob keeps streaming.
	modelDo(t, modelRequest(t, f.url, bobToken, "GET", "models", ""), 200)
	modelDo(t, modelRequest(t, f.url, aliceToken, "GET", "models", ""), 401)
}

func TestInferenceStartRacingAccountDisableDoesNotLeakActiveRequests(t *testing.T) {
	var active atomic.Int32
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active.Add(1)
		defer active.Add(-1)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: ready\n\n")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	admin := f.admin("administrator")
	for range 12 {
		start := make(chan struct{})
		var wg sync.WaitGroup
		for range 2 {
			wg.Go(func() {
				<-start
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				req := modelRequest(t, f.url, token, "POST", "responses", `{"stream":true}`).WithContext(ctx)
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					if ctx.Err() != nil {
						t.Error("inference survived account disable")
					}
					return
				}
				defer res.Body.Close()
				_, err = io.Copy(io.Discard, res.Body)
				if err != nil && ctx.Err() != nil {
					t.Error("active stream survived account disable")
				}
			})
		}
		close(start)
		f.request(admin, "PATCH", "/api/admin/users/owner", map[string]bool{"disabled": true}, 200)
		wg.Wait()
		deadline := time.Now().Add(time.Second)
		for active.Load() != 0 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if active.Load() != 0 {
			t.Fatal("active upstream leaked after disable")
		}
		f.request(admin, "PATCH", "/api/admin/users/owner", map[string]bool{"disabled": false}, 200)
	}
}

func TestDisableInterruptsPartialHTTPUploadsAndReleasesGatewaySlots(t *testing.T) {
	var uploads atomic.Int32
	f, token := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploads.Add(1)
		}
		io.WriteString(w, `{}`)
	}))
	admin := f.admin("administrator")
	if err := f.db.CreateUser(store.User{ID: "other", Username: "other"}); err != nil {
		t.Fatal(err)
	}
	_, otherToken, err := f.db.NewCodexToken("other", "other", "")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(f.url)
	if err != nil {
		t.Fatal(err)
	}
	connections := []net.Conn{}
	for range 2 {
		conn, err := net.DialTimeout("tcp", u.Host, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		connections = append(connections, conn)
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		_, err = fmt.Fprintf(conn, "POST %s/codex/v1/responses HTTP/1.1\r\nHost: %s\r\nAuthorization: Bearer %s\r\nContent-Type: application/json\r\nContent-Length: 4096\r\nExpect: 100-continue\r\n\r\n", u.Path, u.Host, token)
		if err != nil {
			t.Fatal(err)
		}
		// net/http sends 100 Continue only when the handler starts reading the
		// body. Both continuations prove that both proxy slots are occupied.
		res, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusContinue {
			t.Fatalf("upload did not reach body read: %d", res.StatusCode)
		}
		if _, err := io.WriteString(conn, "{"); err != nil {
			t.Fatal(err)
		}
		conn.SetDeadline(time.Time{})
	}
	modelDo(t, modelRequest(t, f.url, otherToken, "GET", "models", ""), http.StatusTooManyRequests)
	f.request(admin, "PATCH", "/api/admin/users/owner", map[string]bool{"disabled": true}, 200)
	client := &http.Client{Timeout: time.Second}
	released := false
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		res, err := client.Do(modelRequest(t, f.url, otherToken, "GET", "models", ""))
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if res.StatusCode == http.StatusOK {
			released = true
			break
		}
		if res.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("unexpected healthy-account status: %d", res.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !released {
		t.Fatal("disabled partial uploads retained gateway slots")
	}
	for _, conn := range connections {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		_, err := io.Copy(io.Discard, conn)
		if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
			t.Fatal("disabled partial-upload connection remained open")
		}
	}
	if uploads.Load() != 0 {
		t.Fatal("incomplete upload reached upstream")
	}
}

func TestGatewayConfigurationRequiresLoopbackAndServiceCredential(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, cfg := range []server.Config{{GatewayURL: "https://example.com", GatewayToken: "key"}, {GatewayURL: "http://127.0.0.1:8088"}, {GatewayURL: "http://127.0.0.1:8088/base", GatewayToken: "key"}, {GatewayURL: "http://localhost:8088", GatewayToken: "key"}} {
		cfg.PublicURL = "http://localhost:8080"
		if _, _, err := server.New(db, cfg, web.Files); err == nil {
			t.Fatalf("unsafe config accepted: %#v", cfg)
		}
	}
}
