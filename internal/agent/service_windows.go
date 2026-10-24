package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/string1225/remote-agent-cli/internal/i18n"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const WindowsServiceName = "RemoteAgent"

func serviceProfilePath() string {
	return filepath.Join(os.Getenv("ProgramData"), "RemoteAgent", "service.json")
}

func installedServiceExecutable() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\RemoteAgent`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer key.Close()
	dir, _, err := key.GetStringValue("InstallDir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dir) {
		return "", errors.New("invalid MSI installation directory")
	}
	return filepath.Join(dir, "remote-agent.exe"), nil
}

func msiManagedExecutable(exe string) bool {
	installed, err := installedServiceExecutable()
	return err == nil && strings.EqualFold(filepath.Clean(exe), filepath.Clean(installed))
}

func openWindowsService(access uint32) (*mgr.Service, error) {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return nil, err
	}
	defer windows.CloseServiceHandle(scm)
	name, _ := windows.UTF16PtrFromString(WindowsServiceName)
	handle, err := windows.OpenService(scm, name, access)
	if err != nil {
		return nil, fmt.Errorf("open Windows Service (configuration/control requires an administrator): %w", err)
	}
	return &mgr.Service{Name: WindowsServiceName, Handle: handle}, nil
}

func checkServiceOwner(s *mgr.Service) (mgr.Config, error) {
	c, err := s.Config()
	if err != nil {
		return c, err
	}
	installed, err := installedServiceExecutable()
	if err != nil {
		return c, err
	}
	args, err := windows.DecomposeCommandLine(c.BinaryPathName)
	if err != nil || len(args) != 3 || !strings.EqualFold(args[0], installed) || args[1] != "service" || args[2] != "run" || !strings.EqualFold(c.ServiceStartName, `NT AUTHORITY\LocalService`) {
		return c, errors.New("RemoteAgent service belongs to another installation or identity; it was not changed")
	}
	return c, nil
}

func waitWindowsService(s *mgr.Service, state svc.State) error {
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Query()
		if err != nil {
			return err
		}
		if status.State == state {
			return nil
		}
		if state == svc.Running && status.State == svc.Stopped {
			return fmt.Errorf("Windows Service stopped during startup (exit code %d); inspect agent.log", status.Win32ExitCode)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("timed out waiting for Windows Service")
}

func stopWindowsService(s *mgr.Service) error {
	status, err := s.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Stopped {
		return nil
	}
	if status.State != svc.StopPending {
		if _, err = s.Control(svc.Stop); err != nil {
			return err
		}
	}
	return waitWindowsService(s, svc.Stopped)
}

func startWindowsService(s *mgr.Service) error {
	status, err := s.Query()
	if err != nil {
		return err
	}
	if status.State == svc.Running {
		return nil
	}
	if status.State != svc.StartPending {
		if err = s.Start(); err != nil {
			return err
		}
	}
	return waitWindowsService(s, svc.Running)
}

func WindowsServiceStatus() (any, error) {
	s, err := openWindowsService(windows.SERVICE_QUERY_CONFIG | windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	c, err := checkServiceOwner(s)
	if err != nil {
		return nil, err
	}
	status, err := s.Query()
	if err != nil {
		return nil, err
	}
	p, profileErr := ReadServiceProfile(serviceProfilePath())
	return map[string]any{"name": WindowsServiceName, "state": status.State, "pid": status.ProcessId, "automatic": c.StartType == mgr.StartAutomatic, "account": c.ServiceStartName, "config": p.Config, "configured": profileErr == nil}, nil
}

func windowsServiceAutostart(action, config string) (bool, error) {
	if _, err := installedServiceExecutable(); errors.Is(err, registry.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return true, err
	}
	p, err := ReadServiceProfile(serviceProfilePath())
	if err != nil {
		return true, errors.New("MSI service is not configured; run remote-agent service configure as administrator")
	}
	config, err = filepath.Abs(config)
	if err != nil {
		return true, err
	}
	if !strings.EqualFold(config, p.Config) {
		return true, errors.New("Windows Service belongs to a different configuration; it was not changed")
	}
	s, err := openWindowsService(windows.SERVICE_QUERY_CONFIG | windows.SERVICE_QUERY_STATUS | windows.SERVICE_CHANGE_CONFIG | windows.SERVICE_START | windows.SERVICE_STOP)
	if err != nil {
		return true, err
	}
	defer s.Close()
	c, err := checkServiceOwner(s)
	if err != nil {
		return true, err
	}
	if action == "remove" {
		c.StartType = mgr.StartDisabled
		if err = s.UpdateConfig(c); err == nil {
			err = stopWindowsService(s)
		}
	} else {
		c.StartType = mgr.StartAutomatic
		if err = s.UpdateConfig(c); err == nil {
			err = startWindowsService(s)
		}
	}
	return true, err
}

func serviceACL(path, sid, permissions string) error {
	// An inherited grant on a large workspace can visit hundreds of thousands
	// of children even without icacls /T. Killing it after one minute leaves a
	// partially propagated ACL and previously produced an empty exit-status 1.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	started := time.Now()
	log.Printf(i18n.Text("granting Windows Service directory access: %s"), path)
	out, err := command(ctx, filepath.Join(os.Getenv("SystemRoot"), "System32", "icacls.exe"), []string{path, "/grant:r", "*" + sid + ":" + permissions}).CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			err = fmt.Errorf("directory permission propagation timed out after %s: %w", time.Since(started).Round(time.Second), ctx.Err())
		}
		return fmt.Errorf("grant service access to %s: %w: %s", path, err, out)
	}
	log.Printf(i18n.Text("Windows Service directory access granted in %s: %s"), time.Since(started).Round(time.Millisecond), path)
	return nil
}

// PrepareWindowsService is the MSI's architecture-native deferred action. The
// standard MSI service tables own create/delete/rollback; this adds SID/recovery.
func PrepareWindowsService() error {
	s, err := openWindowsService(windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return err
	}
	defer s.Close()
	c, err := checkServiceOwner(s)
	if err != nil {
		return err
	}
	c.SidType = windows.SERVICE_SID_TYPE_UNRESTRICTED
	if err = s.UpdateConfig(c); err != nil {
		return err
	}
	if err = s.SetRecoveryActions([]mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 5 * time.Second}}, 86400); err != nil {
		return err
	}
	return s.SetRecoveryActionsOnNonCrashFailures(true)
}

// ConfigureWindowsService keeps the user's files in place. The new service
// gets explicit access as LocalService + its service SID, never LocalSystem.
func ConfigureWindowsService(p ServiceProfile, noStart bool) (resultErr error) {
	stage := "validate original configuration"
	defer func() {
		if resultErr != nil {
			resultErr = fmt.Errorf("service setup step %s: %w", stage, resultErr)
		}
	}()
	if err := p.Validate(); err != nil {
		return err
	}
	c, err := Load(p.Config)
	if err != nil {
		return err
	}
	stage = "open installed service"
	s, err := openWindowsService(windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return err
	}
	defer s.Close()
	sc, err := checkServiceOwner(s)
	if err != nil {
		return err
	}
	oldState, err := s.Query()
	if err != nil {
		return err
	}
	stage = "stop service before configuration"
	if err = stopWindowsService(s); err != nil {
		return err
	}
	committed := false
	originalConfig := sc
	defer func() {
		if !committed {
			_ = s.UpdateConfig(originalConfig)
			if oldState.State == svc.Running {
				_ = startWindowsService(s)
			}
		}
	}()
	stage = "configure service identity and recovery"
	sc.SidType = windows.SERVICE_SID_TYPE_UNRESTRICTED
	sc.StartType = mgr.StartAutomatic
	if noStart {
		sc.StartType = mgr.StartDisabled
	}
	if err = s.UpdateConfig(sc); err != nil {
		return err
	}
	if err = s.SetRecoveryActions([]mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: 5 * time.Second}}, 86400); err != nil {
		return err
	}
	if err = s.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		return err
	}
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\`+WindowsServiceName)
	if err != nil {
		return err
	}
	// Existing per-user ACLs and data are retained. Inheritance propagates this
	// service-specific grant to existing history, recordings and package assets.
	type directoryGrant struct{ path, permission string }
	grants := map[string]directoryGrant{}
	grantDirectory := func(path, permission string) error {
		canonical, e := canonicalDir(path)
		if e != nil {
			return e
		}
		key := directoryKey(canonical)
		if previous, ok := grants[key]; !ok || previous.permission != "M" {
			grants[key] = directoryGrant{canonical, permission}
		}
		return nil
	}
	stage = "resolve original data and tool directories"
	if err = grantDirectory(filepath.Dir(c.Path), "M"); err != nil {
		return err
	}
	for _, root := range c.AllowedRoots {
		permission := "RX"
		if c.AllowWrite {
			permission = "M"
		}
		if err = grantDirectory(root, permission); err != nil {
			return err
		}
	}
	for _, path := range []string{p.Environment["CODEX_HOME"], filepath.Join(p.Environment["USERPROFILE"], ".qoder")} {
		if info, e := os.Stat(path); e == nil && info.IsDir() {
			if err = grantDirectory(path, "M"); err != nil {
				return err
			}
		}
	}
	stage = "grant service access to Codex credentials"
	for _, file := range []string{"config.toml", "auth.json"} {
		path := filepath.Join(p.Environment["CODEX_HOME"], file)
		if info, e := os.Stat(path); e == nil && !info.IsDir() {
			if err = serviceACL(path, sid.String(), "M"); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"RA_CODEX_PATH", "RA_QODER_PATH"} {
		if path := p.Environment[key]; path != "" {
			if info, e := os.Stat(path); e == nil && !info.IsDir() {
				if err = grantDirectory(filepath.Dir(path), "RX"); err != nil {
					return err
				}
			}
		}
	}
	for _, grant := range grants {
		stage = "grant service access to " + grant.path
		if err = serviceACL(grant.path, sid.String(), "(OI)(CI)"+grant.permission); err != nil {
			return err
		}
	}
	stage = "grant service access to device configuration"
	if err = serviceACL(c.Path, sid.String(), "M"); err != nil {
		return err
	}
	stage = "save protected machine service profile"
	dir := filepath.Dir(serviceProfilePath())
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	// Only administrators/SYSTEM may redirect the machine service. Readers see
	// paths and the allowlisted environment, not the original configuration secret.
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)(A;OICI;FRFX;;;BU)(A;OICI;FRFX;;;" + sid.String() + ")")
	if err != nil {
		return err
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return err
	}
	if err = windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		return err
	}
	previous, readErr := os.ReadFile(serviceProfilePath())
	if readErr != nil && !os.IsNotExist(readErr) {
		return readErr
	}
	if err = p.Save(serviceProfilePath()); err != nil {
		return err
	}
	restore := func(cause error) error {
		cause = errors.Join(cause, stopWindowsService(s))
		if readErr == nil {
			cause = errors.Join(cause, os.WriteFile(serviceProfilePath(), previous, 0600))
		} else {
			cause = errors.Join(cause, os.Remove(serviceProfilePath()))
		}
		return cause
	}
	// Remove only the task whose action matches this configuration. A failed
	// service activation restores its XML and prior enabled/running state.
	stage = "migrate matching legacy startup task"
	legacyExe := filepath.Join(filepath.Dir(c.Path), "bin", "remote-agent.exe")
	shell, legacyArgs := windowsTaskAction(legacyExe, c.Path)
	legacy, err := updatePowerShell("$ErrorActionPreference='Stop'; $task=Get-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -ErrorAction SilentlyContinue; if ($task) { if (@($task.Actions).Count -ne 1 -or $task.Actions[0].Execute -ne " + quotePowerShell(shell) + " -or $task.Actions[0].Arguments -ne " + quotePowerShell(legacyArgs) + ") { throw 'Existing scheduled task belongs to another configuration.' }; [ordered]@{xml=(Export-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent');running=($task.State -eq 'Running')} | ConvertTo-Json -Compress }; exit 0")
	if err != nil {
		return restore(err)
	}
	var taskSnapshot struct {
		XML     string `json:"xml"`
		Running bool   `json:"running"`
	}
	if legacy != "" {
		if err = json.Unmarshal([]byte(legacy), &taskSnapshot); err != nil {
			return restore(err)
		}
	}
	restoreAll := func(cause error) error {
		cause = restore(cause)
		if taskSnapshot.XML != "" {
			script := "Register-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent' -Xml " + quotePowerShell(taskSnapshot.XML) + " -Force | Out-Null; "
			if taskSnapshot.Running {
				script += "Start-ScheduledTask -TaskPath '\\' -TaskName 'RemoteAgent'"
			}
			_, rollbackErr := updatePowerShell(script)
			cause = errors.Join(cause, rollbackErr)
		}
		return cause
	}
	if taskSnapshot.XML != "" {
		if _, err = updatePowerShell(windowsTaskRemoveScript(legacyExe, c.Path)); err != nil {
			return restoreAll(err)
		}
	}
	if !noStart {
		stage = "start configured service"
		err = startWindowsService(s)
	}
	if err != nil {
		return restoreAll(err)
	}
	committed = true
	return nil
}

type windowsServiceHandler struct {
	run func(context.Context, chan<- error) error
}

func (h windowsServiceHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready, done := make(chan error, 1), make(chan error, 1)
	current := svc.Status{State: svc.StartPending, WaitHint: 30000}
	statuses <- current
	go func() { done <- h.run(ctx, ready) }()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-ready:
			if err != nil {
				cancel()
				<-done
				return true, 1
			}
			if current.State != svc.StopPending {
				current = svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
				statuses <- current
			}
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				statuses <- current
			case svc.Stop, svc.Shutdown:
				current = svc.Status{State: svc.StopPending, WaitHint: 30000}
				statuses <- current
				cancel()
			}
		case <-ticker.C:
			if current.State == svc.StartPending || current.State == svc.StopPending {
				current.CheckPoint++
				statuses <- current
			}
		case err := <-done:
			if err != nil {
				log.Printf("Windows Service stopped: %v", err)
				return true, 1
			}
			return false, 0
		}
	}
}

func RunWindowsService() error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if !isService {
		return errors.New("service run is reserved for the Windows Service Control Manager; use run for a foreground agent")
	}
	// Hold this handle for the process lifetime. If the daemon is terminated,
	// Windows also terminates its CLI children, preventing orphaned tasks from
	// overlapping the replacement daemon. Do not close it while returning a
	// graceful SCM Stop: main's process exit releases the handle after Stopped.
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
	process, err := windows.GetCurrentProcess()
	if err != nil {
		windows.CloseHandle(job)
		return err
	}
	if err = windows.AssignProcessToJobObject(job, process); err != nil {
		windows.CloseHandle(job)
		return err
	}
	return svc.Run(WindowsServiceName, windowsServiceHandler{run: func(ctx context.Context, ready chan<- error) error {
		err := runWindowsServiceAgent(ctx, ready)
		if err != nil {
			if events, openErr := eventlog.Open(WindowsServiceName); openErr == nil {
				_ = events.Error(1, err.Error())
				events.Close()
			}
		}
		return err
	}})
}

func runWindowsServiceAgent(ctx context.Context, ready chan<- error) error {
	p, err := ReadServiceProfile(serviceProfilePath())
	if errors.Is(err, os.ErrNotExist) {
		ready <- nil // Fresh MSI installations wait for local enrollment/configuration.
		<-ctx.Done()
		return nil
	}
	if err != nil {
		ready <- err
		return err
	}
	for key, value := range p.Environment {
		if err = os.Setenv(key, value); err != nil {
			ready <- err
			return err
		}
	}
	c, err := Load(p.Config)
	if err != nil {
		ready <- err
		return err
	}
	unlock, err := LockRun(c.Path)
	if err != nil {
		ready <- err
		return err
	}
	defer unlock()
	path := filepath.Join(filepath.Dir(c.Path), "agent.log")
	if info, e := os.Stat(path); e == nil && info.Size() > 10<<20 {
		_ = os.Rename(path, path+".old")
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		ready <- err
		return err
	}
	defer file.Close()
	log.SetOutput(file)
	log.Printf("Windows Service starting as LocalService (pid %d), data directory %s", os.Getpid(), filepath.Dir(c.Path))
	ready <- nil
	return New(c).Run(ctx)
}
