package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestSaysoCEOActionDurableRemoteDispatchRecovery(t *testing.T) {
	s, p := soFixture(t)
	s.a.Config.ID = "ceo-host"
	s.a.Config.Services[0].Scope = "account"
	s.a.Config.Services[0].OwnerAccount = "account"
	discussion := soNewSession(t, s, p)
	var action soAction
	_ = json.Unmarshal(soCall(t, s, p, "POST", "/api/sessions/"+discussion.ID+"/actions", map[string]string{"title": "remote action", "detail": "verify output"}), &action)
	soCall(t, s, p, "PATCH", "/api/actions/"+action.ID, map[string]string{"status": "ready"})
	target := protocol.AssistantRef{Scope: "project", ClientID: "remote", ProjectID: "original-project", Name: "项目管家"}
	var run soRun
	_ = json.Unmarshal(soCall(t, s, p, "POST", "/api/actions/"+action.ID+"/dispatch", map[string]any{"target": target}), &run)
	if run.Command == nil || run.WorkID != "" || run.Command.Source.SessionID == discussion.ID {
		t.Fatal("not a durable real source", run)
	}
	if _, e := s.a.historyStore().get(run.Command.Source.SessionID); e != nil {
		t.Fatal("source is not a real web session", e)
	}
	accepted := protocol.WorkAccepted{WorkID: store.Hash("accepted"), ReplyHandle: store.Token(), Target: protocol.AssistantRef{Scope: "project", ClientID: "remote", ProjectID: "project-steward", SessionID: "remote-session"}, TaskID: "remote-task"}
	attempts := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var command protocol.CoordinationCommand
		_ = json.NewDecoder(r.Body).Decode(&command)
		if command.Operation == "dispatch" {
			attempts++
			if command.RequestID != run.ID || command.Target.ClientID != "remote" {
				t.Error("retry identity changed")
			}
			if attempts == 1 {
				w.WriteHeader(409)
				_, _ = w.Write([]byte(`{"error":"offline"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(accepted)
			return
		}
		_ = json.NewEncoder(w).Encode(protocol.SessionState{TaskID: "remote-task", Complete: true, Captured: true, Result: "verified action", Session: protocol.Conversation{TaskStatus: "done"}})
	}))
	defer site.Close()
	s.a.Config.Server = site.URL
	s.a.flushSaysoCoordination(context.Background())
	// Loading after a restart retains remote running work; it does not mark it
	// interrupted or dispatch a new task with a new idempotency key.
	recovered := &Agent{Config: s.a.Config, Providers: s.a.Providers}
	recovered.flushSaysoCoordination(context.Background())
	recovered.flushSaysoCoordination(context.Background())
	state, e := recovered.saysoStore()
	if e != nil {
		t.Fatal(e)
	}
	if attempts != 2 || state.state.Runs[0].WebSessionID != "remote-session" || state.state.Runs[0].Status != "done" || state.state.Actions[0].Status != "done" || state.state.Runs[0].Output != "verified action" {
		t.Fatal("remote action not recovered", attempts, state.state.Runs[0])
	}
}
