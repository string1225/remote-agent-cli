package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func TestBindingSetupPreservesDeviceAndReportsAccountWithoutSecret(t *testing.T) {
	token := "ra_codex_fixture-private-token"
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agent/codex" || r.URL.Query().Get("id") != "device" || r.Header.Get("Authorization") != "Bearer fixture-device" {
			t.Errorf("invalid device exchange")
		}
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", 503)
			return
		}
		fmt.Fprintf(w, `{"baseUrl":"http://%s/codex/v1","token":%q,"model":"fixture","accountId":"alice-id","accountUsername":"alice"}`, r.Host, token)
	}))
	defer server.Close()
	root := t.TempDir()
	config := agent.Config{Server: server.URL, ID: "device", Credential: "fixture-device", Path: filepath.Join(root, "config.json")}
	if err := config.Save(); err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(config.Path)
	var output bytes.Buffer
	if err := applyBoundCodex(server.Client(), config, filepath.Join(root, "codex"), &output); err == nil {
		t.Fatal("failed gateway was accepted")
	}
	if err := applyBoundCodex(server.Client(), config, filepath.Join(root, "codex"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "alice") || strings.Contains(output.String(), token) {
		t.Fatal("account missing or token leaked", output.String())
	}
	after, _ := os.ReadFile(config.Path)
	if !bytes.Equal(original, after) {
		t.Fatal("setup changed original device credentials")
	}
	data, _ := os.ReadFile(filepath.Join(root, "codex", "config.toml"))
	if !bytes.Contains(data, []byte(token)) {
		t.Fatal("credential not securely persisted in local config")
	}
}

func TestDefaultCodexSetupUsesBoundAccount(t *testing.T) {
	root := t.TempDir()
	helper := filepath.Join(root, "codex")
	content := "#!/bin/sh\necho exec\n"
	if runtime.GOOS == "windows" {
		helper += ".exe"
		// A PowerShell-backed .cmd probe can exceed the detector's timeout on
		// loaded CI runners. Keep account setup independent of shell startup.
		source := filepath.Join(root, "fixture.go")
		if err := os.WriteFile(source, []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"exec\")}\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if out, err := exec.Command("go", "build", "-o", helper, source).CombinedOutput(); err != nil {
			t.Fatal("build native CLI fixture", err, string(out))
		}
	} else {
		if err := os.WriteFile(helper, []byte(content), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("RA_CODEX_PATH", helper)
	t.Setenv("RA_QODER_PATH", filepath.Join(root, "missing"))
	t.Setenv("CODEX_HOME", filepath.Join(root, "codex-home"))
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/api/agent/codex" {
			t.Errorf("unexpected setup endpoint: %s", r.URL.Path)
		}
		fmt.Fprintf(w, `{"baseUrl":"http://%s/codex/v1","token":"fixture","accountId":"alice-id","accountUsername":"alice","model":"fixture"}`, r.Host)
	}))
	defer server.Close()
	c := agent.Config{Server: server.URL, ID: "device", Credential: "fixture", Path: filepath.Join(root, "config.json")}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	if err := runCodex([]string{"setup", "--config", c.Path}); err != nil {
		t.Fatal("default setup failed", err)
	}
	if requests != 1 {
		t.Fatal("default setup did not use binding")
	}
}
