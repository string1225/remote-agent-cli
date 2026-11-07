package codexsetup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUserSkillMigratesLegacyAndTracksVersion(t *testing.T) {
	root := t.TempDir()
	codex := filepath.Join(root, ".codex")
	legacy := filepath.Join(codex, "skills", "remote-agent")
	if err := os.MkdirAll(legacy, 0700); err != nil {
		t.Fatal(err)
	}
	old := []byte("customized previous remote-agent instructions\n")
	if err := os.WriteFile(filepath.Join(legacy, "SKILL.md"), old, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "extra.txt"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := SyncUserRemoteAgentSkill(root, codex, "fixture-one")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != filepath.Join(root, ".agents", "skills", "remote-agent", "SKILL.md") {
		t.Fatal("not user-global shared location", result.Path)
	}
	if len(result.Backups) != 1 {
		t.Fatal("legacy folder was not backed up")
	}
	got, err := os.ReadFile(filepath.Join(result.Backups[0], "SKILL.md"))
	if err != nil || !bytes.Equal(got, old) {
		t.Fatal("legacy instructions lost", err)
	}
	if _, err = os.Stat(filepath.Join(result.Backups[0], "extra.txt")); err != nil {
		t.Fatal("legacy resource lost", err)
	}
	if _, err = os.Stat(filepath.Join(result.Backups[0], ".install.lock")); !os.IsNotExist(err) {
		t.Fatal("migration lock was retained in backup", err)
	}
	if _, err = os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("duplicate legacy skill still discoverable")
	}
	info, _ := os.Stat(result.Path)
	result, err = SyncUserRemoteAgentSkill(root, codex, "fixture-two")
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(result.Path)
	if !info.ModTime().Equal(after.ModTime()) {
		t.Fatal("unchanged skill body rewritten")
	}
	status, err := UserRemoteAgentSkillStatus(root, "fixture-two")
	if err != nil || !status.Current || status.Version != "fixture-two" {
		t.Fatal("version did not follow client", status, err)
	}
	if len(result.Backups) != 0 {
		t.Fatal("idempotent sync repeated migration")
	}
}

func TestUserSkillPreservesCanonicalEditsAndRejectsBusyLegacy(t *testing.T) {
	root := t.TempDir()
	codex := filepath.Join(root, "profile")
	result, err := SyncUserRemoteAgentSkill(root, codex, "one")
	if err != nil {
		t.Fatal(err)
	}
	custom := []byte("user customized skill\n")
	if err = os.WriteFile(result.Path, custom, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = SyncUserRemoteAgentSkill(root, codex, "two"); err != nil {
		t.Fatal(err)
	}
	backups, _ := filepath.Glob(result.Path + ".backup-*")
	if len(backups) != 1 {
		t.Fatal("custom instructions not backed up")
	}
	got, _ := os.ReadFile(backups[0])
	if !bytes.Equal(got, custom) {
		t.Fatal("custom skill backup changed")
	}
	legacy := filepath.Join(codex, "skills", "remote-agent")
	os.MkdirAll(legacy, 0700)
	os.WriteFile(filepath.Join(legacy, "SKILL.md"), custom, 0600)
	os.WriteFile(filepath.Join(legacy, ".install.lock"), nil, 0600)
	if _, err = SyncUserRemoteAgentSkill(root, codex, "three"); err == nil {
		t.Fatal("migrated an active legacy installation")
	}
	got, _ = os.ReadFile(filepath.Join(legacy, "SKILL.md"))
	if !bytes.Equal(got, custom) {
		t.Fatal("busy legacy skill changed")
	}
}

func TestUserSkillRejectsLinkedLegacyRootAndSupportsSharedCodexHome(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	legacy := filepath.Join(external, "skills", "remote-agent")
	os.MkdirAll(legacy, 0700)
	os.WriteFile(filepath.Join(legacy, "SKILL.md"), []byte("external instructions"), 0600)
	if err := os.Symlink(external, filepath.Join(root, ".codex")); err != nil {
		t.Skip("symlink fixture unavailable")
	}
	if _, err := SyncUserRemoteAgentSkill(root, filepath.Join(root, ".codex"), "fixture"); err == nil {
		t.Fatal("followed legacy root link")
	}
	if _, err := os.Stat(filepath.Join(legacy, "SKILL.md")); err != nil {
		t.Fatal("external skill was moved", err)
	}
	other := t.TempDir()
	if _, err := SyncUserRemoteAgentSkill(other, filepath.Join(other, ".agents"), "fixture"); err != nil {
		t.Fatal("shared Codex home incorrectly migrated itself", err)
	}
}
