//go:build linux || darwin

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
)

// Unix services re-exec in place. No child updater can be killed by systemd's
// control-group cleanup, and no launchd/systemd registration needs rewriting.
func remoteClientUpdateSupported() bool {
	return len(os.Args) > 1 && os.Args[1] == "run"
}

func remoteClientRelease(ctx context.Context, server string) (clientupdate.Release, error) {
	return clientupdate.Fetch(ctx, clientupdate.HTTPClient(), server, runtime.GOOS, runtime.GOARCH, clientupdate.TrustKeys())
}

type unixUpdateState struct {
	SchemaVersion   int                  `json:"schemaVersion"`
	AgentID         string               `json:"agentId"`
	Server          string               `json:"server"`
	Executable      string               `json:"executable"`
	OperationID     string               `json:"operationId"`
	Phase           string               `json:"phase"`
	PreviousSHA256  string               `json:"previousSha256"`
	HighestSequence uint64               `json:"highestSequence"`
	HighestSHA256   string               `json:"highestSha256"`
	Release         clientupdate.Release `json:"release"`
}

func unixUpdatePath(config string) string { return config + ".runtime-update.json" }
func (s unixUpdateState) pending() bool   { return s.Phase == "switching" || s.Phase == "booting" }

func loadUnixUpdate(config string) (unixUpdateState, error) {
	var state unixUpdateState
	info, err := os.Lstat(unixUpdatePath(config))
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if !info.Mode().IsRegular() || info.Size() > 32768 {
		return state, errors.New("invalid Unix update journal")
	}
	f, err := os.Open(unixUpdatePath(config))
	if err != nil {
		return state, err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 32769))
	d.DisallowUnknownFields()
	if d.Decode(&state) != nil || d.Decode(&struct{}{}) != io.EOF || state.SchemaVersion != 1 || !filepath.IsAbs(state.Executable) || state.OperationID == "" || state.HighestSequence < state.Release.Sequence || len(state.PreviousSHA256) != 64 {
		return state, errors.New("invalid Unix update journal")
	}
	if state.Phase != "switching" && state.Phase != "booting" && state.Phase != "committed" && state.Phase != "rolled_back" && state.Phase != "abandoned" {
		return state, errors.New("invalid Unix update phase")
	}
	if err := state.Release.Verify(clientupdate.TrustKeys(), runtime.GOOS, runtime.GOARCH); err != nil {
		return state, err
	}
	if state.HighestSHA256 != state.Release.SHA256 || state.HighestSequence != state.Release.Sequence {
		return state, errors.New("invalid Unix update sequence floor")
	}
	return state, nil
}

func saveUnixUpdate(config string, state unixUpdateState) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err = writeMaintenanceFile(unixUpdatePath(config), raw); err != nil {
		return err
	}
	return syncUpdateDirectory(filepath.Dir(config))
}

func syncUpdateDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func unixExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

// Hard-link then rename: installed path is never missing, even across a crash.
// Both files live in the executable directory and retain the original mode.
func linkUpdateFile(source, target string) error {
	f, err := os.CreateTemp(filepath.Dir(target), ".remote-agent-link-*")
	if err != nil {
		return err
	}
	name := f.Name()
	if err = f.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err = os.Remove(name); err != nil {
		return err
	}
	defer os.Remove(name)
	if err = os.Link(source, name); err != nil {
		return err
	}
	if err = os.Rename(name, target); err != nil {
		return err
	}
	return syncUpdateDirectory(filepath.Dir(target))
}

func checkRemoteUpdateTransaction(exe, config string) error {
	state, err := loadUnixUpdate(config)
	if err != nil {
		return err
	}
	if state.pending() {
		return errors.New("a remote update is starting; wait for its result before running another update")
	}
	return nil
}

func remoteRuntimeState(config string) (string, bool, error) {
	state, err := loadUnixUpdate(config)
	return buildinfo.Current(), state.pending(), err
}

type unixUpdateProbe struct {
	Protocol int    `json:"protocol"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
}

// ProbeRemoteUpdate validates local startup inputs without connecting to a site,
// installing a service, modifying configuration, or taking the active run lock.
func ProbeRemoteUpdate(config string, output io.Writer) error {
	if _, err := Load(config); err != nil {
		return err
	}
	if err := validateWorkerStorage(config); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(unixUpdateProbe{1, buildinfo.Current(), runtime.GOOS, runtime.GOARCH})
}

type unixDownloadProgress struct {
	io.Writer
	n, total int64
	last     time.Time
	report   func(int64, int64)
}

func (w *unixDownloadProgress) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.n += int64(n)
	if time.Since(w.last) > 500*time.Millisecond || w.n == w.total {
		w.last = time.Now()
		w.report(w.n, w.total)
	}
	return n, err
}

func (a *Agent) applyRemoteClientUpdate(ctx context.Context, op MaintenanceOperation, release clientupdate.Release) error {
	if !remoteClientUpdateSupported() {
		return errors.New("remote Unix updates require the standard remote-agent run process")
	}
	exe, err := unixExecutable()
	if err != nil {
		return err
	}
	return a.applyUnixUpdate(ctx, op, release, exe, syscall.Exec)
}

func (a *Agent) applyUnixUpdate(ctx context.Context, op MaintenanceOperation, release clientupdate.Release, exe string, reexec func(string, []string, []string) error) error {
	a.mu.Lock()
	c := a.Config
	a.mu.Unlock()
	if err := release.Verify(clientupdate.TrustKeys(), runtime.GOOS, runtime.GOARCH); err != nil {
		return err
	}
	if err := clientupdate.Compatible(release); err != nil {
		return err
	}
	unlock, err := lockUpdate(exe + ".update.lock")
	if err != nil {
		return err
	}
	defer unlock()
	if err := checkRemoteUpdateTransaction(exe, c.Path); err != nil {
		return err
	}
	previous, err := loadUnixUpdate(c.Path)
	if err != nil {
		return err
	}
	if release.Sequence < previous.HighestSequence || release.Sequence == previous.HighestSequence && release.SHA256 != previous.HighestSHA256 {
		return errors.New("published release is older than the accepted sequence floor")
	}
	before, err := fileDigest(exe)
	if err != nil {
		return err
	}
	if before == release.SHA256 {
		if release.Version != buildinfo.Current() {
			return errors.New("published version differs from the running client")
		}
		previousSHA := previous.PreviousSHA256
		if previousSHA == "" {
			previousSHA = before
		}
		return saveUnixUpdate(c.Path, unixUpdateState{SchemaVersion: 1, AgentID: c.ID, Server: c.Server, Executable: exe, OperationID: op.ID, Phase: "committed", PreviousSHA256: previousSHA, HighestSequence: release.Sequence, HighestSHA256: release.SHA256, Release: release})
	}
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&(os.ModeSetuid|os.ModeSetgid) != 0 {
		return errors.New("remote updates require a regular executable without set-ID bits")
	}
	f, err := os.CreateTemp(filepath.Dir(exe), ".remote-agent-update-*")
	if err != nil {
		return fmt.Errorf("cannot write the installed client directory: %w", err)
	}
	candidate := f.Name()
	defer os.Remove(candidate)
	progress := &unixDownloadProgress{Writer: f, total: release.Size, report: func(n, total int64) {
		_ = a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) {
			v.Stage = "download"
			v.Downloaded = n
			v.Total = total
			v.Message = "正在下载并验证客户端更新"
		})
	}}
	err = clientupdate.Download(ctx, clientupdate.HTTPClient(), c.Server, release, progress)
	if err == nil {
		err = f.Chmod(info.Mode().Perm())
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	if err = a.PauseUpdates(ctx, false); err != nil {
		return err
	}
	defer a.ResumeUpdates()
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	out, err := command(probeCtx, candidate, []string{"internal-update-probe", c.Path}).Output()
	var probe unixUpdateProbe
	if err != nil || len(out) > 4096 || json.Unmarshal(out, &probe) != nil || probe.Protocol != 1 || probe.Version != release.Version || probe.OS != runtime.GOOS || probe.Arch != runtime.GOARCH {
		return errors.New("downloaded client failed its local startup check; installed client is unchanged")
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	state := unixUpdateState{SchemaVersion: 1, AgentID: c.ID, Server: c.Server, Executable: exe, OperationID: op.ID, Phase: "switching", PreviousSHA256: before, HighestSequence: release.Sequence, HighestSHA256: release.SHA256, Release: release}
	if err = saveUnixUpdate(c.Path, state); err != nil {
		return err
	}
	if err = linkUpdateFile(exe, exe+".previous"); err != nil {
		state.Phase = "rolled_back"
		return errors.Join(err, saveUnixUpdate(c.Path, state))
	}
	if err = os.Rename(candidate, exe); err != nil {
		state.Phase = "rolled_back"
		return errors.Join(err, saveUnixUpdate(c.Path, state))
	}
	if err = syncUpdateDirectory(filepath.Dir(exe)); err == nil {
		err = a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) {
			v.Stage = "restarting"
			v.Message = "更新已校验，正在重新加载客户端；稍后将自动连接"
		})
	}
	if err == nil {
		args := append([]string{}, os.Args...)
		args[0] = exe
		err = reexec(exe, args, os.Environ())
		if err == nil {
			err = errors.New("client re-exec returned unexpectedly")
		}
	}
	if restoreErr := rollbackUnixUpdate(c.Path, exe, state, "重新加载失败，已恢复上一版本"); restoreErr != nil {
		return errors.Join(err, restoreErr)
	}
	return fmt.Errorf("client reload failed; previous binary restored: %w", err)
}

func rollbackUnixUpdate(config, exe string, state unixUpdateState, message string) error {
	if state.Executable != exe {
		return errors.New("update journal belongs to another executable")
	}
	backup, err := fileDigest(exe + ".previous")
	if err != nil || backup != state.PreviousSHA256 {
		return errors.New("previous client failed verification; retain update files for manual recovery")
	}
	current, err := fileDigest(exe)
	if err != nil {
		return err
	}
	if current != state.PreviousSHA256 {
		if current != state.Release.SHA256 {
			return errors.New("installed client changed outside the update transaction")
		}
		if err = linkUpdateFile(exe, exe+".failed"); err != nil {
			return err
		}
		if err = linkUpdateFile(exe+".previous", exe); err != nil {
			return err
		}
	}
	state.Phase = "rolled_back"
	if err = saveUnixUpdate(config, state); err != nil {
		return err
	}
	m := (&Agent{Config: Config{Path: config, ID: state.AgentID}}).maintenanceState()
	return m.change(state.OperationID, func(v *MaintenanceOperation) { v.Status = "failed"; v.Stage = "rollback"; v.Message = message })
}

// The run lock must already be held. It is close-on-exec and will be reacquired
// by the replacement image. A second launch can never trigger a rollback.
func RunWithRemoteUpdateRecovery(config string, run func() error) error {
	exe, err := unixExecutable()
	if err != nil {
		return err
	}
	return runWithUnixRecovery(config, exe, buildinfo.Current(), run, syscall.Exec)
}

func runWithUnixRecovery(config, exe, version string, run func() error, reexec func(string, []string, []string) error) (err error) {
	state, err := loadUnixUpdate(config)
	if err != nil {
		return err
	}
	if !state.pending() {
		return run()
	}
	unlock, err := lockUpdate(exe + ".update.lock")
	if err != nil {
		return err
	}
	state, err = loadUnixUpdate(config)
	if err != nil {
		unlock()
		return err
	}
	if state.pending() {
		if state.Executable != exe {
			unlock()
			return errors.New("update journal belongs to another executable")
		}
		c, loadErr := Load(config)
		if loadErr == nil && (c.ID != state.AgentID || strings.TrimSuffix(c.Server, "/") != strings.TrimSuffix(state.Server, "/")) {
			state.Phase = "abandoned"
			err = saveUnixUpdate(config, state)
		} else if state.Phase == "booting" || loadErr != nil {
			err = rollbackUnixUpdate(config, exe, state, "新版本未完成本地启动，已恢复上一版本")
			unlock()
			if err != nil {
				return err
			}
			args := append([]string{}, os.Args...)
			args[0] = exe
			return reexec(exe, args, os.Environ())
		} else {
			digest, e := fileDigest(exe)
			if e != nil {
				err = e
			} else if digest == state.PreviousSHA256 {
				state.Phase = "rolled_back"
				err = saveUnixUpdate(config, state)
			} else if digest != state.Release.SHA256 || version != state.Release.Version {
				err = errors.New("starting client differs from signed update")
			} else {
				state.Phase = "booting"
				err = saveUnixUpdate(config, state)
			}
		}
	}
	unlock()
	if err != nil {
		return err
	}
	defer func() {
		failure := recover()
		if failure != nil {
			err = fmt.Errorf("client startup panicked: %v", failure)
		}
		current, e := loadUnixUpdate(config)
		if e == nil && current.Phase == "booting" {
			if e = rollbackUnixUpdate(config, exe, current, "新版本启动中断，已恢复上一版本"); e != nil {
				err = errors.Join(err, e)
			} else if err != nil {
				args := append([]string{}, os.Args...)
				args[0] = exe
				err = errors.Join(err, reexec(exe, args, os.Environ()))
			}
		}
		if failure != nil && (e != nil || current.Phase != "booting") {
			panic(failure)
		}
	}()
	return run()
}

func (a *Agent) commitRemoteRuntime() error {
	state, err := loadUnixUpdate(a.Config.Path)
	if err != nil || state.Phase != "booting" {
		return err
	}
	if state.AgentID != a.Config.ID || state.Server != a.Config.Server {
		return errors.New("binding changed during client startup")
	}
	exe, err := unixExecutable()
	if err != nil {
		return err
	}
	digest, err := fileDigest(exe)
	if err != nil {
		return err
	}
	if exe != state.Executable || digest != state.Release.SHA256 || buildinfo.Current() != state.Release.Version {
		return errors.New("client startup did not match the verified release")
	}
	state.Phase = "committed"
	return saveUnixUpdate(a.Config.Path, state)
}
