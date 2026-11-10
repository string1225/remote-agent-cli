package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestMCPVerifiedManagerInstructionsOnEveryInvocationAndExecutorScopeIsolation(t *testing.T) {
	for _, scope := range []string{"account", "client", "project", ""} {
		t.Run(scope, func(t *testing.T) {
			root := t.TempDir()
			original := "原项目规则\n"
			if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			c := agent.Config{Path: filepath.Join(root, "config.json"), ID: "fixture", Credential: "fake-only", Server: "https://example.test/agents", Services: []protocol.Service{{ID: "project", Scope: scope, Workspace: root}}}
			if err := c.Save(); err != nil {
				t.Fatal(err)
			}
			for turn := 0; turn < 2; turn++ {
				var out bytes.Buffer
				// --scope is deliberately wrong; the configuration is authoritative.
				err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "project", "--session", "session", "--scope", "account"}, strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}\n"), &out)
				if err != nil {
					t.Fatal(err)
				}
				var result struct {
					Result struct {
						Instructions string `json:"instructions"`
					} `json:"result"`
				}
				if err := json.Unmarshal(out.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				guidance := result.Result.Instructions
				if scope == "" {
					if !strings.Contains(guidance, agent.SessionInstructions) || strings.Contains(guidance, "你是会话管理员") {
						t.Fatal("ordinary executor was promoted by a CLI flag")
					}
				} else if !strings.Contains(guidance, "你是会话管理员") || strings.Contains(guidance, "你可以直接完成任务") || strings.Contains(guidance, "小事可以直接处理") {
					t.Fatal("manager invocation kept conflicting autonomy guidance")
				}
			}
			if scope == "project" || scope == "" {
				got, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
				if string(got) != original {
					t.Fatal("shared project rules overwritten")
				}
			}
		})
	}
}

func TestMCPDiscoversTopicRecoveryAndSchedulingTools(t *testing.T) {
	found := map[string]bool{}
	for _, raw := range clientTools() {
		tool := raw.(map[string]any)
		name := tool["name"].(string)
		found[name] = true
		if name == "start_web_session" {
			schema := tool["inputSchema"].(map[string]any)
			if schema["properties"].(map[string]any)["topic"] == nil {
				t.Fatal("manager cannot pass a stable topic")
			}
		}
	}
	for _, name := range []string{"report_execution_event", "list_execution_topics", "bind_execution_topic", "start_web_session", "delegate_work", "reply_work"} {
		if !found[name] {
			t.Fatal("missing manager capability", name)
		}
	}
}
