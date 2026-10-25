package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxInferenceBody = 32 * 1024 * 1024
const maxInferenceDuration = 30 * time.Minute
const maxConcurrentInference = 20

type inferenceCall struct{ cancel context.CancelFunc }

func (s *Server) trackInference(userID string, cancel context.CancelFunc) func() {
	call := &inferenceCall{cancel: cancel}
	s.gatewayMu.Lock()
	if s.gatewayCalls[userID] == nil {
		s.gatewayCalls[userID] = map[*inferenceCall]struct{}{}
	}
	s.gatewayCalls[userID][call] = struct{}{}
	s.gatewayMu.Unlock()
	return func() {
		s.gatewayMu.Lock()
		defer s.gatewayMu.Unlock()
		delete(s.gatewayCalls[userID], call)
		if len(s.gatewayCalls[userID]) == 0 {
			delete(s.gatewayCalls, userID)
		}
	}
}

// An empty user ID cancels every request during server shutdown.
func (s *Server) cancelInference(userID string) {
	s.gatewayMu.Lock()
	var calls []*inferenceCall
	for id, active := range s.gatewayCalls {
		if userID == "" || id == userID {
			for call := range active {
				calls = append(calls, call)
			}
		}
	}
	s.gatewayMu.Unlock()
	for _, call := range calls {
		call.cancel()
	}
}

func validateGateway(cfg Config) error {
	if cfg.GatewayURL == "" {
		if cfg.GatewayAdminToken != "" {
			return errors.New("account control requires a model gateway")
		}
		return nil
	}
	u, err := url.Parse(cfg.GatewayURL)
	if err != nil || u.Scheme != "http" || u.Host == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return errors.New("CODEX_GATEWAY_URL must be an HTTP root URL with a literal loopback address")
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || !ip.IsLoopback() {
		return errors.New("CODEX_GATEWAY_URL must use a literal loopback address")
	}
	if strings.TrimSpace(cfg.GatewayToken) == "" || strings.ContainsAny(cfg.GatewayToken, "\r\n") {
		return errors.New("a valid private gateway service token is required")
	}
	if cfg.GatewayAdminToken != "" && (len(cfg.GatewayAdminToken) < 32 || strings.ContainsAny(cfg.GatewayAdminToken, "\r\n") || cfg.GatewayAdminToken == cfg.GatewayToken) {
		return errors.New("account control requires a separate private credential")
	}
	return nil
}

func newGatewayClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	// The loopback broker may refresh credentials and retry its upstream once.
	// Do not cut that work off at the former 30-second header deadline.
	transport.ResponseHeaderTimeout = 3 * time.Minute
	transport.MaxConnsPerHost = maxConcurrentInference
	transport.MaxIdleConnsPerHost = maxConcurrentInference
	transport.IdleConnTimeout = 90 * time.Second
	return &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func modelError(w http.ResponseWriter, status int, message, kind string) {
	w.Header().Set("Cache-Control", "no-store")
	reply(w, status, map[string]any{"error": map[string]string{"message": message, "type": kind}})
}

func credentialQuery(r *http.Request) bool {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return true
	}
	for key := range values {
		key = strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(key))
		for _, sensitive := range []string{"token", "apikey", "authorization", "credential", "password", "account", "organization", "project"} {
			if strings.Contains(key, sensitive) {
				return true
			}
		}
		switch key {
		case "key", "auth", "userid":
			return true
		}
	}
	return false
}

func (s *Server) gateway(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" || r.Header.Get("Sec-Fetch-Site") != "" {
		modelError(w, 403, "browser inference requests are not permitted", "invalid_request_error")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/codex/v1/")
	allowedMethod := ""
	switch path {
	case "models":
		allowedMethod = http.MethodGet
	case "responses", "responses/compact":
		allowedMethod = http.MethodPost
	default:
		modelError(w, 404, "unsupported inference endpoint", "invalid_request_error")
		return
	}
	if r.Method != allowedMethod {
		w.Header().Set("Allow", allowedMethod)
		modelError(w, 405, "method not allowed", "invalid_request_error")
		return
	}
	if credentialQuery(r) {
		modelError(w, 400, "credential and account-routing query parameters are not permitted", "invalid_request_error")
		return
	}
	info, err := s.Store.AuthenticateCodexToken(bearer(r))
	if err != nil {
		modelError(w, 401, "invalid or revoked access token", "authentication_error")
		return
	}
	if s.Config.GatewayURL == "" {
		modelError(w, 503, "Codex gateway is not configured", "api_error")
		return
	}
	select {
	case s.gatewaySlots <- struct{}{}:
		defer func() { <-s.gatewaySlots }()
	default:
		w.Header().Set("Retry-After", "2")
		modelError(w, 429, "gateway is busy; retry shortly", "rate_limit_error")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), maxInferenceDuration)
	defer cancel()
	untrack := s.trackInference(info.UserID, cancel)
	defer untrack()
	// Registration precedes the account recheck: disabling either cancels the
	// registered request or leaves disabled state for this check to reject.
	owner, err := s.Store.UserByID(info.UserID)
	if err != nil || owner.Disabled || ctx.Err() != nil {
		modelError(w, 401, "account is disabled", "authentication_error")
		return
	}
	r = r.WithContext(ctx)
	if r.Method == http.MethodPost && !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		modelError(w, 415, "application/json required", "invalid_request_error")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxInferenceBody)
	requestBody := r.Body
	controller := http.NewResponseController(w)
	ioCanceled := make(chan struct{})
	stopIOCancel := context.AfterFunc(ctx, func() {
		defer close(ioCanceled)
		// HTTP/1 body Close can wait behind a blocked Read. Interrupt the
		// transport first, including a client that stopped reading SSE output.
		now := time.Now()
		_ = controller.SetWriteDeadline(now)
		_ = controller.SetReadDeadline(now)
		_ = requestBody.Close()
	})
	defer func() {
		if !stopIOCancel() {
			<-ioCanceled
		}
	}()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			modelError(w, 413, "request body exceeds 32 MiB", "invalid_request_error")
		} else {
			modelError(w, 400, "cannot read request body", "invalid_request_error")
		}
		return
	}
	target := strings.TrimSuffix(s.Config.GatewayURL, "/") + "/" + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	upstream, err := http.NewRequestWithContext(r.Context(), r.Method, target, bytes.NewReader(body))
	if err != nil {
		modelError(w, 502, "cannot create gateway request", "api_error")
		return
	}
	// An allowlist prevents unknown account-routing or credential override headers
	// from reaching the shared gateway as new clients add headers over time.
	for _, header := range []string{"Content-Type", "Accept", "OpenAI-Beta", "X-Client-Request-ID"} {
		if value := r.Header.Get(header); value != "" {
			upstream.Header.Set(header, value)
		}
	}
	upstream.Header.Set("Authorization", "Bearer "+s.Config.GatewayToken)
	response, err := s.gatewayClient.Do(upstream)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			modelError(w, 504, "inference request timed out", "api_error")
		} else if errors.Is(err, context.Canceled) {
			modelError(w, 401, "inference request canceled", "authentication_error")
		} else {
			modelError(w, 502, "Codex gateway is unavailable", "api_error")
		}
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		status := 502
		kind := "api_error"
		if response.StatusCode == 429 {
			status = 429
			kind = "rate_limit_error"
		}
		modelError(w, status, "Codex gateway could not complete the request", kind)
		return
	}
	for _, header := range []string{"Content-Type", "X-Request-ID"} {
		if value := response.Header.Get(header); value != "" {
			w.Header().Set(header, value)
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(response.StatusCode)
	_ = http.NewResponseController(w).Flush()
	_, _ = io.Copy(flushWriter{w}, response.Body)
}

type flushWriter struct{ w http.ResponseWriter }

func (f flushWriter) Write(p []byte) (int, error) {
	n, err := f.w.Write(p)
	if err == nil {
		err = http.NewResponseController(f.w).Flush()
	}
	return n, err
}
