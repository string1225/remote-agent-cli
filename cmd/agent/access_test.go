package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/string1225/remote-agent-cli/internal/access"
	"strings"
	"testing"
)

func TestServersCLIPasswordNeverAppearsInArgumentsOrOutput(t *testing.T) {
	var output bytes.Buffer
	called := false
	call := func(_ context.Context, _ string, r access.Request) (json.RawMessage, error) {
		called = true
		if r.Password != "mock-stdin-secret" || r.Action != "servers.add" {
			t.Fatal("password was not delivered to credential handler")
		}
		return json.RawMessage(`{"name":"fixture"}`), nil
	}
	args := []string{"add", "--name", "fixture", "--host", "127.0.0.1", "--user", "fixture", "--password-stdin"}
	if err := runServers(context.Background(), args, strings.NewReader("mock-stdin-secret\n"), &output, call); err != nil {
		t.Fatal(err)
	}
	if !called || strings.Contains(output.String(), "mock-stdin-secret") {
		t.Fatal("missing handler or password in output")
	}
	for _, extra := range [][]string{{"--password", "must-not-be-an-argument"}, {"--unknown", "x"}} {
		if err := runServers(context.Background(), append(args, extra...), strings.NewReader(""), &output, call); err == nil {
			t.Fatal("plaintext password flag accepted")
		}
	}
}
func TestGitHubCLIPrintsDeviceCodeAndReturnsWithoutOwningGrant(t *testing.T) {
	var output bytes.Buffer
	call := func(_ context.Context, _ string, r access.Request) (json.RawMessage, error) {
		if r.Action != "github.start" {
			t.Fatal(r.Action)
		}
		return json.RawMessage(`{"status":"waiting","url":"https://github.com/login/device","code":"A1B2-C3D4","message":"waiting"}`), nil
	}
	if err := runGitHub(context.Background(), []string{"login"}, &output, call); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "A1B2-C3D4") || !strings.Contains(output.String(), "https://github.com/login/device") {
		t.Fatal("missing user instructions")
	}
}
