package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestProviderToolsAreInvocationScopedAndPreserveLocalPermissions(t *testing.T) {
	root := t.TempDir()
	service := protocol.Service{ID: "project", Provider: "codex", ToolBinary: filepath.Join(root, "remote-agent.exe"), ToolConfig: filepath.Join(root, "配置'目录", "config.json"), ToolSession: "web-session", ToolSource: "web"}
	service.ToolNotifySession = "origin-im"
	if err := os.MkdirAll(filepath.Dir(service.ToolConfig), 0700); err != nil {
		t.Fatal(err)
	}
	args, cleanup, err := providerTools(service, providerSessionArgs("codex", false, "thread"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if args[len(args)-1] != "-" || !strings.Contains(strings.Join(args, " "), "sandbox_mode=\"danger-full-access\"") {
		t.Fatal("changed prompt stdin or sandbox policy")
	}
	values := map[string]any{}
	var document strings.Builder
	for i, arg := range args {
		if arg == "-c" {
			document.WriteString(args[i+1])
			document.WriteByte('\n')
		}
	}
	if err := toml.Unmarshal([]byte(document.String()), &values); err != nil {
		t.Fatal("native CLI cannot parse overrides", err)
	}
	servers := values["mcp_servers"].(map[string]any)
	entry := servers["remote_agent"].(map[string]any)
	if entry["command"] != service.ToolBinary {
		t.Fatal("wrong MCP executable")
	}
	if !strings.Contains(document.String(), "--notify-session") || !strings.Contains(document.String(), "origin-im") {
		t.Fatal("provider lost the original IM reply context")
	}
	service.Provider = "qoder"
	args, cleanup, err = providerTools(service, providerSessionArgs("qoder", false, "thread"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(args, " "), "--permission-mode bypass_permissions") {
		t.Fatal("changed Qoder local permissions")
	}
	path := args[len(args)-1]
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if json.Unmarshal(data, &config) != nil {
		t.Fatal("invalid Qoder MCP config")
	}
	if !strings.Contains(string(data), "web-session") {
		t.Fatal("lost caller session")
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("temporary MCP config retained")
	}
}
