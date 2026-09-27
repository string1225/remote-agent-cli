package store

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestEnrollmentAtomicAndRevocation(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a, token, err := db.Enroll("alice", "workstation")
	if err != nil {
		t.Fatal(err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			if _, _, err := db.Redeem(token, "windows", "amd64"); err == nil {
				successes.Add(1)
			}
		})
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("expected exactly one redemption, got %d", successes.Load())
	}
	if err := db.Revoke("bob", a.ID); err == nil {
		t.Fatal("cross-account revocation accepted")
	}
	if err := db.Revoke("alice", a.ID); err != nil {
		t.Fatal(err)
	}
	after, _ := db.Agent(a.ID)
	if !after.Revoked || after.TokenHash != "" {
		t.Fatal("credential not removed")
	}
	_, pending, err := db.Enroll("alice", "pending")
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := db.Agents("alice")
	if len(rows) != 1 {
		t.Fatal(rows)
	}
	if err := db.Revoke("alice", rows[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.Redeem(pending, "darwin", "arm64"); err == nil {
		t.Fatal("revoked invitation redeemed")
	}
}

func TestExpiredCredentialsDenied(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	a, invite, err := db.Enroll("alice", "old invitation")
	if err != nil {
		t.Fatal(err)
	}
	session, err := db.NewSession("alice")
	if err != nil {
		t.Fatal(err)
	}
	err = db.db.Update(func(tx *bolt.Tx) error {
		past := time.Now().Add(-time.Minute).Unix()
		if err := put(tx, "enrollments", Hash(invite), enrollment{a.ID, past}); err != nil {
			return err
		}
		return put(tx, "sessions", Hash(session), Session{"alice", past})
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.Redeem(invite, "darwin", "arm64"); err == nil {
		t.Fatal("expired invite accepted")
	}
	if _, err := db.Session(session); err == nil {
		t.Fatal("expired session accepted")
	}
	if err := db.Prune(); err != nil {
		t.Fatal(err)
	}
}
func TestSessionsPersistAndLogout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	token, err := db.NewSession("alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s, err := db.Session(token)
	if err != nil || s.UserID != "alice" {
		t.Fatal("session did not survive restart")
	}
	if err = db.Logout(token); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Session(token); err == nil {
		t.Fatal("logged out token accepted")
	}
}
