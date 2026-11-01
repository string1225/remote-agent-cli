package gitauth

import (
	"errors"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

func PrepareStore(path, serviceSID string) error {
	if !filepath.IsAbs(path) {
		return errors.New("Git credential store must be an absolute path")
	}
	if _, err := windows.StringToSid(serviceSID); err != nil {
		return err
	}
	// DPAPI CurrentUser belongs to LocalService. OWNER RIGHTS also prevents a
	// different service sharing that user SID from rewriting the directory DACL.
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;RC;;;OW)(A;OICI;FA;;;" + serviceSID + ")")
	if err != nil {
		return err
	}
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	if err = windows.CreateDirectory(ptr, &sa); err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	attributes, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return err
	}
	if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || attributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		return errors.New("Git credential store must be a real directory, not a reparse point")
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	err = windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
	runtime.KeepAlive(sd)
	return err
}
