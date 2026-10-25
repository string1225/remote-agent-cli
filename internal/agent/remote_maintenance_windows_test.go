package agent

import (
	"os/user"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestRemoteMaintenanceStateRetainsInheritedAccess(t *testing.T) {
	root := t.TempDir()
	owner, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	// A fictitious owner grant verifies ACL inheritance without changing a real
	// account, service or machine binding.
	other, err := windows.StringToSid("S-1-5-21-1-2-3-9876")
	if err != nil {
		t.Fatal(err)
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;" + owner.Uid + ")(A;OICI;FA;;;" + other.String() + ")")
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err = windows.SetNamedSecurityInfo(root, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "maintenance.json")
	m := &maintenanceState{agentID: "fixture", path: path}
	op, _, err := m.start(maintenanceInput{Kind: "codex", RequestID: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.change(op.ID, func(v *MaintenanceOperation) { v.Status = "succeeded" }); err != nil {
		t.Fatal(err)
	}
	sd, err = windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err = sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	for i := uint32(0); i < uint32(dacl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, i, &ace); err != nil {
			t.Fatal(err)
		}
		if ace.Header.AceType == windows.ACCESS_ALLOWED_ACE_TYPE && ace.Header.AceFlags&windows.INHERITED_ACE != 0 && (*windows.SID)(unsafe.Pointer(&ace.SidStart)).Equals(other) && ace.Mask&(windows.DELETE|windows.GENERIC_ALL) != 0 {
			return
		}
	}
	t.Fatal("service-written maintenance state lost the enrolled owner's inherited access")
}
