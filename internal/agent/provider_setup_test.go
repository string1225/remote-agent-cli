package agent

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestMissingCodexProjectProvidesInstallGuidanceWithoutChangingConfig(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), AllowedRoots: []string{root}}, Providers: map[string]Provider{}}
	_, err := a.registerProject(protocol.Service{Name: "Project", Provider: "codex", Workspace: root})
	if err == nil || !strings.Contains(err.Error(), "remote-agent codex setup") || !strings.Contains(err.Error(), "目标客户端") {
		t.Fatalf("missing actionable guidance: %v", err)
	}
	if len(a.Config.Services) != 0 {
		t.Fatal("missing provider created a project")
	}
}

func TestUnavailableRunnerProvidesProviderSpecificGuidance(t *testing.T) {
	for _, test := range []struct{ provider, guidance string }{{"codex", "remote-agent codex setup"}, {"qoder", "Qoder CLI"}} {
		err := runProcess(context.Background(), Provider{}, protocol.Service{Provider: test.provider}, "fixture", true, nil, "fixture")
		if err == nil || !strings.Contains(err.Error(), test.guidance) || !strings.Contains(err.Error(), "目标客户端") {
			t.Fatalf("%s: %v", test.provider, err)
		}
	}
}
