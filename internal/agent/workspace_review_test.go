package agent

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func reviewTestGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "safe.directory=" + root, "-c", "core.hooksPath=" + filepath.Join(root, ".no-hooks"), "-c", "commit.gpgsign=false", "-C", root}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, output)
	}
}

func TestWorkspaceReviewIsLiteralBoundedAndLeavesRepositoryUntouched(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	reviewTestGit(t, root, "init", "--initial-branch=main")
	name := "file[1].txt"
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(name, "original\n")
	write("other.txt", "other\n")
	write(".gitattributes", "*.txt filter=review\n")
	reviewTestGit(t, root, "add", ".")
	reviewTestGit(t, root, "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-m", "initial")
	write(name, "staged\n")
	reviewTestGit(t, root, "add", "--", name)
	write(name, "working\n")
	write("new.txt", strings.Repeat("x", workspaceReviewLimit+5))
	reviewTestGit(t, root, "config", "diff.external", "invalid-external-diff")
	reviewTestGit(t, root, "config", "core.fsmonitor", "invalid-fsmonitor")
	sentinel := filepath.Join(t.TempDir(), "filter-executed")
	reviewTestGit(t, root, "config", "filter.review.clean", "echo FILTER_RAN > '"+filepath.ToSlash(sentinel)+"'")
	reviewTestGit(t, root, "config", "filter.review.required", "true")
	indexBefore, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	changes, err := workspaceChanges(context.Background(), root)
	if err != nil || !changes.Git || changes.Branch != "main" || len(changes.Files) != 2 {
		t.Fatal(changes, err)
	}
	diff, err := workspaceDiff(context.Background(), root, name)
	if err != nil || !strings.Contains(diff.Staged, "+staged") || !strings.Contains(diff.Unstaged, "+working") || strings.Contains(diff.Staged, "other") {
		t.Fatal(diff, err)
	}
	untracked, err := workspaceDiff(context.Background(), root, "new.txt")
	if err != nil || !untracked.Truncated || len(untracked.Untracked) > workspaceReviewLimit {
		t.Fatal(len(untracked.Untracked), err)
	}
	for _, name := range []string{"../other.txt", ".git/config", "file*.txt", filepath.Join(root, "other.txt")} {
		if _, err := workspaceDiff(context.Background(), root, name); err == nil {
			t.Fatalf("unauthorized/nonliteral path accepted: %s", name)
		}
	}
	indexAfter, _ := os.ReadFile(filepath.Join(root, ".git", "index"))
	if !bytes.Equal(indexBefore, indexAfter) {
		t.Fatal("read-only review modified Git index")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("read-only review executed a configured Git filter", err)
	}
	content, _ := os.ReadFile(filepath.Join(root, name))
	if string(content) != "working\n" {
		t.Fatal("review modified worktree")
	}
}

func TestWorkspaceReviewRestrictsSubdirectoryAndNonGitProjects(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	reviewTestGit(t, root, "init", "--initial-branch=main")
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "inside.txt"), []byte("inside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "outside.txt"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	changes, err := workspaceChanges(context.Background(), sub)
	if err != nil || len(changes.Files) != 1 || changes.Files[0].Path != "inside.txt" {
		t.Fatal(changes, err)
	}
	diff, err := workspaceDiff(context.Background(), sub, "inside.txt")
	if err != nil || diff.Untracked != "inside" {
		t.Fatal(diff, err)
	}
	plain, err := workspaceChanges(context.Background(), t.TempDir())
	if err != nil || plain.Git {
		t.Fatal(plain, err)
	}
}
