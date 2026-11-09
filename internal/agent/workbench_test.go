package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestWorkbenchValidatesParentProjectAndRedactsDiagnostics(t *testing.T) {
	a, project := concurrencyFixture(t)
	a.Config.ID = "client"
	a.Config.Credential = "PRIVATE_CREDENTIAL"
	parent, err := a.historyStore().create(project, "PRIVATE_PROMPT")
	if err != nil {
		t.Fatal(err)
	}
	other := project
	other.ID = "other"
	a.Config.Services = append(a.Config.Services, other)
	request := func(req protocol.Request) protocol.Event {
		t.Helper()
		var event protocol.Event
		a.handleWorkbench(context.Background(), req, func(value protocol.Event) error { event = value; return nil })
		return event
	}
	for _, id := range []string{"other", "missing"} {
		if event := request(protocol.Request{Type: "sessions.children", ServiceID: id, SessionID: parent.ID}); event.Type != "error" {
			t.Fatal("accepted parent from a different project", event)
		}
	}
	child, err := a.historyStore().create(project, "Child")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.historyStore().update(child.ID, func(session *protocol.Conversation) {
		session.Kind = "web"
		session.ParentAgentID, session.ParentSessionID = a.Config.ID, parent.ID
		session.NativeID = "PRIVATE_NATIVE_ID"
	}); err != nil {
		t.Fatal(err)
	}
	event := request(protocol.Request{Type: "sessions.children", ServiceID: project.ID, SessionID: parent.ID})
	page, ok := event.Data.(protocol.ConversationPage)
	if !ok || len(page.Sessions) != 1 || page.Sessions[0].ID != child.ID || page.Sessions[0].NativeID != "" || page.Sessions[0].Workspace != "" {
		t.Fatal(event)
	}
	event = request(protocol.Request{Type: "diagnostics.snapshot"})
	data, err := json.Marshal(event.Data)
	if err != nil || event.Type != "result" || strings.Contains(string(data), "PRIVATE") || strings.Contains(string(data), project.Workspace) {
		t.Fatal(string(data), err)
	}
}
