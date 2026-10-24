package personalgateway

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const MaxBody = 32 << 20

type Server struct {
	RegistryPath string
	Credentials  CredentialSource
	Client       *http.Client
	Upstream     string // fixed by the service administrator, never supplied by a client
	Slots        chan struct{}
	Admin        *Admin
	Accounts     *AccountManager
}

func NewServer(registry string, credentials CredentialSource, proxy string, concurrency int) (*Server, error) {
	if _, err := LoadRegistry(registry); err != nil {
		return nil, err
	}
	if concurrency < 1 || concurrency > 20 {
		return nil, fmt.Errorf("concurrency must be 1-20")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 45 * time.Second
	transport.MaxIdleConnsPerHost = concurrency
	if proxy != "" {
		parsed, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(parsed)
	}
	return &Server{RegistryPath: registry, Credentials: credentials, Client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, Upstream: "https://chatgpt.com/backend-api/codex", Slots: make(chan struct{}, concurrency)}, nil
}

func fail(w http.ResponseWriter, status int, message, kind string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if status == 401 {
		w.Header().Set("WWW-Authenticate", `Bearer realm="sunny-personal"`)
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"message": message, "type": kind}})
}

func (server *Server) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if strings.HasPrefix(request.URL.Path, "/internal/codex/") {
		if server.Accounts == nil {
			http.NotFound(w, request)
			return
		}
		server.Accounts.ServeHTTP(w, request)
		return
	}
	if request.URL.Path == "/openai/admin" || strings.HasPrefix(request.URL.Path, "/openai/admin/") {
		if server.Admin == nil {
			fail(w, 503, "Administration is not configured", "gateway_unavailable")
			return
		}
		server.Admin.serve(server, w, request)
		return
	}
	if request.URL.Path == "/healthz" && request.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"status":"ok"}`)
		return
	}
	registry, err := LoadRegistry(server.RegistryPath)
	if err != nil {
		fail(w, 503, "Device authentication is unavailable", "gateway_unavailable")
		return
	}
	auth := request.Header.Get("Authorization")
	if len(request.Header.Values("Authorization")) != 1 || len(auth) < 8 || !strings.EqualFold(auth[:7], "Bearer ") || !registry.Authorize(auth[7:]) {
		fail(w, 401, "A valid personal device token is required", "authentication_error")
		return
	}
	if request.Header.Get("Origin") != "" {
		fail(w, 403, "Browser cross-origin requests are not supported", "permission_error")
		return
	}
	path := request.URL.Path
	if path != "/models" && path != "/responses" && path != "/responses/compact" {
		fail(w, 404, "Only model inference endpoints are available", "not_found_error")
		return
	}
	if path == "/models" && request.Method != "GET" || path != "/models" && request.Method != "POST" {
		fail(w, 405, "Method not supported; use HTTP Responses streaming", "invalid_request_error")
		return
	}
	select {
	case server.Slots <- struct{}{}:
		defer func() { <-server.Slots }()
	default:
		w.Header().Set("Retry-After", "2")
		fail(w, 429, "Personal gateway concurrency limit reached", "rate_limit_error")
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), 30*time.Minute)
	defer cancel()
	var body []byte
	nonstream := false
	if request.Method == "POST" {
		request.Body = http.MaxBytesReader(w, request.Body, MaxBody)
		body, err = io.ReadAll(request.Body)
		if err != nil {
			fail(w, 413, "Request body too large", "invalid_request_error")
			return
		}
		if path == "/responses" {
			body, nonstream, err = prepareResponse(body)
		}
		if err != nil {
			fail(w, 400, err.Error(), "invalid_request_error")
			return
		}
	}
	source := server.Credentials
	if manager, ok := source.(*AccountManager); ok {
		source = manager.ActiveSource()
	}
	credential, err := source.Get(ctx, "")
	if err != nil {
		fail(w, 503, "Codex on the server needs login; contact the owner", "upstream_login_required")
		return
	}
	response, err := server.forward(ctx, request, path, body, credential)
	if err != nil {
		fail(w, 502, "Failed to connect to the Codex backend", "upstream_error")
		return
	}
	if response.StatusCode == 401 {
		io.Copy(io.Discard, io.LimitReader(response.Body, 65536))
		response.Body.Close()
		credential, err = source.Get(ctx, credential.Token)
		if err != nil {
			fail(w, 503, "Server login expired; sign in to Codex again on 139", "upstream_login_required")
			return
		}
		response, err = server.forward(ctx, request, path, body, credential)
		if err != nil {
			fail(w, 502, "Codex backend retry failed", "upstream_error")
			return
		}
	}
	defer response.Body.Close()
	if response.StatusCode == 401 {
		fail(w, 503, "Server credential was rejected; sign in to Codex again on 139", "upstream_login_required")
		return
	}
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		fail(w, 502, "Unexpected upstream redirect was blocked", "upstream_error")
		return
	}
	if path == "/models" && response.StatusCode == 200 {
		server.models(w, response)
		return
	}
	if nonstream && response.StatusCode == 200 {
		server.collect(w, response)
		return
	}
	copyResponseHeaders(w, response)
	w.WriteHeader(response.StatusCode)
	buffer := make([]byte, 16<<10)
	for {
		n, readErr := response.Body.Read(buffer)
		if n > 0 {
			if _, err := w.Write(buffer[:n]); err != nil {
				return
			}
			http.NewResponseController(w).Flush()
		}
		if readErr != nil {
			return
		}
	}
}

func prepareResponse(body []byte) ([]byte, bool, error) {
	var object map[string]json.RawMessage
	if json.Unmarshal(body, &object) != nil || object == nil {
		return nil, false, fmt.Errorf("Expected a JSON object")
	}
	var model string
	if json.Unmarshal(object["model"], &model) != nil || strings.TrimSpace(model) == "" {
		return nil, false, fmt.Errorf("model is required")
	}
	for _, key := range []string{"background", "conversation", "max_output_tokens", "max_tool_calls", "metadata", "moderation", "multi_agent", "prompt", "prompt_cache_retention", "safety_identifier", "temperature", "top_logprobs", "top_p", "truncation", "user", "previous_response_id"} {
		if raw, exists := object[key]; exists && string(raw) != "null" {
			return nil, false, fmt.Errorf("%s is not supported by the Codex subscription route", key)
		}
		delete(object, key)
	}
	if raw, exists := object["store"]; exists && string(raw) != "false" {
		return nil, false, fmt.Errorf("The Codex subscription route requires store=false")
	}
	stream := false
	if raw, exists := object["stream"]; exists && json.Unmarshal(raw, &stream) != nil {
		return nil, false, fmt.Errorf("stream must be boolean")
	}
	object["stream"] = json.RawMessage("true")
	object["store"] = json.RawMessage("false")
	if _, exists := object["instructions"]; !exists {
		object["instructions"] = json.RawMessage(`""`)
	}
	if input, exists := object["input"]; !exists {
		return nil, false, fmt.Errorf("input is required")
	} else {
		var text string
		if json.Unmarshal(input, &text) == nil {
			normalized, _ := json.Marshal([]any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "input_text", "text": text}}}})
			object["input"] = normalized
		} else if len(input) == 0 || input[0] != '[' {
			return nil, false, fmt.Errorf("input must be a string or array")
		}
	}
	prepared, err := json.Marshal(object)
	return prepared, !stream, err
}

func (server *Server) forward(ctx context.Context, incoming *http.Request, path string, body []byte, credential Credential) (*http.Response, error) {
	target := server.Upstream + path
	query := incoming.URL.Query()
	if path == "/models" && query.Get("client_version") == "" {
		query.Set("client_version", "0.156.1")
	}
	// Credential-like URL parameters are not supported by this gateway.
	for _, key := range []string{"access_token", "api_key", "key", "token"} {
		query.Del(key)
	}
	if encoded := query.Encode(); encoded != "" {
		target += "?" + encoded
	}
	request, err := http.NewRequestWithContext(ctx, incoming.Method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// Copy only model-protocol metadata; caller credentials and workspace overrides stay local.
	for _, key := range []string{"OpenAI-Beta", "X-Codex-Turn-Metadata", "X-Codex-Turn-State", "X-Client-Request-Id", "Session-Id", "Conversation-Id"} {
		if value := incoming.Header.Get(key); value != "" {
			request.Header.Set(key, value)
		}
	}
	request.Header.Set("Authorization", "Bearer "+credential.Token)
	request.Header.Set("ChatGPT-Account-Id", credential.AccountID)
	request.Header.Set("User-Agent", "SunnyPersonalGateway/0.1.0")
	request.Header.Set("originator", "sunny_personal_gateway")
	request.Header.Set("Accept", "text/event-stream, application/json")
	if incoming.Method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}
	return server.Client.Do(request)
}

func copyResponseHeaders(w http.ResponseWriter, response *http.Response) {
	for _, key := range []string{"Content-Type", "X-Request-Id", "Retry-After", "X-Codex-Turn-State", "X-RateLimit-Limit-Requests", "X-RateLimit-Remaining-Requests", "X-RateLimit-Reset-Requests"} {
		if value := response.Header.Get(key); value != "" {
			w.Header().Set(key, value)
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
}

func (server *Server) models(w http.ResponseWriter, response *http.Response) {
	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		fail(w, 502, "Failed to read model catalog", "upstream_error")
		return
	}
	var catalog map[string]json.RawMessage
	if json.Unmarshal(data, &catalog) != nil {
		fail(w, 502, "Invalid model catalog", "upstream_error")
		return
	}
	var models []struct {
		Slug string `json:"slug"`
	}
	if json.Unmarshal(catalog["models"], &models) == nil {
		items := []map[string]any{}
		for _, model := range models {
			if model.Slug != "" {
				items = append(items, map[string]any{"id": model.Slug, "object": "model", "owned_by": "openai"})
			}
		}
		catalog["object"] = json.RawMessage(`"list"`)
		catalog["data"], _ = json.Marshal(items)
	}
	copyResponseHeaders(w, response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}

func (server *Server) collect(w http.ResponseWriter, response *http.Response) {
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), MaxBody)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var event struct {
			Type     string          `json:"type"`
			Response json.RawMessage `json:"response"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &event) != nil {
			continue
		}
		if event.Type == "response.completed" && len(event.Response) > 0 {
			copyResponseHeaders(w, response)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(event.Response)))
			w.Write(event.Response)
			return
		}
		if event.Type == "response.failed" || event.Type == "error" {
			fail(w, 502, "The Codex backend could not complete this response", "upstream_error")
			return
		}
	}
	fail(w, 502, "The Codex stream ended without response.completed", "upstream_error")
}
