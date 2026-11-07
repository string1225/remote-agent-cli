package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func TestSkillCLIWorksWithoutBindingAndOnlyUsesExplicitHome(t *testing.T) {
	root := t.TempDir()
	unrelated := t.TempDir()
	t.Setenv("CODEX_HOME", unrelated)
	var output bytes.Buffer
	if err := runSkill([]string{"install", "--home", root}, &output); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "remote-agent", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(unrelated); len(entries) != 0 {
		t.Fatal("explicit home modified unrelated CODEX_HOME")
	}
	output.Reset()
	if err := runSkill([]string{"status", "--home", root}, &output); err != nil {
		t.Fatal(err)
	}
	var status codexsetup.UserSkillStatus
	if json.Unmarshal(output.Bytes(), &status) != nil || !status.Current {
		t.Fatal("CLI does not report current skill", output.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".remote-agent", "config.json")); !os.IsNotExist(err) {
		t.Fatal("skill installation created a device binding")
	}
}
