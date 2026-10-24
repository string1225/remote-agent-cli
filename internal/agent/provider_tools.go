package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// Configure the built-in capability server for this invocation only; do not
// rewrite the user's Codex/Qoder configuration or include credentials in argv.
func providerTools(s protocol.Service, args []string) ([]string, func(), error) {
	cleanup := func() {}
	if s.ToolBinary == "" || s.ToolConfig == "" || s.ToolSession == "" {
		return args, cleanup, nil
	}
	toolArgs := []string{"mcp", "--config", s.ToolConfig, "--project", s.ID, "--session", s.ToolSession, "--source", s.ToolSource, "--notify-client", s.ToolNotifyAgent, "--notify-project", s.ToolNotifyProject}
	if s.Provider == "codex" {
		command := tomlToolString(s.ToolBinary)
		quoted := make([]string, len(toolArgs))
		for i, arg := range toolArgs {
			quoted[i] = tomlToolString(arg)
		}
		argv := "[" + strings.Join(quoted, ",") + "]"
		overrides := []string{"-c", "mcp_servers.remote_agent.command=" + string(command), "-c", "mcp_servers.remote_agent.args=" + string(argv), "-c", "mcp_servers.remote_agent.enabled=true"}
		args = append(append(append([]string{}, args[:len(args)-1]...), overrides...), args[len(args)-1])
		return args, cleanup, nil
	}
	config := map[string]any{"mcpServers": map[string]any{"remote_agent": map[string]any{"command": s.ToolBinary, "args": toolArgs, "trust": true, "includeTools": []string{"list_clients_projects", "start_web_session", "notify_im_session"}}}}
	data, _ := json.Marshal(config)
	f, err := os.CreateTemp(filepath.Dir(s.ToolConfig), ".mcp-*.json")
	if err != nil {
		return nil, cleanup, err
	}
	name := f.Name()
	cleanup = func() { _ = os.Remove(name) }
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	return append(args, "--mcp-config", name), cleanup, nil
}

// Literal strings survive PowerShell 5.1 -> npm .cmd shims, which consume
// double quotes inside JSON/TOML arrays before Codex sees them.
func tomlToolString(value string) string {
	if !strings.ContainsAny(value, "'\r\n") {
		return "'" + value + "'"
	}
	if !strings.Contains(value, "'''") {
		return "'''" + value + "'''"
	}
	quoted, _ := json.Marshal(value)
	return string(quoted)
}
