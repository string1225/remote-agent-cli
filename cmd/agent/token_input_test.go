package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenInputRejectsAmbiguousMultilineAndOversizedSecrets(t *testing.T) {
	for _, test := range []struct {
		name, input              string
		prompt, stdin, wantError bool
	}{
		{"stdin", "fixture-token\r\n", false, true, false},
		{"both", "fixture-token", true, true, true},
		{"none", "fixture-token", false, false, true},
		{"empty", "  \n", false, true, true},
		{"multiline", "one\ntwo", false, true, true},
		{"oversized", strings.Repeat("x", 4097), false, true, true},
		{"prompt without terminal", "fixture-token", true, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input")
			if err := os.WriteFile(path, []byte(test.input), 0600); err != nil {
				t.Fatal(err)
			}
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			previous := os.Stdin
			os.Stdin = f
			defer func() { os.Stdin = previous }()
			got, err := readSetupToken(test.prompt, test.stdin)
			if (err != nil) != test.wantError {
				t.Fatal("token validation result", err)
			}
			if err == nil && got != strings.TrimSpace(test.input) {
				t.Fatal("token altered")
			}
			if err != nil && strings.Contains(err.Error(), "fixture-token") {
				t.Fatal("secret leaked in error")
			}
		})
	}
}
