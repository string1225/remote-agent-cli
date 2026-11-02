package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestMCPTodosUseCurrentSessionAndStablePublicationIdentity(t *testing.T) {
	commands := []protocol.TodoCommand{}
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agent/todos" || r.URL.Query().Get("id") != "current-client" || r.Header.Get("Authorization") != "Bearer mock-secret" {
			t.Error("missing endpoint/auth identity")
		}
		var in protocol.TodoCommand
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			t.Error("bad command")
		}
		commands = append(commands, in)
		io.WriteString(w, `{"todos":[],"id":"todo","version":1}`)
	}))
	defer site.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "current-client", Credential: "mock-secret", Server: site.URL}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_todos","arguments":{"status":"open","pageCursor":"next"}}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"publish_todo","arguments":{"title":"choose","body":"A or B?","projectId":"foreign","sessionId":"foreign"}}}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"publish_todo","arguments":{"title":"choose","body":"A or B?"}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"update_todo","arguments":{"todoId":"todo","operation":"complete","version":8}}}`)
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "current-project", "--session", "current-session", "--turn", "turn"}, input, &out); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 4 || commands[0].Operation != "list" || commands[0].Cursor != "next" || commands[1].ProjectID != "current-project" || commands[1].SessionID != "current-session" || commands[1].RequestID != commands[2].RequestID || commands[3].Operation != "complete" || commands[3].Version != 8 {
		t.Fatal("MCP todos lost context/version/retry identity", commands)
	}
	if strings.Contains(out.String(), "mock-secret") || !strings.Contains(out.String(), "publish_todo") {
		t.Fatal("tool response is incomplete or leaked device credentials")
	}
}
