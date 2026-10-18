package codexsetup

import (
	"fmt"
	"os/exec"
	"os/user"
	"syscall"
)

func protect(path string) error {
	current, err := user.Current()
	if err != nil {
		return err
	}
	cmd := exec.Command("icacls.exe", path, "/inheritance:r", "/grant:r", "*"+current.Uid+":F", "*S-1-5-18:F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protect Codex credentials: %w", err)
	}
	return nil
}
