package servicehost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

type broker struct {
	mu  sync.Mutex
	log io.Writer
}
type eventWriter struct {
	e   *json.Encoder
	log io.Writer
}

func (w eventWriter) Write(p []byte) (int, error) {
	if w.log != nil {
		fmt.Fprintf(w.log, "%s %s", time.Now().Format(time.RFC3339), p)
	}
	if w.e == nil {
		return len(p), nil
	}
	err := w.e.Encode(Event{Message: string(p)})
	return len(p), err
}

func (b *broker) execute(ctx context.Context, request Request, output io.Writer) error {
	if request.Action != "apply" && request.Action != "rollback" {
		return errors.New("unsupported updater operation")
	}
	if !b.mu.TryLock() {
		return errors.New("another runtime update is already in progress")
	}
	defer b.mu.Unlock()
	store, err := runtimeStore()
	if err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(store.Path, "update.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	overlap := &windows.Overlapped{}
	if err = windows.LockFileEx(windows.Handle(lock.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, overlap); err != nil {
		return errors.New("runtime update lock is held")
	}
	defer windows.UnlockFileEx(windows.Handle(lock.Fd()), 0, 1, 0, overlap)
	control, err := openController()
	if err != nil {
		return err
	}
	defer control.s.Close()
	engine := clientupdate.Engine{Store: store, Service: control, Output: output}
	if err = engine.Reconcile(); err != nil {
		return err
	}
	if request.Action == "rollback" {
		return engine.Rollback(request.Force)
	}
	server, err := sourceServer(request.Server)
	if err != nil {
		return err
	}
	release, err := clientupdate.Fetch(ctx, clientupdate.HTTPClient(), server, "windows", store.Arch, store.Keys)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "Verified signed runtime %s (release %d).\n", release.Version, release.Sequence)
	return engine.Apply(ctx, release, func(w io.Writer) error {
		return clientupdate.Download(ctx, clientupdate.HTTPClient(), server, release, w)
	}, request.Force)
}

func authorizedClient(conn net.Conn) (bool, error) {
	fd, ok := conn.(interface{ Fd() uintptr })
	if !ok {
		return false, errors.New("unsupported local pipe")
	}
	var pid uint32
	if err := windows.GetNamedPipeClientProcessId(windows.Handle(fd.Fd()), &pid); err != nil || pid == 0 {
		return false, errors.New("remote pipe clients are denied")
	}
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(process)
	var token windows.Token
	if err = windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token); err != nil {
		return false, err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		return false, err
	}
	p, err := readProfile()
	if os.IsNotExist(err) {
		return false, nil
	} // Fresh, unbound installation has no user work to interrupt.
	if err != nil {
		return false, err
	}
	if user.User.Sid.String() == p.OwnerSID || token.IsElevated() {
		return false, nil
	}
	if user.User.Sid.String() == "S-1-5-19" {
		serviceSID, _, _, e := windows.LookupSID("", `NT SERVICE\`+ServiceName)
		groups, e2 := token.GetTokenGroups()
		if e == nil && e2 == nil {
			for _, group := range groups.AllGroups() {
				if group.Sid.Equals(serviceSID) && group.Attributes&windows.SE_GROUP_ENABLED != 0 && group.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
					return true, nil
				}
			}
		}
	}
	return false, errors.New("only the configured device user or an administrator may update this installation")
}

func (b *broker) serve(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Minute))
	e := json.NewEncoder(conn)
	serviceCaller, err := authorizedClient(conn)
	if err != nil {
		_ = e.Encode(Event{Done: true, Error: err.Error()})
		return
	}
	var req Request
	d := json.NewDecoder(io.LimitReader(conn, 8193))
	d.DisallowUnknownFields()
	if err := d.Decode(&req); err != nil {
		_ = e.Encode(Event{Done: true, Error: "invalid update request"})
		return
	}
	// Service-initiated remote updates use the protected enrolled source and
	// never force-cancel work or request rollback/system maintenance.
	if serviceCaller && (req.Action != "apply" || req.Server != "" || req.Force) {
		_ = e.Encode(Event{Done: true, Error: "remote service may only update from its protected enrolled source without force"})
		return
	}
	// Completing or rolling back the transaction is independent of the CLI
	// connection; closing a terminal cannot leave an uncommitted service switch.
	err = b.execute(ctx, req, eventWriter{e: e, log: b.log})
	result := Event{Done: true}
	if err != nil {
		result.Error = err.Error()
		fmt.Fprintf(b.log, "%s update failed: %s\n", time.Now().Format(time.RFC3339), err)
	}
	_ = e.Encode(result)
}

type updaterHandler struct{}

func (updaterHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	statuses <- svc.Status{State: svc.StartPending, WaitHint: 10000}
	root, err := machineRoot()
	if err != nil {
		return true, 1
	}
	logPath := filepath.Join(root, "update-logs", "updater.log")
	if st, e := os.Stat(logPath); e == nil && st.Size() > 10<<20 {
		_ = os.Rename(logPath, logPath+".old")
	}
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return true, 1
	}
	defer log.Close()
	listener, err := winio.ListenPipe(PipeName, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GRGW;;;AU)", InputBufferSize: 8192, OutputBufferSize: 8192})
	if err != nil {
		fmt.Fprintln(log, "updater pipe unavailable")
		return true, 1
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := &broker{log: log}
	var workers sync.WaitGroup
	acceptDone := make(chan struct{})
	slots := make(chan struct{}, 16)
	go func() {
		defer close(acceptDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			select {
			case slots <- struct{}{}:
			default:
				conn.Close()
				continue
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer func() { <-slots }()
				done := make(chan struct{})
				defer close(done)
				go func() {
					select {
					case <-ctx.Done():
						conn.Close()
					case <-done:
					}
				}()
				b.serve(ctx, conn)
			}()
		}
	}()
	// Updates are idle-only in the background. A signed release is checked at
	// startup and every six hours; manual update uses exactly the same engine.
	autoDone := make(chan struct{})
	go func() {
		defer close(autoDone)
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := b.execute(ctx, Request{Action: "apply"}, eventWriter{log: log}); err != nil && ctx.Err() == nil {
					fmt.Fprintf(log, "%s automatic update deferred: %s\n", time.Now().Format(time.RFC3339), err)
				}
				timer.Reset(6 * time.Hour)
			}
		}
	}()
	status := svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	statuses <- status
	for req := range requests {
		switch req.Cmd {
		case svc.Interrogate:
			statuses <- status
		case svc.Stop, svc.Shutdown:
			statuses <- svc.Status{State: svc.StopPending, WaitHint: 120000}
			listener.Close()
			cancel()
			<-acceptDone
			finished := make(chan struct{})
			go func() { workers.Wait(); <-autoDone; close(finished) }()
			select {
			case <-finished:
				return false, 0
			case <-time.After(110 * time.Second):
				return true, 1
			}
		}
	}
	return false, 0
}
func runUpdater() error {
	if err := ownInstallation(); err != nil {
		return err
	}
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if !isService {
		return errors.New("updater run requires Windows SCM")
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return err
	}
	if user.User.Sid.String() != "S-1-5-18" {
		return errors.New("updater requires the MSI-installed SYSTEM service identity")
	}
	return svc.Run(UpdaterName, updaterHandler{})
}
