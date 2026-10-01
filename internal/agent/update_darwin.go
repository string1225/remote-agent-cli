package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func launchUpdate(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("update autostart: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func prepareServiceUpdate(exe, config string) (serviceUpdate, error) {
	service := serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}
	home, err := os.UserHomeDir()
	if err != nil {
		return service, err
	}
	plist := filepath.Join(home, "Library", "LaunchAgents", "com.remote-agent.cli.plist")
	if _, err := os.Stat(plist); os.IsNotExist(err) {
		return service, nil
	} else if err != nil {
		return service, err
	}
	b, err := launchUpdate("/usr/bin/plutil", "-extract", "ProgramArguments", "json", "-o", "-", plist)
	if err != nil {
		return service, err
	}
	var args []string
	if json.Unmarshal(b, &args) != nil || len(args) < 4 || args[0] != exe || args[1] != "run" || args[2] != "--config" || args[3] != config {
		return service, nil
	}
	domain := fmt.Sprintf("gui/%d", os.Getuid())
	label := domain + "/com.remote-agent.cli"
	if _, err := launchUpdate("/bin/launchctl", "print", label); err != nil {
		return service, nil // An installed but unloaded LaunchAgent stays unloaded.
	}
	service.managed = true
	service.stop = func() error {
		state, err := launchUpdate("/bin/launchctl", "print", label)
		if err == nil {
			pid := regexp.MustCompile(`(?m)^\s*pid = ([0-9]+)\s*$`).FindSubmatch(state)
			if len(pid) == 2 {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := exec.CommandContext(ctx, "/usr/bin/pgrep", "-P", string(pid[1])).Run(); err == nil {
					return fmt.Errorf("an agent task is active; finish it before updating")
				} else if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
					return fmt.Errorf("cannot determine whether the agent has active tasks: %w", err)
				}
			}
		} else {
			return nil // Already unloaded (e.g. failed bootstrap during rollback).
		}
		_, err = launchUpdate("/bin/launchctl", "bootout", domain, plist)
		return err
	}
	service.start = func() error {
		if _, err := launchUpdate("/bin/launchctl", "print", label); err == nil {
			return nil
		}
		_, err := launchUpdate("/bin/launchctl", "bootstrap", domain, plist)
		return err
	}
	return service, nil
}
