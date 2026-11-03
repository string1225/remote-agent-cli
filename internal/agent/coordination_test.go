package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func waitWork(t *testing.T, a *Agent, accepted protocol.WorkAccepted) protocol.SessionState {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s, e := a.historyStore().get(accepted.Target.SessionID)
		if e == nil && s.TaskID == accepted.TaskID && s.TaskStatus != "running" && s.TaskStatus != "queued" {
			state, e := a.sessionState(s, 0)
			if e != nil {
				t.Fatal(e)
			}
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("work did not finish")
	return protocol.SessionState{}
}

func TestCoordinationIdempotentDispatchReplyOutboxAndNativeContext(t *testing.T) {
	a, project := managedSessionFixture(t)
	var executions atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		if prompt == "continue" && s.ResumeID != "native-work" {
			t.Error("lost native context")
		}
		item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": "result " + prompt}})
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"native-work\"}\n" + string(item) + "\n"}})
	}
	command := protocol.CoordinationCommand{Operation: "dispatch", AccountID: "account", WorkID: store.Hash("work"), ReplyHandle: store.Token(), Prompt: "first", Source: protocol.AssistantRef{ClientID: "parent", ProjectID: "project", SessionID: "source"}, Target: protocol.AssistantRef{Scope: "project", ProjectID: project.ID}}
	v, e := a.handleCoordination(context.Background(), command)
	if e != nil {
		t.Fatal(e)
	}
	accepted := v.(protocol.WorkAccepted)
	again, e := a.handleCoordination(context.Background(), command)
	if e != nil || again.(protocol.WorkAccepted).TaskID != accepted.TaskID {
		t.Fatal("duplicate dispatch", e)
	}
	if state := waitWork(t, a, accepted); !state.Captured || state.Result != "result first" {
		t.Fatal(state)
	}
	reply := protocol.CoordinationCommand{Operation: "reply", AccountID: "account", WorkID: command.WorkID, ReplyHandle: command.ReplyHandle, Target: accepted.Target, RequestID: "reply-one", Prompt: "continue"}
	v, e = a.handleCoordination(context.Background(), reply)
	if e != nil {
		t.Fatal(e)
	}
	continued := v.(protocol.WorkAccepted)
	if continued.Target.SessionID != accepted.Target.SessionID {
		t.Fatal("new session on reply")
	}
	waitWork(t, a, continued)
	for i := 0; i < 2; i++ {
		v, e = a.handleCoordination(context.Background(), reply)
		if e != nil || v.(protocol.WorkAccepted).TaskID != continued.TaskID {
			t.Fatal("duplicate continuation", e)
		}
	}
	reply.Prompt = "conflict"
	if _, e = a.handleCoordination(context.Background(), reply); e == nil {
		t.Fatal("accepted changed idempotent request")
	}
	rows, e := a.historyStore().coordinationRows("coordination_pending")
	if e != nil || len(rows) != 2 {
		t.Fatal("lost a turn's outbox", len(rows), e)
	}
	var attempts atomic.Int32
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agent/coordination" {
			t.Error(r.URL.Path)
		}
		var c protocol.CoordinationCommand
		_ = json.NewDecoder(r.Body).Decode(&c)
		if c.Report == nil || !strings.HasPrefix(c.Report.Summary, "result ") {
			t.Error("invalid final report")
		}
		if attempts.Add(1) <= 2 {
			http.Error(w, `{"error":"offline"}`, 409)
			return
		}
		_, _ = w.Write([]byte(`{"accepted":true}`))
	}))
	defer site.Close()
	a.Config.Server = site.URL
	a.flushCoordination(context.Background())
	// Recreate just the Agent; delivery state comes from history.db, not memory.
	recovered := &Agent{Config: a.Config, Providers: a.Providers}
	recovered.flushCoordination(context.Background())
	recovered.flushCoordination(context.Background())
	if attempts.Load() != 4 || executions.Load() != 2 {
		t.Fatal("lost retry/dedup", attempts.Load(), executions.Load())
	}
}

func TestCoordinationInboxDedupAndHiddenInternalPrompt(t *testing.T) {
	a, project := managedSessionFixture(t)
	a.Execute = func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error {
		return nil
	}
	source, e := a.historyStore().create(project, "CEO source")
	if e != nil {
		t.Fatal(e)
	}
	report := &protocol.WorkReport{ID: store.Hash("report"), WorkID: store.Hash("work"), Summary: "final result", ReplyHandle: store.Token(), Origin: protocol.AssistantRef{ClientID: "target", SessionID: "leaf"}}
	in := protocol.CoordinationCommand{Operation: "report", AccountID: "account", Source: protocol.AssistantRef{ProjectID: project.ID, SessionID: source.ID}, Report: report}
	for i := 0; i < 3; i++ {
		if _, e = a.handleCoordination(context.Background(), in); e != nil {
			t.Fatal(e)
		}
	}
	a.flushCoordination(context.Background())
	a.flushCoordination(context.Background())
	page, e := a.historyStore().read(source.ID, 0)
	if e != nil {
		t.Fatal(e)
	}
	reports := 0
	for _, entry := range page.Entries {
		if entry.Kind == "report" {
			reports++
		}
		if entry.Role == "user" {
			t.Fatal("internal engineering prompt leaked as user message")
		}
	}
	if reports != 1 {
		t.Fatal("duplicate result card", reports)
	}
	rows, e := a.historyStore().coordinationRows("coordination_inbox")
	if e != nil || len(rows) != 1 {
		t.Fatal(e, len(rows))
	}
}
