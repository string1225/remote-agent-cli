package agent

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemdUnitEscapesLiteralPathsAndMatchesOnlyItsInstallation(t *testing.T) {
	exe := filepath.Join(t.TempDir(), `agent $USER %n "quoted"`)
	config := filepath.Join(t.TempDir(), "中文 project", "config.json")
	t.Setenv("PATH", "/bin:/project %n/$HOME")
	t.Setenv("RA_CODEX_PATH", `/tools/a "quoted"/codex`)
	unit := linuxUnit(exe, config)
	if !strings.Contains(unit, "ExecStart=:") || !strings.Contains(unit, "%%n") || !strings.Contains(unit, "$USER") || !strings.Contains(unit, `Environment="PATH=/bin:/project %%n/$HOME"`) {
		t.Fatalf("unsafe unit quoting: %s", unit)
	}
	if !linuxUnitMatches([]byte(unit), exe, config) {
		t.Fatal("own unit not recognized")
	}
	for _, changed := range []string{
		strings.ReplaceAll(unit, linuxUnitMarker, "# custom service"),
		unit + "\nExecStart=/bin/false\n",
		strings.Replace(unit, linuxExecStart(exe, config), linuxExecStart(exe, config+".other"), 1),
	} {
		if linuxUnitMatches([]byte(changed), exe, config) {
			t.Fatal("accepted another installation or overridden ExecStart")
		}
	}
	for _, setting := range []string{"WorkingDirectory=~", "Restart=always", "RestartSec=5", "StartLimitIntervalSec=0", "UMask=0077", "WantedBy=default.target"} {
		if !strings.Contains(unit, setting+"\n") {
			t.Fatalf("missing %s", setting)
		}
	}
}
