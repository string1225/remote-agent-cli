package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func topicFixture(t *testing.T) (*Agent, Config, protocol.Service, string, *atomic.Bool) {
	t.Helper()
	a, project := managedSessionFixture(t)
	manager := project
	manager.ID, manager.Scope = "manager", "project"
	manager.ManagedProject = project.ID
	a.Config.Services = append(a.Config.Services, manager)
	source, err := a.historyStore().create(manager, "topic manager")
	if err != nil {
		t.Fatal(err)
	}
	loseAck := new(atomic.Bool)
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/agent/coordination" {
			json.NewEncoder(w).Encode(map[string]bool{"accepted": true})
			return
		}
		var in struct {
			AgentID string          `json:"agentId"`
			Type    string          `json:"type"`
			Body    json.RawMessage `json:"body"`
		}
		if r.URL.Path == "/api/agent/catalog" {
			json.NewEncoder(w).Encode(map[string]any{"clients": []any{map[string]any{"id": "remote", "projects": []protocol.Service{project, manager}}}})
			return
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || (in.AgentID != "client" && in.AgentID != "remote") {
			t.Error("bad relay request")
			w.WriteHeader(400)
			return
		}
		var result any
		var err error
		switch in.Type {
		case "sessions.spawn":
			var start protocol.TaskStart
			json.Unmarshal(in.Body, &start)
			start.ParentAgentID = "client"
			result, err = a.startTask(context.Background(), start, nil)
			if loseAck.Swap(false) && err == nil {
				w.WriteHeader(503)
				return
			}
		case "sessions.control":
			var control protocol.SessionControl
			json.Unmarshal(in.Body, &control)
			control.CallerAgentID = "client"
			result, err = a.sessionControl(r.Context(), control)
		default:
			t.Error("unexpected topic operation", in.Type)
		}
		if err != nil {
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	}))
	t.Cleanup(site.Close)
	c := a.Config
	c.Server, c.Credential = site.URL, "test-only-secret"
	return a, c, project, source.ID, loseAck
}

func waitTopic(t *testing.T, a *Agent, accepted ExecutionTopicResult) protocol.SessionState {
	t.Helper()
	return waitWork(t, a, protocol.WorkAccepted{Target: protocol.AssistantRef{SessionID: accepted.Session.ID}, TaskID: accepted.TaskID})
}

func TestExecutionTopicsPersistReuseNativeContextAndRecoverLostAcceptance(t *testing.T) {
	a, c, project, source, loseAck := topicFixture(t)
	var executions atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		if prompt == "continue" && s.ResumeID != "topic-native" {
			t.Error("topic continuation lost native context")
		}
		item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": "verified " + prompt}})
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"topic-native\"}\n" + string(item) + "\n"}})
	}
	start := protocol.TaskStart{ServiceID: project.ID, Prompt: "first", Source: "web"}
	loseAck.Store(true)
	if _, err := StartExecutionTopic(context.Background(), c, "manager", source, "initial", "CEO sessions", c.ID, start); err == nil {
		t.Fatal("lost acknowledgement not reported")
	}
	page, err := ListExecutionTopics(c, "manager", source, 0)
	if err != nil || len(page.Topics) != 1 || page.Topics[0].SessionID != "" {
		t.Fatal("lost pending routing claim", page, err)
	}
	// A fresh caller after restart recovers the original unconfirmed dispatch,
	// without substituting or losing a later prompt.
	start.Prompt = "continue"
	recovered, err := StartExecutionTopic(context.Background(), c, "manager", source, "next", "ceo   sessions", c.ID, start)
	if err != nil || !recovered.Recovered || recovered.Session.ID == "" {
		t.Fatal("failed to recover original task", recovered, err)
	}
	waitTopic(t, a, recovered)
	next, err := StartExecutionTopic(context.Background(), c, "manager", source, "next", "CEO sessions", c.ID, start)
	if err != nil || next.Session.ID != recovered.Session.ID || next.Recovered {
		t.Fatal("same topic did not continue original", next, err)
	}
	if state := waitTopic(t, a, next); state.Result != "verified continue" || !state.Captured {
		t.Fatal("continuation result", state)
	}
	for i := 0; i < 2; i++ {
		duplicate, err := StartExecutionTopic(context.Background(), c, "manager", source, "next", "CEO sessions", c.ID, start)
		if err != nil || duplicate.TaskID != next.TaskID {
			t.Fatal("duplicate continuation", duplicate, err)
		}
	}
	start.Prompt = "different"
	if _, err = StartExecutionTopic(context.Background(), c, "manager", source, "next", "CEO sessions", c.ID, start); err == nil {
		t.Fatal("changed content reused receipt")
	}
	other, err := StartExecutionTopic(context.Background(), c, "manager", source, "other", "Windows service authorization", c.ID, start)
	if err != nil || other.Session.ID == next.Session.ID {
		t.Fatal("different topic did not create separate session", err)
	}
	waitTopic(t, a, other)
	page, err = ListExecutionTopics(c, "manager", source, 0)
	if err != nil || len(page.Topics) != 2 || executions.Load() != 3 {
		t.Fatal("topic mapping/retry failure", page, executions.Load(), err)
	}
	encoded, _ := json.Marshal(page)
	if strings.Contains(string(encoded), "test-only-secret") || strings.Contains(string(encoded), "topic-native") || strings.Contains(string(encoded), "verified") {
		t.Fatal("topic listing leaked transcript, native context or credentials")
	}
}

func TestExecutionTopicConcurrentCreateBusyConflictDeletionAndExecutorIsolation(t *testing.T) {
	a, c, project, source, _ := topicFixture(t)
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var executions atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, _ string, _ bool, _ func(protocol.Event) error, _ string) error {
		executions.Add(1)
		select {
		case started <- struct{}{}:
		default:
		}
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil
	}
	start := protocol.TaskStart{ServiceID: project.ID, Prompt: "work", Source: "web"}
	var results [2]ExecutionTopicResult
	var failures [2]error
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], failures[i] = StartExecutionTopic(context.Background(), c, "manager", source, "create", "one topic", c.ID, start)
		}(i)
	}
	wg.Wait()
	for _, err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if results[0].Session.ID != results[1].Session.ID {
		t.Fatal("concurrent callers created two sessions")
	}
	<-started
	if _, err := StartExecutionTopic(context.Background(), c, "manager", source, "busy-followup", "one topic", c.ID, start); err == nil {
		t.Fatal("busy session accepted a concurrent turn")
	}
	if _, err := StartExecutionTopic(context.Background(), c, "manager", source, "move", "one topic", "remote", start); err == nil {
		t.Fatal("topic silently changed client")
	}
	managerStart := start
	managerStart.ServiceID = "manager"
	if _, err := StartExecutionTopic(context.Background(), c, "manager", source, "recursive", "recursive", c.ID, managerStart); err == nil {
		t.Fatal("manager project accepted as execution target")
	}
	if _, err := StartExecutionTopic(context.Background(), c, project.ID, results[0].Session.ID, "executor", "recursive", c.ID, start); err == nil {
		t.Fatal("executor inherited manager topic scheduling")
	}
	close(release)
	waitTopic(t, a, results[0])
	a.historyStore().update(results[0].Session.ID, func(s *protocol.Conversation) { s.DeletedAt = 1 })
	if _, err := StartExecutionTopic(context.Background(), c, "manager", source, "deleted", "one topic", c.ID, start); err == nil {
		t.Fatal("deleted topic silently replaced")
	}
	page, err := ListExecutionTopics(c, "manager", source, 0)
	if err != nil || len(page.Topics) != 1 || page.Topics[0].SessionID != results[0].Session.ID || executions.Load() != 1 {
		t.Fatal("failed request lost mapping or duplicated work", page, err)
	}
}

func TestBindExecutionTopicAdoptsOnlyExistingOwnedChildrenWithoutRunningWork(t *testing.T) {
	a, c, project, source, _ := topicFixture(t)
	child, err := a.historyStore().create(project, "existing Windows work")
	if err != nil {
		t.Fatal(err)
	}
	a.historyStore().update(child.ID, func(s *protocol.Conversation) {
		s.Kind, s.ParentAgentID, s.ParentSessionID, s.TaskID = "web", c.ID, source, "existing-task"
	})
	mapping, err := BindExecutionTopic(context.Background(), c, "manager", source, "Windows service authorization", c.ID, project.ID, child.ID)
	if err != nil || mapping.SessionID != child.ID || mapping.TaskID != "existing-task" {
		t.Fatal("existing execution session not adopted", mapping, err)
	}
	if _, err = BindExecutionTopic(context.Background(), c, "manager", source, "unrelated direction", c.ID, project.ID, child.ID); err == nil {
		t.Fatal("different topics shared the same execution session")
	}
	// Sibling parent cannot discover/adopt an unrelated execution session.
	sibling, _ := a.historyStore().create(protocol.Service{ID: "manager", Provider: "codex", Workspace: project.Workspace}, "sibling")
	if _, err = BindExecutionTopic(context.Background(), c, "manager", sibling.ID, "Windows", c.ID, project.ID, child.ID); err == nil {
		t.Fatal("adopted unrelated child")
	}
	page, err := ListExecutionTopics(c, "manager", sibling.ID, 0)
	if err != nil || len(page.Topics) != 0 {
		t.Fatal("sibling inherited another manager session's topic map")
	}
	for i := 0; i < 51; i++ {
		h, owner, _ := executionTopicStore(c, "manager", source)
		key := store.Hash(string(rune(i)) + "topic")
		h.coordinationSave("execution_topics", key, executionTopicRecord{Owner: owner, ExecutionTopic: ExecutionTopic{Topic: key}})
	}
	page, err = ListExecutionTopics(c, "manager", source, 0)
	if err != nil || len(page.Topics) != 50 || page.NextCursor == 0 {
		t.Fatal("unbounded topic metadata response", page, err)
	}
	last, err := ListExecutionTopics(c, "manager", source, page.NextCursor)
	if err != nil || len(last.Topics) != 2 || last.NextCursor != 0 {
		t.Fatal("topic pagination lost records", last, err)
	}
}

func TestExecutionTopicCrossClientUsesOriginalProjectAndReusesRemoteSession(t *testing.T) {
	a, c, project, source, _ := topicFixture(t)
	a.Execute = func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error {
		return nil
	}
	start := protocol.TaskStart{ServiceID: project.ID, Prompt: "remote work", Source: "web"}
	first, err := StartExecutionTopic(context.Background(), c, "manager", source, "remote-first", "remote topic", "remote", start)
	if err != nil || first.Mapping.ClientID != "remote" || first.Mapping.ProjectID != project.ID {
		t.Fatal("cross-client original project route", first, err)
	}
	waitTopic(t, a, first)
	next, err := StartExecutionTopic(context.Background(), c, "manager", source, "remote-next", "remote topic", "remote", start)
	if err != nil || next.Session.ID != first.Session.ID {
		t.Fatal("cross-client continuation changed session", next, err)
	}
	waitTopic(t, a, next)
}
