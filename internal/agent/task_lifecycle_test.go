package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestTaskAcceptanceRollsBackIdentityPromptReceiptAndQueueTogether(t *testing.T) {
	a, project := concurrencyFixture(t)
	h := a.historyStore()
	previous, err := h.create(project, "original")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.rememberTask("duplicate", protocol.TaskStarted{Session: previous, TaskID: "previous"}); err != nil {
		t.Fatal(err)
	}
	task := queuedTask{newSession: true, requestKey: "duplicate", recorded: true, start: protocol.TaskStart{ServiceID: project.ID, Prompt: "new prompt", Source: "web"}, result: protocol.TaskStarted{TaskID: "new-task", Session: protocol.Conversation{ID: "new-session", ServiceID: project.ID, Provider: project.Provider, Title: "New", TaskID: "new-task", TaskStatus: "queued"}}}
	if err := a.persistQueuedTask(task); err == nil {
		t.Fatal("duplicate receipt accepted")
	}
	if _, err := h.get("new-session"); err == nil {
		t.Fatal("failed acceptance left a session")
	}
	if _, found, err := h.queueRecord("new-task"); err != nil || found {
		t.Fatal(found, err)
	}
	task.requestKey = "unique"
	if err := a.persistQueuedTask(task); err != nil {
		t.Fatal(err)
	}
	accepted, found, err := h.acceptedTask("unique")
	if err != nil || !found || accepted.TaskID != "new-task" {
		t.Fatal(accepted, found, err)
	}
	page, err := h.read("new-session", 0)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Text != "new prompt" {
		t.Fatal(page, err)
	}
	if err := h.markTaskRunning(task); err != nil {
		t.Fatal(err)
	}
	running, _ := h.get("new-session")
	if running.TaskStartedAt == 0 || running.TaskAcceptedAt == 0 || running.TaskCompletedAt != 0 {
		t.Fatal(running)
	}
	if err := h.finishQueuedTask(task, "failed", errors.New("simulated failure")); err != nil {
		t.Fatal(err)
	}
	err = h.transact(false, func(tx *bolt.Tx) error {
		session, err := sessionFrom(tx, "new-session")
		if err != nil {
			return err
		}
		row, found, err := queueRecordFrom(tx, "new-task")
		if err != nil {
			return err
		}
		if !found || row.Status != "failed" || session.TaskStatus != "failed" || session.TaskError != "simulated failure" || session.TaskCompletedAt < session.TaskStartedAt {
			t.Fatal(row, session)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFullQueueRejectsBeforeCreatingDurableTask(t *testing.T) {
	a, project := concurrencyFixture(t)
	if _, err := a.historyStore().create(project, "existing"); err != nil {
		t.Fatal(err)
	}
	a.tasks = make(chan queuedTask, 1)
	a.tasks <- queuedTask{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	task := queuedTask{ctx: ctx, cancel: cancel, newSession: true, requestKey: "full", start: protocol.TaskStart{ServiceID: project.ID, Prompt: "retain"}, result: protocol.TaskStarted{TaskID: "full-task", Session: protocol.Conversation{ID: "full-session", ServiceID: project.ID}}}
	if a.enqueueTask(task) {
		t.Fatal("full queue accepted task")
	}
	if _, err := a.historyStore().get("full-session"); err == nil {
		t.Fatal("rejected task left a session")
	}
	if _, found, err := a.historyStore().acceptedTask("full"); err != nil || found {
		t.Fatal(found, err)
	}
}
