package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestExternalProjectSessionCreationAndDeferredMessage(t *testing.T) {
	a, project := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	started := make(chan string, 4)
	a.Execute = func(_ context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		return nil
	}
	create := func(id string, in protocol.ProjectSessionCreate) protocol.TaskStarted {
		t.Helper()
		body, _ := json.Marshal(protocol.ProjectSessionRelay{Title: in.Title, Message: []byte(in.Message), Execute: in.Execute})
		var event protocol.Event
		a.handle(&peer{ctx: ctx}, protocol.Request{ID: id, Type: "sessions.create-external", ServiceID: project.ID, Body: body}, func(e protocol.Event) error { event = e; return nil })
		if event.Type != "result" {
			t.Fatal(event)
		}
		result, ok := event.Data.(protocol.TaskStarted)
		if !ok {
			t.Fatalf("invalid response %#v", event.Data)
		}
		return result
	}
	empty := create("empty", protocol.ProjectSessionCreate{Title: "Empty"})
	if empty.Session.ServiceID != project.ID || empty.TaskID != "" {
		t.Fatal(empty)
	}
	deferred := create("deferred", protocol.ProjectSessionCreate{Title: "Deferred title", Message: "saved first message"})
	page, err := a.historyStore().read(deferred.Session.ID, 0)
	if err != nil || len(page.Session.QueuedMessages) != 1 || page.Session.QueuedMessages[0].Prompt != "saved first message" || page.Session.QueuedMessages[0].Status != "paused" || page.Session.Title != "Deferred title" {
		t.Fatal(page, err)
	}
	select {
	case <-started:
		t.Fatal("deferred message executed")
	case <-time.After(50 * time.Millisecond):
	}
	// Reopening history and startup queue recovery must not execute saved work.
	if err := a.historyStore().recoverQueue(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.changeQueuedTask(protocol.Request{Type: "tasks.promote", ServiceID: project.ID, SessionID: deferred.Session.ID, RunID: deferred.TaskID}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "saved first message" {
		t.Fatal("deferred message not resumable")
	}
	waitSessionStatus(t, a, deferred.Session.ID, "done")
	listed, listErr := a.historyStore().list(project.ID, "", 0)
	if listErr != nil {
		t.Fatal(listErr)
	}
	for _, session := range listed.Sessions {
		if session.ID == deferred.Session.ID && len(session.QueuedMessages) != 0 {
			t.Fatal("completed message still appears paused in session list")
		}
	}
	page, err = a.historyStore().read(deferred.Session.ID, 0)
	if err != nil || len(page.Entries) == 0 || page.Entries[0].Text != "saved first message" {
		t.Fatal(page, err)
	}
	immediate := create("immediate", protocol.ProjectSessionCreate{Title: "Run", Message: "execute first message", Execute: true})
	if immediate.TaskID == "" || awaitPrompt(t, started) != "execute first message" {
		t.Fatal(immediate)
	}
	waitSessionStatus(t, a, immediate.Session.ID, "done")
}
