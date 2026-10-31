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
)

func TestMCPSessionOwnsDelegationAndExplicitMessages(t *testing.T) {
	var operations []string
	var spawnIDs []string
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-device-secret" || r.URL.Query().Get("id") != "source-client" {
			t.Error("lost device authentication")
		}
		if r.URL.Path == "/api/agent/catalog" {
			operations = append(operations, "catalog")
			io.WriteString(w, `{"clients":[]}`)
			return
		}
		var in struct {
			AgentID string         `json:"agentId"`
			Type    string         `json:"type"`
			Body    map[string]any `json:"body"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			t.Error("invalid tool request")
		}
		operations = append(operations, in.Type)
		switch in.Type {
		case "sessions.spawn":
			if in.AgentID != "target" || in.Body["parentSessionId"] != "current-im" || in.Body["serviceId"] != "target-project" || in.Body["notify"] == true || in.Body["notifySessionId"] != nil {
				t.Error("spawn inherited automatic routing", in)
			}
			spawnIDs = append(spawnIDs, in.Body["requestId"].(string))
			io.WriteString(w, `{"session":{"id":"child"},"taskId":"task"}`)
		case "sessions.control":
			if in.Body["callerSessionId"] != "current-im" || in.Body["sessionId"] != "child" || in.Body["taskId"] != "task" {
				t.Error("lost management identity", in)
			}
			io.WriteString(w, `{"complete":true,"result":"child-result"}`)
		case "channels.notify":
			if in.AgentID != "source-client" || in.Body["sessionId"] != "current-im" || in.Body["serviceId"] != "current-project" {
				t.Error("message escaped current IM session", in)
			}
			io.WriteString(w, `{"queued":true}`)
		default:
			t.Error("unexpected request", in.Type)
		}
	}))
	defer site.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "source-client", Credential: "local-device-secret", Server: site.URL}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_clients_projects","arguments":{}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"start_web_session","arguments":{"clientId":"target","projectId":"target-project","prompt":"work"}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"start_web_session","arguments":{"clientId":"target","projectId":"target-project","prompt":"work"}}}
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"wait_session","arguments":{"clientId":"target","sessionId":"child","taskId":"task","waitSeconds":1}}}
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"explicit message"}}}`)
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "current-project", "--session", "current-im", "--turn", "turn-1", "--source", "im"}, input, &out); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 5 || len(spawnIDs) != 2 || spawnIDs[0] != spawnIDs[1] {
		t.Fatal("operation sequence or retry identity changed", operations, spawnIDs)
	}
	if strings.Contains(out.String(), "local-device-secret") || !strings.Contains(out.String(), "child-result") {
		t.Fatal("missing result or exposed credential")
	}
	var init struct {
		Result struct {
			Instructions string `json:"instructions"`
		} `json:"result"`
	}
	if json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&init) != nil || !strings.Contains(init.Result.Instructions, agent.SessionInstructions) || strings.Contains(init.Result.Instructions, "先列出") {
		t.Fatal("incorrect MCP guidance", out.String())
	}
}

func TestMCPWebSessionDoesNotInheritIMDelivery(t *testing.T) {
	calls := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; io.WriteString(w, `{}`) }))
	defer site.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "worker", Credential: "mock", Server: site.URL}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "work", "--session", "web-child", "--source", "web", "--notify-client", "old-origin", "--notify-project", "old-project", "--notify-session", "old-im"}, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"must not auto-route"}}}`), &out)
	if err != nil || calls != 0 || !strings.Contains(out.String(), "isError") {
		t.Fatal("legacy route sent an unsolicited message", err, out.String())
	}
}
