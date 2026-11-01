package gitauth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func PipeName(config string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(config))))
	return `\\.\pipe\RemoteAgent.GitAuth.` + hex.EncodeToString(hash[:16])
}

type identity struct {
	SID               string
	Elevated, Service bool
}

func tokenIdentity(token windows.Token, serviceSID string) (identity, error) {
	user, err := token.GetTokenUser()
	if err != nil {
		return identity{}, err
	}
	i := identity{SID: user.User.Sid.String(), Elevated: token.IsElevated()}
	if i.SID == "S-1-5-19" {
		groups, err := token.GetTokenGroups()
		if err != nil {
			return identity{}, err
		}
		for _, group := range groups.AllGroups() {
			if group.Sid.String() == serviceSID && group.Attributes&windows.SE_GROUP_ENABLED != 0 && group.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
				i.Service = true
			}
		}
	}
	return i, nil
}

func IsService(serviceSID string) bool {
	i, err := tokenIdentity(windows.GetCurrentProcessToken(), serviceSID)
	return err == nil && i.Service
}

var impersonatePipeClient = windows.NewLazySystemDLL("advapi32.dll").NewProc("ImpersonateNamedPipeClient")

func clientIdentity(conn net.Conn, serviceSID string) (identity, error) {
	fd, ok := conn.(interface{ Fd() uintptr })
	if !ok {
		return identity{}, errors.New("unsupported local Git authorization pipe")
	}
	runtime.LockOSThread()
	result, _, callErr := impersonatePipeClient.Call(fd.Fd())
	if result == 0 {
		runtime.UnlockOSThread()
		return identity{}, callErr
	}
	defer func() {
		if err := windows.RevertToSelf(); err != nil {
			// Never return an impersonating OS thread to Go's scheduler.
			runtime.Goexit()
		}
		runtime.UnlockOSThread()
	}()
	var token windows.Token
	if err := windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_QUERY, true, &token); err != nil {
		return identity{}, err
	}
	defer token.Close()
	return tokenIdentity(token, serviceSID)
}

func serviceWorker(conn net.Conn, servicePID uint32) error {
	fd, ok := conn.(interface{ Fd() uintptr })
	if !ok || servicePID == 0 {
		return errors.New("invalid Git authorization service process")
	}
	var pid uint32
	if err := windows.GetNamedPipeServerProcessId(windows.Handle(fd.Fd()), &pid); err != nil || pid == 0 {
		return errors.New("cannot verify Git authorization server process")
	}
	if pid == servicePID {
		return nil // Legacy service executes the core in its own process.
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ProcessID == pid && entry.ParentProcessID == servicePID && strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), "remote-agent-core.exe") {
			return nil
		}
	}
	return errors.New("Git authorization pipe does not belong to the installed RemoteAgent service or its core worker")
}

func allowClient(i identity, ownerSID string) error {
	if i.SID == ownerSID || i.SID == "S-1-5-18" || i.Elevated {
		return nil
	}
	return errors.New("only the configured device user or an administrator may manage service Git authorization")
}

func Listen(parent context.Context, name, ownerSID, serviceSID string, manager Manager) (func(), error) {
	for _, sid := range []string{ownerSID, serviceSID} {
		if _, err := windows.StringToSid(sid); err != nil {
			return nil, errors.New("invalid Git authorization owner identity")
		}
	}
	// go-winio rejects remote clients and exclusively creates the first instance.
	listener, err := winio.ListenPipe(name, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;RC;;;OW)(A;;GA;;;" + ownerSID + ")(A;;GA;;;" + serviceSID + ")", InputBufferSize: 4096, OutputBufferSize: 4096})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parent)
	server := &Server{Manager: manager, Authorize: func(conn net.Conn) error {
		i, err := clientIdentity(conn, serviceSID)
		if err != nil {
			return errors.New("cannot verify Git authorization caller identity")
		}
		return allowClient(i, ownerSID)
	}}
	slots := make(chan struct{}, 4)
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			select {
			case slots <- struct{}{}:
				workers.Add(1)
				go func() { defer workers.Done(); defer func() { <-slots }(); server.Serve(ctx, conn) }()
			default:
				conn.Close()
			}
		}
	}()
	var once sync.Once
	stop := func() { once.Do(func() { cancel(); listener.Close(); workers.Wait() }) }
	go func() { <-ctx.Done(); listener.Close() }()
	return stop, nil
}

func Call(ctx context.Context, name string, servicePID uint32, request Request, output io.Writer) error {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	conn, err := winio.DialPipeAccessImpLevel(connectCtx, name, windows.GENERIC_READ|windows.GENERIC_WRITE, winio.PipeImpLevelIdentification)
	cancel()
	if err != nil {
		return errors.New("service Git authorization is unavailable; start or update RemoteAgent, then retry remote-agent git login")
	}
	defer conn.Close()
	if err := serviceWorker(conn, servicePID); err != nil {
		return err
	}
	return Exchange(ctx, conn, request, output)
}
