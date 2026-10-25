package store

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestDeviceTokenIsIdempotentAndCannotCrossAccounts(t *testing.T) {
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
	a, invite, err := db.Enroll("alice", "Desktop")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = db.Redeem(invite, "windows", "amd64"); err != nil {
		t.Fatal(err)
	}
	first, secret, err := db.EnsureDeviceCodexToken("alice", "Desktop", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, again, err := db.EnsureDeviceCodexToken("alice", "Desktop", a.ID)
	if err != nil || first.ID != second.ID || secret != again {
		t.Fatal("device token was rotated", err)
	}
	var calls sync.WaitGroup
	for i := 0; i < 8; i++ {
		calls.Add(1)
		go func() {
			defer calls.Done()
			same, value, err := db.EnsureDeviceCodexToken("alice", "Desktop", a.ID)
			if err != nil || same.ID != first.ID || value != secret {
				t.Error("concurrent retry rotated token", err)
			}
		}()
	}
	calls.Wait()
	if _, _, err = db.EnsureDeviceCodexToken("bob", "Desktop", a.ID); err == nil {
		t.Fatal("cross-account device token allowed")
	}
	if err = db.RevokeCodexToken("alice", first.ID, false); err != nil {
		t.Fatal(err)
	}
	third, replacement, err := db.EnsureDeviceCodexToken("alice", "Desktop", a.ID)
	if err != nil || third.ID == first.ID || replacement == secret {
		t.Fatal("revoked token reused", err)
	}
}
