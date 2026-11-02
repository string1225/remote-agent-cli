package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func todoTestDevice(t *testing.T, f *fixture, owner *http.Client) (store.Agent, agent.Config, chan string, func()) {
	t.Helper()
	row, token := f.enroll(owner, runtime.GOOS)
	root := t.TempDir()
	cfg, err := agent.Enroll(f.url, token, filepath.Join(root, "config.json"), []string{root}, true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Services = []protocol.Service{{ID: "project-" + row.ID, Name: "test", Provider: "codex", Workspace: root}}
	if err = cfg.Save(); err != nil {
		t.Fatal(err)
	}
	ran := make(chan string, 10)
	a := &agent.Agent{Config: cfg, Providers: map[string]agent.Provider{"codex": {Name: "codex", Available: true}}, Execute: func(ctx context.Context, _ agent.Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		ran <- prompt
		return nil
	}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = a.Run(ctx); close(done) }()
	stop := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(6 * time.Second):
			t.Error("test client did not stop")
		}
	}
	t.Cleanup(stop)
	deadline := time.Now().Add(5 * time.Second)
	for {
		data := f.request(owner, "GET", "/api/agents", nil, 200)
		var rows []struct {
			ID     string `json:"id"`
			Online bool   `json:"online"`
		}
		_ = json.Unmarshal(data, &rows)
		online := false
		for _, v := range rows {
			if v.ID == row.ID {
				online = v.Online
			}
		}
		if online {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("test client offline")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return row, cfg, ran, stop
}

func TestAccountTodosRelayReplyMigrationAndIsolation(t *testing.T) {
	f := setupAt(t, "/agents")
	alice, bob := f.user("alice"), f.user("bob")
	host, _, _, stopHost := todoTestDevice(t, f, alice)
	source, cfg, ran, _ := todoTestDevice(t, f, alice)
	foreign, _, _, _ := todoTestDevice(t, f, bob)
	f.request(alice, "GET", "/api/todos", nil, 409)
	f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": foreign.ID}, 404)
	f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 200)
	p := f.peer(alice, source.ID)
	p.send(t, protocol.Request{ID: "create-session", Type: "sessions.create", ServiceID: cfg.Services[0].ID})
	e := p.next(t, "create-session", "result")
	raw, _ := json.Marshal(e.Data)
	var session protocol.Conversation
	if json.Unmarshal(raw, &session) != nil || session.ID == "" {
		t.Fatal("no source session", string(raw))
	}
	command := protocol.TodoCommand{Operation: "create", RequestID: "decision", Title: "Human decision", Body: "A or B?", ClientID: foreign.ID, ProjectID: cfg.Services[0].ID, SessionID: session.ID}
	raw, err := agent.ClientAPI(context.Background(), cfg, "POST", "/api/agent/todos", command)
	if err != nil {
		t.Fatal(err)
	}
	var row protocol.Todo
	if json.Unmarshal(raw, &row) != nil || row.ClientID != source.ID {
		t.Fatal("source identity was spoofed", string(raw))
	}
	list, err := agent.ClientAPI(context.Background(), cfg, "GET", "/api/agent/todos", nil)
	if err != nil || !strings.Contains(string(list), row.ID) {
		t.Fatal("global client query failed", err)
	}
	f.request(bob, "GET", "/api/todos", nil, 409)
	// A browser cannot manufacture the server-owned account scope.
	forged, _ := json.Marshal(protocol.TodoStorageCall{AccountID: "alice", Generation: "forged", TodoCommand: protocol.TodoCommand{Operation: "list"}})
	p.send(t, protocol.Request{ID: "forged", Type: "todos.call", Body: forged})
	p.next(t, "forged", "error")
	replyCommand := protocol.TodoCommand{Operation: "reply", ID: row.ID, Version: row.Version, RequestID: "same-reply", Message: "Choose A"}
	for i := 0; i < 2; i++ {
		f.request(alice, "POST", "/api/todos", replyCommand, 200)
	}
	select {
	case text := <-ran:
		if !strings.Contains(text, "Choose A") {
			t.Fatal(text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reply never reached original session")
	}
	select {
	case <-ran:
		t.Fatal("duplicate reply executed")
	case <-time.After(50 * time.Millisecond):
	}
	for i := 0; i < 36; i++ {
		f.request(alice, "POST", "/api/todos", protocol.TodoCommand{Operation: "create", RequestID: store.Token(), Title: "note", Body: strings.Repeat("note ", 500)}, 200)
	}
	before, err := f.db.DefaultHost(host.UserID)
	if err != nil || before.Generation == "" {
		t.Fatal("host generation not persisted", err)
	}
	f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": source.ID}, 200)
	count, cursor := 0, ""
	for {
		raw = f.request(alice, "GET", "/api/todos?cursor="+cursor, nil, 200)
		var page protocol.TodoPage
		if err = json.Unmarshal(raw, &page); err != nil {
			t.Fatal(err)
		}
		count += len(page.Todos)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if count != 37 {
		t.Fatal("host migration lost tasks", count)
	}
	// Model a site restart after writing a migration marker. The next operation
	// recovers the canonical host before accepting another mutation.
	current, err := f.db.DefaultHost(host.UserID)
	if err != nil {
		t.Fatal(err)
	}
	interrupted := current
	interrupted.MigrationID = store.Token()
	if err = f.db.SetDefaultHost(host.UserID, current, interrupted); err != nil {
		t.Fatal(err)
	}
	f.request(alice, "GET", "/api/todos", nil, 200)
	recovered, err := f.db.DefaultHost(host.UserID)
	if err != nil || recovered != current {
		t.Fatal("site restart failed to recover interrupted migration", recovered, err)
	}
	f.request(alice, "POST", "/api/todos", replyCommand, 200)
	select {
	case <-ran:
		t.Fatal("migration lost reply receipt")
	case <-time.After(50 * time.Millisecond):
	}
	f.request(alice, "DELETE", "/api/agents/"+source.ID, nil, 409)
	stopHost()
	f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 409)
	raw = f.request(alice, "GET", "/api/default-host", nil, 200)
	if !strings.Contains(string(raw), source.ID) {
		t.Fatal("failed migration changed canonical host")
	}
	f.request(alice, "GET", "/api/todos", nil, 200)
}
