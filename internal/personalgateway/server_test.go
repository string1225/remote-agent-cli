package personalgateway

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeCredentials struct {
	calls atomic.Int32
	fail  bool
}

func (fake *fakeCredentials) Get(_ context.Context, previous string) (Credential, error) {
	fake.calls.Add(1)
	if fake.fail {
		return Credential{}, ErrLoginRequired
	}
	token := "upstream-secret-original"
	if previous != "" {
		token = "upstream-secret-refreshed"
	}
	return Credential{Token: token, AccountID: "owner-workspace", Expires: time.Now().Add(time.Hour)}, nil
}

func testServer(t *testing.T, upstream http.HandlerFunc) (*Server, string, *fakeCredentials) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "devices.json")
	token, err := AddDevice(path, "pc-one")
	if err != nil {
		t.Fatal(err)
	}
	credentials := &fakeCredentials{}
	server, err := NewServer(path, credentials, "", 2)
	if err != nil {
		t.Fatal(err)
	}
	backend := httptest.NewServer(upstream)
	t.Cleanup(backend.Close)
	server.Upstream = backend.URL
	return server, token, credentials
}

func request(t *testing.T, handler http.Handler, token, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestAuthenticationAndEndpointIsolation(t *testing.T) {
	server, token, credentials := testServer(t, func(w http.ResponseWriter, r *http.Request) { t.Error("Unexpected upstream access") })
	for _, invalid := range []string{"", "wrong-token", token + "x"} {
		if w := request(t, server, invalid, "GET", "/models", ""); w.Code != 401 {
			t.Fatalf("invalid auth: %d", w.Code)
		}
	}
	for _, path := range []string{"/backend-api/wham/accounts/check", "/account/login/start", "/auth.json", "/responses/other-user", "/chat/completions"} {
		if w := request(t, server, token, "GET", path, ""); w.Code != 404 {
			t.Fatalf("unexpected endpoint %s: %d", path, w.Code)
		}
	}
	if credentials.calls.Load() != 0 {
		t.Fatal("Unauthenticated or non-model requests accessed credentials")
	}
	if err := RevokeDevice(server.RegistryPath, "pc-one"); err != nil {
		t.Fatal(err)
	}
	if w := request(t, server, token, "GET", "/models", ""); w.Code != 401 {
		t.Fatal("Revoked device remained authorized")
	}
}

func TestReplaceCredentialAndRemoveCallerIdentity(t *testing.T) {
	server, token, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer upstream-secret-original" || r.Header.Get("ChatGPT-Account-Id") != "owner-workspace" {
			t.Error("Upstream credential not applied")
		}
		for _, key := range []string{"Proxy-Authorization", "Cookie", "OpenAI-Organization", "OpenAI-Project", "X-OpenAI-Account-Routing-Override", "X-Forwarded-For"} {
			if r.Header.Get(key) != "" {
				t.Errorf("Caller header escaped: %s", key)
			}
		}
		if r.Header.Get("X-Client-Request-Id") != "client-trace" {
			t.Error("Protocol trace lost")
		}
		if r.URL.Query().Get("client_version") != "0.156.1" || r.URL.Query().Get("token") != "" {
			t.Error("Unexpected query forwarding")
		}
		w.Header().Set("Set-Cookie", "secret=upstream")
		io.WriteString(w, `{"models":[{"slug":"available-model"}]}`)
	})
	r := httptest.NewRequest("GET", "/models?token=never-forward", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("ChatGPT-Account-Id", "different-owner")
	for _, key := range []string{"Proxy-Authorization", "Cookie", "OpenAI-Organization", "OpenAI-Project", "X-OpenAI-Account-Routing-Override", "X-Forwarded-For"} {
		r.Header.Set(key, "must-not-leave")
	}
	r.Header.Set("X-Client-Request-Id", "client-trace")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("Set-Cookie") != "" {
		t.Fatalf("Unexpected response %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"id":"available-model"`) || strings.Contains(w.Body.String(), "upstream-secret") {
		t.Fatal("Invalid or leaking model catalog")
	}
}

func TestRefreshOn401AndPreserveStreaming(t *testing.T) {
	var calls atomic.Int32
	server, token, credentials := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("Authorization") == "Bearer upstream-secret-original" {
			w.WriteHeader(401)
			io.WriteString(w, "do not expose this credential error")
			return
		}
		var payload map[string]any
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["stream"] != true || payload["store"] != false || payload["instructions"] != "" {
			t.Error("Request not normalized")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Request-Id", "upstream-trace")
		io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"OK\"}\n\n")
		w.(http.Flusher).Flush()
		io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"response-one\",\"output\":[]}}\n\n")
	})
	w := request(t, server, token, "POST", "/responses", `{"model":"available-model","input":"hello","stream":true}`)
	if w.Code != 200 || !w.Flushed || !strings.Contains(w.Body.String(), "response.completed") || strings.Contains(w.Body.String(), "do not expose") {
		t.Fatalf("Invalid stream %d", w.Code)
	}
	if calls.Load() != 2 || credentials.calls.Load() != 2 {
		t.Fatal("Expected exactly one authenticated retry")
	}
	if w.Header().Get("X-Request-Id") != "upstream-trace" {
		t.Fatal("Request ID lost")
	}
}

func TestNonStreamingAndUnsupportedParameters(t *testing.T) {
	server, token, credentials := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"object\":\"response\",\"id\":\"completed\"}}\n\n")
	})
	w := request(t, server, token, "POST", "/responses", `{"model":"available-model","input":"hello"}`)
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || !strings.Contains(w.Body.String(), "completed") {
		t.Fatal("Non-streaming response failed")
	}
	for _, body := range []string{`{"model":"x","input":"x","store":true}`, `{"model":"x","input":"x","temperature":0.5}`, `{"model":"x","input":"x","previous_response_id":"other-device"}`, `{"input":"x"}`} {
		if w := request(t, server, token, "POST", "/responses", body); w.Code != 400 {
			t.Fatalf("Unsupported payload accepted: %s", body)
		}
	}
	if credentials.calls.Load() != 1 {
		t.Fatal("Invalid parameters accessed upstream credentials")
	}
}

func TestNoCredentialLeaksAndRedirects(t *testing.T) {
	server, token, credentials := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://example.invalid/steal")
		w.WriteHeader(307)
	})
	w := request(t, server, token, "GET", "/models", "")
	if w.Code != 502 || w.Header().Get("Location") != "" {
		t.Fatal("Redirect not blocked")
	}
	credentials.fail = true
	w = request(t, server, token, "GET", "/models", "")
	if w.Code != 503 || strings.Contains(w.Body.String(), "upstream-secret") {
		t.Fatal("Login failure leaked credentials")
	}
}

func TestConcurrencyAndIncompleteStream(t *testing.T) {
	server, token, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"type\":\"response.created\"}\n\n")
	})
	server.Slots <- struct{}{}
	server.Slots <- struct{}{}
	if w := request(t, server, token, "GET", "/models", ""); w.Code != 429 {
		t.Fatal("Concurrency not bounded")
	}
	<-server.Slots
	<-server.Slots
	if w := request(t, server, token, "POST", "/responses", `{"model":"x","input":"x"}`); w.Code != 502 {
		t.Fatal("Incomplete stream reported success")
	}
}

func TestTwentyConcurrentRequests(t *testing.T) {
	started := make(chan struct{}, 20)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	original, token, credentials := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-release
		io.WriteString(w, `{"models":[{"slug":"available-model"}]}`)
	})
	t.Cleanup(unblock)
	server, err := NewServer(original.RegistryPath, credentials, "", 20)
	if err != nil {
		t.Fatal(err)
	}
	server.Upstream = original.Upstream
	completed := make(chan int, 20)
	for range 20 {
		go func() {
			completed <- request(t, server, token, "GET", "/models", "").Code
		}()
	}
	deadline := time.After(5 * time.Second)
	for range 20 {
		select {
		case <-started:
		case <-deadline:
			t.Fatal("Twenty requests could not reach the upstream concurrently")
		}
	}
	if w := request(t, server, token, "GET", "/models", ""); w.Code != 429 || w.Header().Get("Retry-After") != "2" {
		t.Fatalf("Overflow request should be rejected with a retry delay: %d", w.Code)
	}
	unblock()
	for range 20 {
		select {
		case code := <-completed:
			if code != 200 {
				t.Errorf("Concurrent request failed: %d", code)
			}
		case <-deadline:
			t.Fatal("Concurrent requests did not finish")
		}
	}
	if w := request(t, server, token, "GET", "/models", ""); w.Code != 200 {
		t.Fatalf("Released slots could not be reused: %d", w.Code)
	}
}

func TestCredentialCacheAndRefreshDeduplication(t *testing.T) {
	var calls int
	source := &CodexSource{Fetch: func(_ context.Context, refresh bool) (Credential, error) {
		calls++
		token := "old"
		if refresh {
			token = "new"
		}
		return Credential{Token: token, AccountID: "owner", Expires: time.Now().Add(time.Hour)}, nil
	}}
	first, err := source.Get(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	source.Get(context.Background(), "")
	second, _ := source.Get(context.Background(), first.Token)
	third, _ := source.Get(context.Background(), first.Token)
	if calls != 2 || second.Token != "new" || third.Token != "new" {
		t.Fatal("Concurrent 401 retry would refresh twice")
	}
	payload := fmt.Sprintf(`{"exp":%d,"https://api.openai.com/auth":{"chatgpt_account_id":"owner"}}`, time.Now().Add(time.Hour).Unix())
	parsed, err := parseCredential("header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature")
	if err != nil || parsed.AccountID != "owner" {
		t.Fatal("Managed ChatGPT token parse failed")
	}
}

func TestRegistryMutationLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	unlock, err := lockRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddDevice(path, "blocked"); err == nil {
		t.Fatal("Concurrent registry mutation was allowed")
	}
	unlock()
	if _, err := AddDevice(path, "available"); err != nil {
		t.Fatal("Lock was not released")
	}
}
