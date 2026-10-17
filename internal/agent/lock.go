package agent

import (
	"errors"
	"fmt"
)

var errLockHeld = errors.New("file lock is held")
var ErrAgentRunning = errors.New("an agent is already running with this configuration")

func lockUpdate(path string) (func(), error) {
	unlock, err := lockFile(path)
	if err != nil {
		return nil, fmt.Errorf("another update is running or the update lock is unavailable: %w", err)
	}
	return unlock, nil
}

// LockRun prevents watchdog launches or an orphaned launcher from creating two
// agents that continuously replace each other's connection to the control server.
// The OS releases the lock even when the process crashes or is forcefully stopped.
func LockRun(configPath string) (func(), error) {
	unlock, err := lockFile(configPath + ".run.lock")
	if errors.Is(err, errLockHeld) {
		return nil, ErrAgentRunning
	}
	return unlock, err
}
