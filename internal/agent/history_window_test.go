package agent

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestHistoryTurnWindows(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	s, err := h.create(protocol.Service{ID: "project", Provider: "codex"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var expected []protocol.ChatEntry
	for turn := range 7 {
		id := fmt.Sprintf("turn-%d", turn)
		for _, entry := range []protocol.ChatEntry{
			{TurnID: id, Role: "user", Kind: "message", Text: strings.Repeat("问题🙂", 900)},
			{TurnID: id, Role: "system", Kind: "tool", Text: "$ pwd"},
			{TurnID: id, Role: "assistant", Kind: "stdout", Text: strings.Repeat("回答🙂", 4000)},
			{TurnID: id, Role: "system", Kind: "done"},
		} {
			rows, err := h.append(s.ID, entry)
			if err != nil {
				t.Fatal(err)
			}
			expected = append(expected, rows...)
		}
	}
	var before uint64
	var all []protocol.ChatEntry
	for window, wantTurns := range []int{2, 2, 2, 1} {
		var cursor, end, older uint64
		var rows []protocol.ChatEntry
		pages := 0
		for {
			page, err := h.readWindow(s.ID, cursor, before, 2)
			if err != nil {
				t.Fatal(err)
			}
			if page.OlderCursor == nil || page.Session.NativeID != "" {
				t.Fatal("missing window metadata or leaked native ID")
			}
			if cursor == 0 {
				older, end = *page.OlderCursor, page.HistoryEnd
			} else if end != page.HistoryEnd {
				t.Fatal("window moved during pagination")
			}
			rows = append(rows, page.Entries...)
			pages++
			cursor, before = page.NextCursor, end
			if cursor == 0 {
				break
			}
		}
		users := map[string]bool{}
		for i, entry := range rows {
			if i > 0 && rows[i-1].Seq >= entry.Seq {
				t.Fatal("window entries out of order")
			}
			if entry.Role == "user" {
				users[entry.TurnID] = true
			}
		}
		if len(users) != wantTurns || pages < 2 {
			t.Fatalf("window %d: got %d turns, %d pages", window, len(users), pages)
		}
		all = append(rows, all...)
		before = older
		if (older == 0) != (window == 3) {
			t.Fatalf("incorrect older cursor in window %d: %d", window, older)
		}
	}
	if len(all) != len(expected) {
		t.Fatalf("lost or duplicated chunks: got %d, want %d", len(all), len(expected))
	}
	for i, entry := range all {
		if entry.Seq != expected[i].Seq || entry.Text != expected[i].Text {
			t.Fatalf("entry %d changed across windows", i)
		}
	}
}

func TestHistoryTurnWindowSnapshotAndSkippedOldHistory(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	s, err := h.create(protocol.Service{ID: "project"}, "")
	if err != nil {
		t.Fatal(err)
	}
	page, err := h.readWindow(s.ID, 0, 0, 2)
	if err != nil || len(page.Entries) != 0 || page.OlderCursor == nil || *page.OlderCursor != 0 {
		t.Fatalf("empty window: %+v, %v", page, err)
	}
	for turn := range 4 {
		for _, role := range []string{"user", "assistant"} {
			_, err := h.append(s.ID, protocol.ChatEntry{TurnID: fmt.Sprint(turn), Role: role, Text: strings.Repeat("字", 16000)})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	// A recent read must not decode unrelated ancient history.
	err = h.transact(true, func(tx *bolt.Tx) error {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, 1)
		return tx.Bucket([]byte("entries")).Bucket([]byte(s.ID)).Put(key, []byte("old corrupt record"))
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err = h.readWindow(s.ID, 0, 0, 2)
	if err != nil || page.NextCursor == 0 {
		t.Fatalf("recent read scanned old history or was not paginated: %v", err)
	}
	_, err = h.append(s.ID, protocol.ChatEntry{TurnID: "new", Role: "user", Text: "arrived during read"})
	if err != nil {
		t.Fatal(err)
	}
	for page.NextCursor != 0 {
		page, err = h.readWindow(s.ID, page.NextCursor, page.HistoryEnd, 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range page.Entries {
			if entry.TurnID == "new" {
				t.Fatal("new turn leaked into an existing snapshot")
			}
		}
	}
	if _, err = h.readWindow("missing", 0, 0, 2); err == nil {
		t.Fatal("missing session accepted")
	}
	for _, turns := range []int{-1, 11} {
		if _, err = h.readWindow(s.ID, 0, 0, turns); err == nil {
			t.Fatalf("invalid turn count %d accepted", turns)
		}
	}
}
