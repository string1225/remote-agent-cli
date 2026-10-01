package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestDirectorySearchAndCompletion(t *testing.T) {
	root := t.TempDir()
	code := filepath.Join(root, "Code")
	project := filepath.Join(code, "Remote Agent 中文")
	other := filepath.Join(code, "Other")
	for _, path := range []string{project, other, filepath.Join(root, "node_modules", "HiddenProject")} {
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(code, "RemoteFile"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &Agent{Config: Config{AllowedRoots: []string{root}, Services: []protocol.Service{{Workspace: project}}}}
	canonicalProject, _ := canonicalDir(project)
	canonicalRoot, _ := canonicalDir(root)
	canonicalCode, _ := canonicalDir(code)
	contains := func(page directoryPage, path string) bool {
		for _, item := range page.Directories {
			if item.Path == path {
				return true
			}
		}
		return false
	}
	page, err := a.directories(context.Background(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(page, canonicalRoot) || !contains(page, canonicalProject) {
		t.Fatal("initial roots and registered projects missing")
	}
	page, err = a.directories(context.Background(), "", "remote agent")
	if err != nil {
		t.Fatal(err)
	}
	if !page.Recursive || !contains(page, canonicalProject) {
		t.Fatalf("directory name search: %+v", page)
	}
	page, err = a.directories(context.Background(), "", filepath.Join(code, "Rem"))
	if err != nil {
		t.Fatal(err)
	}
	if page.Recursive || page.Base != canonicalCode || len(page.Directories) != 1 || !contains(page, canonicalProject) {
		t.Fatalf("path completion leaked files or missed result: %+v", page)
	}
	page, err = a.directories(context.Background(), code, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Directories) != 2 || page.Parent != canonicalRoot {
		t.Fatalf("browse failed: %+v", page)
	}
	page, err = a.directories(context.Background(), root, "")
	if err != nil {
		t.Fatal(err)
	}
	if page.Parent != "" {
		t.Fatal("allowed root has navigable outside parent")
	}
	page, err = a.directories(context.Background(), "", "HiddenProject")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Directories) != 0 {
		t.Fatal("recursive search traversed node_modules")
	}
	page, err = a.directories(context.Background(), filepath.Join(root, "node_modules"), "")
	if err != nil || len(page.Directories) != 1 {
		t.Fatal("explicit browse should still allow dependency directories")
	}
	// Typing part of an authorized root must not enumerate its unauthorized parent.
	page, err = a.directories(context.Background(), "", root[:len(root)-1])
	if err != nil || !contains(page, canonicalRoot) {
		t.Fatal("authorized root prefix completion failed")
	}
}

func TestDirectorySearchBoundariesAndBudgets(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	a := &Agent{Config: Config{AllowedRoots: []string{root}}}
	if _, err := a.directories(context.Background(), outside, ""); err == nil {
		t.Fatal("browse escaped root")
	}
	if _, err := a.directories(context.Background(), "", filepath.Join(outside, "unknown")); err == nil {
		t.Fatal("completion escaped root")
	}
	if _, err := a.directories(context.Background(), "", strings.Repeat("x", 4097)); err == nil {
		t.Fatal("oversized query accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.directories(ctx, "", ""); err == nil {
		t.Fatal("cancellation ignored")
	}
	if err := os.Symlink(outside, filepath.Join(root, "escaped-link")); err == nil {
		page, err := a.directories(context.Background(), root, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Directories) != 0 {
			t.Fatal("outside symlink exposed")
		}
		if _, err = a.directories(context.Background(), filepath.Join(root, "escaped-link"), ""); err == nil {
			t.Fatal("outside symlink browsed")
		}
	}
	for i := range 80 {
		if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("project-%03d", i)), 0700); err != nil {
			t.Fatal(err)
		}
	}
	page, err := a.directories(context.Background(), root, "")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(page)
	if !page.Truncated || len(page.Directories) > 40 || len(data) > 32*1024 {
		t.Fatalf("unbounded response: %d entries / %d bytes", len(page.Directories), len(data))
	}
	page, err = a.directories(context.Background(), root, "project-079")
	if err != nil || len(page.Directories) != 1 {
		t.Fatal("filtered browsing did not reach later directory")
	}
}
