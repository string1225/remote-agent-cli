package agent

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsStopRemovesOnlyOwnedAgentTree(t *testing.T) {
	exe, config := filepath.Join(t.TempDir(), "agent's test.exe"), filepath.Join(t.TempDir(), "中文 config.json")
	shell, args := windowsTaskAction(exe, config)
	for _, scenario := range []string{"owned", "missing", "unrelated", "reused"} {
		t.Run(scenario, func(t *testing.T) {
			script := "$ErrorActionPreference='Stop'; $global:disabled=0; $global:stopped=0; $global:removed=0; $global:killed=@(); " +
				"$global:fixture=[pscustomobject]@{Actions=@([pscustomobject]@{Execute=" + quotePowerShell(shell) + "; Arguments=" + quotePowerShell(args) + "})}; " +
				"$global:processes=@([pscustomobject]@{ProcessId=100;ParentProcessId=1;ExecutablePath=" + quotePowerShell(shell) + ";CommandLine=" + quotePowerShell(args) + ";CreationDate='launcher'}, " +
				"[pscustomobject]@{ProcessId=200;ParentProcessId=100;ExecutablePath=" + quotePowerShell(exe) + ";CommandLine='agent run --config fixture';CreationDate='owned'}, " +
				"[pscustomobject]@{ProcessId=201;ParentProcessId=1;ExecutablePath=" + quotePowerShell(exe) + ";CommandLine='agent run --config fixture';CreationDate='foreground'}, " +
				"[pscustomobject]@{ProcessId=300;ParentProcessId=200;ExecutablePath='codex.exe';CommandLine='codex exec';CreationDate='child'}); " +
				"$scenario=" + quotePowerShell(scenario) + "; if ($scenario -eq 'missing') {$global:fixture=$null}; if ($scenario -eq 'unrelated') {$global:fixture.Actions[0].Arguments='another configuration'}; " +
				"function Get-ScheduledTask { param($TaskPath,$TaskName,$ErrorAction); $global:fixture }; " +
				"function Disable-ScheduledTask { param($TaskPath,$TaskName); $global:disabled++ }; " +
				"function Stop-ScheduledTask { param($TaskPath,$TaskName,$ErrorAction); $global:stopped++; $global:processes=@($global:processes | Where-Object ProcessId -ne 100); if ($scenario -eq 'reused') {$global:processes[0]=[pscustomobject]@{ProcessId=200;ParentProcessId=1;ExecutablePath=" + quotePowerShell(exe) + ";CommandLine='agent run';CreationDate='replacement'}} }; " +
				"function Unregister-ScheduledTask { param($TaskPath,$TaskName,[switch]$Confirm); $global:removed++; $global:fixture=$null }; " +
				"function Get-CimInstance { param($ClassName,$Filter); if ($Filter -match '^ProcessId=([0-9]+)$') {$global:processes | Where-Object ProcessId -eq ([int]$matches[1])} else {$global:processes} }; " +
				"function taskkill.exe { if ($args.Count -ne 4 -or $args[0] -ne '/PID' -or $args[2] -ne '/T' -or $args[3] -ne '/F') {throw 'unsafe taskkill arguments'}; $target=[int]$args[1]; $global:killed+= $target; $global:processes=@($global:processes | Where-Object {$_.ProcessId -ne $target -and $_.ParentProcessId -ne $target}); $global:LASTEXITCODE=0 }; " +
				"function Invoke-Remove { " + windowsTaskRemoveScript(exe, config) + " }; $failure=''; try {Invoke-Remove; if ($scenario -eq 'owned') {Invoke-Remove}} catch {$failure=$_.Exception.Message}; " +
				"[ordered]@{disabled=$global:disabled; stopped=$global:stopped; removed=$global:removed; killed=@($global:killed); remaining=@($global:processes | ForEach-Object ProcessId); failure=$failure} | ConvertTo-Json -Compress"
			out, err := updatePowerShell(script)
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				Disabled, Stopped, Removed int
				Killed, Remaining          []int
				Failure                    string
			}
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatal(err, out)
			}
			switch scenario {
			case "owned":
				if got.Disabled != 1 || got.Stopped != 1 || got.Removed != 1 || len(got.Killed) != 1 || got.Killed[0] != 200 || len(got.Remaining) != 1 || got.Remaining[0] != 201 || got.Failure != "" {
					t.Fatalf("owned tree survived or unrelated foreground process changed: %s", out)
				}
			case "missing":
				if got.Disabled != 0 || got.Stopped != 0 || got.Removed != 0 || len(got.Killed) != 0 || got.Failure != "" {
					t.Fatalf("missing task removal was not idempotent: %s", out)
				}
			case "unrelated":
				if got.Disabled != 0 || got.Stopped != 0 || got.Removed != 0 || len(got.Killed) != 0 || !strings.Contains(got.Failure, "another") {
					t.Fatalf("unrelated task was modified: %s", out)
				}
			case "reused":
				if got.Removed != 1 || len(got.Killed) != 0 || len(got.Remaining) != 3 || got.Failure != "" {
					t.Fatalf("reused process ID was terminated: %s", out)
				}
			}
		})
	}
}

func TestWindowsAutostartPolicyAndScopedMigration(t *testing.T) {
	exe, config := filepath.Join(t.TempDir(), "agent's test.exe"), filepath.Join(t.TempDir(), "中文 config.json")
	shell, args := windowsTaskAction(exe, config)
	// Construct real Scheduler settings and triggers, but mock registration so the
	// test never changes the user's task. Run twice to check migration idempotency.
	script := "$ErrorActionPreference='Stop'; Import-Module ScheduledTasks; " +
		"$global:updates=0; $global:fixture=[pscustomobject]@{ Actions=@([pscustomobject]@{Execute=" + quotePowerShell(shell) + "; Arguments=" + quotePowerShell(args) + "}); Settings=[pscustomobject]@{Enabled=$true}; Triggers=@() }; " +
		"function Get-ScheduledTask { param($TaskPath,$TaskName,$ErrorAction); $global:fixture }; " +
		"function Set-ScheduledTask { param($TaskPath,$TaskName,$Trigger,$Settings); $global:updates++; $global:fixture.Triggers=$Trigger; $global:fixture.Settings=$Settings }; " +
		"function Invoke-Repair { " + windowsTaskRepairScript(exe, config) + " }; Invoke-Repair; Invoke-Repair; " +
		"$migrations=$global:updates; $policy=$global:fixture; " +
		"$global:fixture.Actions[0].Arguments='another configuration'; Invoke-Repair; " +
		"$global:fixture.Actions[0].Arguments=" + quotePowerShell(args) + "; $global:fixture.Settings.Enabled=$false; Invoke-Repair; " +
		"[ordered]@{ migrations=$migrations; updates=$global:updates; triggers=@($policy.Triggers).Count; interval=$policy.Triggers[1].Repetition.Interval; duration=$policy.Triggers[1].Repetition.Duration; stop=$policy.Triggers[1].Repetition.StopAtDurationEnd; available=$policy.Settings.StartWhenAvailable; limit=$policy.Settings.ExecutionTimeLimit; instances=[string]$policy.Settings.MultipleInstances; batteryStart=$policy.Settings.DisallowStartIfOnBatteries; batteryStop=$policy.Settings.StopIfGoingOnBatteries } | ConvertTo-Json -Compress"
	out, err := updatePowerShell(script)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Migrations, Updates, Triggers              int
		Interval, Duration, Limit, Instances       string
		Stop, Available, BatteryStart, BatteryStop bool
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got.Migrations != 1 || got.Updates != 1 {
		t.Fatalf("migration not idempotent or changed a disabled/unrelated task: %s", out)
	}
	if got.Triggers != 2 || got.Interval != "PT1M" || got.Duration != "" || got.Stop || !got.Available || got.Limit != "PT0S" || got.Instances != "IgnoreNew" || got.BatteryStart || got.BatteryStop {
		t.Fatalf("incomplete recovery policy: %s", out)
	}
}

func TestWindowsUpdateDefersWatchdogWithoutDisablingIt(t *testing.T) {
	out, err := updatePowerShell("$ErrorActionPreference='Stop'; Import-Module ScheduledTasks; " + windowsTaskPolicy() +
		"$task=[pscustomobject]@{Triggers=@($t,$watchdog)}; $global:writes=0; function Set-ScheduledTask { param($TaskPath,$TaskName,$Trigger); $global:writes++ }; " +
		deferWindowsWatchdog(5) + "$paused=([datetime]$task.Triggers[1].StartBoundary)-(Get-Date); " +
		deferWindowsWatchdog(1) + "$resumed=([datetime]$task.Triggers[1].StartBoundary)-(Get-Date); " +
		"[ordered]@{writes=$global:writes; paused=$paused.TotalSeconds; resumed=$resumed.TotalSeconds; enabled=$task.Triggers[1].Enabled; interval=$task.Triggers[1].Repetition.Interval; duration=$task.Triggers[1].Repetition.Duration} | ConvertTo-Json -Compress")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Writes             int
		Paused, Resumed    float64
		Enabled            bool
		Interval, Duration string
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err, out)
	}
	if got.Writes != 2 || got.Paused < 290 || got.Paused > 301 || got.Resumed < 50 || got.Resumed > 61 || !got.Enabled || got.Interval != "PT1M" || got.Duration != "" {
		t.Fatalf("unsafe watchdog pause: %s", out)
	}
}
