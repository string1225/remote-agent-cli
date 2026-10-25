package main

import (
	"bytes"
	"github.com/string1225/remote-agent-cli/internal/agent"
	"strings"
	"testing"
)

func TestRepairMismatchRequiresExplicitSecondConfirmation(t *testing.T) {
	for _, language := range []string{"en", "zh", "fr"} {
		t.Run(language, func(t *testing.T) {
			t.Setenv("REMOTE_AGENT_LANGUAGE", language)
			backup := "backed up"
			if language == "zh" {
				backup = "备份"
			}
			for _, in := range []string{"\n", "yes\n", "n\n", "BIND\n", "绑定\n"} {
				var out bytes.Buffer
				ok := confirmRepair(strings.NewReader(in), &out, agent.Config{ID: "source", Name: "Actual"}, agent.RepairTarget{ID: "target", Name: "Target", Hostname: "old-pc"}, "actual-pc")
				want := in == "BIND\n" || in == "绑定\n"
				if ok != want || !strings.Contains(out.String(), "source") || !strings.Contains(out.String(), "target") || !strings.Contains(out.String(), backup) {
					t.Fatal(in, ok, out.String())
				}
			}
		})
	}
}
