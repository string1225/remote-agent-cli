package agent

import (
	"bytes"
	"context"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"testing"
)

func TestIdleQueueRecoveryPreservesHistoryDatabaseBytes(t *testing.T) {
	for _, emptyBucket := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "empty-queue"}[emptyBucket], func(t *testing.T) {
			a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
			h := a.historyStore()
			session, err := h.create(protocol.Service{ID: "project", Provider: "codex", Workspace: t.TempDir()}, "retained")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = h.append(session.ID, protocol.ChatEntry{Role: "user", Text: "retained message"}); err != nil {
				t.Fatal(err)
			}
			if emptyBucket {
				if err = h.transact(true, func(tx *bolt.Tx) error { _, err := tx.CreateBucketIfNotExists([]byte("task_queue")); return err }); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(h.path)
			if err != nil {
				t.Fatal(err)
			}
			a.recoverTasks()
			a.initNotifications(context.Background())
			after, err := os.ReadFile(h.path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("idle recovery changed existing history database bytes")
			}
		})
	}
}
