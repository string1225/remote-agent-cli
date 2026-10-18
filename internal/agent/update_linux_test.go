package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func linuxFixture(t *testing.T, active bool) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SYSTEMD_FIXTURE_DIR", dir)
	t.Setenv("SYSTEMD_FIXTURE_PID", strconv.Itoa(os.Getpid()))
	script := `#!/bin/sh
printf '%s\n' "$*" >> "$SYSTEMD_FIXTURE_DIR/calls"
[ "$1" = --user ] || exit 9
shift
case "$1" in
show) printf 'ActiveState=%s\nMainPID=%s\n' "$(cat "$SYSTEMD_FIXTURE_DIR/state")" "$(cat "$SYSTEMD_FIXTURE_DIR/pid")" ;;
start|enable) echo active > "$SYSTEMD_FIXTURE_DIR/state"; echo "$SYSTEMD_FIXTURE_PID" > "$SYSTEMD_FIXTURE_DIR/pid" ;;
stop|disable) echo inactive > "$SYSTEMD_FIXTURE_DIR/state"; echo 0 > "$SYSTEMD_FIXTURE_DIR/pid" ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "systemctl"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	state, pid := "inactive", "0"
	if active {
		state, pid = "active", strconv.Itoa(os.Getpid())
	}
	for name, value := range map[string]string{"state": state, "pid": pid} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	exe, _ := os.Executable()
	config := filepath.Join(dir, "agent-config.json")
	return dir, exe, config
}

func writeLinuxFixtureUnit(t *testing.T, exe, config string) string {
	t.Helper()
	path, err := linuxUnitPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(linuxUnit(exe, config)), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLinuxServiceUpdateScopesStateAndRefusesBusyAgents(t *testing.T) {
	dir, exe, config := linuxFixture(t, true)
	path := writeLinuxFixtureUnit(t, exe, config)
	service, err := prepareServiceUpdate(exe, config, false)
	if err != nil || !service.managed {
		t.Fatal("active service not managed:", err)
	}
	child := exec.Command("sleep", "60")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	if err := service.stop(); err == nil || !strings.Contains(err.Error(), "task is active") {
		t.Fatalf("busy service was stopped: %v", err)
	}
	_ = child.Process.Kill()
	_ = child.Wait()
	if err := service.stop(); err != nil {
		t.Fatal(err)
	}
	state, _, _ := linuxServiceState()
	if state != "inactive" {
		t.Fatal("service not stopped")
	}
	if err := service.start(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[Service]\nExecStart=/bin/false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := service.stop(); err == nil {
		t.Fatal("changed unit was stopped")
	}
	unmanaged, err := prepareServiceUpdate(exe, config, true)
	if err != nil || unmanaged.managed {
		t.Fatal("custom supervisor was managed", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "calls")); err != nil {
		t.Fatal(err)
	}
}

func TestLinuxAutostartAndStoppedUpdatePreserveData(t *testing.T) {
	_, exe, config := linuxFixture(t, false)
	if err := os.WriteFile(config, []byte("private binding"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := autostartLinux("install", exe, config); err != nil {
		t.Fatal(err)
	}
	path, _ := linuxUnitPath()
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("unit permissions", err)
	}
	if _, err := systemdUser("stop", linuxUnitName); err != nil {
		t.Fatal(err)
	}
	service, err := prepareServiceUpdate(exe, config, false)
	if err != nil || service.managed {
		t.Fatal("normal update changed a stopped service", err)
	}
	service, err = prepareServiceUpdate(exe, config, true)
	if err != nil || !service.managed {
		t.Fatal("installer cannot recover a stopped service", err)
	}
	if err := service.start(); err != nil {
		t.Fatal(err)
	}
	if err := autostartLinux("remove", exe, config); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("unit not removed", err)
	}
	data, _ := os.ReadFile(config)
	if string(data) != "private binding" {
		t.Fatal("autostart changed the binding")
	}
}
