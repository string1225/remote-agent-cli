package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWindowsCMDShimStdin(t *testing.T) {
	t.Setenv("RA_TEST_HELPER", "1")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(t.TempDir(), "shim with spaces.cmd")
	if err := os.WriteFile(shim, []byte("@echo off\r\n\""+exe+"\" -test.run=TestHelperProcess\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := command(ctx, shim, nil)
	prompt := "中文 $(not-a-command) & ; 'quotes'\nsecond line"
	cmd.Stdin = strings.NewReader(prompt)
	var output boundedBuffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("%v: %s", err, output.String())
	}
	if output.String() != prompt {
		t.Fatalf("CMD shim corrupted stdin: %q", output.String())
	}
}
