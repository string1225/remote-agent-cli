package agent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestRunActivitiesKeepLifecycleWithoutToolBodies(t *testing.T) {
	var activities []protocol.RunActivity
	var text strings.Builder
	decoder := transcript{native: func(string) error { return nil }, text: func(_, value string) error { text.WriteString(value); return nil }, activity: func(activity protocol.RunActivity) error { activities = append(activities, activity); return nil }}
	frames := []string{
		`{"type":"item.started","item":{"id":"cmd","type":"command_execution","command":"go test ./...\nPRIVATE SECOND LINE","status":"in_progress"}}`,
		`{"type":"item.updated","item":{"id":"cmd","type":"command_execution","command":"go test ./...\nPRIVATE SECOND LINE","aggregated_output":"PRIVATE FILE BODY"}}`,
		`{"type":"item.completed","item":{"id":"cmd","type":"command_execution","command":"go test ./...\nPRIVATE SECOND LINE","status":"completed","exit_code":1,"aggregated_output":"PRIVATE FILE BODY"}}`,
		`{"type":"item.started","item":{"id":"mcp","type":"mcp_tool_call","server":"remote-agent","tool":"read_session","arguments":{"token":"PRIVATE TOKEN"}}}`,
		`{"type":"item.updated","item":{"id":"plan","type":"todo_list","items":[{"text":"检查目录","completed":true},{"text":"完成修改","completed":false}]}}`,
	}
	for _, frame := range frames {
		if err := decoder.write("stdout", frame+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	if err := decoder.closeActivities(); err != nil {
		t.Fatal(err)
	}
	if len(activities) != 6 {
		t.Fatalf("unexpected events: %+v", activities)
	}
	if activities[0].Status != "running" || activities[1].Status != "failed" || activities[0].StartedAt != activities[1].StartedAt || activities[1].CompletedAt == 0 {
		t.Fatal(activities[:2])
	}
	if activities[3].Kind != "todo_list" || len(activities[3].Steps) != 2 || !activities[3].Steps[0].Completed {
		t.Fatal(activities[3])
	}
	encoded, _ := json.Marshal(activities)
	if strings.Contains(string(encoded)+text.String(), "PRIVATE") {
		t.Fatalf("tool body escaped normalization: %s %s", encoded, text.String())
	}
	for _, activity := range activities[4:] {
		if activity.Status != "stopped" {
			t.Fatal(activity)
		}
	}
}

func TestQoderMixedMessagesRetainTextAndNormalizeToolResult(t *testing.T) {
	var activities []protocol.RunActivity
	var text strings.Builder
	decoder := transcript{native: func(string) error { return nil }, text: func(_, value string) error { text.WriteString(value); return nil }, activity: func(activity protocol.RunActivity) error { activities = append(activities, activity); return nil }}
	for _, frame := range []string{
		`{"type":"assistant","message":{"content":[{"type":"text","text":"我会检查"},{"type":"tool_use","id":"call","name":"Read","input":{"path":"PRIVATE PATH"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"call","is_error":false,"content":"PRIVATE BODY"}]}}`,
	} {
		if err := decoder.write("stdout", frame+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	if len(activities) != 2 || activities[1].Name != "Read" || activities[1].Status != "completed" || !strings.Contains(text.String(), "我会检查") {
		t.Fatal(activities, text.String())
	}
	data, _ := json.Marshal(activities)
	if strings.Contains(string(data)+text.String(), "PRIVATE") {
		t.Fatal("raw tool content was retained")
	}
}
