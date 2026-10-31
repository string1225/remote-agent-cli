package agent

import (
	"encoding/json"
	"unicode/utf8"

	bolt "go.etcd.io/bbolt"
)

type taskResult struct {
	Text     string
	Notified bool // Retained only to preserve records written by older clients.
	Captured bool
}

func (h historyStore) taskResult(id string) (result taskResult, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte("task_results")); b != nil {
			if data := b.Get([]byte(id)); data != nil {
				return json.Unmarshal(data, &result)
			}
		}
		return nil
	})
	return
}

func (h historyStore) updateTaskResult(id string, change func(*taskResult)) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("task_results"))
		if err != nil {
			return err
		}
		var result taskResult
		if data := b.Get([]byte(id)); data != nil {
			if err := json.Unmarshal(data, &result); err != nil {
				return err
			}
		}
		change(&result)
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}

func (h historyStore) saveTaskResult(id, text string) error {
	return h.updateTaskResult(id, func(v *taskResult) { v.Text = text; v.Captured = true })
}

// Only model-authored text is returned. Status and errors travel separately;
// this helper never invents a success message or a delivery policy.
func (h historyStore) taskResultText(sessionID, taskID string) (string, bool, error) {
	result, err := h.taskResult(taskID)
	if err != nil {
		return "", false, err
	}
	if result.Captured {
		return result.Text, true, nil
	}
	return h.turnText(sessionID, taskID, 1<<20), false, nil
}

// channels.notify accepts at most 16000 bytes. Preserve every part of a long
// result, including UTF-8 boundaries, rather than replacing it with a URL.
func notificationParts(text string) []string {
	var parts []string
	for len(text) > 14000 {
		n := 14000
		for !utf8.RuneStart(text[n]) {
			n--
		}
		parts = append(parts, text[:n])
		text = text[n:]
	}
	if text != "" {
		parts = append(parts, text)
	}
	return parts
}
