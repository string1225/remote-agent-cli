package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestTodoOfflineReplyCanBeCancelledWithoutDroppingTheQuestion(t *testing.T) {
	f := setup(t)
	owner := f.user("owner")
	host, _, _, _ := todoTestDevice(t, f, owner)
	source, cfg, _, stopSource := todoTestDevice(t, f, owner)
	f.request(owner, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 200)
	p := f.peer(owner, source.ID)
	p.send(t, protocol.Request{ID: "create", Type: "sessions.create", ServiceID: cfg.Services[0].ID})
	event := p.next(t, "create", "result")
	raw, _ := json.Marshal(event.Data)
	var session protocol.Conversation
	_ = json.Unmarshal(raw, &session)
	raw = f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: "create", RequestID: "offline-question", Title: "Which option?", Body: "Keep this context", ClientID: source.ID, ProjectID: cfg.Services[0].ID, SessionID: session.ID}, 200)
	var row protocol.Todo
	_ = json.Unmarshal(raw, &row)
	stopSource()
	command := protocol.TodoCommand{Operation: "reply", ID: row.ID, Version: row.Version, RequestID: "reply-offline", Message: "Choose B"}
	f.request(owner, "POST", "/api/todos", command, 409)
	raw = f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: "get", ID: row.ID}, 200)
	_ = json.Unmarshal(raw, &row)
	if row.Reply == nil || row.Reply.State != "pending" || row.Reply.Text != "Choose B" {
		t.Fatal("offline reply was not retained", string(raw))
	}
	f.request(owner, "POST", "/api/todos", protocol.TodoCommand{Operation: "cancel", ID: row.ID, Version: row.Version}, 200)
	f.request(owner, "POST", "/api/todos", command, 409)
	raw = f.request(owner, "GET", "/api/todos", nil, 200)
	if !strings.Contains(string(raw), "Keep this context") || !strings.Contains(string(raw), `"state":"cancelled"`) {
		t.Fatal("cancellation dropped context or allowed delivery", string(raw))
	}
}
