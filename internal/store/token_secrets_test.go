package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestSavedTokenOwnerReopenAndLegacyImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"alice", "bob"} {
		if err := db.CreateUser(User{ID: id, Username: id}); err != nil {
			t.Fatal(err)
		}
	}
	info, secret, err := db.NewCodexToken("alice", "desktop", "")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := db.RevealCodexToken("alice", info.ID); err != nil || got != secret {
		t.Fatal("owner reveal", err)
	}
	if _, err := db.RevealCodexToken("bob", info.ID); !errors.Is(err, ErrDenied) {
		t.Fatal("cross-account reveal", err)
	}
	// Model an old database entry that has only its verifier.
	if err := db.db.Update(func(tx *bolt.Tx) error { return tx.Bucket([]byte("codex_secrets")).Delete([]byte(info.ID)) }); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RevealCodexToken("alice", info.ID); !errors.Is(err, ErrTokenNotSaved) {
		t.Fatal("legacy reveal", err)
	}
	if err := db.SaveCodexToken("bob", info.ID, secret); !errors.Is(err, ErrDenied) {
		t.Fatal("cross-account import", err)
	}
	if err := db.SaveCodexToken("alice", info.ID, "wrong"); !errors.Is(err, ErrDenied) {
		t.Fatal("wrong token import", err)
	}
	if err := db.SaveCodexToken("alice", info.ID, secret); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AuthenticateCodexToken(secret); err != nil {
		t.Fatal("import changed authentication", err)
	}
	db.Close()
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := db.RevealCodexToken("alice", info.ID); err != nil || got != secret {
		t.Fatal("reopen", err)
	}
	if err := db.RevokeCodexToken("alice", info.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RevealCodexToken("alice", info.ID); !errors.Is(err, ErrDenied) {
		t.Fatal("revoked reveal", err)
	}
	db.Close()
	if err := os.Remove(path + ".token-key"); err != nil {
		t.Fatal(err)
	}
	if restored, err := Open(path); err == nil {
		restored.Close()
		t.Fatal("missing key silently replaced")
	}
}
