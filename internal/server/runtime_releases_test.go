package server_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicRuntimeReleasePaths(t *testing.T) {
	f := setupAt(t, "/agents")
	dir := t.TempDir()
	f.s.Config.Downloads = dir
	manifest := filepath.Join(dir, "updates", "v1", "windows")
	if err := os.MkdirAll(manifest, 0755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{filepath.Join(manifest, "amd64.json"), filepath.Join(dir, "remote-agent-core-windows-amd64.exe")} {
		if err := os.WriteFile(file, []byte("fixture-only"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"/updates/v1/windows/amd64.json", "/downloads/remote-agent-core-windows-amd64.exe"} {
		if got := f.request(http.DefaultClient, "GET", path, nil, 200); string(got) != "fixture-only" {
			t.Fatal("wrong release artifact")
		}
	}
	for _, path := range []string{"/updates/v1/linux/amd64.json", "/updates/v1/windows/amd64", "/updates/v1/windows/other.json", "/downloads/remote-agent-core-linux-amd64.exe"} {
		f.request(http.DefaultClient, "GET", path, nil, 404)
	}
}
