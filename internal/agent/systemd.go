package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const linuxUnitName = "remote-agent.service"
const linuxUnitMarker = "# Managed by remote-agent CLI."

func systemdQuote(value string) string {
	return strconv.Quote(strings.ReplaceAll(value, "%", "%%"))
}

func linuxExecStart(exe, config string) string {
	// ':' disables environment expansion; quoted paths may contain literal '$'.
	return "ExecStart=:" + systemdQuote(exe) + " run --config " + systemdQuote(config) + " --log " + systemdQuote(filepath.Join(filepath.Dir(config), "agent.log"))
}

func linuxUnit(exe, config string) string {
	unit := linuxUnitMarker + "\n[Unit]\nDescription=Remote Agent CLI\nStartLimitIntervalSec=0\n\n[Service]\nType=simple\n" +
		linuxExecStart(exe, config) + "\nWorkingDirectory=~" +
		"\nRestart=always\nRestartSec=5\nTimeoutStopSec=15\nKillMode=control-group\nUMask=0077\n"
	for _, key := range []string{"PATH", "RA_CODEX_PATH", "RA_QODER_PATH"} {
		if value := os.Getenv(key); value != "" {
			unit += "Environment=" + systemdQuote(key+"="+value) + "\n"
		}
	}
	return unit + "\n[Install]\nWantedBy=default.target\n"
}

func linuxUnitPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "systemd", "user", linuxUnitName), nil
}

func linuxUnitMatches(data []byte, exe, config string) bool {
	marker, actions := false, 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == linuxUnitMarker {
			marker = true
		}
		if strings.HasPrefix(line, "ExecStart=") {
			if line != linuxExecStart(exe, config) {
				return false
			}
			actions++
		}
	}
	return marker && actions == 1
}

func systemdUser(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	out, err := command(ctx, "systemctl", append([]string{"--user"}, args...)).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("systemd user service: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func autostartLinux(action, exe, config string) error {
	path, err := linuxUnitPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil && !linuxUnitMatches(data, exe, config) {
		return fmt.Errorf("%s belongs to another installation or a custom supervisor; refusing to overwrite it", path)
	}
	if action == "remove" {
		if os.IsNotExist(err) {
			return nil
		}
		if _, err := systemdUser("disable", "--now", linuxUnitName); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		_, err = systemdUser("daemon-reload")
		return err
	}
	if _, err := systemdUser("show-environment"); err != nil {
		return fmt.Errorf("a systemd user login session is required: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(linuxUnit(exe, config)), 0600); err != nil {
		return err
	}
	if _, err := systemdUser("daemon-reload"); err != nil {
		return err
	}
	_, err = systemdUser("enable", "--now", linuxUnitName)
	return err
}
