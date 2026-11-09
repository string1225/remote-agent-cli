package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func activityName(value string) string {
	return soBound(strings.SplitN(strings.ReplaceAll(value, "\r", ""), "\n", 2)[0], 600)
}

// Provider events are normalized here, without keeping arguments, file contents,
// tool results, or reasoning. Older transcript consumers keep their text path.
func (t *transcript) recordActivity(data []byte) (bool, error) {
	if t.activity == nil {
		return false, nil
	}
	var frame struct {
		Type string `json:"type"`
		Item struct {
			ID       string              `json:"id"`
			Type     string              `json:"type"`
			Command  string              `json:"command"`
			Server   string              `json:"server"`
			Tool     string              `json:"tool"`
			Status   string              `json:"status"`
			ExitCode *int                `json:"exit_code"`
			Items    []protocol.PlanStep `json:"items"`
		} `json:"item"`
		Message struct {
			Content []struct {
				Type      string `json:"type"`
				ID        string `json:"id"`
				ToolUseID string `json:"tool_use_id"`
				Name      string `json:"name"`
				IsError   bool   `json:"is_error"`
			} `json:"content"`
		} `json:"message"`
	}
	if json.Unmarshal(data, &frame) != nil {
		return false, nil
	}
	if strings.HasPrefix(frame.Type, "item.") {
		item := frame.Item
		activity := protocol.RunActivity{ID: soBound(item.ID, 160), Kind: item.Type, Status: "running"}
		switch item.Type {
		case "command_execution":
			activity.Name = activityName(item.Command)
		case "mcp_tool_call":
			activity.Name = activityName(item.Server + "/" + item.Tool)
		case "file_change":
			activity.Name = "修改文件"
		case "web_search":
			activity.Name = "搜索网页"
		case "todo_list":
			activity.Name = "执行计划"
			for _, step := range item.Items[:min(len(item.Items), 20)] {
				activity.Steps = append(activity.Steps, protocol.PlanStep{Text: activityName(step.Text), Completed: step.Completed})
			}
		default:
			return false, nil
		}
		if activity.ID == "" {
			sum := sha256.Sum256([]byte(item.Type + "\x00" + activity.Name))
			activity.ID = "synthetic-" + hex.EncodeToString(sum[:8])
		}
		if frame.Type == "item.completed" {
			activity.Status = "completed"
			if item.Status == "failed" || item.Status == "error" || item.Status == "declined" || item.ExitCode != nil && *item.ExitCode != 0 {
				activity.Status = "failed"
			}
		}
		return true, t.emitActivity(activity)
	}
	// Qoder's print stream uses tool_use/tool_result content blocks. A mixed
	// assistant message must still pass through the normal text decoder.
	for _, block := range frame.Message.Content {
		if block.Type == "tool_use" && block.ID != "" {
			if err := t.emitActivity(protocol.RunActivity{ID: soBound(block.ID, 160), Kind: "tool", Name: activityName(block.Name), Status: "running"}); err != nil {
				return false, err
			}
		}
		if block.Type == "tool_result" && block.ToolUseID != "" {
			status := "completed"
			if block.IsError {
				status = "failed"
			}
			if err := t.emitActivity(protocol.RunActivity{ID: soBound(block.ToolUseID, 160), Kind: "tool", Status: status}); err != nil {
				return false, err
			}
		}
	}
	return false, nil
}

func (t *transcript) emitActivity(activity protocol.RunActivity) error {
	if t.activities == nil {
		t.activities = make(map[string]protocol.RunActivity)
	}
	previous, exists := t.activities[activity.ID]
	if !exists && len(t.activities) >= 512 {
		return nil
	}
	if activity.Name == "" {
		activity.Name = previous.Name
	}
	if activity.Name == "" {
		activity.Name = "工具调用"
	}
	activity.StartedAt = previous.StartedAt
	if !exists && activity.Status == "running" {
		activity.StartedAt = time.Now().UnixMilli()
	}
	if previous.CompletedAt != 0 && activity.Status == "running" {
		return nil // Late updates cannot revive a finished invocation.
	}
	if exists && activity.Status == previous.Status && activity.Name == previous.Name {
		before, _ := json.Marshal(previous.Steps)
		after, _ := json.Marshal(activity.Steps)
		if string(before) == string(after) {
			return nil
		}
	}
	if activity.Status != "running" {
		activity.CompletedAt = time.Now().UnixMilli()
	}
	if err := t.activity(activity); err != nil {
		return err
	}
	t.activities[activity.ID] = activity
	return nil
}

func (t *transcript) closeActivities() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, activity := range t.activities {
		if activity.Status == "running" {
			activity.Status = "stopped"
			if err := t.emitActivity(activity); err != nil {
				return err
			}
		}
	}
	return nil
}
