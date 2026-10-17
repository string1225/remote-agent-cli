package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

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
