package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestDirectoryCreateRPCAndRetryPreserveExistingContents(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "Code")
	if err := os.Mkdir(parent, 0755); err != nil {
		t.Fatal(err)
	}
	a := &Agent{Config: Config{AllowedRoots: []string{root}}}
	p := &peer{ctx: context.Background()}
	var event protocol.Event
	send := func(e protocol.Event) error { event = e; return nil }
	a.handle(p, protocol.Request{ID: "capabilities", Type: "services.list"}, send)
	if event.Data.(map[string]any)["directoryCreate"] != true {
		t.Fatal("missing capability")
	}
	req := protocol.Request{ID: "create", Type: "directories.create", Path: parent, Title: "新建 项目"}
	a.handle(p, req, send)
	if event.Type != "result" {
		t.Fatal(event)
	}
	created := event.Data.(directoryCreated)
	expected, _ := canonicalDir(filepath.Join(parent, req.Title))
	if !created.Created || created.Path != expected {
		t.Fatal(created)
	}
	encoded, _ := json.Marshal(created)
	if !strings.Contains(string(encoded), `"path":`) || !strings.Contains(string(encoded), `"created":true`) {
		t.Fatal("unexpected result format", string(encoded))
	}
	marker := filepath.Join(created.Path, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep original"), 0600); err != nil {
		t.Fatal(err)
	}
	a.handle(p, req, send)
	if event.Type != "result" || event.Data.(directoryCreated).Created {
		t.Fatal("retry did not reuse directory", event)
	}
	if raw, err := os.ReadFile(marker); err != nil || string(raw) != "keep original" {
		t.Fatal("retry changed existing contents", err)
	}
	if _, err := a.Config.Workspace(created.Path); err != nil {
		t.Fatal("new directory cannot be used as a project", err)
	}
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := a.createDirectory(context.Background(), parent, "concurrent"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestDirectoryCreateRejectsEscapesInvalidNamesAndFiles(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	a := &Agent{Config: Config{AllowedRoots: []string{root}}}
	for _, name := range []string{"", ".", "..", "../escape", "nested/child", `nested\child`, "/absolute", `C:\absolute`, "file:stream", "bad\x00name", "bad\nname", "trailing.", " space", "space ", "CON", "NUL.txt", "COM1", "LPT9.txt", strings.Repeat("字", 100)} {
		t.Run(name, func(t *testing.T) {
			if _, err := a.createDirectory(context.Background(), root, name); err == nil {
				t.Fatal("invalid folder name accepted")
			}
		})
	}
	for _, parent := range []string{outside, filepath.Join(root, "missing"), "relative"} {
		if _, err := a.createDirectory(context.Background(), parent, "escaped"); err == nil {
			t.Fatal("unauthorized or nonexistent parent accepted", parent)
		}
	}
	file := filepath.Join(root, "existing")
	if err := os.WriteFile(file, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := a.createDirectory(context.Background(), root, "existing"); err == nil {
		t.Fatal("existing file was reused as a directory")
	}
	if data, err := os.ReadFile(file); err != nil || string(data) != "private" {
		t.Fatal("existing file changed", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.createDirectory(ctx, root, "cancelled"); err == nil {
		t.Fatal("cancelled request created a directory")
	}
	if _, err := os.Stat(filepath.Join(root, "cancelled")); !os.IsNotExist(err) {
		t.Fatal("cancelled directory exists", err)
	}
}

func TestDirectoryCreateConfinesSymlinks(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	a := &Agent{Config: Config{AllowedRoots: []string{root}}}
	link := filepath.Join(root, "escaped")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symlink unavailable", err)
	}
	if _, err := a.createDirectory(context.Background(), link, "child"); err == nil {
		t.Fatal("followed parent outside authorized roots")
	}
	if _, err := a.createDirectory(context.Background(), root, "escaped"); err == nil {
		t.Fatal("adopted an existing symlink")
	}
	if _, err := os.Stat(filepath.Join(outside, "child")); !os.IsNotExist(err) {
		t.Fatal("created a directory outside authorized roots", err)
	}
}
