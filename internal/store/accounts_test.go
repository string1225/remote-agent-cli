package store

import (
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestUsernameUniquenessAndValidation(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(User{ID: "one", Username: "  用户 Alice  "}); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateUser(User{ID: "two", Username: "用户 aLiCe"}); !errors.Is(err, ErrExists) {
		t.Fatalf("case-insensitive duplicate accepted: %v", err)
	}
	if err := db.CreateUser(User{ID: "one", Username: "replacement"}); !errors.Is(err, ErrExists) {
		t.Fatalf("immutable ID replaced: %v", err)
	}
	u, err := db.User("  用户 ALICE ")
	if err != nil || u.ID != "one" || u.Username != "用户 Alice" {
		t.Fatalf("normalized lookup: %#v %v", u, err)
	}
	for _, name := range []string{"", "  ", "a\nb", "\nalice", string(make([]rune, 65))} {
		if err := db.CreateUser(User{ID: Token(), Username: name}); err == nil {
			t.Fatalf("invalid name accepted: %q", name)
		}
	}
}

func TestUnicodeCaseFoldUsernameUniqueness(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, pair := range [][2]string{{"Σ", "ς"}, {"K", "K"}} {
		id := Token()
		if err := db.CreateUser(User{ID: id, Username: pair[0]}); err != nil {
			t.Fatal(err)
		}
		if err := db.CreateUser(User{ID: Token(), Username: pair[1]}); !errors.Is(err, ErrExists) {
			t.Fatalf("Unicode equivalent %q accepted separately from %q: %v", pair[1], pair[0], err)
		}
		if u, err := db.User(pair[1]); err != nil || u.ID != id {
			t.Fatalf("Unicode fold lookup failed %q: %#v %v", pair[1], u, err)
		}
	}
}

func TestUserSearchUsesUnicodeIdentityFolding(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, u := range []User{{ID: "greek", Username: "ος"}, {ID: "latin", Username: "Kelvin"}} {
		if err := db.CreateUser(u); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ query, id string }{{" ΟΣ ", "greek"}, {"Σ", "greek"}, {"σ", "greek"}, {"ς", "greek"}, {"Kel", "latin"}} {
		list, err := db.Users(tc.query)
		if err != nil || len(list) != 1 || list[0].ID != tc.id {
			t.Fatalf("search %q: got %#v, %v", tc.query, list, err)
		}
	}
	for _, query := range []string{"", "  "} {
		list, err := db.Users(query)
		if err != nil || len(list) != 2 {
			t.Fatalf("empty search must list all: %#v, %v", list, err)
		}
	}
}

func TestConcurrentPasswordResetDeniesStaleLoginSnapshot(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(User{ID: "owner", Username: "owner", Password: []byte("old hash")}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := db.User("owner")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.UpdateUser("", "owner", nil, nil, []byte("new hash")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.NewSessionForUser(snapshot); !errors.Is(err, ErrDenied) {
		t.Fatalf("stale login survived password reset: %v", err)
	}
}

func TestAccountChangeInvalidatesEverySession(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(User{ID: "owner", Username: "owner", Password: []byte("hash")}); err != nil {
		t.Fatal(err)
	}
	tokens := []string{}
	for range 12 {
		token, err := db.NewSession("owner")
		if err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, token)
	}
	if err := db.ChangeAccount("owner", "renamed", nil, []byte("hash")); err != nil {
		t.Fatal(err)
	}
	for _, token := range tokens {
		if _, err := db.Session(token); err == nil {
			t.Error("one of multiple sessions survived account change")
		}
	}
}

func TestAccountChangeKeepsOwnershipAndInvalidatesSessions(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(User{ID: "owner", Username: "Original", Password: []byte("hash")}); err != nil {
		t.Fatal(err)
	}
	session, err := db.NewSession("owner")
	if err != nil {
		t.Fatal(err)
	}
	a, _, err := db.Enroll("owner", "owned device")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.ChangeAccount("owner", "新用户", nil, []byte("wrong")); !errors.Is(err, ErrDenied) {
		t.Fatalf("stale credential accepted: %v", err)
	}
	if err := db.ChangeAccount("owner", "新用户", nil, []byte("hash")); err != nil {
		t.Fatal(err)
	}
	u, err := db.User("新用户")
	if err != nil || u.ID != "owner" {
		t.Fatalf("ID changed: %#v %v", u, err)
	}
	if _, err := db.User("Original"); err == nil {
		t.Fatal("old name still resolves")
	}
	agents, err := db.Agents(u.ID)
	if err != nil || len(agents) != 1 || agents[0].ID != a.ID {
		t.Fatal("ownership lost", agents, err)
	}
	if _, err := db.Session(session); err == nil {
		t.Fatal("old session accepted")
	}
}

func TestLastActiveAdministratorProtectedAtomically(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, id := range []string{"a", "b"} {
		if err := db.CreateUser(User{ID: id, Username: id, Admin: true}); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	var success atomic.Int32
	yes := true
	for _, id := range []string{"a", "b"} {
		wg.Go(func() {
			if _, err := db.UpdateUser("", id, nil, &yes, nil); err == nil {
				success.Add(1)
			} else if !errors.Is(err, ErrLastAdmin) {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatalf("expected one disable, got %d", success.Load())
	}
	users, err := db.Users("")
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range users {
		if !u.Disabled {
			no := false
			if _, err := db.UpdateUser("", u.ID, &no, nil, nil); !errors.Is(err, ErrLastAdmin) {
				t.Fatalf("last admin demoted: %v", err)
			}
		}
	}
}

func TestDisabledAccountsDenyCredentialsAndPendingEnrollments(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateUser(User{ID: "owner", Username: "owner"}); err != nil {
		t.Fatal(err)
	}
	session, _ := db.NewSession("owner")
	_, invite, _ := db.Enroll("owner", "pending")
	yes := true
	if _, err := db.UpdateUser("", "owner", nil, &yes, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Session(session); err == nil {
		t.Fatal("disabled session accepted")
	}
	if _, err := db.NewSession("owner"); err == nil {
		t.Fatal("disabled user signed in")
	}
	if _, _, err := db.Redeem(invite, "windows", "amd64"); err == nil {
		t.Fatal("disabled enrollment accepted")
	}
	if _, _, err := db.Enroll("owner", "new"); err == nil {
		t.Fatal("disabled user enrolled device")
	}
}

func TestLegacyAccountsPreserveSessionsAndOwnership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	raw, err := bolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = raw.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"users", "sessions", "agents", "enrollments"} {
			if _, e := tx.CreateBucketIfNotExists([]byte(name)); e != nil {
				return e
			}
		}
		if err := put(tx, "sessions", Hash("legacy-session"), Session{UserID: "stable-id", ExpiresAt: time.Now().Add(time.Hour).Unix()}); err != nil {
			return err
		}
		if err := put(tx, "agents", "legacy-device", Agent{ID: "legacy-device", UserID: "stable-id", Name: "existing device", Enrolled: true, TokenHash: Hash("legacy-agent-credential")}); err != nil {
			return err
		}
		return put(tx, "users", "Alice", User{ID: "stable-id", Username: "Alice", Password: []byte("existing-hash")})
	})
	if err != nil {
		t.Fatal(err)
	}
	raw.Close()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	u, err := db.User(" alice ")
	if err != nil || u.ID != "stable-id" || string(u.Password) != "existing-hash" {
		t.Fatalf("migration lost user: %#v %v", u, err)
	}
	if session, err := db.Session("legacy-session"); err != nil || session.UserID != u.ID {
		t.Fatal("migration lost session", session, err)
	}
	if a, err := db.AuthenticateAgent("legacy-device", "legacy-agent-credential"); err != nil || a.UserID != u.ID {
		t.Fatal("migration lost device", a, err)
	}
}
