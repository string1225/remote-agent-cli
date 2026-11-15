package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func TestEndpointActivationNeverChangesBindingAndAppIsOptIn(t *testing.T) {
	for _, app := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "config.json")
		original := []byte("existing binding must not even be parsed")
		if err := os.WriteFile(path, original, 0600); err != nil {
			t.Fatal(err)
		}
		o := activationOptions{Server: "https://example.test", APIToken: "ra_codex_fixture", ConfigureCodex: true, InstallCodex: true, InstallApp: app, AutoStart: true, WebProxy: true}
		var calls []string
		forbidden := func(agent.Config) error { t.Fatal("device path called for personal token"); return nil }
		steps := activationSteps{
			Personal: func(server, token string) (codexsetup.Connection, error) {
				calls = append(calls, "validate")
				if server != o.Server || token != o.APIToken {
					t.Fatal("credentials changed")
				}
				return codexsetup.Connection{}, nil
			},
			Install:   func(agent.Config) error { calls = append(calls, "install"); return nil },
			Apply:     func(codexsetup.Connection) error { calls = append(calls, "apply"); return nil },
			App:       func() error { calls = append(calls, "app"); return nil },
			Configure: forbidden, Online: forbidden, Service: func(c agent.Config, _ bool) error { return forbidden(c) }, Proxy: func(c agent.Config, _ bool) error { return forbidden(c) },
		}
		var result activationEvent
		_ = activateDevice(path, o, steps, func(e activationEvent) { result = e })
		expected := []string{"validate", "install", "apply"}
		if app {
			expected = append(expected, "app")
		}
		if !result.Success || result.Bound || !reflect.DeepEqual(calls, expected) {
			t.Fatal(result, calls)
		}
		after, _ := os.ReadFile(path)
		if string(after) != string(original) {
			t.Fatal("binding overwritten")
		}
		if _, err := os.Stat(activationPreferences(path)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("personal token preferences persisted")
		}
		calls = nil
		steps.Personal = func(_, token string) (codexsetup.Connection, error) {
			return codexsetup.Connection{}, errors.New("invalid " + token)
		}
		_ = activateDevice(path, o, steps, func(e activationEvent) { result = e })
		if result.Success || len(calls) != 0 || strings.Contains(result.Message, o.APIToken) {
			t.Fatal("failed validation continued or exposed token")
		}
	}
}
