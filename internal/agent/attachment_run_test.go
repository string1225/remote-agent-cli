package agent

import (
	"context"
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestWebAttachmentReachesRunnerAndPersistsHistory(t *testing.T) {
	root := t.TempDir()
	service := protocol.Service{ID: "project", Name: "Demo", Provider: "codex", Workspace: root}
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), AllowedRoots: []string{root}, Services: []protocol.Service{service}}, Providers: map[string]Provider{"codex": {Available: true}}}
	session, err := a.historyStore().create(service, "Upload")
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := a.uploadAttachment(session.ID, attachmentUpload{Name: "notes.txt", Size: 5, Data: base64.StdEncoding.EncodeToString([]byte("hello")), Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	called := make(chan string, 1)
	a.Execute = func(_ context.Context, _ Provider, _ protocol.Service, prompt string, write bool, _ func(protocol.Event) error, _ string) error {
		if !write {
			t.Error("attachment run was read-only")
		}
		called <- prompt
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &peer{ctx: ctx}
	done := make(chan protocol.Event, 8)
	a.handle(p, protocol.Request{ID: "turn", Type: "run", ServiceID: service.ID, SessionID: session.ID, Prompt: "Read this", Attachments: []string{file.ID}}, func(e protocol.Event) error { done <- e; return nil })
	select {
	case prompt := <-called:
		if !strings.Contains(prompt, "notes.txt") || !strings.Contains(prompt, file.ID) {
			t.Fatal("runner did not receive attachment path")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runner was not called")
	}
	for {
		select {
		case event := <-done:
			if event.Type == "done" {
				page, err := a.historyStore().read(session.ID, 0)
				if err != nil || len(page.Entries[0].Attachments) != 1 || page.Entries[0].Attachments[0].ID != file.ID {
					t.Fatal("history omitted attachment", err)
				}
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatal("run did not complete")
		}
	}
}
