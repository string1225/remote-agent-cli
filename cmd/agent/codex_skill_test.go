package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexSkillSetupUsesExplicitCodexHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CODEX_HOME", filepath.Join(home, "custom-codex"))
	if err := installRemoteAgentSkill(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "remote-agent", "SKILL.md")); err != nil {
		t.Fatal("skill missing from shared user home", err)
	}
}
