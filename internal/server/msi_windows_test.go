package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMSIBootstrapRetainsBindingAndKeepsSecretsOutOfElevation(t *testing.T) {
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	template, err := installers.ReadFile("installers/windows.ps1")
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"existing", "fresh", "corrupt", "foreign"} {
		t.Run(scenario, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "agent's data")
			if err := os.MkdirAll(root, 0700); err != nil {
				t.Fatal(err)
			}
			config := filepath.Join(root, "config.json")
			binding := []byte(`{"server":"http://localhost:8080/agents","id":"existing-device","credential":"fixture-secret-token","services":[{"id":"saved-project"}]}`)
			if scenario == "foreign" {
				binding = []byte(strings.ReplaceAll(string(binding), "localhost:8080", "localhost:8081"))
			}
			if scenario != "fresh" {
				if err := os.WriteFile(config, binding, 0600); err != nil {
					t.Fatal(err)
				}
			}
			script := strings.NewReplacer("__SERVER__", "http://localhost:8080/agents", "__TOKEN__", "fixture-secret-token", "__MODE__", "agent", "__SETUP_CODEX__", "false").Replace(string(template))
			script = strings.Replace(script, "$root = Join-Path $HOME '.remote-agent'", "$root = "+quote(root), 1)
			// Native calls and elevation are replaced, so this test can never
			// install a package, configure a service or touch Scheduled Tasks.
			script = strings.ReplaceAll(script, "& $exe ", "Invoke-FixtureAgent ")
			path := filepath.Join(root, "bootstrap.ps1")
			if err := os.WriteFile(path, []byte(script), 0600); err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("%x", sha256.Sum256([]byte("fixture-package")))
			if scenario == "corrupt" {
				digest = strings.Repeat("0", 64)
			}
			prefix := `$ErrorActionPreference='Stop'; $global:elevations=0; $global:enrollments=0; $global:configure='';
function Write-Host {}
function icacls.exe { $global:LASTEXITCODE=0 }
function Invoke-WebRequest { param($Uri,$OutFile,[switch]$UseBasicParsing); if ($OutFile) { [IO.File]::WriteAllText($OutFile,'fixture-package'); return }; [pscustomobject]@{Content=($global:digest+'  '+(($Uri -split '/')[-1] -replace '\.sha256$',''))} }
function Start-Process { param($FilePath,$ArgumentList,$Verb,$WindowStyle,[switch]$Wait,[switch]$PassThru); if ($Verb -ne 'RunAs') { throw 'Unexpected native process' }; $global:elevations++; $decoded=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($ArgumentList[-1])); $parseErrors=$null; [Management.Automation.Language.Parser]::ParseInput($decoded,[ref]$null,[ref]$parseErrors)|Out-Null; if ($parseErrors.Count) { throw ($parseErrors|Out-String) }; if ($decoded.Contains('fixture-secret-token')) { throw 'Credential leaked into UAC command line' }; if ($global:elevations -eq 1 -and (!$decoded.Contains('msiexec.exe') -or !$decoded.Contains('Get-PackageSHA256'))) { throw 'Missing MSI verification' }; if ($global:elevations -eq 2) { $global:configure=$decoded }; [pscustomobject]@{ExitCode=0} }
function Invoke-FixtureAgent { if ($args[0] -eq 'enroll') { $global:enrollments++; [IO.File]::WriteAllText($global:config,'{"server":"http://localhost:8080/agents","id":"fixture-new"}') }; if ($args[0] -eq 'service' -and $args[1] -eq 'profile') { $target=$args[([array]::IndexOf($args,'--output')+1)]; [IO.File]::WriteAllText($target,'{}') }; $global:LASTEXITCODE=0 }
`
			prefix += "$global:digest=" + quote(digest) + "; $global:config=" + quote(config) + "; $failure=''; try { & " + quote(path)
			if scenario == "existing" {
				prefix += " -NoStart"
			}
			prefix += " } catch { $failure=$_.Exception.Message }; [ordered]@{elevations=$global:elevations;enrollments=$global:enrollments;noStart=$global:configure.Contains('--no-start');failure=$failure}|ConvertTo-Json -Compress"
			cmd := exec.Command(filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe"), "-NoProfile", "-NonInteractive", "-Command", prefix)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("bootstrap harness: %v\n%s", err, output)
			}
			var got struct {
				Elevations, Enrollments int
				NoStart                 bool
				Failure                 string
			}
			if err = json.Unmarshal([]byte(strings.TrimSpace(string(output))), &got); err != nil {
				t.Fatalf("bootstrap result: %v\n%s", err, output)
			}
			if scenario == "corrupt" || scenario == "foreign" {
				if got.Failure == "" || got.Elevations != 0 || got.Enrollments != 0 {
					t.Fatalf("unsafe failure handling: %+v", got)
				}
			} else {
				if got.Failure != "" || got.Elevations != 2 {
					t.Fatalf("MSI bootstrap failed: %+v", got)
				}
				if scenario == "existing" && (got.Enrollments != 0 || !got.NoStart) {
					t.Fatalf("existing binding/start preference lost: %+v", got)
				}
				if scenario == "fresh" && got.Enrollments != 1 {
					t.Fatalf("fresh device was not enrolled once: %+v", got)
				}
			}
			if scenario != "fresh" {
				retained, _ := os.ReadFile(config)
				if string(retained) != string(binding) {
					t.Fatal("bootstrap changed existing device/project configuration")
				}
			}
		})
	}
}
