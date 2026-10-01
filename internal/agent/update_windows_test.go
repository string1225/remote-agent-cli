package agent

import (
	"strings"
	"testing"
)

func TestUpdatePowerShellSeparatesStateFromDiagnostics(t *testing.T) {
	state, err := updatePowerShell(`Write-Progress -Activity 'fixture' -Status 'loading' -PercentComplete 50; [Console]::Error.WriteLine('module diagnostic'); 'running'`)
	if err != nil || state != "running" {
		t.Fatalf("diagnostics contaminated service state: %q, %v", state, err)
	}
	if _, err := updatePowerShell(`throw 'restart-check-failed'`); err == nil || !strings.Contains(err.Error(), "restart-check-failed") {
		t.Fatalf("PowerShell error diagnostics were lost: %v", err)
	}
}
