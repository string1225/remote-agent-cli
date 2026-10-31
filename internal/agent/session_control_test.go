package agent

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func managedSessionFixture(t *testing.T) (*Agent, protocol.Service) {
	t.Helper()
	a, project := concurrencyFixture(t)
	a.Config.ID = "client"
	ctx, cancel := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { cancel(); <-a.taskDone; <-a.notifyDone })
	return a, project
}

func TestParentManagesChildWithoutIMDelivery(t *testing.T) {
	a, project := managedSessionFixture(t)
	var executions atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		if prompt == "continue" && s.ResumeID != "native-child" {
			t.Error("continuation lost native context")
		}
		item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": "child says " + prompt}})
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"native-child\"}\n" + string(item) + "\n"}})
	}
	input := protocol.TaskStart{RequestID: "spawn-1", ServiceID: project.ID, Prompt: "work", Source: "im", ParentAgentID: "client", ParentSessionID: "parent-im", Notify: true, NotifyAgentID: "client", NotifyServiceID: project.ID, NotifySessionID: "parent-im"}
	child, err := a.startTask(context.Background(), input, nil)
	if err != nil {
		t.Fatal(err)
	}
	control := protocol.SessionControl{Operation: "wait", CallerAgentID: "client", CallerSessionID: "parent-im", SessionID: child.Session.ID, TaskID: child.TaskID}
	got, err := a.sessionControl(context.Background(), control)
	if err != nil {
		t.Fatal(err)
	}
	state := got.(protocol.SessionState)
	if !state.Complete || !state.Captured || state.Result != "child says work" || state.Session.TaskStatus != "done" || state.Session.NotifySessionID != "" {
		t.Fatal("incorrect child result", state)
	}
	for _, wrong := range []protocol.SessionControl{
		{Operation: "read", CallerAgentID: "other-client", CallerSessionID: "parent-im", SessionID: child.Session.ID},
		{Operation: "read", CallerAgentID: "client", CallerSessionID: "sibling", SessionID: child.Session.ID},
		{Operation: "cancel", CallerAgentID: "client", CallerSessionID: "parent-im", SessionID: child.Session.ID, TaskID: "other-task"},
	} {
		if _, err := a.sessionControl(context.Background(), wrong); err == nil {
			t.Fatal("accepted unrelated control", wrong)
		}
	}
	retry, err := a.startTask(context.Background(), input, nil)
	if err != nil || retry.Session.ID != child.Session.ID || retry.TaskID != child.TaskID || executions.Load() != 1 {
		t.Fatal("spawn retry reexecuted work", retry, err)
	}
	control.Operation, control.RequestID, control.Prompt = "send", "continue-1", "continue"
	continued, err := a.sessionControl(context.Background(), control)
	if err != nil {
		t.Fatal(err)
	}
	next := continued.(protocol.TaskStarted)
	control.Operation, control.TaskID = "wait", next.TaskID
	if _, err = a.sessionControl(context.Background(), control); err != nil {
		t.Fatal(err)
	}
	control.Operation = "send"
	if _, err = a.sessionControl(context.Background(), control); err != nil {
		t.Fatal("continuation retry failed", err)
	}
	if executions.Load() != 2 || next.Session.ID != child.Session.ID {
		t.Fatal("continuation duplicated or changed session")
	}
	control.Operation = "list"
	listed, err := a.sessionControl(context.Background(), control)
	if err != nil || len(listed.(protocol.ConversationPage).Sessions) != 1 {
		t.Fatal("lost direct child", err)
	}
	if err := a.historyStore().transact(false, func(tx *bolt.Tx) error {
		for _, name := range []string{"remote_notifications", "channel_outbox"} {
			if b := tx.Bucket([]byte(name)); b != nil && b.Stats().KeyN != 0 {
				t.Error("completion created an IM delivery")
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestChildResultsPaginateExactlyAndWaitHonorsCancellation(t *testing.T) {
	a, project := managedSessionFixture(t)
	child, err := a.historyStore().create(project, "child")
	if err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().update(child.ID, func(s *protocol.Conversation) {
		s.Kind = "web"
		s.ParentAgentID = "remote-parent"
		s.ParentSessionID = "parent"
		s.TaskID = "task"
		s.TaskStatus = "running"
	}); err != nil {
		t.Fatal(err)
	}
	control := protocol.SessionControl{Operation: "wait", CallerAgentID: "remote-parent", CallerSessionID: "parent", SessionID: child.ID, TaskID: "task"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = a.sessionControl(ctx, control); err == nil {
		t.Fatal("wait ignored cancellation")
	}
	text := strings.Repeat("中文🙂\x00\"", 9000)
	if err = a.historyStore().saveTaskResult("task", text); err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().update(child.ID, func(s *protocol.Conversation) { s.TaskStatus = "failed"; s.TaskError = "executor error" }); err != nil {
		t.Fatal(err)
	}
	var joined strings.Builder
	for {
		got, err := a.sessionControl(context.Background(), control)
		if err != nil {
			t.Fatal(err)
		}
		state := got.(protocol.SessionState)
		encoded, _ := json.Marshal(state)
		if len(encoded) > 48000 || !utf8.ValidString(state.Result) || !state.Complete || state.Session.TaskError != "executor error" {
			t.Fatal("invalid result page")
		}
		joined.WriteString(state.Result)
		if state.NextCursor == 0 {
			break
		}
		control.Cursor = state.NextCursor
	}
	if joined.String() != text {
		t.Fatal("result lost content across pages")
	}
}

func TestCancelChildOnlyCancelsItsVerifiedTask(t *testing.T) {
	a, project := managedSessionFixture(t)
	started := make(chan struct{}, 1)
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, _ string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	}
	child, err := a.startTask(context.Background(), protocol.TaskStart{ServiceID: project.ID, Prompt: "work", ParentAgentID: "client", ParentSessionID: "parent"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("child did not start")
	}
	control := protocol.SessionControl{Operation: "cancel", CallerAgentID: "client", CallerSessionID: "parent", SessionID: child.Session.ID, TaskID: child.TaskID}
	got, err := a.sessionControl(context.Background(), control)
	if err != nil || !got.(map[string]bool)["cancelled"] {
		t.Fatal("could not cancel child", err)
	}
	control.Operation = "wait"
	got, err = a.sessionControl(context.Background(), control)
	if err != nil || got.(protocol.SessionState).Session.TaskStatus != "cancelled" {
		t.Fatal("cancellation not persisted", got, err)
	}
}

func TestUpgradeRetainsLegacyNotificationsWithoutSendingThem(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	old := []byte(`{"ID":"task-completion-0000","Text":"legacy-result"}`)
	manual := []byte(`{"ID":"manual-message","Text":"explicit-message"}`)
	if err := a.historyStore().transact(true, func(tx *bolt.Tx) error {
		for _, name := range []string{"remote_notifications", "channel_outbox"} {
			b, e := tx.CreateBucketIfNotExists([]byte(name))
			if e != nil {
				return e
			}
			if e = b.Put([]byte("old"), old); e != nil {
				return e
			}
		}
		return tx.Bucket([]byte("channel_outbox")).Put([]byte("manual"), manual)
	}); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := a.historyStore().retireAutomaticNotifications(); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.historyStore().transact(false, func(tx *bolt.Tx) error {
		for _, name := range []string{"remote_notifications", "channel_outbox"} {
			if string(tx.Bucket([]byte("retired_"+name)).Get([]byte("old"))) != string(old) || tx.Bucket([]byte(name)).Get([]byte("old")) != nil {
				t.Error("legacy delivery lost or remains active")
			}
		}
		if string(tx.Bucket([]byte("channel_outbox")).Get([]byte("manual"))) != string(manual) {
			t.Error("explicit message was changed")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestIMTransportsOnlyItsSessionFinalAnswer(t *testing.T) {
	for _, provider := range []string{"feishu", "dingtalk"} {
		for _, answer := range []string{"", strings.Repeat("中文🙂", 5000) + " https://site.test/agents/#client=c&project=p&session=s"} {
			name := "empty"
			if answer != "" {
				name = "long"
			}
			t.Run(provider+"/"+name, func(t *testing.T) {
				a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex"}}}, tasks: make(chan queuedTask, 1), taskCancels: map[string]context.CancelFunc{}}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				r := &channelRuntime{a: a, ctx: ctx, config: ChannelConfig{Provider: provider, ServiceID: "project", AppID: "mock"}, adapter: &fakeIM{}, wake: make(chan struct{}, 1)}
				if err := r.receive(channelMessage{ID: "message", Text: "original user message", ChatID: "oc_chat", ChatType: "user"}); err != nil {
					t.Fatal(err)
				}
				task := <-a.tasks
				defer task.cancel()
				if _, err := a.historyStore().append(task.result.Session.ID, protocol.ChatEntry{TurnID: task.result.TaskID, Role: "assistant", Text: "commentary must not be delivered"}); err != nil {
					t.Fatal(err)
				}
				if err := a.historyStore().saveTaskResult(task.result.TaskID, answer); err != nil {
					t.Fatal(err)
				}
				task.complete(nil)
				items, err := a.historyStore().channelPendingItems(r.config, true)
				if err != nil {
					t.Fatal(err)
				}
				var delivered strings.Builder
				for _, item := range items {
					if len(item.Text) > 16000 || !utf8.ValidString(item.Text) {
						t.Fatal("invalid IM segment")
					}
					delivered.WriteString(item.Text)
				}
				if delivered.String() != answer {
					t.Fatal("transport rewrote the session answer")
				}
			})
		}
	}
}
