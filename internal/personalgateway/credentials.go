package personalgateway

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var ErrLoginRequired = errors.New("server Codex login is required")

type Credential struct {
	Token, AccountID string
	Expires          time.Time
}
type CredentialSource interface {
	Get(context.Context, string) (Credential, error)
}

// CodexSource delegates OAuth refresh to the native CLI; it never reads auth.json.
// previous is the rejected token on a 401 retry, or empty for a normal request.
type CodexSource struct {
	Binary, Proxy, Home string
	mu                  sync.Mutex
	cached              Credential
	validUntil          time.Time
	Fetch               func(context.Context, bool) (Credential, error)
}

func (source *CodexSource) Get(ctx context.Context, previous string) (Credential, error) {
	source.mu.Lock()
	defer source.mu.Unlock()
	if ctx.Err() != nil {
		return Credential{}, ctx.Err()
	}
	if source.cached.Token != "" && time.Now().Before(source.validUntil) && (previous == "" || previous != source.cached.Token) {
		return source.cached, nil
	}
	force := previous != ""
	fetch := source.Fetch
	if fetch == nil {
		fetch = source.fetch
	}
	credential, err := fetch(ctx, force)
	if err == nil && !force && credential.Expires.Before(time.Now().Add(time.Minute)) {
		credential, err = fetch(ctx, true)
	}
	if err != nil || credential.Token == "" || credential.AccountID == "" || !credential.Expires.After(time.Now()) {
		source.cached = Credential{}
		source.validUntil = time.Time{}
		return Credential{}, ErrLoginRequired
	}
	source.cached = credential
	source.validUntil = time.Now().Add(time.Minute)
	if until := credential.Expires.Add(-time.Minute); until.Before(source.validUntil) {
		source.validUntil = until
	}
	return credential, nil
}

func (source *CodexSource) fetch(parent context.Context, refresh bool) (Credential, error) {
	ctx, cancel := context.WithTimeout(parent, 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, source.Binary, "app-server")
	cmd.Env = os.Environ()
	if source.Home != "" {
		cmd.Env = append(cmd.Env, "CODEX_HOME="+source.Home)
	}
	if source.Proxy != "" {
		cmd.Env = append(cmd.Env, "HTTP_PROXY="+source.Proxy, "HTTPS_PROXY="+source.Proxy, "http_proxy="+source.Proxy, "https_proxy="+source.Proxy)
	}
	cmd.Stderr = io.Discard
	input, err := cmd.StdinPipe()
	if err != nil {
		return Credential{}, ErrLoginRequired
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		return Credential{}, ErrLoginRequired
	}
	if err := cmd.Start(); err != nil {
		return Credential{}, ErrLoginRequired
	}
	defer func() { input.Close(); cmd.Process.Kill(); cmd.Wait() }()
	encoder := json.NewEncoder(input)
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	request := func(id int, method string, params any) (json.RawMessage, error) {
		if err := encoder.Encode(map[string]any{"id": id, "method": method, "params": params}); err != nil {
			return nil, ErrLoginRequired
		}
		for scanner.Scan() {
			var reply struct {
				ID     *int            `json:"id"`
				Result json.RawMessage `json:"result"`
				Error  json.RawMessage `json:"error"`
			}
			if json.Unmarshal(scanner.Bytes(), &reply) != nil || reply.ID == nil || *reply.ID != id {
				continue
			}
			if len(reply.Error) > 0 && string(reply.Error) != "null" {
				return nil, ErrLoginRequired
			}
			return reply.Result, nil
		}
		return nil, ErrLoginRequired
	}
	_, err = request(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "sunny_personal_gateway", "version": "0.1.0"}, "capabilities": map[string]bool{"experimentalApi": true}})
	if err != nil {
		return Credential{}, err
	}
	if encoder.Encode(map[string]string{"method": "initialized"}) != nil {
		return Credential{}, ErrLoginRequired
	}
	data, err := request(2, "getAuthStatus", map[string]bool{"includeToken": true, "refreshToken": refresh})
	if err != nil {
		return Credential{}, err
	}
	var status struct {
		Method string `json:"authMethod"`
		Token  string `json:"authToken"`
	}
	if json.Unmarshal(data, &status) != nil || status.Method != "chatgpt" {
		return Credential{}, ErrLoginRequired
	}
	return parseCredential(status.Token)
}

func parseCredential(token string) (Credential, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Credential{}, ErrLoginRequired
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Credential{}, ErrLoginRequired
	}
	var claims struct {
		Expires int64 `json:"exp"`
		Auth    struct {
			Account  string `json:"chatgpt_account_id"`
			Fallback string `json:"account_id"`
		} `json:"https://api.openai.com/auth"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return Credential{}, ErrLoginRequired
	}
	account := claims.Auth.Account
	if account == "" {
		account = claims.Auth.Fallback
	}
	if account == "" || claims.Expires == 0 {
		return Credential{}, ErrLoginRequired
	}
	return Credential{Token: token, AccountID: account, Expires: time.Unix(claims.Expires, 0)}, nil
}
