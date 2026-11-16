package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func TestCodexInstallServerDefaultsAndOverrides(t *testing.T) {
	root := t.TempDir()
	config := agent.Config{Server: "https://enrolled.example.test/agents/", ID: "fixture", Credential: "fixture-secret", Path: filepath.Join(root, "config.json")}
	if err := config.Save(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, server, config, want string }{
		{"default", "", filepath.Join(root, "missing.json"), "https://codex.sunny-string.cn/agents"},
		{"enrolled", "", config.Path, "https://enrolled.example.test/agents"},
		{"override", "https://override.example.test/site/", config.Path, "https://override.example.test/site"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := codexInstallServer(test.server, test.config); got != test.want {
				t.Fatalf("server = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCodexInstallProgressDisplaysStepsAndBytes(t *testing.T) {
	for _, language := range []string{"en", "zh"} {
		t.Run(language, func(t *testing.T) {
			t.Setenv("REMOTE_AGENT_LANGUAGE", language)
			var output bytes.Buffer
			reporter := &codexInstallProgress{out: &output}
			reporter.update(codexsetup.Progress{Stage: "release"})
			reporter.tick()
			reporter.update(codexsetup.Progress{Stage: "download", Downloaded: 1 << 20, Total: 4 << 20})
			reporter.tick()
			reporter.update(codexsetup.Progress{Stage: "download", Downloaded: 4 << 20, Total: 4 << 20})
			for _, stage := range []string{"verify", "extract", "publish", "complete"} {
				reporter.update(codexsetup.Progress{Stage: stage})
			}
			for _, want := range []string{"[1/6]", "[2/6]", "[3/6]", "[4/6]", "[5/6]", "[6/6]", "1.0 / 4.0 MiB", "25%", "100%"} {
				if !strings.Contains(output.String(), want) {
					t.Fatalf("missing %q: %s", want, output.String())
				}
			}
			if language == "zh" && (!strings.Contains(output.String(), "正在下载") || !strings.Contains(output.String(), "已用时")) {
				t.Fatal("progress was not localized", output.String())
			}
			before := output.String()
			reporter.tick()
			if output.String() != before {
				t.Fatal("completed installer still reports work")
			}
		})
	}
}

func TestCodexInstallProgressHandlesUnknownDownloadSize(t *testing.T) {
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	var output bytes.Buffer
	reporter := &codexInstallProgress{out: &output}
	reporter.update(codexsetup.Progress{Stage: "download", Downloaded: 2 << 20, Total: -1})
	reporter.tick()
	if !strings.Contains(output.String(), "2.0 MiB (total size unknown") || strings.Contains(output.String(), "%") {
		t.Fatal("unknown content length reported a percentage", output.String())
	}
}

func TestCodexInstallReportsActivityWhileWaitingForServer(t *testing.T) {
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	var output bytes.Buffer
	_, err := installCodex(ctx, server.URL, t.TempDir(), "windows", "amd64", server.Client(), &output)
	if err == nil || !strings.Contains(output.String(), "elapsed") || strings.Contains(output.String(), "[6/6]") {
		t.Fatalf("missing waiting/failure progress: %v, %s", err, output.String())
	}
}
