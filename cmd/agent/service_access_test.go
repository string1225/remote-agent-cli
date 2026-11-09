package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/serviceaccess"
)

func TestAccessCommandParsesExplicitGrant(t *testing.T) {
	var got serviceaccess.Options
	invoke := func(_ context.Context, o serviceaccess.Options, _ io.Writer) error { got = o; return nil }
	var out bytes.Buffer
	err := runAccess(context.Background(), []string{"grant", "--path", `C:\Work\O'Brien`, "--permission", "modify", "--recursive", "--git-trust"}, &out, invoke)
	if err != nil || got.Path != `C:\Work\O'Brien` || got.Permission != "modify" || !got.Recursive || !got.GitTrust {
		t.Fatal(got, err)
	}
}

func TestAccessCommandDoesNotInvokeForInvalidInputOrHelp(t *testing.T) {
	called := false
	invoke := func(context.Context, serviceaccess.Options, io.Writer) error { called = true; return nil }
	for _, args := range [][]string{{"grant"}, {"grant", "--path", `C:\repo`}, {"grant", "--path", `C:\*`, "--permission", "read"}, {"list", "--path", `C:\repo`}, {"restore", "--backup", "../x"}, {"restore", "--backup", strings.Repeat("a", 32), "--git-trust"}, {"grant", "--path", `C:\repo`, "--permission", "read", "extra"}} {
		if err := runAccess(context.Background(), args, &bytes.Buffer{}, invoke); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	var out bytes.Buffer
	if err := runAccess(context.Background(), []string{"--help"}, &out, invoke); err != nil {
		t.Fatal(err)
	}
	if called || !strings.Contains(out.String(), "DPAPI") || !strings.Contains(out.String(), "restore") || !strings.Contains(out.String(), "read|modify") {
		t.Fatal(called, out.String())
	}
}
