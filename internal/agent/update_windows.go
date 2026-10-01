package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func lockUpdate(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	overlap := &windows.Overlapped{}
	if err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlap); err != nil {
		f.Close()
		return nil, errors.New("another update is running or the update lock is unavailable")
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
	out, err := command(ctx, shell, []string{"-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script)}).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("update autostart: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func prepareServiceUpdate(exe, config string) (serviceUpdate, error) {
	service := serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	launch := "& " + quotePowerShell(exe) + " run --config " + quotePowerShell(config) + " --log " + quotePowerShell(filepath.Join(filepath.Dir(config), "agent.log")) + "; exit $LASTEXITCODE"
	args := "-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand " + encodedPowerShell(launch)
	prelude := "$ErrorActionPreference='Stop'; $task=Get-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -ErrorAction SilentlyContinue; " +
		"$taskMatches=$task -and @($task.Actions).Count -eq 1 -and $task.Actions[0].Execute -eq " + quotePowerShell(shell) + " -and $task.Actions[0].Arguments -eq " + quotePowerShell(args) + "; "
	state, err := updatePowerShell(prelude + "if ($taskMatches -and $task.State -eq 'Running') { 'running' } else { 'unmanaged' }")
	if err != nil || state != "running" {
		return service, err
	}
	processes := "$agentProcesses=@(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq " + quotePowerShell(exe) + " -and $_.ProcessId -ne " + strconv.Itoa(os.Getpid()) + ` -and $_.CommandLine -match '\srun(\s|$)' }); `
	busy := processes + "foreach ($agentProcess in $agentProcesses) { if (Get-CimInstance Win32_Process -Filter ('ParentProcessId=' + $agentProcess.ProcessId)) { throw 'An agent task is active. Finish it before updating.' } }; "
	service.managed = true
	service.stop = func() error {
		_, err := updatePowerShell(prelude + "if (!$taskMatches) { throw 'Autostart configuration changed; update cancelled.' }; " + busy +
			"Stop-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'; " +
			"for ($attempt=0; $attempt -lt 50; $attempt++) { " + processes + "if (!$agentProcesses.Count) { exit 0 }; Start-Sleep -Milliseconds 100 }; throw 'Agent is still running. Stop any manually launched agent before updating.'")
		return err
	}
	service.start = func() error {
		_, err := updatePowerShell(prelude + "if (!$taskMatches) { throw 'Autostart configuration changed; cannot restart.' }; Start-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'; " +
			"for ($attempt=0; $attempt -lt 50; $attempt++) { " + processes + "if ($agentProcesses.Count) { exit 0 }; Start-Sleep -Milliseconds 100 }; throw 'Agent did not start.'")
		return err
	}
	return service, nil
}
