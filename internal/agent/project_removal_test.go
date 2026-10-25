package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestProjectRemovalRetainsLocalDataAndRequiresChannelConsent(t *testing.T) {
	root := t.TempDir()
	p := protocol.Service{ID: "project", Name: "Demo", Provider: "codex", Workspace: root}
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{p}}}
	if err := a.Config.Save(); err != nil {
		t.Fatal(err)
	}
	s, err := a.historyStore().create(p, "Keep history")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "keep.txt")
	os.WriteFile(file, []byte("keep"), 0600)
	a.Config.Services[0].Channel = &protocol.ChannelInfo{Provider: "feishu"}
	if err := a.removeProject(context.Background(), p.ID, false); err == nil {
		t.Fatal("channel removal did not require consent")
	}
	a.Config.Services[0].Channel = nil
	a.busy = true
	if err := a.removeProject(context.Background(), p.ID, true); err == nil {
		t.Fatal("removed a busy project")
	}
	a.busy = false
	if err := a.removeProject(context.Background(), p.ID, true); err != nil {
		t.Fatal(err)
	}
	if len(a.Config.Services) != 0 {
		t.Fatal("registration retained")
	}
	if data, err := os.ReadFile(file); err != nil || string(data) != "keep" {
		t.Fatal("work file removed")
	}
	if _, err := a.historyStore().get(s.ID); err != nil {
		t.Fatal("conversation history removed", err)
	}
	backups, _ := filepath.Glob(a.Config.Path + ".project-remove-backup-*")
	if len(backups) != 1 {
		t.Fatal("configuration backup missing")
	}
}
