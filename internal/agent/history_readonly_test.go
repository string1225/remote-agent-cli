package agent

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestHistoryReadsPreserveDatabaseBytes(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json")}}
	h := a.historyStore()
	session, err := h.create(protocol.Service{ID: "project", Provider: "codex", Workspace: root}, "保留会话")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.nativeID(session.ID, "native-session"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.append(session.ID, protocol.ChatEntry{Role: "user", Text: "保留原消息"}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(h.path)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := h.get(session.ID)
	if err != nil || stored.NativeID != "native-session" {
		t.Fatal("stored session changed", stored, err)
	}
	if _, err := h.list("project", "", 0); err != nil {
		t.Fatal(err)
	}
	page, err := h.read(session.ID, 0)
	if err != nil || page.Session.ID != session.ID || len(page.Entries) != 1 || page.Entries[0].Text != "保留原消息" {
		t.Fatal("history changed", page, err)
	}
	if _, err := h.coordinationRows("coordination_pending"); err != nil {
		t.Fatal(err)
	}
	// Exercise the same idle recovery/query paths as an installed service,
	// without starting a real service or a network connection.
	a.recoverTasks()
	a.restoreExecutionSubscriptions(context.Background())
	a.flushCoordination(context.Background())
	after, err := os.ReadFile(h.path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("read-only history or idle recovery rewrote the database", err)
	}
}

func TestHistoryReadInitializesMissingBuckets(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "new", true: "legacy"}[existing], func(t *testing.T) {
			h := historyStore{filepath.Join(t.TempDir(), "history.db")}
			if existing {
				db, err := bolt.Open(h.path, 0600, nil)
				if err != nil {
					t.Fatal(err)
				}
				if err := db.Update(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucket([]byte("legacy-data"))
					if err != nil {
						return err
					}
					return b.Put([]byte("keep"), []byte("original"))
				}); err != nil {
					t.Fatal(err)
				}
				if err := db.Close(); err != nil {
					t.Fatal(err)
				}
			}
			page, err := h.list("project", "", 0)
			if err != nil || len(page.Sessions) != 0 {
				t.Fatal("cannot read a new or legacy history", err)
			}
			if existing {
				if err := h.transact(false, func(tx *bolt.Tx) error {
					if !bytes.Equal(tx.Bucket([]byte("legacy-data")).Get([]byte("keep")), []byte("original")) {
						t.Error("schema initialization changed existing data")
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
