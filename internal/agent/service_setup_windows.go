package agent

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// Capture the original user's environment before elevation. Only a private
// profile path crosses UAC; device/model credentials never appear in argv.
func SetupWindowsService(config string, noStart bool) error {
	p, err := CaptureServiceProfile(config)
	if err != nil {
		return err
	}
	exe, err := installedServiceExecutable()
	if err != nil {
		return err
	}
	if windows.GetCurrentProcessToken().IsElevated() {
		return ConfigureWindowsService(p, noStart)
	}
	dir, err := os.MkdirTemp("", "remote-agent-service-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	profile := filepath.Join(dir, "profile.json")
	if err = p.Save(profile); err != nil {
		return err
	}
	args := []string{"service", "configure", "--profile", profile}
	if noStart {
		args = append(args, "--no-start")
	}
	nativeArgs := ""
	for _, arg := range args {
		if nativeArgs != "" {
			nativeArgs += " "
		}
		nativeArgs += windows.EscapeArg(arg)
	}
	script := "$ErrorActionPreference='Stop'; $p=Start-Process -FilePath " + quotePowerShell(exe) + " -ArgumentList " + quotePowerShell(nativeArgs) + " -Verb RunAs -WindowStyle Hidden -Wait -PassThru; exit $p.ExitCode"
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err = cmd.Run(); err != nil {
		return errors.New("管理员授权被取消或服务配置失败；已有设备凭证保留")
	}
	return nil
}
