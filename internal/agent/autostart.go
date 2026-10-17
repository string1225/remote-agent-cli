package agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := command(ctx, name, args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("autostart: %w: %s", err, out)
	}
	return nil
}

func windowsTaskAction(exe, configPath string) (shell, args string) {
	shell = filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	launch := "& " + quotePowerShell(exe) + " run --config " + quotePowerShell(configPath) + " --log " + quotePowerShell(filepath.Join(filepath.Dir(configPath), "agent.log")) + "; exit $LASTEXITCODE"
	args = "-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand " + encodedPowerShell(launch)
	return
}

func windowsTaskPolicy() string {
	// Failure retries alone do not cover a launcher interrupted or stopped by Windows.
	// An indefinite trigger also restarts a stopped task; IgnoreNew keeps a running one.
	return "$user=[System.Security.Principal.WindowsIdentity]::GetCurrent().Name; " +
		"$t=New-ScheduledTaskTrigger -AtLogOn -User $user; " +
		"$watchdog=New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 1); $watchdog.Repetition.StopAtDurationEnd=$false; " +
		"$s=New-ScheduledTaskSettingsSet -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit ([TimeSpan]::Zero) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -MultipleInstances IgnoreNew; "
}

func windowsTaskRepairScript(exe, configPath string) string {
	shell, args := windowsTaskAction(exe, configPath)
	return "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; " +
		"$task=Get-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -ErrorAction SilentlyContinue; " +
		"if (!$task -or !$task.Settings.Enabled -or @($task.Actions).Count -ne 1 -or $task.Actions[0].Execute -ne " + quotePowerShell(shell) + " -or $task.Actions[0].Arguments -ne " + quotePowerShell(args) + ") { return }; " +
		"$watchdogs=@($task.Triggers | Where-Object { $_.Enabled -and $_.Repetition.Interval -eq 'PT1M' -and !$_.Repetition.Duration -and !$_.EndBoundary -and !$_.Repetition.StopAtDurationEnd }); " +
		"if ($watchdogs.Count -and $task.Settings.StartWhenAvailable -and $task.Settings.ExecutionTimeLimit -eq 'PT0S' -and $task.Settings.RestartCount -eq 999 -and $task.Settings.RestartInterval -eq 'PT1M' -and !$task.Settings.DisallowStartIfOnBatteries -and !$task.Settings.StopIfGoingOnBatteries -and $task.Settings.MultipleInstances -eq 'IgnoreNew') { return }; " +
		windowsTaskPolicy() + "Set-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -Trigger @($t,$watchdog) -Settings $s | Out-Null"
}

// RepairAutostart migrates only this installation's existing, enabled Windows task.
// Running a new CLI after an update upgrades the old recovery policy automatically.
func RepairAutostart(configPath string) error {
	if runtime.GOOS != "windows" {
		return nil
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
	shell, _ := windowsTaskAction(exe, configPath)
	return runSetup(shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(windowsTaskRepairScript(exe, configPath)))
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
		// Keep the action stable so existing installers/updaters still recognize it.
		_, args := windowsTaskAction(exe, configPath)
		script := "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; " + windowsTaskPolicy() + "$a=New-ScheduledTaskAction -Execute " + quotePowerShell(shell) + " -Argument " + quotePowerShell(args) + "; $p=New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited; Register-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -Action $a -Trigger @($t,$watchdog) -Principal $p -Settings $s -Force | Out-Null; Start-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'"
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
