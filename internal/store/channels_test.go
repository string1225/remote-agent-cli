package store

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRobotBindingUniquenessReservationsAndRecovery(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"alice", "bob"} {
		if err := db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	a, _, _ := db.Enroll("alice", "a")
	b, _, _ := db.Enroll("bob", "b")
	first, second := Hash("feishu:first"), Hash("feishu:second")
	bind := func(user, id, project, fp string, final bool) error {
		return db.BindChannel(user, ChannelBinding{id, project, fp}, final)
	}
	if err := bind("alice", a.ID, "project-a", first, true); err != nil {
		t.Fatal(err)
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if bind("bob", b.ID, "project-b", first, true) == nil {
				accepted.Add(1)
			}
		})
	}
	wg.Wait()
	if accepted.Load() != 0 {
		t.Fatal("same robot bound to another account/project")
	}
	if err := bind("alice", a.ID, "project-a", second, false); err != nil {
		t.Fatal(err)
	}
	if bind("bob", b.ID, "project-b", first, true) == nil {
		t.Fatal("released old robot before client commit")
	}
	if err := bind("alice", a.ID, "project-a", first, true); err != nil {
		t.Fatal("cannot restore previous reservation", err)
	}
	if err := bind("bob", b.ID, "project-b", second, true); err != nil {
		t.Fatal("rollback did not release unused new robot", err)
	}
	if err := db.SyncChannels("alice", a.ID, []ChannelBinding{{ServiceID: "one", Fingerprint: first}, {ServiceID: "two", Fingerprint: second}}); err == nil {
		t.Fatal("sync stole another client's robot")
	}
	if bind("bob", b.ID, "third", first, true) == nil {
		t.Fatal("failed sync damaged old binding")
	}
	if err := db.Revoke("alice", a.ID); err != nil {
		t.Fatal(err)
	}
	if err := bind("bob", b.ID, "third", first, true); err != nil {
		t.Fatal("revoke did not release robot reservation", err)
	}
}
