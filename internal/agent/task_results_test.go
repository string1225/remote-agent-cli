package agent

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestCompletionSelectsFinalAssistantMessage(t *testing.T) {
	for _, tc := range []struct{ name, output, want string }{
		{"codex", `{"type":"item.completed","item":{"type":"reasoning","text":"internal reasoning"}}
{"type":"item.completed","item":{"type":"agent_message","phase":"commentary","text":"working"}}
{"type":"item.updated","item":{"id":"final","type":"agent_message","text":"partial"}}
{"type":"item.completed","item":{"id":"final","type":"agent_message","phase":"final_answer","text":"结论：测试通过，提交 abc123"}}
{"type":"item.completed","item":{"type":"command_execution","command":"go test","aggregated_output":"tool output"}}
`, "结论：测试通过，提交 abc123"},
		{"qoder", `{"type":"stream_event","event":{"delta":{"text":"streamed"}}}
{"type":"assistant","message":{"content":[{"text":"assistant candidate"}]}}
{"type":"result","result":"最终结论：检查通过"}
`, "最终结论：检查通过"},
		{"commentary-only", `{"type":"item.completed","item":{"type":"agent_message","phase":"commentary","text":"仍在处理"}}
`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
			s, err := a.historyStore().create(protocol.Service{ID: "project"}, "work")
			if err != nil {
				t.Fatal(err)
			}
			decoder := &transcript{native: func(string) error { return nil }, final: func(text string) error { return a.historyStore().saveTaskResult("task", text) }, text: func(stream, text string) error {
				role := "system"
				if stream == "stdout" {
					role = "assistant"
				}
				_, err := a.historyStore().append(s.ID, protocol.ChatEntry{TurnID: "task", Role: role, Kind: stream, Text: text})
				return err
			}}
			// Exercise JSONL split across arbitrary transport chunks.
			mid := len(tc.output) / 2
			for _, chunk := range []string{tc.output[:mid], tc.output[mid:]} {
				if err := decoder.write("stdout", chunk); err != nil {
					t.Fatal(err)
				}
			}
			if err := decoder.flush(); err != nil {
				t.Fatal(err)
			}
			text, _, err := a.historyStore().taskResultText(s.ID, "task")
			if err != nil {
				t.Fatal(err)
			}
			if text != tc.want || strings.Contains(text, "internal reasoning") || strings.Contains(text, "tool output") || strings.Contains(text, "working") || strings.Contains(text, "streamed") {
				t.Fatalf("incorrect final result: %s", text)
			}
		})
	}
}
