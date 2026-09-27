//go:build !windows

package agent

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

func command(ctx context.Context, path string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}
	cmd.WaitDelay = 3 * time.Second
	return cmd
}
