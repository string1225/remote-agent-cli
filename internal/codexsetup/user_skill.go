package codexsetup

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const userSkillManifest = ".remote-agent-install.json"

type UserSkillStatus struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Exists  bool   `json:"exists"`
	Current bool   `json:"current"`
}

type UserSkillInstall struct {
	Path    string   `json:"path"`
	Version string   `json:"version"`
	Backups []string `json:"backups,omitempty"`
}

type skillManifest struct {
	Schema  int    `json:"schema"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

func userSkillPath(home string) string {
	return filepath.Join(home, ".agents", "skills", "remote-agent", "SKILL.md")
}

func skillSourceDigest() string {
	sum := sha256.Sum256(remoteAgentSkill)
	return hex.EncodeToString(sum[:])
}

func skillPathKey(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}

func UserRemoteAgentSkillStatus(home, version string) (UserSkillStatus, error) {
	status := UserSkillStatus{Path: userSkillPath(home)}
	if !filepath.IsAbs(home) {
		return status, errors.New("user skill home must be an absolute path")
	}
	data, err := os.ReadFile(status.Path)
	if os.IsNotExist(err) {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	status.Exists = true
	manifest, err := os.ReadFile(filepath.Join(filepath.Dir(status.Path), userSkillManifest))
	if os.IsNotExist(err) {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	var installed skillManifest
	if json.Unmarshal(manifest, &installed) != nil || installed.Schema != 1 {
		return status, errors.New("remote-agent skill version marker is invalid")
	}
	status.Version = installed.Version
	status.Current = installed.Version == version && installed.SHA256 == skillSourceDigest() && bytes.Equal(data, remoteAgentSkill)
	return status, nil
}

// SyncUserRemoteAgentSkill never chooses a profile implicitly. Callers pass the
// owning user's home and Codex home, which may differ from the service account.
// Legacy folders move outside skill discovery roots, retaining all resources.
func SyncUserRemoteAgentSkill(home, codexHome, version string) (UserSkillInstall, error) {
	result := UserSkillInstall{Path: userSkillPath(home), Version: version}
	if !filepath.IsAbs(home) || !filepath.IsAbs(codexHome) || version == "" {
		return result, errors.New("explicit absolute user/Codex homes and client version are required")
	}
	agentsHome := filepath.Join(home, ".agents")
	directory := filepath.Dir(result.Path)
	for _, path := range []string{agentsHome, filepath.Join(agentsHome, "skills"), directory} {
		if info, err := os.Lstat(path); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return result, fmt.Errorf("skill root must be a real directory: %s", path)
			}
		} else if !os.IsNotExist(err) {
			return result, err
		}
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return result, err
	}
	lockPath := filepath.Join(directory, ".remote-agent.sync.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, fmt.Errorf("cannot lock remote-agent skill synchronization: %w", err)
	}
	defer func() { lock.Close(); os.Remove(lockPath) }()
	type legacyMigration struct {
		root     string
		lockPath string
	}
	var legacyRoots []*legacyMigration
	defer func() {
		for _, legacy := range legacyRoots {
			os.Remove(legacy.lockPath)
		}
	}()
	seen := map[string]bool{skillPathKey(agentsHome): true}
	for _, root := range []string{codexHome, filepath.Join(home, ".codex")} {
		root = filepath.Clean(root)
		if seen[skillPathKey(root)] {
			continue
		}
		seen[skillPathKey(root)] = true
		for _, ancestor := range []string{root, filepath.Join(root, "skills")} {
			if info, err := os.Lstat(ancestor); err == nil {
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return result, errors.New("legacy skill root is linked or unsafe; it was preserved")
				}
			} else if !os.IsNotExist(err) {
				return result, err
			}
		}
		path := filepath.Join(root, "skills", "remote-agent")
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return result, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return result, errors.New("legacy remote-agent skill must be a real directory; it was preserved")
		}
		if info, err = os.Lstat(filepath.Join(path, "SKILL.md")); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return result, err
		} else if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return result, errors.New("legacy skill source is not a regular file; it was preserved")
		}
		legacyLockPath := filepath.Join(path, ".install.lock")
		legacyLock, err := os.OpenFile(legacyLockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return result, fmt.Errorf("cannot lock legacy remote-agent skill; source retained: %w", err)
		}
		// The file's exclusive existence is the legacy installer's lock. Close
		// the handle so Windows can rename its directory while keeping it held.
		if err = legacyLock.Close(); err != nil {
			os.Remove(legacyLockPath)
			return result, err
		}
		legacyRoots = append(legacyRoots, &legacyMigration{root, legacyLockPath})
	}
	if _, err = InstallRemoteAgentSkill(agentsHome); err != nil {
		return result, err
	}
	for _, legacy := range legacyRoots {
		root := legacy.root
		backupRoot := filepath.Join(root, "remote-agent-skill-backups")
		if info, err := os.Lstat(backupRoot); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return result, errors.New("legacy skill backup directory is unsafe; source retained")
			}
		} else if !os.IsNotExist(err) {
			return result, err
		}
		if err = os.MkdirAll(backupRoot, 0700); err != nil {
			return result, err
		}
		backup := filepath.Join(backupRoot, "remote-agent-"+time.Now().UTC().Format("20060102T150405.000000000"))
		if err = os.Rename(filepath.Join(root, "skills", "remote-agent"), backup); err != nil {
			return result, fmt.Errorf("shared skill installed, but legacy migration failed; source retained: %w", err)
		}
		legacy.lockPath = filepath.Join(backup, ".install.lock")
		result.Backups = append(result.Backups, backup)
	}
	path := filepath.Join(filepath.Dir(result.Path), userSkillManifest)
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return result, errors.New("skill version marker is not a regular file")
		}
	} else if !os.IsNotExist(err) {
		return result, err
	}
	manifest, _ := json.MarshalIndent(skillManifest{1, version, skillSourceDigest()}, "", "  ")
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return result, err
	}
	if !bytes.Equal(old, manifest) {
		if err = writeSkillFile(path, manifest, true); err != nil {
			return result, err
		}
	}
	return result, nil
}
