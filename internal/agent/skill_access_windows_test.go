package agent

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestUserSkillServiceAccessIsLimitedAndInherited(t *testing.T) {
	home := t.TempDir()
	directory, err := userSkillServiceDirectory(home, true)
	if err != nil {
		t.Fatal(err)
	}
	// Only this temporary profile is changed; no real service is opened.
	sid, err := windows.StringToSid("S-1-5-19") // LocalService, a built-in identity.
	if err != nil {
		t.Fatal(err)
	}
	if err = grantUserSkillServiceAccess(home, sid.String()); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(directory, "SKILL.md")
	if err = os.WriteFile(child, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Dir(directory), directory, child} {
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
			if err = windows.GetAce(dacl, i, &ace); err != nil {
				t.Fatal(err)
			}
			if ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE && (*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(sid) && ace.Mask&windows.FILE_WRITE_DATA != 0 {
				found = true
				if path == child && ace.Header.AceFlags&windows.INHERITED_ACE == 0 {
					t.Fatal("skill files must inherit service write access")
				}
			}
		}
		if found != (path != filepath.Dir(directory)) {
			t.Fatalf("unexpected service access on %s: %t", path, found)
		}
	}
}
