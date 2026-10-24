//go:build windows

package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestMCPArgumentHelper(t *testing.T) {
	if os.Getenv("RA_MCP_ARGUMENT_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			_ = json.NewEncoder(os.Stdout).Encode(os.Args[i+1:])
			os.Exit(0)
		}
	}
	os.Exit(2)
}
func TestCodexMCPOverridesSurvivePowerShellAndNpmCommandShims(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	script := filepath.Join(root, "codex.cmd")
	if err = os.WriteFile(script, []byte("@echo off\r\n\""+exe+"\" -test.run=TestMCPArgumentHelper -- %*\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RA_MCP_ARGUMENT_HELPER", "1")
	s := protocol.Service{ID: "project", Provider: "codex", ToolBinary: filepath.Join(root, "remote-agent.exe"), ToolConfig: filepath.Join(root, "中文 ' profile", "config.json"), ToolSession: "session", ToolSource: "web"}
	args, cleanup, err := providerTools(s, providerSessionArgs("codex", false, ""))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := command(ctx, script, args).Output()
	if err != nil {
		t.Fatal(err)
	}
	var received []string
	if json.Unmarshal(output, &received) != nil {
		t.Fatal("shim did not preserve argument framing")
	}
	var doc strings.Builder
	for i, arg := range received {
		if arg == "-c" && strings.HasPrefix(received[i+1], "mcp_servers.") {
			doc.WriteString(received[i+1])
			doc.WriteByte('\n')
		}
	}
	var config struct {
		MCP map[string]struct {
			Command string   `toml:"command"`
			Args    []string `toml:"args"`
		} `toml:"mcp_servers"`
	}
	if err := toml.Unmarshal([]byte(doc.String()), &config); err != nil {
		t.Fatal("native Codex cannot parse MCP override after npm shim", err)
	}
	server := config.MCP["remote_agent"]
	if server.Command != s.ToolBinary || !strings.Contains(strings.Join(server.Args, "|"), s.ToolConfig) {
		t.Fatal("lost executable or configuration path through shim")
	}
}
