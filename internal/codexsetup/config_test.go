package codexsetup

import (
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyPreservesSettingsAndSupportsBackgroundClients(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "config.toml")
	before := "model = \"my-model\"\nnotify = [\"notify.exe\"]\n[projects.\"/work\"]\ntrust_level = \"trusted\"\n[model_providers.other]\nname = \"Keep me\"\n"
	if err := os.WriteFile(path, []byte(before), 0600); err != nil {
		t.Fatal(err)
	}
	connection := Connection{BaseURL: "https://example.com/agents/codex/v1", Token: "ra_codex_test-secret", Model: "gpt-6.1-sol"}
	backup, err := Apply(home, connection)
	if err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(backup)
	if string(original) != before {
		t.Fatal("backup lost original configuration")
	}
	data, _ := os.ReadFile(path)
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if config["model"] != "my-model" || config["model_provider"] != "sunny_agents" || config["notify"] == nil {
		t.Fatalf("settings not preserved: %s", data)
	}
	providers := config["model_providers"].(map[string]any)
	provider := providers["sunny_agents"].(map[string]any)
	if providers["other"] == nil || provider["base_url"] != connection.BaseURL || provider["supports_websockets"] != false || provider["requires_openai_auth"] != false {
		t.Fatal("invalid provider configuration")
	}
	headers := provider["http_headers"].(map[string]any)
	if headers["Authorization"] != "Bearer "+connection.Token || provider["env_key"] != nil {
		t.Fatal("GUI/background clients cannot read credential")
	}
	if _, err := Apply(home, connection); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRejectsBrokenConfigurationAndInvalidEndpoint(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "config.toml")
	before := []byte("[broken\n")
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Apply(home, Connection{BaseURL: "https://example.com/agents/codex/v1", Token: "ra_codex_secret", Model: "model"})
	if err == nil || strings.Contains(err.Error(), "ra_codex_secret") {
		t.Fatal("broken TOML accepted or secret leaked")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("broken config was overwritten")
	}
	if _, err := Apply(t.TempDir(), Connection{BaseURL: "http://public.example/codex", Token: "secret"}); err == nil {
		t.Fatal("public HTTP accepted")
	}
}

func TestApplyUpdatesSelectedProfile(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "config.toml")
	before := "profile = 'work'\n[profiles.work]\nmodel_provider = 'old'\nmodel = 'keep-model'\nmodel_reasoning_effort = 'extrahigh'\n[profiles.other]\nmodel_provider = 'other'\n"
	if err := os.WriteFile(path, []byte(before), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(home, Connection{BaseURL: "https://example.com/agents/codex/v1", Token: "test", Model: "default"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	var config map[string]any
	if err := toml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	profiles := config["profiles"].(map[string]any)
	work := profiles["work"].(map[string]any)
	if work["model_provider"] != Provider || work["model"] != "keep-model" || work["model_reasoning_effort"] != "xhigh" || profiles["other"].(map[string]any)["model_provider"] != "other" {
		t.Fatal("selected profile did not switch, or unrelated profile changed")
	}
}
