package systemproxy

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeBackend struct {
	state    Settings
	failNext bool
	writes   int
}

func (b *fakeBackend) Read() (Settings, error) { return b.state, nil }
func (b *fakeBackend) Write(state Settings) error {
	b.state = state
	b.writes++
	if b.failNext {
		b.failNext = false
		return errors.New("fixture partial failure")
	}
	return nil
}

func TestManagerRestoresOriginalProxyAndPACAcrossRepeatedEnable(t *testing.T) {
	original := Settings{Flags: Direct | Proxy | 4 | 8, Server: "old-proxy.example:8080", Bypass: "localhost;<local>", AutoConfigURL: "https://pac.example.test/user.pac"}
	backend := &fakeBackend{state: original}
	m := Manager{Backend: backend, Path: filepath.Join(t.TempDir(), "system.json")}
	for _, address := range []string{"127.0.0.1:17891", "127.0.0.1:17891", "127.0.0.1:17892"} {
		if err := m.Enable(address); err != nil {
			t.Fatal(err)
		}
	}
	if backend.state.Flags != Direct|Proxy || backend.state.AutoConfigURL != "" || backend.state.Server != "127.0.0.1:17892" {
		t.Fatal("system proxy was not applied", backend.state)
	}
	if err := m.Disable("127.0.0.1:17892"); err != nil {
		t.Fatal(err)
	}
	if backend.state != original {
		t.Fatal("original settings were lost", backend.state)
	}
	if _, err := os.Stat(m.Path + ".restored"); err != nil {
		t.Fatal("recovery snapshot missing", err)
	}
}

func TestManagerRetainsUserChangesAndRollsBackPartialFailures(t *testing.T) {
	backend := &fakeBackend{state: Settings{Flags: Direct}}
	m := Manager{Backend: backend, Path: filepath.Join(t.TempDir(), "system.json")}
	backend.failNext = true
	if err := m.Enable("127.0.0.1:17891"); err == nil || backend.state != (Settings{Flags: Direct}) {
		t.Fatal("failed enable left partial settings", backend.state, err)
	}
	if err := m.Enable("127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	changed := Settings{Flags: Direct | Proxy, Server: "user-proxy.example:9090"}
	backend.state = changed
	if err := m.Disable("127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	if backend.state != changed {
		t.Fatal("external changes were overwritten")
	}
}

func TestManagerOffDisablesMatchingManualProxyButPreservesOthers(t *testing.T) {
	for _, same := range []bool{true, false} {
		backend := &fakeBackend{state: Settings{Flags: Direct | Proxy, Server: "127.0.0.1:17891"}}
		if !same {
			backend.state.Server = "different.example:8080"
		}
		m := Manager{Backend: backend, Path: filepath.Join(t.TempDir(), "system.json")}
		if err := m.Disable("127.0.0.1:17891"); err != nil {
			t.Fatal(err)
		}
		if (backend.state.Flags&Proxy == 0) != same {
			t.Fatal("wrong manual proxy restored", backend.state)
		}
	}
}

func TestManagerRejectsBrokenBackupWithoutChangingSettings(t *testing.T) {
	backend := &fakeBackend{state: Settings{Flags: Direct}}
	m := Manager{Backend: backend, Path: filepath.Join(t.TempDir(), "system.json")}
	os.WriteFile(m.Path, []byte("{broken"), 0600)
	if err := m.Enable("127.0.0.1:17891"); err == nil || backend.writes != 0 {
		t.Fatal("invalid backup ignored")
	}
}
