package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStandaloneInstallersKeepSecretsOutOfScripts(t *testing.T) {
	f := setupAt(t, "/agents")
	for _, mode := range []string{"cli", "codex"} {
		for _, platform := range []string{"macos.sh", "linux.sh"} {
			response, err := http.Get(f.url + "/install/" + mode + "-" + platform + "?token=INJECTED_SECRET")
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()
			if response.StatusCode != 200 || strings.Contains(string(body), "INJECTED_SECRET") || strings.Contains(string(body), "__MODE__") || strings.Contains(string(body), "__SETUP_CODEX__") {
				t.Fatalf("bad %s/%s installer: status %d", mode, platform, response.StatusCode)
			}
			if !strings.Contains(string(body), "'"+mode+"'") {
				t.Fatal("wrong installer mode")
			}
		}
	}
}

func TestWindowsUsesMSIDownloadsAndNativeTokenCommands(t *testing.T) {
	f := setupAt(t, "/agents")
	for _, mode := range []string{"", "cli-", "codex-"} {
		body := f.request(http.DefaultClient, "GET", "/install/"+mode+"windows.ps1?token=INJECTED_SECRET", nil, 410)
		if strings.Contains(string(body), "INJECTED_SECRET") || strings.Contains(string(body), "Invoke-WebRequest") {
			t.Fatal("Windows script/secret still served")
		}
	}
	owner := f.user("owner")
	body := f.request(owner, "POST", "/api/agents", map[string]string{"name": "Windows", "os": "windows"}, 201)
	var got struct{ Windows, WindowsMsi, WindowsArm64Msi string }
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Windows, " | remote-agent connect --server '") || !strings.HasSuffix(got.Windows, "--token-stdin") || strings.Contains(got.Windows, "Invoke-WebRequest") || !strings.HasSuffix(got.WindowsMsi, "remote-agent-windows-amd64.msi") || !strings.HasSuffix(got.WindowsArm64Msi, "remote-agent-windows-arm64.msi") {
		t.Fatal("MSI/native enrollment flow not returned")
	}
}
