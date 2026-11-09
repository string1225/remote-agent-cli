package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestTodoOptionsAndReplyCompleteBeforeExecutionFinishes(t *testing.T) {
	f := setupAt(t, "/agents")
	owner := f.user("decisions")
	started, finish := make(chan string, 2), make(chan struct{})
	host, cfg, _, _ := todoTestDevice(t, f, owner, func(ctx context.Context, _ agent.Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		select {
		case <-finish:
		case <-ctx.Done():
		}
		return nil
	})
	defer close(finish)
	f.request(owner, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 200)
	p := f.peer(owner, host.ID)
	p.send(t, protocol.Request{ID: "session", Type: "sessions.create", ServiceID: cfg.Services[0].ID})
	event := p.next(t, "session", "result")
	data, _ := json.Marshal(event.Data)
	var session protocol.Conversation
	json.Unmarshal(data, &session)
	create := protocol.TodoCommand{Operation: "create", RequestID: "decision", Title: "发布窗口", Body: "选一个窗口", ClientID: host.ID, ProjectID: session.ServiceID, SessionID: session.ID,
		Options: []protocol.TodoOption{{Label: "A · 周五", Description: "保留周末观察时间"}, {Label: "B · 周一", Description: "工作日处理"}}}
	var row protocol.Todo
	json.Unmarshal(f.request(owner, "POST", "/api/todos", create, 200), &row)
	if len(row.Options) != 2 || row.Options[0] != create.Options[0] || row.Body != create.Body {
		t.Fatal("options/body did not survive client storage", row)
	}
	json.Unmarshal(f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: "update", ID: row.ID, Version: row.Version, Title: "新标题", Body: "新正文"}, 200), &row)
	if len(row.Options) != 2 || row.Body != "新正文" {
		t.Fatal("inline edit lost choices", row)
	}
	command := protocol.TodoCommand{Operation: "reply", ID: row.ID, Version: row.Version, RequestID: "choose-A", Message: "A · 周五\n补充：检查健康"}
	json.Unmarshal(f.request(owner, "POST", "/api/todos", command, 200), &row)
	if row.Status != "completed" || row.Reply == nil || row.Reply.State != "delivered" || len(row.Options) != 2 {
		t.Fatal("human reply did not immediately process the todo", row)
	}
	select {
	case text := <-started:
		if !strings.Contains(text, command.Message) {
			t.Fatal("lost original reply", text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reply never entered original session")
	}
	if strings.Contains(string(f.request(owner, "GET", "/api/todos?status=open", nil, 200)), row.ID) {
		t.Fatal("processed question stayed in pending while execution is blocked")
	}
	f.request(owner, "POST", "/api/todos", command, 200)
	select {
	case <-started:
		t.Fatal("retry created another execution")
	default:
	}
	// Notification completion and deletion must never submit a message.
	for _, operation := range []string{"complete", "delete"} {
		create.RequestID = operation
		create.Options = nil
		var notice protocol.Todo
		json.Unmarshal(f.request(owner, "POST", "/api/todos", create, 200), &notice)
		json.Unmarshal(f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: operation, ID: notice.ID, Version: notice.Version}, 200), &notice)
		if notice.Reply != nil || len(notice.Options) != 0 || operation == "delete" && notice.Status != "cancelled" || operation == "complete" && notice.Status != "completed" {
			t.Fatal("notification action submitted a reply or lost state", operation, notice)
		}
	}
	select {
	case <-started:
		t.Fatal("complete/delete woke the agent")
	default:
	}
	// Choices remain on the client and follow the existing complete-generation migration.
	target, _, _, _ := todoTestDevice(t, f, owner)
	f.request(owner, "PUT", "/api/default-host", map[string]string{"agentId": target.ID}, 200)
	json.Unmarshal(f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: "get", ID: row.ID}, 200), &row)
	if len(row.Options) != 2 || row.Options[1] != createOptionB() || row.Body != "新正文" {
		t.Fatal("host migration lost structured choices", row)
	}
}

func createOptionB() protocol.TodoOption {
	return protocol.TodoOption{Label: "B · 周一", Description: "工作日处理"}
}
