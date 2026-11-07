package agent

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// PrepareUserSkillServiceAccess runs as the owning user after skill installation.
// Only this skill's directory is writable by the installed service; sibling
// user skills and their parent directories keep their existing permissions.
func PrepareUserSkillServiceAccess(home string) error {
	if _, err := installedServiceExecutable(); errors.Is(err, registry.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\`+WindowsServiceName)
	if err != nil {
		return err
	}
	return grantUserSkillServiceAccess(home, sid.String())
}

func userSkillServiceDirectory(home string, create bool) (string, error) {
	if !filepath.IsAbs(home) {
		return "", errors.New("user skill home must be absolute")
	}
	path := filepath.Join(home, ".agents", "skills", "remote-agent")
	for _, directory := range []string{filepath.Join(home, ".agents"), filepath.Dir(path), path} {
		info, err := os.Lstat(directory)
		if os.IsNotExist(err) && create {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("user skill service access requires real directories")
		}
	}
	if create {
		if err := os.MkdirAll(path, 0700); err != nil {
			return "", err
		}
	}
	return path, nil
}

func grantUserSkillServiceAccess(home, sid string) error {
	path, err := userSkillServiceDirectory(home, false)
	if err != nil {
		return err
	}
	return serviceACL(path, sid, "(OI)(CI)M")
}
