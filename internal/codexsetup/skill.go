package codexsetup

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

//go:embed skills/remote-agent/SKILL.md
var remoteAgentSkill []byte

// InstallRemoteAgentSkill installs instructions only, without any machine
// identity or credentials. The caller supplies the intended Codex home so tests
// and service/user setup never implicitly write another profile's configuration.
func InstallRemoteAgentSkill(home string) (string, error) {
	if home == "" {
		return "", errors.New("Codex home is required for skill installation")
	}
	directory := filepath.Join(home, "skills", "remote-agent")
	for _, path := range []string{filepath.Join(home, "skills"), directory} {
		if info, err := os.Lstat(path); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return "", fmt.Errorf("skill directory must be a real directory: %s", path)
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	lockPath := filepath.Join(directory, ".install.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("cannot lock remote-agent skill installation: %w", err)
	}
	defer func() { lock.Close(); os.Remove(lockPath) }()
	path := filepath.Join(directory, "SKILL.md")
	if info, err := os.Lstat(path); err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return "", errors.New("existing remote-agent skill is not a regular file; it was not changed")
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	before, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if bytes.Equal(before, remoteAgentSkill) {
		return path, nil
	}
	if len(before) > 0 {
		backup := path + ".backup-" + time.Now().UTC().Format("20060102T150405.000000000")
		if err := writeSkillFile(backup, before, false); err != nil {
			return "", err
		}
	}
	latest, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if !bytes.Equal(before, latest) {
		return "", errors.New("remote-agent skill changed during installation; original file retained")
	}
	if err := writeSkillFile(path, remoteAgentSkill, true); err != nil {
		return "", err
	}
	return path, nil
}

func writeSkillFile(path string, data []byte, replace bool) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".skill-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	if !replace {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return errors.New("skill backup already exists")
		}
	}
	return os.Rename(f.Name(), path)
}
