package agent

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func command(ctx context.Context, path string, args []string) *exec.Cmd {
	ext := strings.ToLower(filepath.Ext(path))
	var cmd *exec.Cmd
	if ext == ".cmd" || ext == ".bat" || ext == ".ps1" {
		script := "& " + psQuote(path)
		for _, arg := range args {
			script += " " + psQuote(arg)
		}
		script += "; exit $LASTEXITCODE"
		shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		cmd = exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-Command", script)
	} else {
		cmd = exec.CommandContext(ctx, path, args...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		kill := exec.Command(filepath.Join(os.Getenv("SystemRoot"), "System32", "taskkill.exe"), "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F")
		kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		if kill.Run() != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
	cmd.WaitDelay = 3 * time.Second
	return cmd
}
