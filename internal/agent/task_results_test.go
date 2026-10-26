package agent

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestCompletionSelectsFinalAssistantMessage(t *testing.T) {
	for _, tc := range []struct{ name, output, want string }{
		{"codex", `{"type":"item.completed","item":{"type":"reasoning","text":"internal reasoning"}}
{"type":"item.completed","item":{"type":"agent_message","phase":"commentary","text":"working"}}
{"type":"item.updated","item":{"id":"final","type":"agent_message","text":"partial"}}
{"type":"item.completed","item":{"id":"final","type":"agent_message","phase":"final_answer","text":"结论：测试通过，提交 abc123"}}
{"type":"item.completed","item":{"type":"command_execution","command":"go test","aggregated_output":"tool output"}}
`, "结论：测试通过，提交 abc123"},
		{"qoder", `{"type":"stream_event","event":{"delta":{"text":"streamed"}}}
{"type":"assistant","message":{"content":[{"text":"assistant candidate"}]}}
{"type":"result","result":"最终结论：检查通过"}
`, "最终结论：检查通过"},
		{"commentary-only", `{"type":"item.completed","item":{"type":"agent_message","phase":"commentary","text":"仍在处理"}}
`, "未获得可回传的最终结果"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
			s, err := a.historyStore().create(protocol.Service{ID: "project"}, "work")
			if err != nil {
				t.Fatal(err)
			}
			decoder := &transcript{native: func(string) error { return nil }, final: func(text string) error { return a.historyStore().saveTaskResult("task", text) }, text: func(stream, text string) error {
				role := "system"
				if stream == "stdout" {
					role = "assistant"
				}
				_, err := a.historyStore().append(s.ID, protocol.ChatEntry{TurnID: "task", Role: role, Kind: stream, Text: text})
				return err
			}}
			// Exercise JSONL split across arbitrary transport chunks.
			mid := len(tc.output) / 2
			for _, chunk := range []string{tc.output[:mid], tc.output[mid:]} {
				if err := decoder.write("stdout", chunk); err != nil {
					t.Fatal(err)
				}
			}
			if err := decoder.flush(); err != nil {
				t.Fatal(err)
			}
			text := a.taskCompletionText(queuedTask{result: protocol.TaskStarted{TaskID: "task", Session: s}}, nil)
			if !strings.Contains(text, tc.want) || strings.Contains(text, "internal reasoning") || strings.Contains(text, "tool output") || strings.Contains(text, "working") || strings.Contains(text, "streamed") {
				t.Fatalf("incorrect final result: %s", text)
			}
		})
	}
}

func TestCompletionReportsFailuresNoOutputAndPreservesLongResults(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Server: "https://site.test/agents"}}
	task := queuedTask{result: protocol.TaskStarted{TaskID: "task", Session: protocol.Conversation{ID: "missing", Title: "work"}}}
	for _, tc := range []struct {
		err  error
		want string
	}{{nil, "未获得可回传的最终结果"}, {errors.New("executor unavailable"), "executor unavailable"}, {context.Canceled, "任务取消"}} {
		if got := a.taskCompletionText(task, tc.err); !strings.Contains(got, tc.want) || strings.Contains(got, "#client=") {
			t.Fatal(got)
		}
	}
	result := strings.Repeat("中文结论🙂", 5000) + "\n验证和提交：abc123\nhttps://github.com/org/repo"
	if err := a.historyStore().saveTaskResult("task", result); err != nil {
		t.Fatal(err)
	}
	text := a.taskCompletionText(task, nil)
	parts := notificationParts(text)
	if len(parts) < 2 || strings.Join(parts, "") != text || !strings.Contains(text, result) {
		t.Fatal("long result truncated")
	}
	for _, part := range parts {
		if len(part) > 16000 || !utf8.ValidString(part) {
			t.Fatal("invalid IM part")
		}
	}
	failure := a.taskCompletionText(task, errors.New("timed out"))
	if !strings.Contains(failure, "可能不完整") || !strings.Contains(failure, result) {
		t.Fatal("failure discarded useful partial results")
	}
}

func TestIMSessionLinkFilteringPreservesOtherLinksAndWebHistory(t *testing.T) {
	input := "已排队 [查看会话](https://site.test/agents/#client=c&project=p&session=s)\nhttps://site.test/agents/#client=c&project=p&session=s\n[仓库](https://github.com/o/r) https://site.test/agents/download https://other.test/#client=c&project=p&session=s"
	got := withoutSessionLinks(input, "https://site.test/agents")
	if strings.Contains(got, "https://site.test/agents/#") || !strings.Contains(got, "[仓库](https://github.com/o/r)") || !strings.Contains(got, "https://site.test/agents/download") || !strings.Contains(got, "https://other.test/#") {
		t.Fatal(got)
	}
}

func TestLocalCompletionSurvivesRestartAndTargetsSavedConversationOnce(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), ID: "local"}}
	c := ChannelConfig{Provider: "feishu", ServiceID: "project", AppID: "mock"}
	project := protocol.Service{ID: "project"}
	origin, route, err := a.historyStore().channelSession(c, project, channelMessage{ChatID: "original", ChatType: "user"})
	if err != nil {
		t.Fatal(err)
	}
	_, other, err := a.historyStore().channelSession(c, project, channelMessage{ChatID: "other-group", ChatType: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.historyStore().saveTaskResult("task", "最终结果：已验证并提交"); err != nil {
		t.Fatal(err)
	}
	task := queuedTask{start: protocol.TaskStart{ServiceID: "project", Source: "im", Notify: true, NotifyAgentID: "local", NotifyServiceID: "project", NotifySessionID: origin.ID}, result: protocol.TaskStarted{TaskID: "task", Session: protocol.Conversation{ID: "web", Title: "work"}}}
	a.notifyTask(task, nil) // No channel or notification worker exists yet.
	restarted := &Agent{Config: a.Config}
	// Even a different project default must not redirect the saved source.
	restarted.channels = map[string]*channelRuntime{"project": {a: restarted, config: other, wake: make(chan struct{}, 1)}}
	ctx, cancel := context.WithCancel(context.Background())
	restarted.initNotifications(ctx)
	t.Cleanup(func() { cancel(); <-restarted.notifyDone })
	wait := func(check func() bool) {
		t.Helper()
		for deadline := time.Now().Add(5 * time.Second); !check(); {
			if time.Now().After(deadline) {
				t.Fatal("delivery did not settle")
			}
			time.Sleep(time.Millisecond)
		}
	}
	wait(func() bool { items, _ := restarted.historyStore().channelPending(route); return len(items) == 1 })
	items, _ := restarted.historyStore().channelPending(route)
	if items[0].SessionID != origin.ID || !strings.Contains(items[0].Text, "已验证并提交") {
		t.Fatal(items)
	}
	if wrong, _ := restarted.historyStore().channelPending(other); len(wrong) != 0 {
		t.Fatal("broadcast or wrong default destination")
	}
	if err := restarted.historyStore().channelDelivered(items[0].ID); err != nil {
		t.Fatal(err)
	}
	restarted.notifyTask(task, nil)
	wait(func() bool { pending, _ := restarted.historyStore().remotePending(); return len(pending) == 0 })
	if duplicates, _ := restarted.historyStore().channelPending(route); len(duplicates) != 0 {
		t.Fatal("duplicate result after retry")
	}
	page, _ := restarted.historyStore().read(origin.ID, 0)
	if len(page.Entries) != 1 {
		t.Fatal("duplicate result in IM history", len(page.Entries))
	}
}

func TestExplicitFinalNotificationSuppressesOnlySuccessfulAutomaticResult(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	s, err := a.historyStore().create(protocol.Service{ID: "project"}, "work")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.historyStore().update(s.ID, func(v *protocol.Conversation) { v.TaskID = "task" }); err != nil {
		t.Fatal(err)
	}
	if err := RecordFinalNotification(a.Config, s.ID); err != nil {
		t.Fatal(err)
	}
	// Later assistant text must preserve the durable notification acknowledgement.
	if err := a.historyStore().saveTaskResult("task", "已发送的最终结果"); err != nil {
		t.Fatal(err)
	}
	task := queuedTask{result: protocol.TaskStarted{TaskID: "task", Session: s}}
	if got := a.taskCompletionText(task, nil); got != "" {
		t.Fatal("duplicate final result", got)
	}
	if got := a.taskCompletionText(task, errors.New("later failure")); !strings.Contains(got, "later failure") {
		t.Fatal("suppressed subsequent failure")
	}
}

func TestInterruptedTaskReportsToPersistedOriginWithoutReplay(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "worker"}}
	s, err := a.historyStore().create(protocol.Service{ID: "work"}, "work")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.historyStore().update(s.ID, func(v *protocol.Conversation) {
		v.TaskID = "task"
		v.TaskStatus = "running"
		v.NotifyAgentID = "origin-client"
		v.NotifyServiceID = "origin-project"
		v.NotifySessionID = "origin-im"
	}); err != nil {
		t.Fatal(err)
	}
	a.recoverTasks()
	a.recoverTasks()
	pending, err := a.historyStore().remotePending()
	if err != nil || len(pending) != 1 || pending[0].SessionID != "origin-im" || pending[0].AgentID != "origin-client" || !strings.Contains(pending[0].Text, "重启") {
		t.Fatal(pending, err)
	}
}
