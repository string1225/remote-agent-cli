package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func TestStandaloneMCPAdvertisesNotificationAndUploadWithoutSession(t *testing.T) {
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "fixture", Credential: "mock-secret", Server: "https://relay.example.test"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"not authorized to an implicit recipient"}}}
`
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path}, strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatal("standalone handshake failed")
	}
	var list struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &list); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range list.Result.Tools {
		names[tool.Name] = true
	}
	if !names["upload_file"] || !names["notify_im_session"] {
		t.Fatal("standalone agent cannot discover sharing capabilities")
	}
	var call struct {
		Result struct {
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[2]), &call); err != nil || !call.Result.IsError {
		t.Fatal("standalone notification accepted an implicit recipient", err)
	}
	if strings.Contains(out.String(), c.Credential) {
		t.Fatal("MCP advertised a credential")
	}
}
