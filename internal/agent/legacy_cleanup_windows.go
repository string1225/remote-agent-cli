package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// MSI calls this only after committing the new installation. Configuration,
// history, recordings, projects and other tools are outside the removal list.
func CleanupLegacyWindows(userSID string) error {
	p, err := ReadServiceProfile(serviceProfilePath())
	if errors.Is(err, os.ErrNotExist) {
		if userSID == "" {
			return nil
		}
		if _, err = windows.StringToSid(userSID); err != nil {
			return errors.New("invalid original user SID")
		}
		key, e := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProfileList\`+userSID, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if e != nil {
			return e
		}
		home, _, e := key.GetStringValue("ProfileImagePath")
		key.Close()
		if e != nil {
			return e
		}
		home, e = registry.ExpandString(home)
		if e != nil {
			return e
		}
		p = ServiceProfile{Config: filepath.Join(home, ".remote-agent", "config.json"), Environment: map[string]string{"USERPROFILE": home, "CODEX_HOME": filepath.Join(home, ".codex")}}
	} else if err != nil {
		return err
	}
	if err = p.Validate(); err != nil {
		return err
	}
	home, err := canonicalDir(p.Environment["USERPROFILE"])
	if err != nil {
		return err
	}
	bin := filepath.Join(home, ".remote-agent", "bin")
	actual, err := canonicalDir(bin)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, bin) {
		return errors.New("legacy binary directory is redirected; cleanup skipped")
	}
	legacy := filepath.Join(bin, "remote-agent.exe")
	// Migration checks the task action, configuration and owned process tree.
	if _, err = updatePowerShell(windowsTaskRemoveScript(legacy, p.Config)); err != nil {
		return err
	}
	return removeLegacyAgentFiles(bin)
}

func removeLegacyAgentFiles(bin string) error {
	// No recursive deletion, wildcards or execution of user-writable programs.
	for _, name := range []string{"remote-agent.exe", "remote-agent.exe.previous", "remote-agent.exe.sha256", "remote-agent.exe.update.lock", "remote-agent-before-directories.exe", "remote-agent-before-sayso.exe", "remote-agent-before-selfupdate.exe"} {
		path := filepath.Join(bin, name)
		st, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if !st.Mode().IsRegular() {
			return fmt.Errorf("legacy artifact is not a regular file: %s", name)
		}
		if err = os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}
