package store

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestCodexTokensHaveHashedVerifiersAndTenantIsolation(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"alice", "bob"} {
		if err := db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	info, secret, err := db.NewCodexToken("alice", "Desktop", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(secret, "ra_codex_") || len(secret) != 52 {
		t.Fatalf("unexpected token format %q", secret)
	}
	err = db.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(_ []byte, b *bolt.Bucket) error {
			return b.ForEach(func(k, v []byte) error {
				if bytes.Contains(k, []byte(secret)) || bytes.Contains(v, []byte(secret)) {
					t.Fatal("plaintext token persisted")
				}
				return nil
			})
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := db.CodexTokens("bob")
	if err != nil || len(list) != 0 {
		t.Fatal("cross-user token visible", list, err)
	}
	if err := db.RevokeCodexToken("bob", info.ID, false); err == nil {
		t.Fatal("cross-user token revoked")
	}
	if _, err := db.AuthenticateCodexToken(secret); err != nil {
		t.Fatal(err)
	}
	if err := db.RevokeCodexToken("alice", info.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateCodexToken(secret); err == nil {
		t.Fatal("revoked token accepted")
	}
}

func TestAgentCodexSetupFailurePreservesPriorTokenAndRevokeDisablesAll(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"alice", "bob"} {
		if err := db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	a, invite, _ := db.Enroll("alice", "first")
	if _, _, err := db.Redeem(invite, "linux", "amd64"); err != nil {
		t.Fatal(err)
	}
	_, old, err := db.NewCodexToken("alice", "device", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, fresh, err := db.NewCodexToken("alice", "device", a.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Issuing a replacement precedes installation and the atomic TOML write.
	// A failed setup must leave the credential in the old TOML usable.
	if _, err := db.AuthenticateCodexToken(old); err != nil {
		t.Fatal("issuing a setup token broke the existing configuration", err)
	}
	_, personal, _ := db.NewCodexToken("alice", "personal", "")
	_, bob, _ := db.NewCodexToken("bob", "bob", "")
	if err := db.Revoke("alice", a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateCodexToken(fresh); err == nil {
		t.Fatal("revoked device token accepted")
	}
	if _, err := db.AuthenticateCodexToken(old); err == nil {
		t.Fatal("revoked device's earlier token accepted")
	}
	if _, err := db.AuthenticateCodexToken(personal); err != nil {
		t.Fatal("device revoke affected personal token", err)
	}
	yes := true
	if _, err := db.UpdateUser("", "alice", nil, &yes, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateCodexToken(personal); err == nil {
		t.Fatal("disabled account token accepted")
	}
	if _, err := db.AuthenticateCodexToken(bob); err != nil {
		t.Fatal("disable leaked across accounts", err)
	}
}
