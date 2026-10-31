package agent

import (
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestInterruptedTasksKeepPromptAndIdentityOnRestart(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json")}}
	s, err := a.historyStore().create(protocol.Service{ID: "project", Provider: "codex", Workspace: root}, "queued")
	if err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().update(s.ID, func(v *protocol.Conversation) {
		v.Kind = "web"
		v.Source = "im"
		v.TaskID = "task"
		v.TaskStatus = "queued"
		v.ParentSessionID = "channel"
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = a.historyStore().append(s.ID, protocol.ChatEntry{TurnID: "task", Role: "user", Text: "retained task prompt", Kind: "message"}); err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().nativeID(s.ID, "native-retained"); err != nil {
		t.Fatal(err)
	}
	a.recoverTasks()
	got, err := a.historyStore().get(s.ID)
	if err != nil || got.NativeID != "native-retained" || got.ParentSessionID != "channel" || got.TaskStatus != "failed" {
		t.Fatal("restart lost identity/context or replayed unfinished work")
	}
	page, err := a.historyStore().read(s.ID, 0)
	if err != nil || len(page.Entries) != 2 || page.Entries[0].Text != "retained task prompt" {
		t.Fatal("restart lost prompt or interruption record")
	}
}
