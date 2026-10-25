package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestTokenAPIEnforcesOwnershipAndNeverListsSecrets(t *testing.T) {
	f := setup(t)
	alice, bob := f.user("alice"), f.user("bob")
	admin := f.admin("administrator")
	created := f.request(alice, "POST", "/api/tokens", map[string]string{"name": "Desktop"}, 201)
	var token struct {
		store.CodexToken
		Token string `json:"token"`
	}
	if err := json.Unmarshal(created, &token); err != nil || !strings.HasPrefix(token.Token, "ra_codex_") {
		t.Fatalf("bad token response %s", created)
	}
	listed := f.request(alice, "GET", "/api/tokens", nil, 200)
	if bytes.Contains(listed, []byte(token.Token)) {
		t.Fatal("list exposed secret")
	}
	revealed := f.request(alice, "POST", "/api/tokens/"+token.ID+"/reveal", map[string]string{}, 200)
	if !bytes.Contains(revealed, []byte(token.Token)) {
		t.Fatal("owner cannot recover token")
	}
	f.request(bob, "POST", "/api/tokens/"+token.ID+"/reveal", map[string]string{}, 404)
	f.request(admin, "POST", "/api/tokens/"+token.ID+"/reveal", map[string]string{}, 404)
	f.request(alice, "POST", "/api/tokens/"+token.ID+"/save", map[string]string{"token": "wrong"}, 405)
	if data := f.request(bob, "GET", "/api/tokens", nil, 200); string(bytes.TrimSpace(data)) != "[]" {
		t.Fatal("cross-user token exposed", string(data))
	}
	f.request(bob, "DELETE", "/api/tokens/"+token.ID, nil, 404)
	owner, _ := f.db.User("alice")
	listed = f.request(admin, "GET", "/api/admin/tokens?userId="+owner.ID, nil, 200)
	if !bytes.Contains(listed, []byte(token.ID)) || bytes.Contains(listed, []byte(token.Token)) {
		t.Fatal("admin token metadata incorrect", string(listed))
	}
	f.request(admin, "DELETE", "/api/admin/tokens/"+token.ID, nil, 200)
	f.request(alice, "POST", "/api/tokens/"+token.ID+"/reveal", map[string]string{}, 404)
	if _, err := f.db.AuthenticateCodexToken(token.Token); err == nil {
		t.Fatal("admin revoke ineffective")
	}
	f.request(alice, "POST", "/api/tokens", map[string]string{"name": " "}, 400)
}

func TestAgentSetupPreservesPriorTokenAndReturnsOnlyPersonalCredential(t *testing.T) {
	f, _ := gatewayFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, `{}`) }))
	owner := f.user("alice")
	a, invite := f.enroll(owner)
	enrollment := f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": invite, "os": "linux", "arch": "amd64"}, 200)
	var e struct {
		Credential string `json:"credential"`
	}
	json.Unmarshal(enrollment, &e)
	issue := func() string {
		req, _ := http.NewRequest("POST", f.url+"/api/agent/codex?id="+a.ID, strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer "+e.Credential)
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		if res.StatusCode != 200 || bytes.Contains(b, []byte("private-service-key")) {
			t.Fatalf("unsafe or failed device setup: %d %s", res.StatusCode, b)
		}
		var out struct {
			Token   string `json:"token"`
			BaseURL string `json:"baseUrl"`
			Model   string `json:"model"`
		}
		if json.Unmarshal(b, &out) != nil || out.BaseURL != f.url+"/codex/v1" || out.Model != "gpt-6.1-sol" {
			t.Fatalf("bad setup %s", b)
		}
		return out.Token
	}
	first, second := issue(), issue()
	if first != second {
		t.Fatal("retry created another device token")
	}
	if _, err := f.db.AuthenticateCodexToken(first); err != nil {
		t.Fatal("setup issuance invalidated the working local configuration", err)
	}
	if _, err := f.db.AuthenticateCodexToken(second); err != nil {
		t.Fatal(err)
	}
	f.request(owner, "DELETE", "/api/agents/"+a.ID, nil, 200)
	if _, err := f.db.AuthenticateCodexToken(second); err == nil {
		t.Fatal("revoked device token stayed valid")
	}
	if _, err := f.db.AuthenticateCodexToken(first); err == nil {
		t.Fatal("revoked device's earlier token stayed valid")
	}
}

func TestCodexConfigurationDisabledByDefault(t *testing.T) {
	f := setup(t)
	c := f.user("owner")
	body := f.request(c, "GET", "/api/codex", nil, 200)
	var cfg struct {
		Enabled bool   `json:"enabled"`
		BaseURL string `json:"baseUrl"`
		Model   string `json:"model"`
	}
	if json.Unmarshal(body, &cfg) != nil || cfg.Enabled || cfg.BaseURL != f.url+"/codex/v1" || cfg.Model != "gpt-6.1-sol" {
		t.Fatalf("unexpected disabled config: %s", body)
	}
}
