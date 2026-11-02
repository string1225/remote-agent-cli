package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBindingCodeRevealRenewConsumptionAndRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	s, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	_ = s.CreateUser(User{ID: "owner", Username: "owner"})
	_ = s.CreateUser(User{ID: "other", Username: "other"})
	a, code, e := s.EnrollForOS("owner", "PC", "windows")
	if e != nil {
		t.Fatal(e)
	}
	_, same, _, e := s.BindingCode("owner", a.ID, false)
	if e != nil || same != code {
		t.Fatal("code cannot be revealed", e)
	}
	if _, _, _, e = s.BindingCode("other", a.ID, false); e == nil {
		t.Fatal("cross-account code disclosed")
	}
	s.Close()
	data, _ := os.ReadFile(path)
	if bytes.Contains(data, []byte(code)) {
		t.Fatal("plaintext binding code in database")
	}
	s, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	_, same, _, e = s.BindingCode("owner", a.ID, false)
	if e != nil || same != code {
		t.Fatal("code lost after reopening", e)
	}
	_, fresh, _, e := s.BindingCode("owner", a.ID, true)
	if e != nil || fresh == code {
		t.Fatal("renew failed", e)
	}
	if _, _, e = s.Redeem(code, "windows", "amd64"); e == nil {
		t.Fatal("superseded code worked")
	}
	credential := Token()
	for range 2 {
		bound, e := s.RedeemActivation(fresh, "windows", "amd64", credential)
		if e != nil || bound.ID != a.ID {
			t.Fatal("lost-response retry failed", e)
		}
	}
	if _, e = s.RedeemActivation(fresh, "windows", "amd64", Token()); e == nil {
		t.Fatal("code reused by another caller")
	}
	if _, _, _, e = s.BindingCode("owner", a.ID, false); e == nil {
		t.Fatal("consumed code disclosed")
	}
	if _, _, e = s.Redeem(fresh, "windows", "amd64"); e == nil {
		t.Fatal("legacy caller reused activation code")
	}
}
