package agent

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestQueuedWebTasksKeepProjectHistoryAndContinueAfterCallerCloses(t *testing.T) {
	root := t.TempDir()
	daemon, stop := context.WithCancel(context.Background())
	defer stop()
	started := make(chan string, 2)
	release := make(chan struct{})
	finished := make(chan error, 2)
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), AllowedRoots: []string{root}, Services: []protocol.Service{{ID: "project", Name: "work", Provider: "codex", Workspace: root}}}, Providers: map[string]Provider{"codex": {Available: true}}, Execute: func(ctx context.Context, _ Provider, s protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
		if s.ID != "project" || !write {
			t.Error("project or local permissions changed")
		}
		started <- prompt
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "task complete"}})
	}}
	a.initTasks(daemon)
	caller, cancel := context.WithCancel(context.Background())
	first, err := a.startTask(caller, protocol.TaskStart{ServiceID: "project", Prompt: "one", Source: "sayso", ActionID: "action", ParentSessionID: "discussion"}, func(err error) { finished <- err })
	if err != nil {
		t.Fatal(err)
	}
	second, err := a.startTask(caller, protocol.TaskStart{ServiceID: "project", Prompt: "two", Source: "im", ParentSessionID: "channel"}, func(err error) { finished <- err })
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case prompt := <-started:
		if prompt != "one" {
			t.Fatal(prompt)
		}
	case <-time.After(time.Second):
		t.Fatal("first task not started")
	}
	select {
	case <-started:
		t.Fatal("task queue ran two turns concurrently")
	default:
	}
	close(release)
	for range 2 {
		select {
		case err := <-finished:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("background task did not finish")
		}
	}
	for _, task := range []protocol.TaskStarted{first, second} {
		v, err := a.historyStore().get(task.Session.ID)
		if err != nil || v.TaskStatus != "done" || v.ServiceID != "project" || v.Kind != "web" {
			t.Fatal(v, err)
		}
		page, err := a.historyStore().read(v.ID, 0)
		if err != nil || len(page.Entries) < 2 {
			t.Fatal("task not visible in web history", err)
		}
		data, _ := json.Marshal(page)
		if !json.Valid(data) {
			t.Fatal("history invalid")
		}
	}
	if first.Session.ID == second.Session.ID {
		t.Fatal("tasks did not create separate web sessions")
	}
}
