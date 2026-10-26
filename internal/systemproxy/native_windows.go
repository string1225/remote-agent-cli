package systemproxy

import (
	"fmt"
	"os/user"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var wininet = windows.NewLazySystemDLL("wininet.dll")
var queryOption = wininet.NewProc("InternetQueryOptionW")
var setOption = wininet.NewProc("InternetSetOptionW")
var globalFree = windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalFree")

type connectionOption struct {
	Option uint32
	Value  uintptr
}
type connectionOptions struct {
	Size       uint32
	Connection *uint16
	Count      uint32
	Error      uint32
	Options    *connectionOption
}
type nativeBackend struct{}

func Native() (Backend, bool) { return nativeBackend{}, true }

func (nativeBackend) Read() (Settings, error) {
	options := []connectionOption{{Option: 10}, {Option: 2}, {Option: 3}, {Option: 4}}
	list := connectionOptions{Count: uint32(len(options)), Options: &options[0]}
	list.Size = uint32(unsafe.Sizeof(list))
	size := list.Size
	ok, _, err := queryOption.Call(0, 75, uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&size)))
	if ok == 0 {
		return Settings{}, fmt.Errorf("query Windows user proxy: %w", err)
	}
	read := func(index int) string {
		pointer := *(*unsafe.Pointer)(unsafe.Pointer(&options[index].Value))
		if pointer == nil {
			return ""
		}
		defer globalFree.Call(uintptr(pointer))
		return windows.UTF16PtrToString((*uint16)(pointer))
	}
	state := Settings{Flags: uint32(options[0].Value), Server: read(1), Bypass: read(2), AutoConfigURL: read(3)}
	runtime.KeepAlive(options)
	return state, nil
}

func (nativeBackend) Write(state Settings) error {
	server, err := windows.UTF16PtrFromString(state.Server)
	if err != nil {
		return err
	}
	bypass, err := windows.UTF16PtrFromString(state.Bypass)
	if err != nil {
		return err
	}
	auto, err := windows.UTF16PtrFromString(state.AutoConfigURL)
	if err != nil {
		return err
	}
	options := []connectionOption{{Option: 1, Value: uintptr(state.Flags)}, {Option: 2, Value: uintptr(unsafe.Pointer(server))}, {Option: 3, Value: uintptr(unsafe.Pointer(bypass))}, {Option: 4, Value: uintptr(unsafe.Pointer(auto))}}
	list := connectionOptions{Count: uint32(len(options)), Options: &options[0]}
	list.Size = uint32(unsafe.Sizeof(list))
	ok, _, err := setOption.Call(0, 75, uintptr(unsafe.Pointer(&list)), uintptr(list.Size))
	runtime.KeepAlive(options)
	runtime.KeepAlive(server)
	runtime.KeepAlive(bypass)
	runtime.KeepAlive(auto)
	if ok == 0 {
		return fmt.Errorf("set Windows user proxy: %w", err)
	}
	for _, option := range []uintptr{39, 37} {
		if ok, _, err := setOption.Call(0, option, 0, 0); ok == 0 {
			return fmt.Errorf("notify Windows proxy change: %w", err)
		}
	}
	return nil
}

func protectFile(path string) error {
	owner, err := user.Current()
	if err != nil {
		return err
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;SY)(A;;FA;;;" + owner.Uid + ")")
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
}
