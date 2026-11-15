package agent

import (
	"encoding/binary"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestSessionLastMessageSurvivesMetadataChangesAndReopen(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	s, err := h.create(protocol.Service{ID: "project", Provider: "codex"}, "Old messages")
	if err != nil {
		t.Fatal(err)
	}
	last := time.Now().Add(-8 * 24 * time.Hour).UnixMilli()
	// A legacy history has no message timestamp in its session metadata.
	err = h.transact(true, func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("entries")).Bucket([]byte(s.ID))
		for i, at := range []int64{last - 1000, last} {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(i+1))
			entry, _ := json.Marshal(protocol.ChatEntry{Seq: uint64(i + 1), CreatedAt: at, Role: "assistant", Text: "Synthetic message"})
			if err := bucket.Put(key, entry); err != nil {
				return err
			}
		}
		return bucket.SetSequence(2)
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.update(s.ID, func(s *protocol.Conversation) {
		s.Title, s.CustomTitle, s.UpdatedAt = "Renamed today", true, time.Now().UnixMilli()
	})
	if err != nil {
		t.Fatal(err)
	}
	h = historyStore{h.path}
	got, err := h.get(s.ID)
	if err != nil || got.LastMessageAt != last || got.UpdatedAt <= last {
		t.Fatalf("metadata timestamp replaced message timestamp: %+v, %v", got, err)
	}
	page, err := h.list("project", "", 0)
	if err != nil || len(page.Sessions) != 1 || page.Sessions[0].LastMessageAt != last {
		t.Fatalf("list: %+v, %v", page, err)
	}
	read, err := h.read(s.ID, 0)
	if err != nil || read.Session.LastMessageAt != last {
		t.Fatalf("history: %+v, %v", read.Session, err)
	}
	if _, err := h.append(s.ID, protocol.ChatEntry{Role: "user", Kind: "message", Text: "Resume same session"}); err != nil {
		t.Fatal(err)
	}
	got, err = h.get(s.ID)
	if err != nil || got.LastMessageAt <= last {
		t.Fatalf("new message did not refresh recency: %+v, %v", got, err)
	}
	if err := h.update(s.ID, func(s *protocol.Conversation) { s.ArchivedAt = 1 }); err != nil {
		t.Fatal(err)
	}
	page, err = h.listArchived("project", "", 0, true)
	if err != nil || len(page.Sessions) != 1 || page.Sessions[0].LastMessageAt != got.LastMessageAt {
		t.Fatalf("archive metadata: %+v, %v", page, err)
	}
}

func TestEmptySessionHasNoLastMessage(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	s, err := h.create(protocol.Service{ID: "project"}, "Empty")
	if err != nil {
		t.Fatal(err)
	}
	page, err := h.list("project", "", 0)
	if err != nil || len(page.Sessions) != 1 || page.Sessions[0].LastMessageAt != 0 || page.Sessions[0].ID != s.ID {
		t.Fatalf("empty session: %+v, %v", page, err)
	}
}
