package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyCleanupRemovesOnlyOldAgentPrograms(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"remote-agent.exe", "remote-agent.exe.previous", "remote-agent.exe.sha256", "codex.exe", "qoder.exe", "config.json", "history.db", "recording.webm"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("retained fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := removeLegacyAgentFiles(root); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"remote-agent.exe", "remote-agent.exe.previous", "remote-agent.exe.sha256"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatal("old program retained", name)
		}
	}
	for _, name := range []string{"codex.exe", "qoder.exe", "config.json", "history.db", "recording.webm"} {
		if b, err := os.ReadFile(filepath.Join(root, name)); err != nil || string(b) != "retained fixture" {
			t.Fatal("business data or another tool removed", name)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "remote-agent.exe"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := removeLegacyAgentFiles(root); err == nil {
		t.Fatal("cleanup accepted directory for removal")
	}
}
