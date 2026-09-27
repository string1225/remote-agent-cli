package agent

import (
	"encoding/json"
	"strings"
	"sync"
)

// Decode provider JSONL without storing credentials or protocol boilerplate.
// The output remains plain text so the web UI never executes model-produced HTML.
type transcript struct {
	mu      sync.Mutex
	pending map[string]string
	seen    map[string]string
	text    func(string, string) error
	native  func(string) error
}

func (t *transcript) write(stream, text string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pending == nil {
		t.pending = map[string]string{}
		t.seen = map[string]string{}
	}
	t.pending[stream] += text
	for {
		at := strings.IndexByte(t.pending[stream], '\n')
		if at < 0 {
			break
		}
		line := t.pending[stream][:at]
		t.pending[stream] = t.pending[stream][at+1:]
		if err := t.line(stream, line); err != nil {
			return err
		}
	}
	// Non-JSON output can be forwarded immediately; JSONL may span chunks.
	if p := t.pending[stream]; p != "" && (!strings.HasPrefix(strings.TrimSpace(p), "{") || len(p) > 128*1024) {
		t.pending[stream] = ""
		return t.text(stream, p)
	}
	return nil
}
func (t *transcript) flush() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for stream, line := range t.pending {
		if line != "" {
			if err := t.line(stream, line); err != nil {
				return err
			}
		}
		t.pending[stream] = ""
	}
	return nil
}
func (t *transcript) line(stream, line string) error {
	if strings.TrimSpace(line) == "" {
		return nil
	}
	var e struct {
		Type      string          `json:"type"`
		ThreadID  string          `json:"thread_id"`
		SessionID string          `json:"session_id"`
		Result    string          `json:"result"`
		Message   json.RawMessage `json:"message"`
		Item      struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			Text    string `json:"text"`
			Command string `json:"command"`
			Output  string `json:"aggregated_output"`
		} `json:"item"`
		Event struct {
			Type  string `json:"type"`
			Delta struct {
				Text string `json:"text"`
			} `json:"delta"`
		} `json:"event"`
	}
	if json.Unmarshal([]byte(line), &e) != nil {
		return t.text(stream, line+"\n")
	}
	if e.ThreadID != "" {
		if err := t.native(e.ThreadID); err != nil {
			return err
		}
	}
	if e.SessionID != "" {
		if err := t.native(e.SessionID); err != nil {
			return err
		}
	}
	if e.Item.Text != "" {
		if e.Item.ID != "" {
			old := t.seen[e.Item.ID]
			t.seen[e.Item.ID] = e.Item.Text
			if strings.HasPrefix(e.Item.Text, old) {
				e.Item.Text = strings.TrimPrefix(e.Item.Text, old)
			}
		}
		if e.Item.Text != "" {
			return t.text("stdout", e.Item.Text+"\n")
		}
		return nil
	}
	if e.Item.Command != "" && e.Type == "item.completed" {
		return t.text("tool", "$ "+e.Item.Command+"\n"+e.Item.Output+"\n")
	}
	if e.Event.Delta.Text != "" {
		t.seen["qoder-delta"] = "yes"
		t.seen["qoder-message"] = "yes"
		return t.text("stdout", e.Event.Delta.Text)
	}
	if e.Type == "assistant" {
		if t.seen["qoder-delta"] != "" {
			delete(t.seen, "qoder-delta")
			return nil
		}
		var msg struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if json.Unmarshal(e.Message, &msg) == nil {
			for _, c := range msg.Content {
				if c.Text != "" {
					t.seen["qoder-message"] = "yes"
					if err := t.text("stdout", c.Text+"\n"); err != nil {
						return err
					}
				}
			}
		}
	}
	if e.Type == "result" && e.Result != "" && t.seen["qoder-message"] == "" {
		return t.text("stdout", e.Result+"\n")
	}
	if e.Type == "error" || e.Type == "turn.failed" {
		return t.text("stderr", line+"\n")
	}
	if e.Type == "" {
		return t.text(stream, line+"\n")
	}
	return nil
}
