package agent

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestOldIMSessionsStayOnFirstPageAboveRecentWebSessions(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	err := h.transact(true, func(tx *bolt.Tx) error {
		for i := 0; i < 65; i++ {
			s := protocol.Conversation{ID: fmt.Sprintf("session-%03d", i), ServiceID: "project", Title: "Session", Kind: "web", UpdatedAt: int64(100 + i)}
			if i == 0 {
				s.Kind = "im"
				s.UpdatedAt = 1
			}
			if err := saveSession(tx, s); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := h.list("project", "", 0)
	if err != nil || page.Sessions[0].Kind != "im" || page.NextCursor == 0 {
		t.Fatalf("old IM session disappeared behind pagination: %v", err)
	}
}
