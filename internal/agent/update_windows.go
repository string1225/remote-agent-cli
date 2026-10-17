package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func lockFile(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	overlap := &windows.Overlapped{}
	if err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlap); err != nil {
		f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, errLockHeld
		}
		return nil, err
	}
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, overlap)
		f.Close()
	}, nil
}

func updatePowerShell(script string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	// Windows PowerShell serializes module progress to stderr as CLIXML.
	// Only stdout is a machine-readable result; keep diagnostics for errors.
	script = "$ProgressPreference='SilentlyContinue'; " + script
	out, err := command(ctx, shell, []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script)}).Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			out = append(out, exit.Stderr...)
		}
		return "", fmt.Errorf("update autostart: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func deferWindowsWatchdog(minutes int) string {
	// Give the replacement a bounded quiet window without disabling recovery.
	// If the updater itself crashes, the timer still resumes after five minutes.
	return "$watchdogs=@($task.Triggers | Where-Object { $_.Repetition.Interval -eq 'PT1M' -and !$_.Repetition.Duration -and !$_.EndBoundary }); " +
		"foreach ($watchdog in $watchdogs) { $watchdog.StartBoundary=(Get-Date).AddMinutes(" + strconv.Itoa(minutes) + ").ToString('s') }; " +
		"if ($watchdogs.Count) { Set-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -Trigger $task.Triggers | Out-Null }; "
}

func prepareServiceUpdate(exe, config string, startStopped bool) (serviceUpdate, error) {
	service := serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	launch := "& " + quotePowerShell(exe) + " run --config " + quotePowerShell(config) + " --log " + quotePowerShell(filepath.Join(filepath.Dir(config), "agent.log")) + "; exit $LASTEXITCODE"
	args := "-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand " + encodedPowerShell(launch)
	prelude := "$ErrorActionPreference='Stop'; $task=Get-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -ErrorAction SilentlyContinue; " +
		"$taskMatches=$task -and @($task.Actions).Count -eq 1 -and $task.Actions[0].Execute -eq " + quotePowerShell(shell) + " -and $task.Actions[0].Arguments -eq " + quotePowerShell(args) + "; "
	condition := "$taskMatches -and $task.State -eq 'Running'"
	if startStopped {
		condition = "$taskMatches"
	}
	state, err := updatePowerShell(prelude + "if (" + condition + ") { 'running' } else { 'unmanaged' }")
	if err != nil || state != "running" {
		return service, err
	}
	processes := "$agentProcesses=@(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq " + quotePowerShell(exe) + " -and $_.ProcessId -ne " + strconv.Itoa(os.Getpid()) + ` -and $_.CommandLine -match '\srun(\s|$)' }); `
	busy := processes + "foreach ($agentProcess in $agentProcesses) { if (Get-CimInstance Win32_Process -Filter ('ParentProcessId=' + $agentProcess.ProcessId)) { throw 'An agent task is active. Finish it before updating.' } }; "
	// Task Scheduler terminates its PowerShell launcher but may leave the CLI
	// child alive. Capture only children of the matching launcher before stopping.
	owned := "$managedAgents=@(); foreach ($agentProcess in $agentProcesses) { $taskLauncher=Get-CimInstance Win32_Process -Filter ('ProcessId=' + $agentProcess.ParentProcessId); if (!$taskLauncher -or $taskLauncher.ExecutablePath -ne " + quotePowerShell(shell) + " -or !$taskLauncher.CommandLine.Contains(" + quotePowerShell(encodedPowerShell(launch)) + ")) { throw 'A foreground agent is running. Stop it manually before updating.' }; $managedAgents += $agentProcess }; "
	terminate := "foreach ($managedAgent in $managedAgents) { $remainingAgent=Get-CimInstance Win32_Process -Filter ('ProcessId=' + $managedAgent.ProcessId); if ($remainingAgent -and $remainingAgent.ExecutablePath -eq " + quotePowerShell(exe) + " -and $remainingAgent.CreationDate -eq $managedAgent.CreationDate) { if (Get-CimInstance Win32_Process -Filter ('ParentProcessId=' + $remainingAgent.ProcessId)) { throw 'An agent task became active; stop it before updating.' }; Stop-Process -Id $remainingAgent.ProcessId -Force } }; "
	service.managed = true
	service.stop = func() error {
		_, err := updatePowerShell(prelude + "if (!$taskMatches) { throw 'Autostart configuration changed; update cancelled.' }; " + busy + owned +
			deferWindowsWatchdog(5) +
			"Stop-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'; " +
			terminate + "for ($attempt=0; $attempt -lt 50; $attempt++) { " + processes + "if (!$agentProcesses.Count) { exit 0 }; Start-Sleep -Milliseconds 100 }; throw 'Agent is still running. Stop any manually launched agent before updating.'")
		return err
	}
	service.start = func() error {
		_, err := updatePowerShell(prelude + "if (!$taskMatches) { throw 'Autostart configuration changed; cannot restart.' }; " + deferWindowsWatchdog(1) + "Start-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'; " +
			"for ($attempt=0; $attempt -lt 50; $attempt++) { " + processes + "if ($agentProcesses.Count) { exit 0 }; Start-Sleep -Milliseconds 100 }; throw 'Agent did not start.'")
		return err
	}
	return service, nil
}
