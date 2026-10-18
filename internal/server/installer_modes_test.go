package server_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestStandaloneInstallersKeepSecretsOutOfScripts(t *testing.T) {
	f := setupAt(t, "/agents")
	for _, mode := range []string{"cli", "codex"} {
		for _, platform := range []string{"windows.ps1", "macos.sh", "linux.sh"} {
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
