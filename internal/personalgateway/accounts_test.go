package personalgateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func managerFixture(t *testing.T) *AccountManager {
	t.Helper()
	m, err := NewAccountManager(t.TempDir(), "unused-codex", "", t.TempDir(), strings.Repeat("control-only-", 4))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	m.source(m.manifest.Accounts[0].Home).Fetch = func(context.Context, bool) (Credential, error) {
		return Credential{Token: "upstream-secret", AccountID: "original", Expires: time.Now().Add(time.Hour)}, nil
	}
	return m
}
func waitLogin(t *testing.T, m *AccountManager, state string) {
	t.Helper()
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		m.mu.Lock()
		got := m.login.State
		m.mu.Unlock()
		if got == state {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("login did not reach %s", state)
}
func TestAccountPoolPreservesOriginalAndActivatesExplicitly(t *testing.T) {
	m := managerFixture(t)
	oldHome := m.manifest.Accounts[0].Home
	if err := os.WriteFile(filepath.Join(oldHome, "auth.json"), []byte("retained original"), 0600); err != nil {
		t.Fatal(err)
	}
	complete := make(chan struct{})
	m.launch = func(ctx context.Context, home string) (deviceChallenge, func() (Credential, error), error) {
		return deviceChallenge{"native-login", "TEST-CODE", "https://auth.openai.com/codex/device"}, func() (Credential, error) {
			select {
			case <-complete:
				return Credential{Token: "new-secret", AccountID: "second", Expires: time.Now().Add(time.Hour)}, nil
			case <-ctx.Done():
				return Credential{}, ctx.Err()
			}
		}, nil
	}
	login, err := m.begin("", "第二个账号")
	if err != nil || login.UserCode != "TEST-CODE" {
		t.Fatal(err)
	}
	if _, err := m.begin("", "duplicate pending"); err == nil {
		t.Fatal("allowed two concurrent login processes")
	}
	close(complete)
	waitLogin(t, m, "completed")
	if len(m.manifest.Accounts) != 2 || m.manifest.Active != "legacy" {
		t.Fatal("adding an account changed current model account")
	}
	second := m.manifest.Accounts[1]
	if second.Home == oldHome {
		t.Fatal("credentials not isolated")
	}
	m.source(second.Home).Fetch = func(context.Context, bool) (Credential, error) {
		return Credential{Token: "new-secret", AccountID: "second", Expires: time.Now().Add(time.Hour)}, nil
	}
	if err := m.activate(context.Background(), second.ID); err != nil {
		t.Fatal(err)
	}
	c, err := m.Get(context.Background(), "")
	if err != nil || c.AccountID != "second" {
		t.Fatal("active source not changed", err)
	}
	status, _ := json.Marshal(m.status(context.Background()))
	if strings.Contains(string(status), "secret") || strings.Contains(string(status), oldHome) {
		t.Fatal("status exposed credentials or paths")
	}
	reloaded, err := NewAccountManager(m.Root, m.Binary, "", oldHome, m.Token)
	if err != nil || reloaded.manifest.Active != second.ID {
		t.Fatal("active selection not persisted", err)
	}
	reloaded.Close()
	before, _ := os.ReadFile(filepath.Join(m.Root, "accounts.json"))
	// Reauthorizing an existing account with a different identity must fail safely.
	_, err = m.begin("legacy", "original")
	if err != nil {
		t.Fatal(err)
	}
	waitLogin(t, m, "failed")
	after, _ := os.ReadFile(filepath.Join(m.Root, "accounts.json"))
	if string(before) != string(after) {
		t.Fatal("failed authorization changed account registry")
	}
	data, _ := os.ReadFile(filepath.Join(oldHome, "auth.json"))
	if string(data) != "retained original" {
		t.Fatal("overwrote original credentials")
	}
}
func TestAccountControlNeedsDedicatedCredentialAndRetainsFailures(t *testing.T) {
	m := managerFixture(t)
	for _, auth := range []string{"", "Bearer sunny_device_token", "Bearer upstream-secret"} {
		r := httptest.NewRequest("GET", "/internal/codex/accounts", nil)
		r.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatal("control access accepted", auth)
		}
	}
	m.launch = func(context.Context, string) (deviceChallenge, func() (Credential, error), error) {
		return deviceChallenge{}, nil, errors.New("private native error")
	}
	r := httptest.NewRequest("POST", "/internal/codex/login", strings.NewReader(`{"name":"test"}`))
	r.Header.Set("Authorization", "Bearer "+m.Token)
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	if w.Code != http.StatusConflict || strings.Contains(w.Body.String(), "private native error") {
		t.Fatal(w.Code, w.Body.String())
	}
	if len(m.manifest.Accounts) != 1 || m.manifest.Active != "legacy" {
		t.Fatal("failed login damaged working account")
	}
	for _, u := range []string{"http://auth.openai.com/device", "https://evil.invalid/device", "https://auth.openai.com@evil.invalid/", "https://auth.openai.com:444/device"} {
		if validChallenge(deviceChallenge{"id", "code", u}) {
			t.Fatal("accepted unsafe authorization URL", u)
		}
	}
}
