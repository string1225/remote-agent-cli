package codexsetup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallRemoteAgentSkillPreservesCustomInstructionsAndIsIdempotent(t *testing.T) {
	home := t.TempDir()
	directory := filepath.Join(home, "skills", "remote-agent")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "SKILL.md")
	before := []byte("user-maintained instructions\n")
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	installed, err := InstallRemoteAgentSkill(home)
	if err != nil {
		t.Fatal(err)
	}
	if installed != path {
		t.Fatal("wrong skill location", installed)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, remoteAgentSkill) {
		t.Fatal("skill not published", err)
	}
	backups, _ := filepath.Glob(path + ".backup-*")
	if len(backups) != 1 {
		t.Fatal("original instructions were not backed up")
	}
	original, _ := os.ReadFile(backups[0])
	if !bytes.Equal(original, before) {
		t.Fatal("backup changed original instructions")
	}
	info, _ := os.Stat(path)
	if _, err := InstallRemoteAgentSkill(home); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(path)
	if !info.ModTime().Equal(after.ModTime()) {
		t.Fatal("unchanged skill was rewritten")
	}
	backups, _ = filepath.Glob(path + ".backup-*")
	if len(backups) != 1 {
		t.Fatal("idempotent install created extra backups")
	}
}

func TestInstallRemoteAgentSkillRejectsConflictingDirectory(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "skills"), []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallRemoteAgentSkill(home); err == nil {
		t.Fatal("overwrote conflicting path")
	}
	if _, err := InstallRemoteAgentSkill(""); err == nil {
		t.Fatal("accepted implicit profile")
	}
}

func TestInstallRemoteAgentSkillHonorsExistingLock(t *testing.T) {
	home := t.TempDir()
	directory := filepath.Join(home, "skills", "remote-agent")
	os.MkdirAll(directory, 0700)
	os.WriteFile(filepath.Join(directory, ".install.lock"), nil, 0600)
	if _, err := InstallRemoteAgentSkill(home); err == nil {
		t.Fatal("overlapped another installer")
	}
	if _, err := os.Stat(filepath.Join(directory, "SKILL.md")); !os.IsNotExist(err) {
		t.Fatal("published despite lock")
	}
}
