package servicehost

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"
	"unsafe"

	"github.com/Microsoft/go-winio"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"github.com/string1225/remote-agent-cli/internal/i18n"
	"golang.org/x/sys/windows"
)

func Run(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "update":
			return ClientUpdate(context.Background(), args[1:], i18n.Writer(os.Stdout))
		case "runtime-cleanup":
			return cleanup()
		case "update-service":
			if len(args) == 2 && args[1] == "run" {
				return runUpdater()
			}
			return errors.New("unsupported updater entry point")
		case "service":
			if len(args) == 2 && args[1] == "run" {
				return runService()
			}
			if len(args) == 2 && args[1] == "prepare" {
				return prepare()
			}
		case "version":
			fmt.Fprintf(i18n.Writer(os.Stdout), "remote-agent host %d (worker protocol %d)\n", clientupdate.HostVersion, clientupdate.WorkerProtocol)
		}
	}
	store, err := runtimeStore()
	if err != nil {
		return err
	}
	state, err := store.Load()
	if err != nil {
		return err
	}
	seed, err := seedPath()
	if err != nil {
		return err
	}
	exe, err := store.Binary(state.Current, seed)
	if err != nil && state.HasPrevious {
		exe, err = store.Binary(state.Previous, seed)
	}
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ClientUpdate(ctx context.Context, args []string, output io.Writer) error {
	request, check, err := ParseUpdate(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(output, "use update [--check | --rollback] [--force] [--server URL] [--config PATH]")
		return nil
	}
	if err != nil {
		return err
	}
	server, err := sourceServer(request.Server)
	if err != nil {
		return err
	}
	request.Server = server
	if check {
		store, err := runtimeStore()
		if err != nil {
			return err
		}
		release, err := clientupdate.Fetch(ctx, clientupdate.HTTPClient(), server, "windows", store.Arch, store.Keys)
		if err != nil {
			return err
		}
		state, err := store.Load()
		if err != nil {
			return err
		}
		if release.Sequence < state.HighestSequence || release.Sequence == state.HighestSequence && release.SHA256 != state.HighestSHA256 {
			return errors.New("published release is older than the accepted sequence floor")
		}
		fmt.Fprintf(output, "Host version: %d; published runtime: %s (release %d).\n", clientupdate.HostVersion, release.Version, release.Sequence)
		if state.Current != nil && state.Current.SHA256 == release.SHA256 {
			fmt.Fprintln(output, "Runtime is up to date.")
		} else {
			fmt.Fprintln(output, "A signed runtime update is available. Run remote-agent update; no MSI or administrator prompt is needed.")
		}
		return nil
	}
	return requestUpdate(ctx, request, output)
}

func requestUpdate(ctx context.Context, request Request, output io.Writer) error {
	conn, err := winio.DialPipeContext(ctx, PipeName)
	if err != nil {
		return errors.New("the fixed update service is unavailable; install the new host MSI once and check RemoteAgentUpdater")
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Minute))
	fd, ok := conn.(interface{ Fd() uintptr })
	if !ok {
		return errors.New("unsupported local updater pipe")
	}
	var pid uint32
	if err = windows.GetNamedPipeServerProcessId(windows.Handle(fd.Fd()), &pid); err != nil {
		return err
	}
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(scm)
	name, _ := windows.UTF16PtrFromString(UpdaterName)
	service, err := windows.OpenService(scm, name, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(service)
	var status windows.SERVICE_STATUS_PROCESS
	var needed uint32
	err = windows.QueryServiceStatusEx(service, windows.SC_STATUS_PROCESS_INFO, (*byte)(unsafe.Pointer(&status)), uint32(unsafe.Sizeof(status)), &needed)
	if err != nil || status.ProcessId != pid || pid == 0 {
		return errors.New("local updater pipe does not belong to the installed service")
	}
	if err = json.NewEncoder(conn).Encode(request); err != nil {
		return err
	}
	d := json.NewDecoder(io.LimitReader(conn, 1<<20))
	for {
		var event Event
		if err = d.Decode(&event); err != nil {
			return errors.New("update connection ended; consult the fixed updater log before retrying")
		}
		if event.Message != "" {
			fmt.Fprint(output, event.Message)
		}
		if event.Done {
			if event.Error != "" {
				return errors.New(event.Error)
			}
			return nil
		}
	}
}
