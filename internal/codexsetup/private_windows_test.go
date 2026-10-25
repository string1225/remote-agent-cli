package codexsetup

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestReplacementPreservesOnlyPreviouslyGrantedServiceSID(t *testing.T) {
	root := t.TempDir()
	old, replacement := filepath.Join(root, "old.toml"), filepath.Join(root, "replacement.toml")
	for _, path := range []string{old, replacement} {
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := protect(path); err != nil {
			t.Fatal(err)
		}
	}
	sid, err := windows.StringToSid("S-1-5-21-1-2-3-9999")
	if err != nil {
		t.Fatal(err)
	}
	if err = grantFileSID(old, sid, windows.FILE_GENERIC_READ); err != nil {
		t.Fatal(err)
	}
	if err = preserveServiceSIDAccess(old, replacement, sid); err != nil {
		t.Fatal(err)
	}
	mask, err := fileSIDAccess(replacement, sid)
	if err != nil || mask != windows.FILE_GENERIC_READ {
		t.Fatal("service read grant lost or broadened", mask, err)
	}
	absent, _ := windows.StringToSid("S-1-5-21-1-2-3-9998")
	if err = preserveServiceSIDAccess(old, replacement, absent); err != nil {
		t.Fatal(err)
	}
	if mask, err = fileSIDAccess(replacement, absent); err != nil || mask != 0 {
		t.Fatal("new service grant introduced", mask, err)
	}
}
