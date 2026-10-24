package personalgateway

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var registryKernel = syscall.NewLazyDLL("kernel32.dll")
var registryLock = registryKernel.NewProc("LockFileEx")
var registryUnlock = registryKernel.NewProc("UnlockFileEx")

func lockRegistryFile(path string) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	overlapped := &syscall.Overlapped{}
	result, _, _ := registryLock.Call(file.Fd(), 3, 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
	if result == 0 {
		file.Close()
		return nil, errors.New("device registry is busy; retry")
	}
	return func() { registryUnlock.Call(file.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(overlapped))); file.Close() }, nil
}
