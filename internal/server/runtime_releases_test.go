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
	for _, platform := range []string{"windows", "linux", "darwin"} {
		manifest := filepath.Join(dir, "updates", "v1", platform)
		if err := os.MkdirAll(manifest, 0755); err != nil {
			t.Fatal(err)
		}
		for _, arch := range []string{"amd64", "arm64"} {
			artifact := "remote-agent-" + platform + "-" + arch
			if platform == "windows" {
				artifact = "remote-agent-core-windows-" + arch + ".exe"
			}
			for _, file := range []string{filepath.Join(manifest, arch+".json"), filepath.Join(dir, artifact)} {
				if err := os.WriteFile(file, []byte("fixture-only"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range []string{"/updates/v1/" + platform + "/" + arch + ".json", "/downloads/" + artifact} {
				if got := f.request(http.DefaultClient, "GET", path, nil, 200); string(got) != "fixture-only" {
					t.Fatal("wrong release artifact")
				}
			}
		}
	}
	for _, path := range []string{"/updates/v1/freebsd/amd64.json", "/updates/v1/windows/amd64", "/updates/v1/windows/other.json", "/downloads/remote-agent-core-linux-amd64.exe"} {
		f.request(http.DefaultClient, "GET", path, nil, 404)
	}
}
