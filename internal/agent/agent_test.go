package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestWorkspaceBoundary(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	child := filepath.Join(root, "project")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	c := Config{AllowedRoots: []string{root}}
	if _, err := c.Workspace(child); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Workspace(outside); err == nil {
		t.Fatal("accepted directory outside roots")
	}
	if _, err := c.Workspace(filepath.Join(root, "..")); err == nil {
		t.Fatal("accepted traversal")
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err == nil {
		if _, err = c.Workspace(link); err == nil {
			t.Fatal("accepted escaping symlink")
		}
	}
}
func TestRunnerPassesPromptThroughStdin(t *testing.T) {
	// Reuse the test executable as a fake CLI; no model call or installed CLI needed.
	t.Setenv("RA_TEST_HELPER", "1")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := command(ctx, exe, []string{"-test.run=TestHelperProcess"})
	cmd.Stdin = strings.NewReader("$(not-a-command) & 中文; 'quoted'\nsecond line")
	var output boundedBuffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if output.String() != "$(not-a-command) & 中文; 'quoted'\nsecond line" {
		t.Fatalf("prompt changed: %q", output.String())
	}
}
func TestOutputWriterUTF8(t *testing.T) {
	var text strings.Builder
	w := &outputWriter{send: func(e protocol.Event) error { text.WriteString(e.Data.(map[string]string)["text"]); return nil }}
	input := []byte(strings.Repeat("你好🙂abc", 2000))
	for i := 0; i < len(input); i += 7 {
		if _, err := w.Write(input[i:min(i+7, len(input))]); err != nil {
			t.Fatal(err)
		}
	}
	if text.String() != string(input) {
		t.Fatal("stream corrupted split UTF-8 characters")
	}
}
func TestProviderPermissions(t *testing.T) {
	for _, name := range []string{"codex", "qoder"} {
		args := strings.Join(providerArgs(name, false), " ")
		if strings.Contains(args, "bypass") || strings.Contains(args, "yolo") {
			t.Fatal("unsafe defaults")
		}
		if name == "codex" && !strings.Contains(args, "read-only") {
			t.Fatal(args)
		}
		if name == "qoder" && !strings.Contains(args, "plan") {
			t.Fatal(args)
		}
	}
}
