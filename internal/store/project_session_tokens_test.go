package store

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

func TestProjectSessionTokenLifecycleAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, id := range []string{"alice", "bob"} {
		if err := db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	a, invite, err := db.Enroll("alice", "test", "linux")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.Redeem(invite, "linux", "amd64"); err != nil {
		t.Fatal(err)
	}
	projects := []protocol.Service{{ID: "p1"}, {ID: "p2"}, {ID: "manager", Scope: "project"}}
	if err := db.Heartbeat(a.ID, protocol.Signal{Services: projects}); err != nil {
		t.Fatal(err)
	}
	info, err := db.ProjectSessionToken("alice", a.ID, "p1")
	if err != nil || info.Configured {
		t.Fatal(info, err)
	}
	_, token, err := db.SetProjectSessionToken("alice", a.ID, "p1", "")
	if err != nil || len(token) < 32 {
		t.Fatal("generate", err)
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", token); err != nil {
		t.Fatal(err)
	}
	for _, project := range []string{"p2", "manager", "missing"} {
		if _, err := db.AuthenticateProjectSessionToken(a.ID, project, token); err == nil {
			t.Fatal("crossed project boundary", project)
		}
	}
	if _, _, err := db.SetProjectSessionToken("alice", a.ID, "manager", ""); err == nil {
		t.Fatal("accepted system project")
	}
	if _, _, err := db.SetProjectSessionToken("bob", a.ID, "p1", ""); err == nil {
		t.Fatal("foreign owner changed token")
	}
	if _, err := db.RevealProjectSessionToken("bob", a.ID, "p1"); err == nil {
		t.Fatal("foreign owner read token")
	}
	if err := db.DeleteProjectSessionToken("bob", a.ID, "p1"); err == nil {
		t.Fatal("foreign owner revoked token")
	}
	if err := db.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(_ []byte, bucket *bolt.Bucket) error {
			return bucket.ForEach(func(k, v []byte) error {
				if bytes.Contains(k, []byte(token)) || bytes.Contains(v, []byte(token)) {
					t.Fatal("plaintext token in database")
				}
				return nil
			})
		})
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if secret, err := db.RevealProjectSessionToken("alice", a.ID, "p1"); err != nil || secret != token {
		t.Fatal("restart lost token", err)
	}
	for _, invalid := range []string{"short", "has spaces and is long enough for a token", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n"} {
		if _, _, err := db.SetProjectSessionToken("alice", a.ID, "p1", invalid); err == nil {
			t.Fatal("accepted invalid token")
		}
	}
	custom := "test-custom-project-token-0123456789"
	if _, _, err := db.SetProjectSessionToken("alice", a.ID, "p1", custom); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", token); err == nil {
		t.Fatal("old token still valid")
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", custom); err != nil {
		t.Fatal(err)
	}
	if err := db.Heartbeat(a.ID, protocol.Signal{Services: projects[1:]}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", custom); err == nil {
		t.Fatal("removed project authenticated")
	}
	if err := db.Heartbeat(a.ID, protocol.Signal{Services: projects}); err != nil {
		t.Fatal(err)
	}
	if err := db.DeleteProjectSessionToken("alice", a.ID, "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", custom); err == nil {
		t.Fatal("revoked token valid")
	}
	if _, err := db.RevealProjectSessionToken("alice", a.ID, "p1"); err == nil {
		t.Fatal("revoked secret readable")
	}
	_, token, err = db.SetProjectSessionToken("alice", a.ID, "p1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Revoke("alice", a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateProjectSessionToken(a.ID, "p1", token); err == nil {
		t.Fatal("revoked device authenticated")
	}
}
