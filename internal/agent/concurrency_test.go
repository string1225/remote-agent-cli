package agent

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func concurrencyFixture(t *testing.T) (*Agent, protocol.Service) {
	t.Helper()
	root := t.TempDir()
	service := protocol.Service{ID: "project", Provider: "codex", Workspace: root}
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), AllowedRoots: []string{root}, Services: []protocol.Service{service}}, Providers: map[string]Provider{"codex": {Available: true}}}
	return a, service
}

func awaitPrompt(t *testing.T, started <-chan string) string {
	t.Helper()
	select {
	case prompt := <-started:
		return prompt
	case <-time.After(5 * time.Second):
		t.Fatal("task did not start")
		return ""
	}
}

func awaitDone(t *testing.T, events <-chan protocol.Event, id string) protocol.Event {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case event := <-events:
			if event.Type == "done" && event.ID == id {
				return event
			}
		case <-timer.C:
			t.Fatalf("task %s did not finish", id)
			return protocol.Event{}
		}
	}
}

func TestParallelRunsIsolateHistoryCancellationAndMaintenance(t *testing.T) {
	a, service := concurrencyFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &peer{ctx: ctx, cancel: cancel}
	other := &peer{ctx: ctx, cancel: cancel}
	started := make(chan string, 4)
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		started <- prompt
		if err := send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "output-" + prompt}}); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}
	var sessions []protocol.Conversation
	var events []chan protocol.Event
	for i, prompt := range []string{"first", "second", "third"} {
		session, err := a.historyStore().create(service, prompt)
		if err != nil {
			t.Fatal(err)
		}
		sessions = append(sessions, session)
		out := make(chan protocol.Event, 16)
		events = append(events, out)
		caller, id := p, prompt
		if i == 2 {
			caller, id = other, "first" // Request IDs may collide across peers.
		}
		a.handle(caller, protocol.Request{ID: id, Type: "run", ServiceID: service.ID, SessionID: session.ID, Prompt: prompt}, func(e protocol.Event) error { out <- e; return nil })
		if got := awaitPrompt(t, started); got != prompt {
			t.Fatal(got)
		}
	}
	var rejected protocol.Event
	a.handle(other, protocol.Request{ID: "duplicate", Type: "run", ServiceID: service.ID, SessionID: sessions[0].ID, Prompt: "duplicate"}, func(e protocol.Event) error { rejected = e; return nil })
	if rejected.Error != errSessionRunning.Error() {
		t.Fatal("same native session was allowed to run concurrently", rejected)
	}
	a.handle(p, protocol.Request{ID: "cancel-first", Type: "cancel", RunID: "first"}, func(protocol.Event) error { return nil })
	if event := awaitDone(t, events[0], "first"); !strings.Contains(event.Error, "canceled") {
		t.Fatal(event)
	}
	a.mu.Lock()
	remaining := len(a.runs)
	a.mu.Unlock()
	if remaining != 2 {
		t.Fatalf("canceling one task changed other tasks: %d remain", remaining)
	}
	analysisDone := make(chan error, 1)
	go func() {
		analysisDone <- a.runSayso(ctx, p, service.Workspace, "analysis", "", "", true, func(string, string) error { return nil })
	}()
	if got := awaitPrompt(t, started); got != "analysis" {
		t.Fatal(got)
	}
	if err := a.PauseUpdates(ctx, false); err == nil {
		t.Fatal("update ignored remaining parallel tasks")
	}
	updateCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	if err := a.PauseUpdates(updateCtx, true); err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"second", "first"} {
		if event := awaitDone(t, events[i+1], id); !strings.Contains(event.Error, "canceled") {
			t.Fatal(event)
		}
	}
	if err := <-analysisDone; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for i, session := range sessions {
		page, err := a.historyStore().read(session.ID, 0)
		if err != nil {
			t.Fatal(err)
		}
		users, output := 0, ""
		for _, entry := range page.Entries {
			if entry.Role == "user" {
				users++
			}
			if entry.Role == "assistant" {
				output += entry.Text
			}
		}
		if users != 1 || output != "output-"+[]string{"first", "second", "third"}[i] {
			t.Fatal("parallel histories mixed or rejected turn was saved", page)
		}
	}
}

func TestBackgroundTurnsKeepSessionOrderWhileOtherSessionsRun(t *testing.T) {
	a, service := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	started := make(chan string, 3)
	release := make(chan struct{})
	finished := make(chan struct {
		id  string
		err error
	}, 3)
	a.Execute = func(ctx context.Context, _ Provider, service protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		if prompt == "second" && service.ResumeID != "native-first" {
			t.Error("queued turn did not load latest native context", service.ResumeID)
		}
		started <- prompt
		if prompt == "first" {
			if err := send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"native-first\"}\n"}}); err != nil {
				return err
			}
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		} else if prompt == "other" {
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	}
	session, err := a.historyStore().create(service, "shared IM context")
	if err != nil {
		t.Fatal(err)
	}
	independent, err := a.historyStore().create(service, "independent")
	if err != nil {
		t.Fatal(err)
	}
	enqueue := func(id string, session protocol.Conversation) {
		t.Helper()
		taskCtx, cancel := context.WithCancel(ctx)
		if !a.enqueueTask(queuedTask{ctx: taskCtx, cancel: cancel, start: protocol.TaskStart{ServiceID: service.ID, Prompt: id}, result: protocol.TaskStarted{TaskID: id, Session: session}, complete: func(err error) {
			finished <- struct {
				id  string
				err error
			}{id, err}
		}}) {
			cancel()
			t.Fatal("task was rejected")
		}
	}
	enqueue("first", session)
	if got := awaitPrompt(t, started); got != "first" {
		t.Fatal(got)
	}
	enqueue("second", session)
	enqueue("other", independent)
	if got := awaitPrompt(t, started); got != "other" {
		t.Fatal("same-session turn overtook its predecessor", got)
	}
	if !a.cancelTask("other") {
		t.Fatal("independent background task not found")
	}
	select {
	case result := <-finished:
		if result.id != "other" || !errors.Is(result.err, context.Canceled) {
			t.Fatal(result)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("background cancellation did not finish")
	}
	close(release)
	if got := awaitPrompt(t, started); got != "second" {
		t.Fatal(got)
	}
	for _, id := range []string{"first", "second"} {
		select {
		case result := <-finished:
			if result.id != id || result.err != nil {
				t.Fatal(result)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("ordered task did not complete")
		}
	}
}

func TestForcedMaintenanceCancelsRunningAndQueuedBackgroundTurns(t *testing.T) {
	a, service := concurrencyFixture(t)
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	started := make(chan string, 3)
	finished := make(chan error, 3)
	a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		<-ctx.Done()
		return ctx.Err()
	}
	session, err := a.historyStore().create(service, "shared")
	if err != nil {
		t.Fatal(err)
	}
	independent, err := a.historyStore().create(service, "independent")
	if err != nil {
		t.Fatal(err)
	}
	for i, id := range []string{"first", "queued", "other"} {
		target := session
		if i == 2 {
			target = independent
		}
		taskCtx, cancel := context.WithCancel(ctx)
		if !a.enqueueTask(queuedTask{ctx: taskCtx, cancel: cancel, start: protocol.TaskStart{ServiceID: service.ID, Prompt: id}, result: protocol.TaskStarted{TaskID: id, Session: target}, complete: func(err error) { finished <- err }}) {
			cancel()
			t.Fatal("task was rejected")
		}
	}
	prompts := map[string]bool{awaitPrompt(t, started): true, awaitPrompt(t, started): true}
	if !prompts["first"] || !prompts["other"] {
		t.Fatal("queued turn ran concurrently with its own session", prompts)
	}
	if err := a.PauseUpdates(ctx, false); err == nil {
		t.Fatal("ordinary maintenance accepted unfinished background tasks")
	}
	updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := a.PauseUpdates(updateCtx, true); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := <-finished; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	select {
	case prompt := <-started:
		t.Fatal("canceled queued turn started a process", prompt)
	default:
	}
	for _, id := range []string{session.ID, independent.ID} {
		v, err := a.historyStore().get(id)
		if err != nil || v.TaskStatus != "cancelled" {
			t.Fatal("maintenance returned before task state was saved", v, err)
		}
	}
}
