package main

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/gitauth"
)

func TestGitCommandParsesOnlySupportedOperations(t *testing.T) {
	config := filepath.Join(t.TempDir(), "custom config.json")
	var output bytes.Buffer
	called := false
	invoke := func(_ context.Context, path string, req gitauth.Request, w io.Writer) error {
		called = true
		if path != config || req.Version != gitauth.Version || req.Action != "login" || req.Username != "fixture-user" || !req.Force {
			t.Fatal(path, req)
		}
		_, err := io.WriteString(w, "fixture-device-code")
		return err
	}
	if err := runGitCommand([]string{"login", "--config", config, "--username", "fixture-user", "--force"}, &output, invoke); err != nil {
		t.Fatal(err)
	}
	if !called || !strings.Contains(output.String(), "fixture-device-code") {
		t.Fatal(output.String())
	}
	for _, args := range [][]string{{"logout"}, {"status", "--force"}, {"login", "--username", "--help"}, {"login", "--url", "https://attacker.invalid"}, {"status", "extra"}} {
		called = false
		if err := runGitCommand(args, &bytes.Buffer{}, invoke); err == nil || called {
			t.Fatal("invalid command reached service", args, err)
		}
	}
}

func TestGitHelpDoesNotContactService(t *testing.T) {
	for _, args := range [][]string{nil, {"--help"}, {"login", "--help"}} {
		var output bytes.Buffer
		if err := runGitCommand(args, &output, func(context.Context, string, gitauth.Request, io.Writer) error {
			t.Fatal("help contacted service")
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "git login") {
			t.Fatal(output.String())
		}
	}
}

func TestWindowsInitializationReminderIncludesAuthorizationAndConfig(t *testing.T) {
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	for _, noStart := range []bool{false, true} {
		var output bytes.Buffer
		gitAuthReminder(&output, "windows", `C:\Users\O'Brien\custom config.json`, noStart)
		for _, want := range []string{"remote-agent git login", "remote-agent git status", `--config 'C:\Users\O''Brien\custom config.json'`, "desktop Git credentials"} {
			if !strings.Contains(output.String(), want) {
				t.Fatal("missing reminder", want, output.String())
			}
		}
		if noStart && !strings.Contains(output.String(), "remote-agent start") {
			t.Fatal("stopped service guidance missing")
		}
	}
	var output bytes.Buffer
	gitAuthReminder(&output, "linux", "/fixture/config.json", false)
	if output.Len() != 0 {
		t.Fatal("Windows reminder shown on Linux")
	}
	t.Setenv("REMOTE_AGENT_LANGUAGE", "zh")
	gitAuthReminder(&output, "windows", `C:\fixture\config.json`, false)
	if !strings.Contains(output.String(), "授权") {
		t.Fatal("Chinese authorization reminder missing", output.String())
	}
}
