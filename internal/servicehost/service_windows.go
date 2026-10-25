package servicehost

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

const pauseControl svc.Cmd = 128
const forceControl svc.Cmd = 129
const resumeControl svc.Cmd = 130

type serviceHandler struct{}

func (h serviceHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	status := svc.Status{State: svc.StartPending, WaitHint: 90000}
	statuses <- status
	log, err := hostLog()
	if err != nil {
		return true, 1
	}
	defer log.Close()
	store, err := runtimeStore()
	if err != nil {
		return true, 1
	}
	state, err := store.Load()
	if err != nil {
		writeHealth(clientupdate.Health{Error: "invalid protected runtime state"})
		return true, 1
	}
	seed, err := seedPath()
	if err != nil {
		return true, 1
	}
	selected := state.Current
	w, err := startSelected(store, selected, seed, log)
	fallback := false
	if err != nil && state.HasPrevious {
		selected = state.Previous
		w, err = startSelected(store, selected, seed, log)
		fallback = true
	}
	if err != nil {
		writeHealth(clientupdate.Health{Error: "candidate and previous runtime could not prepare"})
		return true, 1
	}
	health := clientupdate.Health{Digest: "seed", Version: w.version, Ready: true, Generation: uint64(time.Now().UnixNano()), HostPID: os.Getpid()}
	if selected != nil {
		health.Digest = selected.SHA256
	}
	if fallback {
		health.Error = "candidate preparation failed; previous approved runtime selected"
	}
	if err = writeHealth(health); err != nil {
		w.kill()
		return true, 1
	}
	status = svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	statuses <- status
	defer func() { health.Ready = false; _ = writeHealth(health) }()
	for {
		select {
		case <-w.done:
			return true, 1
		case req := <-requests:
			switch req.Cmd {
			case svc.Interrogate:
				statuses <- status
			case svc.Stop, svc.Shutdown:
				statuses <- svc.Status{State: svc.StopPending, WaitHint: 30000}
				if err = w.stop(); err != nil {
					return true, 1
				}
				return false, 0
			case pauseControl, forceControl:
				err = w.call("pause", req.Cmd == forceControl)
				health.Generation++
				health.Paused = err == nil
				health.Error = ""
				if err != nil {
					health.Error = err.Error()
				}
				_ = writeHealth(health)
				statuses <- status
			case resumeControl:
				err = w.call("resume", false)
				health.Generation++
				health.Paused = false
				health.Error = ""
				if err != nil {
					health.Error = err.Error()
				}
				_ = writeHealth(health)
				statuses <- status
			default:
				statuses <- status
			}
		}
	}
}

func startSelected(store clientupdate.Store, release *clientupdate.Release, seed string, log io.Writer) (*worker, error) {
	exe, err := store.Binary(release, seed)
	if err != nil {
		return nil, err
	}
	w, err := launchWorker(exe, log)
	if err != nil {
		return nil, err
	}
	if err = w.call("activate", false); err != nil {
		w.kill()
		return nil, err
	}
	return w, nil
}

func writeHealth(health clientupdate.Health) error {
	root, err := machineRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "health")
	file, err := os.CreateTemp(dir, ".health-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(health)
	if err == nil {
		err = file.Sync()
	}
	err = errors.Join(err, file.Close())
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, "status.json"))
}

func runService() error {
	if err := ownInstallation(); err != nil {
		return err
	}
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if !isService {
		return errors.New("service run requires the Windows Service Control Manager")
	}
	// Keep the handle until process exit. Windows kills the complete worker/CLI
	// tree if the fixed SCM host crashes, before recovery starts a replacement.
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		windows.CloseHandle(job)
		return err
	}
	process, _ := windows.GetCurrentProcess()
	if err = windows.AssignProcessToJobObject(job, process); err != nil {
		windows.CloseHandle(job)
		return err
	}
	return svc.Run(ServiceName, serviceHandler{})
}
