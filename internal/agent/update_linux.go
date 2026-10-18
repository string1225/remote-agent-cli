package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func linuxServiceState() (string, int, error) {
	out, err := systemdUser("show", "--property=ActiveState", "--property=MainPID", linuxUnitName)
	if err != nil {
		return "", 0, err
	}
	state, pid := "", 0
	for _, line := range strings.Split(out, "\n") {
		if value, ok := strings.CutPrefix(line, "ActiveState="); ok {
			state = value
		}
		if value, ok := strings.CutPrefix(line, "MainPID="); ok {
			pid, err = strconv.Atoi(value)
			if err != nil || pid < 0 {
				return "", 0, fmt.Errorf("invalid systemd MainPID: %q", value)
			}
		}
	}
	if state == "" {
		return "", 0, fmt.Errorf("systemd returned no service state")
	}
	return state, pid, nil
}

func linuxAgentBusy(pid int) error {
	threads, err := os.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, thread := range threads {
		children, err := os.ReadFile(fmt.Sprintf("/proc/%d/task/%s/children", pid, thread.Name()))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if len(strings.Fields(string(children))) > 0 {
			return fmt.Errorf("an agent task is active; finish it before updating")
		}
	}
	return nil
}

// Only the installer-created user unit is managed. Custom/system units retain
// their existing supervisor behavior and require a manual restart.
func prepareServiceUpdate(exe, config string, startStopped bool) (serviceUpdate, error) {
	service := serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}
	path, err := linuxUnitPath()
	if err != nil {
		return service, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return service, nil
	}
	if err != nil {
		return service, err
	}
	if !linuxUnitMatches(data, exe, config) {
		return service, nil
	}
	state, _, err := linuxServiceState()
	if err != nil {
		return service, err
	}
	if state != "active" && state != "activating" && !startStopped {
		return service, nil
	}
	verify := func() error {
		current, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !linuxUnitMatches(current, exe, config) {
			return fmt.Errorf("autostart configuration changed; update cancelled")
		}
		return nil
	}
	service.managed = true
	service.stop = func() error {
		if err := verify(); err != nil {
			return err
		}
		_, pid, err := linuxServiceState()
		if err != nil {
			return err
		}
		if pid > 0 {
			actual, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err == nil && filepath.Clean(actual) != exe {
				return fmt.Errorf("systemd unit is running a different executable; update cancelled")
			}
			if err := linuxAgentBusy(pid); err != nil {
				return err
			}
		}
		_, err = systemdUser("stop", linuxUnitName)
		return err
	}
	service.start = func() error {
		if err := verify(); err != nil {
			return err
		}
		if _, err := systemdUser("start", linuxUnitName); err != nil {
			return err
		}
		for i := 0; i < 25; i++ {
			state, pid, err := linuxServiceState()
			if err != nil {
				return err
			}
			if state == "active" && pid > 0 {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
		}
		return fmt.Errorf("systemd agent did not start")
	}
	return service, nil
}
