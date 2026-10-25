package store

import (
	"path/filepath"
	"testing"
)

func TestEnrollmentOSIsFixedAndMismatchDoesNotConsumeInvite(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.CreateUser(User{ID: "owner", Username: "owner"}); err != nil {
		t.Fatal(err)
	}
	a, invite, err := s.EnrollForOS("owner", "Work PC", "windows")
	if err != nil {
		t.Fatal(err)
	}
	if a.OS != "windows" {
		t.Fatal("missing selected OS")
	}
	if _, _, err := s.Redeem(invite, "linux", "amd64"); err == nil {
		t.Fatal("mismatched OS accepted")
	}
	bound, _, err := s.Redeem(invite, "windows", "amd64")
	if err != nil || bound.ID != a.ID {
		t.Fatalf("mismatch consumed invite: %v", err)
	}
	if _, _, err := s.EnrollForOS("owner", "Bad", "other"); err == nil {
		t.Fatal("unknown OS accepted")
	}
}
