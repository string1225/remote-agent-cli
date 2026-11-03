package agent

import (
	"strings"
	"testing"
)

func TestOversizedToolEventChunksNeverLeakIntoUserTranscript(t *testing.T) {
	var output strings.Builder
	decoder := transcript{text: func(_, text string) error { output.WriteString(text); return nil }, native: func(string) error { return nil }}
	event := `{"type":"item.completed","item":{"type":"mcp_tool_call","result":"` + strings.Repeat(`C:\workspace\internal\secret.txt `, 9000) + `"}}` + "\n"
	for offset := 0; offset < len(event); offset += 8192 {
		end := offset + 8192
		if end > len(event) {
			end = len(event)
		}
		if e := decoder.write("stdout", event[offset:end]); e != nil {
			t.Fatal(e)
		}
	}
	if e := decoder.write("stdout", "{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"verified result\"}}\n"); e != nil {
		t.Fatal(e)
	}
	_ = decoder.flush()
	if strings.Contains(output.String(), "workspace") || !strings.Contains(output.String(), "verified result") || output.Len() > 200 {
		t.Fatal("raw tool payload leaked or next result lost", output.Len())
	}
}
