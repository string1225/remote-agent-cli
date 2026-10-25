package agent

import (
	"os/user"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestGitHubProxySettingsInheritServiceDirectoryAccess(t *testing.T) {
	root := t.TempDir()
	owner, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	// A fictitious SID exercises file ACL inheritance without installing,
	// opening, or changing any real service or account.
	serviceSID, err := windows.StringToSid("S-1-5-21-1-2-3-9876")
	if err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;" + owner.Uid + ")(A;OICI;GR;;;" + serviceSID.String() + ")")
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(root, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Server: "https://relay.example.test", ID: "fixture", Credential: "test-credential", Path: filepath.Join(root, "config.json")}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := SetGitHubProxy(cfg.Path, DefaultGitHubProxyListen); err != nil {
		t.Fatal(err)
	}
	backup, err := SetGitHubProxy(cfg.Path, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{cfg.Path + ".github-proxy.json", backup} {
		sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		dacl, _, err := sd.DACL()
		if err != nil || dacl == nil {
			t.Fatal("missing ACL", err)
		}
		found := false
		for i := uint32(0); i < uint32(dacl.AceCount); i++ {
			var ace *windows.ACCESS_ALLOWED_ACE
			if err := windows.GetAce(dacl, i, &ace); err != nil {
				t.Fatal(err)
			}
			if ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE && ace.Header.AceFlags&windows.INHERITED_ACE != 0 && (*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(serviceSID) && ace.Mask&(windows.FILE_READ_DATA|windows.GENERIC_READ) != 0 {
				found = true
			}
		}
		if !found {
			t.Fatalf("service directory read grant lost for %s", path)
		}
	}
}
