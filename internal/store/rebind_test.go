package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestRebindRequiresOwnerAndConsumesOnceWithoutLosingData(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"alice", "bob"} {
		if err = db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	a, invite, _ := db.Enroll("alice", "Desktop")
	_, old, _ := db.Redeem(invite, "windows", "amd64")
	if err = db.Heartbeat(a.ID, protocol.Signal{Services: []protocol.Service{{ID: "project", Name: "Retained", Provider: "codex"}}}); err != nil {
		t.Fatal(err)
	}
	info, personal, err := db.EnsureDeviceCodexToken("alice", "Desktop", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = db.RebindEnrollment("bob", a.ID); err == nil {
		t.Fatal("non-owner issued recovery")
	}
	_, recovery, err := db.RebindEnrollment("alice", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.AuthenticateAgent(a.ID, old); err != nil {
		t.Fatal("issuing recovery invalidated binding")
	}
	if _, _, err = db.Redeem(recovery, "windows", "amd64"); err == nil {
		t.Fatal("normal enrollment consumed recovery")
	}
	next := Token()
	if _, err = db.RedeemRebind(recovery, "wrong", "windows", "amd64", next); err == nil {
		t.Fatal("wrong target accepted")
	}
	restored, err := db.RedeemRebind(recovery, a.ID, "windows", "amd64", next)
	if err != nil || restored.ID != a.ID || len(restored.Services) != 1 {
		t.Fatal("rebind lost identity/project", err)
	}
	if _, err = db.AuthenticateAgent(a.ID, old); err == nil {
		t.Fatal("old credential stayed valid")
	}
	if _, err = db.AuthenticateAgent(a.ID, next); err != nil {
		t.Fatal(err)
	}
	if _, err = db.RedeemRebind(recovery, a.ID, "windows", "amd64", next); err == nil {
		t.Fatal("recovery reused")
	}
	if got, err := db.AuthenticateCodexToken(personal); err != nil || got.ID != info.ID {
		t.Fatal("model token changed", err)
	}
	_, expired, _ := db.RebindEnrollment("alice", a.ID)
	if err = db.db.Update(func(tx *bolt.Tx) error {
		return put(tx, "enrollments", Hash(expired), enrollment{AgentID: a.ID, ExpiresAt: time.Now().Add(-time.Minute).Unix(), Recovery: true})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.RedeemRebind(expired, a.ID, "windows", "amd64", Token()); err == nil {
		t.Fatal("expired recovery accepted")
	}
	if _, err = db.AuthenticateAgent(a.ID, next); err != nil {
		t.Fatal("expired redemption changed active credential", err)
	}
}
