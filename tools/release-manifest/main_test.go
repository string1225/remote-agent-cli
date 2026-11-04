package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareUnsignedManifestsAndRejectChangedArtifacts(t *testing.T) {
	dir := t.TempDir()
	platforms := []string{"windows", "darwin", "linux"}
	for _, platform := range platforms {
		for _, arch := range []string{"amd64", "arm64"} {
			name := "remote-agent-" + platform + "-" + arch
			if platform == "windows" {
				name = "remote-agent-core-windows-" + arch + ".exe"
			}
			if err := os.WriteFile(filepath.Join(dir, name), []byte("isolated "+platform+" "+arch), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := prepare(dir, platforms, "fixture", 9, false); err != nil {
		t.Fatal(err)
	}
	if err := prepare(dir, platforms, "fixture", 9, true); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "updates/v1/windows/amd64.json"))
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["signature"] != nil || fields["keyId"] != nil {
		t.Fatal("new releases must not require a signing key")
	}
	if err := prepare(dir, platforms, "different", 9, true); err == nil {
		t.Fatal("different release accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, "remote-agent-core-windows-amd64.exe"), []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := prepare(dir, platforms, "fixture", 9, true); err == nil {
		t.Fatal("damaged artifact accepted")
	}
}

func TestMissingArtifactDoesNotWritePartialManifests(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "remote-agent-core-windows-amd64.exe"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := prepare(dir, []string{"windows"}, "fixture", 9, false); err == nil {
		t.Fatal("missing ARM64 artifact accepted")
	}
	if _, err := os.Stat(filepath.Join(dir, "updates")); !os.IsNotExist(err) {
		t.Fatal("partial release manifests were written")
	}
}
