package codexsetup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func protect(path string) error {
	current, err := user.Current()
	if err != nil {
		return err
	}
	cmd := exec.Command("icacls.exe", path, "/inheritance:r", "/grant:r", "*"+current.Uid+":F", "*S-1-5-18:F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protect Codex credentials: %w", err)
	}
	return nil
}

// Atomic credential refresh must retain an already granted service SID, while
// keeping unrelated inherited/users permissions out of the new private file.
func preserveServiceAccess(original, replacement string) error {
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\RemoteAgent`)
	if err != nil {
		if errors.Is(err, windows.ERROR_NONE_MAPPED) {
			return nil
		}
		return err
	} // There is no installed service identity.
	return preserveServiceSIDAccess(original, replacement, sid)
}

func preserveServiceSIDAccess(original, replacement string, sid *windows.SID) error {
	mask, err := fileSIDAccess(original, sid)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || mask == 0 {
		return err
	}
	return grantFileSID(replacement, sid, mask)
}

func fileSIDAccess(path string, sid *windows.SID) (windows.ACCESS_MASK, error) {
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return 0, err
	}
	dacl, _, err := sd.DACL()
	if err != nil || dacl == nil {
		return 0, err
	}
	var mask, denied windows.ACCESS_MASK
	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, i, &ace); err != nil {
			return 0, err
		}
		if (ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE || ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE) && ace.Header.AceFlags&windows.INHERIT_ONLY_ACE == 0 && (*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(sid) {
			if ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE {
				mask |= ace.Mask
			}
			if ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE {
				denied |= ace.Mask
			}
		}
	}
	runtime.KeepAlive(sd)
	return mask &^ denied, nil
}

func grantFileSID(path string, sid *windows.SID, mask windows.ACCESS_MASK) error {
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	var pin runtime.Pinner
	pin.Pin(sid)
	defer pin.Unpin()
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{AccessPermissions: mask, AccessMode: windows.GRANT_ACCESS, Trustee: windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeType: windows.TRUSTEE_IS_UNKNOWN, TrusteeValue: windows.TrusteeValueFromSID(sid)}}}, dacl)
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
