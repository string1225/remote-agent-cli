package servicehost

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

func trustedDirectory(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := windows.GetFileAttributes(p)
	if err != nil {
		return err
	}
	if attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return errors.New("refusing a reparse point in the machine runtime path")
	}
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return err
	}
	if owner.String() != "S-1-5-18" && owner.String() != "S-1-5-32-544" {
		return errors.New("machine runtime directory must be owned by SYSTEM or Administrators")
	}
	return nil
}

func protectDirectory(path, sddl string) error {
	if err := os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}
	if err := trustedDirectory(path); err != nil {
		return err
	}
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
}

func prepare() error {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("initial host installation requires an administrator")
	}
	if err := ownInstallation(); err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	for _, name := range []string{ServiceName, UpdaterName} {
		s, err := m.OpenService(name)
		if err != nil {
			return err
		}
		cfg, err := s.Config()
		if err != nil {
			s.Close()
			return err
		}
		cfg.SidType = windows.SERVICE_SID_TYPE_UNRESTRICTED
		if err = s.UpdateConfig(cfg); err == nil {
			err = s.SetRecoveryActions([]mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 5 * time.Second}}, 86400)
		}
		if err == nil {
			err = s.SetRecoveryActionsOnNonCrashFailures(true)
		}
		s.Close()
		if err != nil {
			return err
		}
	}
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\`+ServiceName)
	if err != nil {
		return err
	}
	readACL := "D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;GRGX;;;BU)(A;OICI;GRGX;;;" + sid.String() + ")"
	root, err := machineRoot()
	if err != nil {
		return err
	}
	if err = protectDirectory(root, readACL); err != nil {
		return err
	}
	for _, name := range []string{"runtime", "update-logs"} {
		if err = protectDirectory(filepath.Join(root, name), readACL); err != nil {
			return err
		}
	}
	healthACL := "D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;GRGX;;;BU)(A;OICI;GRGWGXSD;;;" + sid.String() + ")"
	return protectDirectory(filepath.Join(root, "health"), healthACL)
}

func cleanup() error {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("uninstall maintenance requires an administrator")
	}
	root, err := machineRoot()
	if err != nil {
		return err
	}
	if _, err = os.Stat(root); os.IsNotExist(err) {
		return nil
	}
	if err = trustedDirectory(root); err != nil {
		return err
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer r.Close()
	// Only these fixed, installation-owned children are removed, after MSI has
	// committed uninstall. service.json and every original user file are retained.
	for _, name := range []string{"runtime", "health"} {
		target := filepath.Join(root, name)
		if !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			return errors.New("invalid cleanup target")
		}
		if _, err = os.Stat(target); os.IsNotExist(err) {
			continue
		}
		if err = trustedDirectory(target); err != nil {
			return err
		}
		if err = r.RemoveAll(name); err != nil {
			return err
		}
	}
	return nil
}
