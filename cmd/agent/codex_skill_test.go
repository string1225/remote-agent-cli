package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexSkillSetupUsesExplicitCodexHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CODEX_HOME", home)
	if err := installRemoteAgentSkill(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "skills", "remote-agent", "SKILL.md")); err != nil {
		t.Fatal("skill missing from configured Codex home", err)
	}
}
