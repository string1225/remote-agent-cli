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

func TestMCPToolsUseEnrolledSiteAndProjectSessionContext(t *testing.T) {
	var operations []string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-device-secret" || r.URL.Query().Get("id") != "source-client" {
			t.Error("wrong device identity")
		}
		operations = append(operations, r.URL.Path)
		if r.URL.Path == "/api/agent/catalog" {
			io.WriteString(w, `{"clients":[{"id":"target","projects":[{"id":"project-b"}]}]}`)
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
		if in.Type == "tasks.start" {
			if in.AgentID != "target" || in.Body["serviceId"] != "project-b" || in.Body["parentSessionId"] != "origin-im" || in.Body["source"] != "im" || in.Body["notify"] != true {
				t.Error("lost caller or target context")
			}
			io.WriteString(w, `{"session":{"id":"new-web","kind":"web"},"taskId":"task"}`)
		} else if in.Type == "channels.notify" {
			if in.AgentID != "source-client" || in.Body["serviceId"] != "project-a" || in.Body["message"] != "done" {
				t.Error("notification not scoped to current project")
			}
			io.WriteString(w, `{"queued":true}`)
		} else {
			t.Error("unexpected operation")
		}
	}))
	defer up.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), Server: up.URL, ID: "source-client", Credential: "local-device-secret"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_clients_projects","arguments":{}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"start_web_session","arguments":{"clientId":"target","projectId":"project-b","prompt":"work"}}}
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"done"}}}
`)
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "project-a", "--session", "origin-im", "--source", "im"}, input, &out); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 3 || !strings.Contains(out.String(), "new-web") || strings.Contains(out.String(), "local-device-secret") {
		t.Fatal("tool calls missing or leaked credentials")
	}
	decoder := json.NewDecoder(&out)
	count := 0
	for decoder.More() {
		var v map[string]any
		if err := decoder.Decode(&v); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 5 {
		t.Fatal("notifications produced replies or lost responses", count)
	}
}
