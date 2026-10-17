package agent

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRunLockPreventsDuplicateButAllowsOtherConfigs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	unlock, err := LockRun(path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if other, err := LockRun(path); !errors.Is(err, ErrAgentRunning) {
		if other != nil {
			other()
		}
		t.Fatalf("duplicate agent was not rejected: %v", err)
	}
	other, err := LockRun(path + ".other")
	if err != nil {
		t.Fatal(err)
	}
	other()
	updater, err := lockUpdate(path + ".update.lock")
	if err != nil {
		t.Fatal("run lock interferes with updates:", err)
	}
	updater()
	if _, err := LockRun(filepath.Join(path, "missing", "config.json")); err == nil || errors.Is(err, ErrAgentRunning) {
		t.Fatalf("I/O error reported as a running agent: %v", err)
	}
}

func TestRunLockReleasedAfterCrash(t *testing.T) {
	if os.Getenv("REMOTE_AGENT_RUN_LOCK_HELPER") == "1" {
		unlock, err := LockRun(os.Getenv("REMOTE_AGENT_RUN_LOCK_PATH"))
		if err != nil {
			t.Fatal(err)
		}
		defer unlock()
		_, _ = os.Stdout.WriteString("locked\n")
		for {
			time.Sleep(time.Hour)
		}
	}
	path := filepath.Join(t.TempDir(), "config.json")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestRunLockReleasedAfterCrash$")
	cmd.Env = append(os.Environ(), "REMOTE_AGENT_RUN_LOCK_HELPER=1", "REMOTE_AGENT_RUN_LOCK_PATH="+path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	if line, err := bufio.NewReader(out).ReadString('\n'); err != nil || line != "locked\n" {
		t.Fatalf("child lock: %q, %v", line, err)
	}
	if unlock, err := LockRun(path); !errors.Is(err, ErrAgentRunning) {
		if unlock != nil {
			unlock()
		}
		t.Fatalf("another process acquired live lock: %v", err)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	unlock, err := LockRun(path)
	if err != nil {
		t.Fatalf("crash left agent permanently locked out: %v", err)
	}
	unlock()
}
