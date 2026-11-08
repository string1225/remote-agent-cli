package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func TestMCPDiscoversIMRecipientsWithoutSourceSession(t *testing.T) {
	calls := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var in struct {
			AgentID string         `json:"agentId"`
			Type    string         `json:"type"`
			Body    map[string]any `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || r.URL.Path != "/api/agent/tools" || in.Type != "channels.sessions" || in.AgentID != "target" || in.Body["serviceId"] != "project" || in.Body["search"] != "小林" || in.Body["cursor"] != float64(50) {
			t.Error("recipient discovery lost routing or filters", in, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clientId":"target","sessions":[{"sessionId":"im","projectId":"project","peerName":"小林"}],"nextCursor":0}`))
	}))
	defer site.Close()
	c := agent.Config{ID: "source", Credential: "mock-secret", Path: filepath.Join(t.TempDir(), "config.json"), Server: site.URL}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_im_sessions","arguments":{"clientId":"target","projectId":"project","search":"小林","cursor":50}}}`
	var output bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path}, strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !strings.Contains(output.String(), "小林") || strings.Contains(output.String(), "mock-secret") || strings.Contains(output.String(), `"isError":true`) {
		t.Fatal("standalone agent cannot discover IM recipients", calls, output.String())
	}
	found := false
	for _, entry := range clientTools() {
		tool := entry.(map[string]any)
		if tool["name"] == "list_im_sessions" {
			found = tool["annotations"].(map[string]bool)["readOnlyHint"]
		}
	}
	if !found {
		t.Fatal("recipient discovery is not advertised as read-only")
	}
}
