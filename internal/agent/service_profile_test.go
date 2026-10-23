package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceProfileRetainsPathsWithoutCopyingCredentials(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config.json")
	if err := (Config{Path: config, Server: "http://localhost:8080", ID: "existing-device", Credential: "private-device-token"}).Save(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)
	t.Setenv("CODEX_HOME", filepath.Join(dir, "custom-codex"))
	t.Setenv("RA_CODEX_PATH", filepath.Join(dir, "absent-codex"))
	t.Setenv("RA_QODER_PATH", filepath.Join(dir, "absent-qoder"))
	t.Setenv("OPENAI_API_KEY", "private-model-token")
	t.Setenv("HTTP_PROXY", "http://private-proxy-token@localhost")
	p, err := CaptureServiceProfile(config)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "profile.json")
	if err = p.Save(output); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(output)
	for _, secret := range []string{"private-device-token", "private-model-token", "private-proxy-token"} {
		if strings.Contains(string(data), secret) {
			t.Fatal("service profile leaked credentials")
		}
	}
	got, err := ReadServiceProfile(output)
	if err != nil || got.Config != config || got.Environment["CODEX_HOME"] != filepath.Join(dir, "custom-codex") {
		t.Fatal("original data/tool paths lost", err)
	}
	before, _ := os.ReadFile(config)
	if !strings.Contains(string(before), "private-device-token") {
		t.Fatal("profile capture changed existing configuration")
	}
	p.Environment["OPENAI_API_KEY"] = "secret"
	if p.Save(output) == nil {
		t.Fatal("arbitrary environment credential accepted")
	}
}
