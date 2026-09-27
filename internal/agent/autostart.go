package agent

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
)

func xmlText(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}
func quotePowerShell(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
func encodedPowerShell(script string) string {
	runes := utf16.Encode([]rune(script))
	data := make([]byte, len(runes)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(data[i*2:], r)
	}
	return base64.StdEncoding.EncodeToString(data)
}
func runSetup(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("autostart: %w: %s", err, out)
	}
	return nil
}
func Autostart(action, configPath string) error {
	if action != "install" && action != "remove" {
		return errors.New("use autostart install or autostart remove")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}
	if action == "install" {
		if _, err := Load(configPath); err != nil {
			return err
		}
	}
	root := filepath.Dir(configPath)
	logPath := filepath.Join(root, "agent.log")
	switch runtime.GOOS {
	case "windows":
		shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if action == "remove" {
			return runSetup(shell, "-NoProfile", "-NonInteractive", "-Command", "$ErrorActionPreference='Stop'; Stop-ScheduledTask -TaskName 'RemoteAgent' -ErrorAction SilentlyContinue; Unregister-ScheduledTask -TaskName 'RemoteAgent' -Confirm:$false")
		}
		// A hidden shell prevents a console window appearing at login.
		launch := "& " + quotePowerShell(exe) + " run --config " + quotePowerShell(configPath) + " --log " + quotePowerShell(logPath) + "; exit $LASTEXITCODE"
		args := "-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand " + encodedPowerShell(launch)
		script := "$ErrorActionPreference='Stop'; $user=[System.Security.Principal.WindowsIdentity]::GetCurrent().Name; $a=New-ScheduledTaskAction -Execute " + quotePowerShell(shell) + " -Argument " + quotePowerShell(args) + "; $t=New-ScheduledTaskTrigger -AtLogOn -User $user; $p=New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited; $s=New-ScheduledTaskSettingsSet -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit ([TimeSpan]::Zero) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries; Register-ScheduledTask -TaskName 'RemoteAgent' -Action $a -Trigger $t -Principal $p -Settings $s -Force | Out-Null; Start-ScheduledTask -TaskName 'RemoteAgent'"
		return runSetup(shell, "-NoProfile", "-NonInteractive", "-Command", script)
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir := filepath.Join(home, "Library", "LaunchAgents")
		path := filepath.Join(dir, "com.remote-agent.cli.plist")
		domain := fmt.Sprintf("gui/%d", os.Getuid())
		if action == "remove" {
			_ = runSetup("launchctl", "bootout", domain, path)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		plist := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>Label</key><string>com.remote-agent.cli</string><key>ProgramArguments</key><array><string>` + xmlText(exe) + `</string><string>run</string><string>--config</string><string>` + xmlText(configPath) + `</string><string>--log</string><string>` + xmlText(logPath) + `</string></array><key>RunAtLoad</key><true/><key>KeepAlive</key><true/><key>ThrottleInterval</key><integer>10</integer><key>EnvironmentVariables</key><dict><key>PATH</key><string>` + xmlText(os.Getenv("PATH")) + `</string></dict><key>StandardErrorPath</key><string>` + xmlText(logPath) + `</string></dict></plist>`
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		_ = runSetup("launchctl", "bootout", domain, path)
		if err := os.WriteFile(path, []byte(plist), 0600); err != nil {
			return err
		}
		return runSetup("launchctl", "bootstrap", domain, path)
	default:
		return errors.New("autostart supports Windows and macOS; on Linux supervise remote-agent run with systemd")
	}
}
