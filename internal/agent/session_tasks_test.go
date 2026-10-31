package agent

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func waitSessionStatus(t *testing.T, a *Agent, id, status string) protocol.Conversation {
	t.Helper()
	until := time.Now().Add(5 * time.Second)
	for time.Now().Before(until) {
		s, err := a.historyStore().get(id)
		if err == nil && s.TaskStatus == status {
			return s
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("session %s did not become %s", id, status)
	return protocol.Conversation{}
}

func TestExistingWebTurnsSurvivePeerCloseRunConcurrentlyAndResume(t *testing.T) {
	a, project := concurrencyFixture(t)
	daemon, stop := context.WithCancel(context.Background())
	a.initTasks(daemon)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	first, _ := a.historyStore().create(project, "first")
	second, _ := a.historyStore().create(project, "second")
	started := make(chan string, 3)
	release := make(chan struct{})
	var executions atomic.Int32
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) < 119*time.Minute || time.Until(deadline) > 2*time.Hour {
			t.Error("turn does not have a two-hour deadline")
		}
		if prompt == "continue" && s.ResumeID != "native-first" {
			t.Error("continuation lost native context", s.ResumeID)
		}
		started <- prompt
		if prompt == "other" {
			<-ctx.Done()
			return ctx.Err()
		}
		if prompt == "first" {
			if err := send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"thread.started\",\"thread_id\":\"native-first\"}\n"}}); err != nil {
				return err
			}
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "result"}})
	}
	caller, closeCaller := context.WithCancel(context.Background())
	p := &peer{ctx: caller, cancel: closeCaller}
	req := protocol.Request{ID: "first-turn", Type: "tasks.run", ServiceID: project.ID, SessionID: first.ID, Prompt: "first"}
	var accepted protocol.Event
	a.handle(p, req, func(e protocol.Event) error { accepted = e; return nil })
	if accepted.Type != "result" {
		t.Fatal("turn not accepted", accepted)
	}
	p.close()
	if _, err := a.startSessionTask(protocol.Request{ID: "other-turn", ServiceID: project.ID, SessionID: second.ID, Prompt: "other"}); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{awaitPrompt(t, started): true, awaitPrompt(t, started): true}
	if !seen["first"] || !seen["other"] {
		t.Fatal("different sessions did not run concurrently", seen)
	}
	if _, err := a.startSessionTask(req); err != nil {
		t.Fatal("lost-ack retry failed", err)
	}
	if _, err := a.startSessionTask(protocol.Request{ID: "overlap", ServiceID: project.ID, SessionID: first.ID, Prompt: "unsafe overlap"}); !errors.Is(err, errSessionRunning) {
		t.Fatal("same native session overlapped", err)
	}
	close(release)
	waitSessionStatus(t, a, first.ID, "done")
	if executions.Load() != 2 {
		t.Fatal("retry replayed execution")
	}
	if _, err := a.startSessionTask(protocol.Request{ID: "next-turn", ServiceID: project.ID, SessionID: first.ID, Prompt: "continue"}); err != nil {
		t.Fatal(err)
	}
	if awaitPrompt(t, started) != "continue" {
		t.Fatal("continuation not started")
	}
	waitSessionStatus(t, a, first.ID, "done")
	if _, err := a.startSessionTask(req); err != nil {
		t.Fatal(err)
	}
	if executions.Load() != 3 {
		t.Fatal("old acknowledged turn replayed after a newer turn")
	}
	if !a.cancelTask("other-turn") {
		t.Fatal("cannot cancel detached task")
	}
	waitSessionStatus(t, a, second.ID, "cancelled")
	page, err := a.historyStore().read(first.ID, 0)
	if err != nil || page.Entries[0].Role != "user" || page.Entries[0].Text != "first" {
		t.Fatal("first user entry lost", page, err)
	}
	users := 0
	for _, e := range page.Entries {
		if e.Role == "user" {
			users++
		}
	}
	if users != 2 {
		t.Fatal("retried user messages duplicated", users)
	}
}

func TestTimeoutExplainsContinuationAndToolRecordsOmitSkillContents(t *testing.T) {
	err := taskTimeoutError()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline cause lost")
	}
	message := err.Error()
	for _, want := range []string{"任务执行超时", "2 小时", "同一会话", "继续"} {
		if !strings.Contains(message, want) {
			t.Fatal(message)
		}
	}
	var tools, assistant strings.Builder
	decoder := &transcript{native: func(string) error { return nil }, text: func(stream, text string) error {
		if stream == "tool" {
			tools.WriteString(text)
		} else {
			assistant.WriteString(text)
		}
		return nil
	}}
	for _, line := range []string{
		`{"type":"item.completed","item":{"type":"command_execution","command":"Get-Content SKILL.md","aggregated_output":"PRIVATE SKILL CONTENT"}}`,
		`{"type":"item.completed","item":{"type":"mcp_tool_call","server":"docs","tool":"read","status":"completed"}}`,
		`{"type":"item.completed","item":{"type":"reasoning","text":"PRIVATE REASONING"}}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"final result"}}`,
	} {
		if err := decoder.write("stdout", line+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(tools.String(), "Get-Content SKILL.md") || !strings.Contains(tools.String(), "docs/read") || strings.Contains(tools.String(), "PRIVATE") {
		t.Fatal(tools.String())
	}
	if assistant.String() != "final result\n" {
		t.Fatal(assistant.String())
	}
}
