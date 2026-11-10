package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

func eventFinal(send func(protocol.Event) error, id, text string) error {
	item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": text}})
	return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": string(item) + "\n"}})
}

func TestExecutionEventsCheckpointsGateDedupFollowupAndRestart(t *testing.T) {
	a, p := managedSessionFixture(t)
	started, release := make(chan string, 2), make(chan struct{})
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		started <- id
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
		return eventFinal(send, id, "evidence verified")
	}
	child, err := a.startTask(context.Background(), protocol.TaskStart{ServiceID: p.ID, Prompt: "work", ParentAgentID: "manager-client", ParentSessionID: "manager-session"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	<-started
	command := protocol.CoordinationCommand{AccountID: "fixture", WorkID: store.Hash("subscription"), ReplyHandle: store.Token(), Topic: "topic", Checkpoints: []string{"plan"}, Source: protocol.AssistantRef{ClientID: "manager-client", SessionID: "manager-session"}, Target: protocol.AssistantRef{ClientID: a.Config.ID, ProjectID: p.ID, SessionID: child.Session.ID}}
	if _, err = a.watchExecution(command); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	reports := []protocol.WorkReport{}
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in protocol.CoordinationCommand
		json.NewDecoder(r.Body).Decode(&in)
		if in.Report == nil {
			t.Error("non-event request")
			return
		}
		mu.Lock()
		reports = append(reports, *in.Report)
		mu.Unlock()
		w.Write([]byte(`{"accepted":true}`))
	}))
	defer site.Close()
	a.Config.Server = site.URL
	event := command
	event.TaskID = child.TaskID
	event.Stage = "validation"
	event.Status = "checkpoint"
	event.Summary = "validated"
	event.Evidence = []string{"test log"}
	value, err := a.executionCheckpoint(event)
	if err != nil || value.(map[string]any)["recorded"] != false {
		t.Fatal("unconfigured checkpoint woke manager", value, err)
	}
	event.Stage = "plan"
	for i := 0; i < 3; i++ {
		if _, err = a.executionCheckpoint(event); err != nil {
			t.Fatal(err)
		}
	}
	a.flushCoordination(context.Background())
	a.flushCoordination(context.Background())
	if len(reports) != 1 || reports[0].Event != "checkpoint" {
		t.Fatal("checkpoint duplicate/wrong event", reports)
	}
	event.Status, event.Stage, event.Summary = "review_required", "predeploy", "review deployment evidence"
	if _, err = a.executionCheckpoint(event); err != nil {
		t.Fatal(err)
	}
	a.flushCoordination(context.Background())
	if len(reports) != 1 {
		t.Fatal("review sent before execution round exited")
	}
	close(release)
	waitWork(t, a, protocol.WorkAccepted{Target: command.Target, TaskID: child.TaskID})
	a.flushCoordination(context.Background())
	a.flushCoordination(context.Background())
	if len(reports) != 2 || reports[1].Event != "review_required" || !reports[1].Captured {
		t.Fatal("review/terminal duplicated or missing", reports)
	}
	continued, err := a.startSessionTask(protocol.Request{ID: "followup", ServiceID: p.ID, SessionID: child.Session.ID, Prompt: "approved continue", RequireIdle: true})
	if err != nil {
		t.Fatal(err)
	}
	waitWork(t, a, protocol.WorkAccepted{Target: command.Target, TaskID: continued.TaskID})
	if _, err = a.executionCheckpoint(event); err == nil {
		t.Fatal("old turn checkpoint accepted")
	}
	a.flushCoordination(context.Background())
	if len(reports) != 3 || reports[2].Event != "completed" || reports[2].TaskID != continued.TaskID {
		t.Fatal("followup not subscribed", reports)
	}
	recovered := &Agent{Config: a.Config, Providers: a.Providers}
	recovered.flushCoordination(context.Background())
	if len(reports) != 3 {
		t.Fatal("restart replayed acknowledged events")
	}
}

func TestExecutionInboxBusyUserRaceDedupAndInterruptedReviewRecovery(t *testing.T) {
	a, p := managedSessionFixture(t)
	source, err := a.historyStore().create(p, "source")
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan string, 5), make(chan struct{})
	var reviews atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		started <- prompt
		if prompt == "user first" {
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if strings.Contains(prompt, "下游工作结果") {
			reviews.Add(1)
		}
		return eventFinal(send, id, "reviewed")
	}
	_, err = a.startSessionTask(protocol.Request{ID: "first", ServiceID: p.ID, SessionID: source.ID, Prompt: "user first"})
	if err != nil {
		t.Fatal(err)
	}
	<-started
	report := protocol.WorkReport{ID: store.Hash("event"), WorkID: store.Hash("work"), TaskID: "child-turn", Summary: "bounded evidence", Topic: "topic", Event: "completed", Captured: true, Origin: protocol.AssistantRef{ClientID: "child", ProjectID: "project", SessionID: "child-session"}}
	in := protocol.CoordinationCommand{Operation: "report", AccountID: "fixture", Source: protocol.AssistantRef{ProjectID: p.ID, SessionID: source.ID}, Report: &report}
	for i := 0; i < 3; i++ {
		if _, err = a.handleCoordination(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	}
	a.flushCoordination(context.Background())
	// Browser/user messages race with reports through the same serial queue;
	// no cancellation, no overwrite, no second manager model while busy.
	_, err = a.startSessionTask(protocol.Request{ID: "second", ServiceID: p.ID, SessionID: source.ID, Prompt: "user second"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case v := <-started:
		t.Fatal("model ran concurrently", v)
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	select {
	case v := <-started:
		if !strings.Contains(v, "下游工作结果") {
			t.Fatal("lost queue ordering", v)
		}
	case <-time.After(time.Second):
		t.Fatal("review missing")
	}
	select {
	case v := <-started:
		if v != "user second" {
			t.Fatal(v)
		}
	case <-time.After(time.Second):
		t.Fatal("user message lost")
	}
	a.flushCoordination(context.Background())
	if reviews.Load() != 1 {
		t.Fatal("duplicate wake", reviews.Load())
	}
	// Model/browser is gone: a paused durable review after daemon restart is
	// retried with a bounded new wake ID; duplicates still share that ID.
	report.ID = store.Hash("crash-event")
	in.Report = &report
	if _, err = a.handleCoordination(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	row := coordinationInbox{Source: in.Source, Report: report, Queued: true}
	if err = a.historyStore().coordinationSave("coordination_inbox", report.ID, row); err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().transact(true, func(tx *bolt.Tx) error {
		return saveQueueRecord(tx, queueRecord{QueuedMessage: protocol.QueuedMessage{ID: reportWakeID(row), Status: "paused"}, SessionID: source.ID, ServiceID: p.ID})
	}); err != nil {
		t.Fatal(err)
	}
	a.flushCoordination(context.Background())
	a.flushCoordination(context.Background())
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("paused review not recovered")
	}
	deadline := time.Now().Add(time.Second)
	for reviews.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if reviews.Load() != 2 {
		t.Fatal("recovery duplicate/missing", reviews.Load())
	}
}

func TestExecutionEventsOfflineFailedTurnRecoveredWithoutModelReplay(t *testing.T) {
	a, p := managedSessionFixture(t)
	var executions atomic.Int32
	a.Execute = func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error {
		executions.Add(1)
		return errors.New("fixture failure")
	}
	child, err := a.startTask(context.Background(), protocol.TaskStart{ServiceID: p.ID, Prompt: "work", ParentAgentID: "source-client", ParentSessionID: "source-session"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := protocol.CoordinationCommand{WorkID: store.Hash("failed"), AccountID: "fixture", ReplyHandle: store.Token(), Source: protocol.AssistantRef{ClientID: "source-client", SessionID: "source-session"}, Target: protocol.AssistantRef{ClientID: a.Config.ID, ProjectID: p.ID, SessionID: child.Session.ID}}
	if _, err = a.watchExecution(w); err != nil {
		t.Fatal(err)
	}
	waitWork(t, a, protocol.WorkAccepted{Target: w.Target, TaskID: child.TaskID})
	var online atomic.Bool
	var deliveries atomic.Int32
	site := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var c protocol.CoordinationCommand
		json.NewDecoder(req.Body).Decode(&c)
		if c.Report == nil || c.Report.Event != "failed" || !c.Report.Complete {
			t.Error("lost failure evidence")
		}
		deliveries.Add(1)
		if !online.Load() {
			http.Error(res, "offline", 409)
			return
		}
		res.Write([]byte(`{"accepted":true}`))
	}))
	defer site.Close()
	a.Config.Server = site.URL
	a.flushCoordination(context.Background())
	online.Store(true)
	recovered := &Agent{Config: a.Config, Providers: a.Providers}
	recovered.flushCoordination(context.Background())
	recovered.flushCoordination(context.Background())
	if executions.Load() != 1 || deliveries.Load() != 2 {
		t.Fatal("outbox lost/double executed", executions.Load(), deliveries.Load())
	}
}

func TestExecutionSubscriptionRestoresPreUpgradeBindingWithoutReadingTranscript(t *testing.T) {
	a, c, p, source, _ := topicFixture(t)
	a.Execute = func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error {
		return nil
	}
	child, err := a.startTask(context.Background(), protocol.TaskStart{ServiceID: p.ID, Prompt: "old work", ParentAgentID: c.ID, ParentSessionID: source}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = BindExecutionTopic(context.Background(), c, "manager", source, "bootstrap", c.ID, p.ID, child.Session.ID); err != nil {
		t.Fatal(err)
	}
	var subscriptions atomic.Int32
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in protocol.CoordinationCommand
		json.NewDecoder(r.Body).Decode(&in)
		if in.Operation != "subscribe" || in.Source.SessionID != source || in.Target.SessionID != child.Session.ID || in.Topic != "bootstrap" {
			t.Error("invented bootstrap identity")
		}
		subscriptions.Add(1)
		w.Write([]byte(`{"accepted":true}`))
	}))
	defer site.Close()
	a.Config.Server = site.URL
	a.restoreExecutionSubscriptions(context.Background())
	a.restoreExecutionSubscriptions(context.Background())
	if subscriptions.Load() != 1 {
		t.Fatal("restore missed or repeatedly subscribed", subscriptions.Load())
	}
}
