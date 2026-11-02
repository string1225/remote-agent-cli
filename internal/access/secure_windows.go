package access

import (
	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"golang.org/x/sys/windows"
	"os"
)

func currentSID() (string, bool, error) {
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\RemoteAgent`)
	if err == nil && gitauth.IsService(sid.String()) {
		return sid.String(), true, nil
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", false, err
	}
	return user.User.Sid.String(), false, nil
}
func secureDir(path string) error {
	if err := realDirectoryPath(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	sid, _, err := currentSID()
	if err != nil {
		return err
	}
	return gitauth.PrepareStore(path, sid)
}
func secureFile(path string) error {
	sid, service, err := currentSID()
	if err != nil {
		return err
	}
	acl := "D:P(A;;FA;;;SY)(A;;FA;;;BA)(A;;FA;;;" + sid + ")"
	if service {
		// Conditional owner rights suppress the shared LocalService owner's implicit
		// WRITE_DAC. Native OpenSSH accepts these conditional service-only grants.
		acl = "D:P(A;;FA;;;SY)(A;;FA;;;BA)(XA;;RC;;;OW;(Member_of {SID(" + sid + ")}))(XA;;FA;;;" + sid + ";(Member_of {SID(" + sid + ")}))"
	}
	sd, err := windows.SecurityDescriptorFromString(acl)
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
}
