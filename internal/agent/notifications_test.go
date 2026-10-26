package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestRemoteTaskCompletionReturnsToOriginChannelAndRetriesDurably(t *testing.T) {
	var unavailable atomic.Bool
	unavailable.Store(true)
	attempts := make(chan struct{}, 4)
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fail := unavailable.Load()
		if r.Header.Get("Authorization") != "Bearer target-device-credential" || r.URL.Query().Get("id") != "worker-client" {
			t.Error("notification lost sending device authentication")
		}
		var in struct {
			AgentID string `json:"agentId"`
			Type    string `json:"type"`
			Body    struct {
				ServiceID string `json:"serviceId"`
				SessionID string `json:"sessionId"`
				Message   string `json:"message"`
			} `json:"body"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.AgentID != "origin-client" || in.Type != "channels.notify" || in.Body.ServiceID != "origin-project" || in.Body.SessionID != "origin-im-session" {
			t.Error("completion not routed to original project channel")
		}
		if !strings.Contains(in.Body.Message, "实质结论：修复完成，测试通过") || strings.Contains(in.Body.Message, "#client=") {
			t.Error("notification omitted the result or included a workspace link")
		}
		attempts <- struct{}{}
		if fail {
			w.WriteHeader(503)
			io.WriteString(w, `{"error":"offline"}`)
		} else {
			io.WriteString(w, `{"queued":true}`)
		}
	}))
	defer site.Close()
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "worker-client", Server: site.URL, Credential: "target-device-credential"}}
	if err := a.historyStore().saveTaskResult("task", "实质结论：修复完成，测试通过"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() {
		cancel()
		for _, done := range []chan struct{}{a.taskDone, a.notifyDone} {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("worker did not stop")
			}
		}
	})
	a.notifyTask(queuedTask{start: protocol.TaskStart{ServiceID: "work-project", Notify: true, NotifyAgentID: "origin-client", NotifyServiceID: "origin-project", NotifySessionID: "origin-im-session", ParentAgentID: "intermediate-client", ParentSessionID: "intermediate-web", Source: "web"}, result: protocol.TaskStarted{TaskID: "task", Session: protocol.Conversation{ID: "web-work", Title: "Work"}}}, nil)
	select {
	case <-attempts:
	case <-time.After(3 * time.Second):
		t.Fatal("notification not attempted")
	}
	pending, err := (historyStore{a.historyStore().path}).remotePending()
	if err != nil || len(pending) != 1 {
		t.Fatal("failed notification not retained on disk", err)
	}
	unavailable.Store(false)
	a.notifyWake <- struct{}{}
	select {
	case <-attempts:
	case <-time.After(3 * time.Second):
		t.Fatal("failed notification not retried")
	}
	for deadline := time.Now().Add(3 * time.Second); ; {
		pending, err = a.historyStore().remotePending()
		if err == nil && len(pending) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("delivery acknowledgement not saved")
		}
		time.Sleep(time.Millisecond)
	}
}
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
