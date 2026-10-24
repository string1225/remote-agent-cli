package agent

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestElevationScriptPreservesDiagnosticsAndQuotes(t *testing.T) {
	script := serviceElevationScript(`C:\Program Files\Remote Agent\remote-agent.exe`, []string{"service", "configure", "--profile", `C:\Users\O'Brien\profile.json`}, `C:\Temp\setup out.log`, `C:\Temp\setup err.log`)
	for _, required := range []string{"-Verb RunAs", "-WindowStyle Hidden", "-EncodedCommand", "NativeErrorCode", "1223"} {
		if !strings.Contains(script, required) {
			t.Errorf("missing %s", required)
		}
	}
	encoded := strings.Split(strings.Split(script, "-EncodedCommand ")[1], "'")[0]
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	inner := decodeUTF16LE(data)
	for _, required := range []string{"O''Brien", "-RedirectStandardOutput 'C:\\Temp\\setup out.log'", "-RedirectStandardError 'C:\\Temp\\setup err.log'", "exit $p.ExitCode"} {
		if !strings.Contains(inner, required) {
			t.Errorf("missing redirected argument %s", required)
		}
	}
}

func TestElevationHandoffCapturesRealChildDiagnosticsWithoutUAC(t *testing.T) {
	t.Setenv("REMOTE_AGENT_ELEVATION_HELPER", "1")
	dir := t.TempDir()
	stdout, stderr := filepath.Join(dir, "out.log"), filepath.Join(dir, "err.log")
	script := serviceElevationScript(os.Args[0], []string{"-test.run=^TestElevationNativeHelper$", "--", "a path with spaces and O'Brien"}, stdout, stderr)
	// Mock only the RunAs boundary. The second PowerShell process and native
	// fixture process run normally, exercising quoting, redirection and status.
	mock := "function Start-Process { param($FilePath,$ArgumentList,$Verb,$WindowStyle,[switch]$Wait,[switch]$PassThru); Microsoft.PowerShell.Management\\Start-Process -FilePath $FilePath -ArgumentList $ArgumentList -WindowStyle Hidden -Wait -PassThru }; "
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	out, err := exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(mock+script)).CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("lost child status: %v %s", err, out)
	}
	output, _ := os.ReadFile(stdout)
	diagnostic, _ := os.ReadFile(stderr)
	if string(output) != "a path with spaces and O'Brien" || !strings.Contains(string(diagnostic), "fixture diagnostic") {
		t.Fatalf("lost native output: %q %q", output, diagnostic)
	}
	// A launch error occurs before the child has stderr, so it must also be
	// persisted by the elevated wrapper rather than disappearing in its window.
	missing := serviceElevationScript(filepath.Join(dir, "missing-fixture.exe"), []string{"service", "configure"}, stdout, stderr)
	_, err = exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(mock+missing)).CombinedOutput()
	if err == nil {
		t.Fatal("missing native executable was accepted")
	}
	diagnostic, _ = os.ReadFile(stderr)
	if !strings.Contains(string(diagnostic), "service configuration launcher:") {
		t.Fatalf("launcher error lost: %q", diagnostic)
	}
}

func TestElevationNativeHelper(t *testing.T) {
	if os.Getenv("REMOTE_AGENT_ELEVATION_HELPER") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, os.Args[len(os.Args)-1])
	fmt.Fprint(os.Stderr, "fixture diagnostic")
	os.Exit(23)
}

func TestSetupLogRetainsChildFailureWithoutProfileContents(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "stderr.log"), []byte("grant service access: Access is denied."), 0600); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(t.TempDir(), "config.json")
	path, detail := retainServiceSetupLog(config, dir, "exit status 1")
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), "Access is denied") || !strings.Contains(detail, "Access is denied") {
		t.Fatal(path, detail, err)
	}
}
