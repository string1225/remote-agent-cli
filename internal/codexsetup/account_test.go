package codexsetup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestPersonalTokenAccountComesFromToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/token/account" || r.Header.Get("Authorization") != "Bearer fixture-personal" || r.URL.RawQuery != "" {
			t.Errorf("unsafe account request: %s", r.URL)
		}
		fmt.Fprintf(w, `{"accountId":"alice-id","accountUsername":"alice","baseUrl":"http://%s/codex/v1","model":"fixture-model"}`, r.Host)
	}))
	defer server.Close()
	c, err := PersonalConnection(server.Client(), server.URL, "fixture-personal")
	if err != nil || c.AccountID != "alice-id" || c.AccountUsername != "alice" || c.Token != "fixture-personal" {
		t.Fatalf("wrong token account: %+v %v", c, err)
	}
}

func TestApplyUsesFullAuthorizationAndDoesNotRotateUnchangedFile(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "config.toml")
	if err := os.WriteFile(path, []byte("profile='work'\nsandbox_mode='read-only'\napproval_policy='on-request'\n[profiles.work]\nsandbox_mode='workspace-write'\napproval_policy='untrusted'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := Connection{BaseURL: "https://example.com/codex/v1", Token: "fixture", Model: "model"}
	if _, err := Apply(home, c); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	for _, settings := range []map[string]any{config, config["profiles"].(map[string]any)["work"].(map[string]any)} {
		if settings["sandbox_mode"] != "danger-full-access" || settings["approval_policy"] != "never" {
			t.Fatal("permission override missing", settings)
		}
	}
	if backup, err := Apply(home, c); err != nil || backup != "" {
		t.Fatal("unchanged setup wrote another backup", backup, err)
	}
}
