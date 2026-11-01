package agent

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestQueuePromoteInterruptsWithoutOverlappingNativeSession(t *testing.T) {
	a, project := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	session, _ := a.historyStore().create(project, "queue")
	started := make(chan string, 10)
	var active atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		if active.Add(1) != 1 {
			t.Error("native session ran concurrently")
		}
		defer active.Add(-1)
		if prompt != "first" && s.ResumeID != "retained-native" {
			t.Error("native context was lost", s.ResumeID)
		}
		if prompt == "first" {
			_ = send(protocol.Event{Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"retained-native\"}\n"}})
		}
		started <- prompt
		if prompt == "first" {
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	}
	start := func(id, prompt string) {
		t.Helper()
		if _, err := a.startSessionTask(protocol.Request{ID: id, ServiceID: project.ID, SessionID: session.ID, Prompt: prompt}); err != nil {
			t.Fatal(err)
		}
	}
	start("one", "first")
	if awaitPrompt(t, started) != "first" {
		t.Fatal("first")
	}
	start("two", "second")
	start("three", "third")
	page, err := a.historyStore().read(session.ID, 0)
	if err != nil || len(page.Session.QueuedMessages) != 2 {
		t.Fatal(page.Session, err)
	}
	for _, entry := range page.Entries {
		if entry.Role == "user" && entry.Text != "first" {
			t.Fatal("waiting message entered native history early", entry)
		}
	}
	if _, err := a.changeQueuedTask(protocol.Request{Type: "tasks.promote", ServiceID: project.ID, SessionID: session.ID, RunID: "three"}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "third" || awaitPrompt(t, started) != "second" {
		t.Fatal("priority or FIFO ordering was lost")
	}
	waitSessionStatus(t, a, session.ID, "done")
	row, _, _ := a.historyStore().queueRecord("one")
	if row.Status != "interrupted" {
		t.Fatal(row.Status)
	}
	if _, err := a.startSessionTask(protocol.Request{ID: "three", ServiceID: project.ID, SessionID: session.ID, Prompt: "third"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
		t.Fatal("lost-ack retry replayed a completed queued task")
	case <-time.After(30 * time.Millisecond):
	}
}

func TestIMIncomingAutomaticallyInterruptsTheRunningTurn(t *testing.T) {
	a, project := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	var failed atomic.Bool
	adapter := &fakeIM{sent: make(chan string, 4), fail: &failed}
	r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "mock-app"}, ctx: ctx, adapter: adapter, wake: make(chan struct{}, 1)}
	started := make(chan string, 3)
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, send func(protocol.Event) error, _ string) error {
		started <- prompt
		if prompt == "first IM" {
			<-ctx.Done()
			return ctx.Err()
		}
		return send(protocol.Event{Type: "output", Data: map[string]string{"stream": "stdout", "text": "answer"}})
	}
	if err := r.receive(channelMessage{ID: "im-one", Text: "first IM", ChatID: "oc_test", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "first IM" {
		t.Fatal("first IM")
	}
	if err := r.receive(channelMessage{ID: "im-two", Text: "second IM", ChatID: "oc_test", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "second IM" {
		t.Fatal("IM did not interrupt")
	}
	page, _ := a.historyStore().list(project.ID, "", 0)
	if len(page.Sessions) != 1 {
		t.Fatal("IM session identity changed")
	}
	waitSessionStatus(t, a, page.Sessions[0].ID, "done")
	for until := time.Now().Add(5 * time.Second); time.Now().Before(until); {
		a.mu.Lock()
		remaining := len(a.taskRecords)
		a.mu.Unlock()
		if remaining == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	items, err := a.historyStore().channelPendingItems(r.config, true)
	if err != nil || len(items) != 1 {
		t.Fatal("interrupted IM emitted a stale reply", items, err)
	}
}

func TestQueuedEditIsRemovedBeforeDraftAndSurvivesLostAcknowledgement(t *testing.T) {
	a, project := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	session, _ := a.historyStore().create(project, "editing")
	started := make(chan string, 3)
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		<-ctx.Done()
		return ctx.Err()
	}
	_, err := a.startSessionTask(protocol.Request{ID: "active", ServiceID: project.ID, SessionID: session.ID, Prompt: "active"})
	if err != nil {
		t.Fatal(err)
	}
	awaitPrompt(t, started)
	_, err = a.startSessionTask(protocol.Request{ID: "edit-me", ServiceID: project.ID, SessionID: session.ID, Prompt: "keep this exact text"})
	if err != nil {
		t.Fatal(err)
	}
	request := protocol.Request{Type: "tasks.edit", ServiceID: project.ID, SessionID: session.ID, RunID: "edit-me"}
	for range 2 {
		row, err := a.changeQueuedTask(request)
		if err != nil || row.Prompt != "keep this exact text" {
			t.Fatal(row, err)
		}
	}
	page, _ := a.historyStore().read(session.ID, 0)
	if len(page.Session.QueuedMessages) != 1 || page.Session.QueuedMessages[0].Status != "editing" {
		t.Fatal("lost edit acknowledgement is not recoverable", page.Session)
	}
	request.Body = json.RawMessage(`{"acknowledged":true}`)
	if _, err := a.changeQueuedTask(request); err != nil {
		t.Fatal(err)
	}
	page, _ = a.historyStore().read(session.ID, 0)
	if len(page.Session.QueuedMessages) != 0 {
		t.Fatal("acknowledged edit remained in queue")
	}
	if _, err := a.changeQueuedTask(protocol.Request{Type: "tasks.promote", ServiceID: project.ID, SessionID: "another-session", RunID: "active"}); err == nil {
		t.Fatal("cross-session queue control accepted")
	}
	a.cancelTask("active")
	waitSessionStatus(t, a, session.ID, "cancelled")
	select {
	case prompt := <-started:
		t.Fatal("edited message still ran", prompt)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestQueuedMessagesRecoverPausedWithoutAutomaticReplay(t *testing.T) {
	a, project := concurrencyFixture(t)
	session, _ := a.historyStore().create(project, "recovery")
	a.tasks = make(chan queuedTask, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if !a.enqueueTask(queuedTask{ctx: ctx, cancel: cancel, start: protocol.TaskStart{ServiceID: project.ID, Prompt: "recover me", Source: "web"}, result: protocol.TaskStarted{TaskID: "recover", Session: session}}) {
		t.Fatal("enqueue")
	}
	a.recoverTasks()
	page, err := a.historyStore().read(session.ID, 0)
	if err != nil || len(page.Session.QueuedMessages) != 1 || page.Session.QueuedMessages[0].Status != "paused" {
		t.Fatal(page.Session, err)
	}
	reopened := &Agent{Config: a.Config, Providers: a.Providers}
	started := make(chan string, 1)
	reopened.Execute = func(_ context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		return nil
	}
	daemon, stop := context.WithCancel(context.Background())
	reopened.initTasks(daemon)
	t.Cleanup(func() { stop(); <-reopened.taskDone; <-reopened.notifyDone })
	select {
	case <-started:
		t.Fatal("restart replayed unreviewed work")
	case <-time.After(30 * time.Millisecond):
	}
	if _, err := reopened.changeQueuedTask(protocol.Request{Type: "tasks.promote", ServiceID: project.ID, SessionID: session.ID, RunID: "recover"}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "recover me" {
		t.Fatal("paused prompt lost")
	}
	waitSessionStatus(t, reopened, session.ID, "done")
}
