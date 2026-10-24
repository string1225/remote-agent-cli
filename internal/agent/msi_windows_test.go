package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// This opt-in acceptance test is for a disposable, elevated Windows runner.
// Ordinary go test never installs an MSI or touches a developer's service/data.
func TestMSIWindowsServiceLifecycle(t *testing.T) {
	if os.Getenv("REMOTE_AGENT_MSI_ACCEPTANCE") != "1" {
		t.Skip("requires explicit opt-in on a disposable elevated Windows runner")
	}
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Fatal("MSI acceptance requires an elevated runner")
	}
	if s, err := openWindowsService(windows.SERVICE_QUERY_STATUS); err == nil {
		s.Close()
		t.Fatal("refusing to alter an existing RemoteAgent service")
	} else if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		t.Fatal(err)
	}
	if _, err := os.Stat(serviceProfilePath()); !os.IsNotExist(err) {
		t.Fatal("refusing to alter an existing service profile")
	}
	baseMSI, goodMSI, badMSI := os.Getenv("REMOTE_AGENT_MSI_BASE"), os.Getenv("REMOTE_AGENT_MSI_UPGRADE"), os.Getenv("REMOTE_AGENT_MSI_FAILURE")
	for _, path := range []string{baseMSI, goodMSI, badMSI} {
		if !filepath.IsAbs(path) {
			t.Fatal("all fixture MSI paths must be absolute")
		}
	}
	dir := t.TempDir()
	logDir := os.Getenv("REMOTE_AGENT_MSI_LOG_DIR")
	if logDir == "" {
		logDir = dir
	}
	if err := os.MkdirAll(logDir, 0700); err != nil {
		t.Fatal(err)
	}
	operationCount := 0
	msi := func(operation, path string, wantFailure bool) {
		t.Helper()
		t.Logf("MSI %s %s starting", operation, filepath.Base(path))
		operationCount++
		logPath := filepath.Join(logDir, fmt.Sprintf("%02d-%s-%s.log", operationCount, strings.TrimPrefix(operation, "/"), filepath.Base(path)))
		args := []string{operation, path, "/qn", "/norestart", "/l*v", logPath}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "msiexec.exe", args...)
		err := cmd.Run()
		code := cmd.ProcessState.ExitCode()
		if wantFailure && code == 1603 {
			t.Log("injected MSI failure returned as expected")
			return
		}
		if !wantFailure && (code == 0 || code == 3010) {
			t.Logf("MSI %s completed", operation)
			return
		}
		logData, _ := os.ReadFile(logPath)
		if len(logData) > 18000 {
			logData = logData[len(logData)-18000:]
		}
		t.Fatalf("MSI %s exit %d (%v):\n%s", operation, code, err, logData)
	}
	msi("/i", baseMSI, false)
	installed := true
	t.Cleanup(func() {
		if installed {
			for _, packagePath := range []string{goodMSI, baseMSI} {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				_ = exec.CommandContext(ctx, "msiexec.exe", "/x", packagePath, "/qn", "/norestart").Run()
				cancel()
			}
		}
		_ = os.Remove(serviceProfilePath())
		_ = os.Remove(filepath.Dir(serviceProfilePath()))
	})
	exe, err := installedServiceExecutable()
	if err != nil {
		t.Fatal(err)
	}
	assertCommandPath := func(wantInstalled bool) {
		t.Helper()
		key, e := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, registry.QUERY_VALUE)
		if e != nil {
			t.Fatal(e)
		}
		machine, _, e := key.GetStringValue("Path")
		key.Close()
		if e != nil {
			t.Fatal(e)
		}
		count := 0
		for _, entry := range filepath.SplitList(machine) {
			if strings.EqualFold(filepath.Clean(entry), filepath.Dir(exe)) {
				count++
			}
		}
		if wantInstalled && count != 1 || !wantInstalled && count != 0 {
			t.Fatal("MSI PATH entry missing, duplicated or retained after uninstall", count)
		}
		if !wantInstalled {
			return
		}
		user := ""
		if key, e := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE); e == nil {
			user, _, _ = key.GetStringValue("Path")
			key.Close()
		}
		freshPath, e := registry.ExpandString(machine + ";" + user)
		if e != nil {
			t.Fatal(e)
		}
		probe := exec.Command(filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe"), "-NoProfile", "-Command", "(Get-Command remote-agent -CommandType Application).Source; remote-agent version")
		probe.Env = os.Environ()
		for i, v := range probe.Env {
			if strings.HasPrefix(strings.ToUpper(v), "PATH=") {
				probe.Env[i] = "PATH=" + freshPath
			}
		}
		out, e := probe.CombinedOutput()
		if e != nil || !strings.Contains(strings.ToLower(string(out)), strings.ToLower(exe)) {
			t.Fatal("new PowerShell cannot resolve MSI command", e, string(out))
		}
	}
	assertCommandPath(true)
	s, err := openWindowsService(windows.SERVICE_ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	config, err := checkServiceOwner(s)
	if err != nil {
		t.Fatal(err)
	}
	if config.StartType != mgr.StartAutomatic || config.SidType != windows.SERVICE_SID_TYPE_UNRESTRICTED {
		t.Fatal("MSI did not configure automatic service with a service SID")
	}
	actions, err := s.RecoveryActions()
	if err != nil || len(actions) == 0 || actions[0].Type != mgr.ServiceRestart || actions[0].Delay != 5*time.Second {
		t.Fatal("crash recovery missing", err)
	}
	s.Close()

	var connections atomic.Int32
	var offline atomic.Bool
	var mu sync.Mutex
	active := map[*websocket.Conn]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if offline.Load() {
			w.WriteHeader(503)
			return
		}
		if r.URL.Path != "/api/agent/connect" || r.URL.Query().Get("id") != "retained-device" || r.Header.Get("Authorization") != "Bearer fixture-private-credential" {
			w.WriteHeader(401)
			return
		}
		conn, e := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if e != nil {
			return
		}
		mu.Lock()
		active[conn] = true
		mu.Unlock()
		defer func() { conn.Close(); mu.Lock(); delete(active, conn); mu.Unlock() }()
		first := true
		for {
			var signal protocol.Signal
			if conn.ReadJSON(&signal) != nil {
				return
			}
			if signal.Type == "heartbeat" && len(signal.Services) == 1 && signal.Services[0].ID == "retained-project" {
				if first {
					connections.Add(1)
					first = false
				}
				if conn.WriteJSON(protocol.Signal{Type: "pong"}) != nil {
					return
				}
			}
		}
	}))
	t.Cleanup(server.Close)
	workspace := filepath.Join(dir, "workspace")
	home := filepath.Join(dir, "profile")
	dataDir := filepath.Join(home, ".remote-agent")
	codexHome := filepath.Join(home, ".codex")
	for _, path := range []string{workspace, dataDir, codexHome, filepath.Join(dataDir, "sayso", "recordings")} {
		if err = os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	service := protocol.Service{ID: "retained-project", Name: "中文项目", Provider: "codex", Workspace: workspace}
	c := Config{Path: filepath.Join(dataDir, "config.json"), Server: server.URL, ID: "retained-device", Credential: "fixture-private-credential", AllowedRoots: []string{workspace}, AllowWrite: true, Services: []protocol.Service{service}}
	if err = c.Save(); err != nil {
		t.Fatal(err)
	}
	history := (&Agent{Config: c}).historyStore()
	conversation, err := history.create(service, "保留会话")
	if err != nil {
		t.Fatal(err)
	}
	if err = history.nativeID(conversation.ID, "retained-native-session"); err != nil {
		t.Fatal(err)
	}
	if _, err = history.append(conversation.ID, protocol.ChatEntry{Role: "user", Text: "升级保留这一条消息"}); err != nil {
		t.Fatal(err)
	}
	so, err := (&Agent{Config: c, Providers: map[string]Provider{}}).saysoStore()
	if err != nil || len(so.state.Projects) != 1 {
		t.Fatal("SaySo fixture failed", err)
	}
	recording := filepath.Join(dataDir, "sayso", "recordings", "retained.webm")
	if err = os.WriteFile(recording, []byte("fixture-recording-bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	codexConfig := filepath.Join(codexHome, "config.toml")
	if err = os.WriteFile(codexConfig, []byte("model_provider = 'fixture'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	witnesses := map[string][]byte{}
	for _, path := range []string{c.Path, history.path, filepath.Join(dataDir, "sayso", "state.json"), recording, codexConfig} {
		witnesses[path], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	assertRetained := func() {
		t.Helper()
		for path, want := range witnesses {
			got, e := os.ReadFile(path)
			if e != nil || !bytes.Equal(got, want) {
				t.Fatalf("business data changed: %s (%v)", path, e)
			}
		}
	}
	probe, envProbe := filepath.Join(workspace, "identity.txt"), filepath.Join(workspace, "environment.txt")
	fakeCLI := filepath.Join(workspace, "fixture-codex.ps1")
	script := "if ($args[0] -eq '--help') { whoami | Set-Content -LiteralPath " + quotePowerShell(probe) + "; [IO.File]::WriteAllText(" + quotePowerShell(envProbe) + ",$env:USERPROFILE); Write-Output 'exec'; exit 0 }; exit 1"
	if err = os.WriteFile(fakeCLI, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	p := ServiceProfile{Config: c.Path, Environment: map[string]string{"USERPROFILE": home, "CODEX_HOME": codexHome, "PATH": os.Getenv("PATH"), "RA_CODEX_PATH": fakeCLI, "RA_QODER_PATH": filepath.Join(workspace, "missing-qoder")}}
	if err = ConfigureWindowsService(p, false); err != nil {
		t.Fatal(err)
	}
	wait := func(label string, predicate func() bool) {
		t.Helper()
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			if predicate() {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
		logs, _ := os.ReadFile(filepath.Join(dataDir, "agent.log"))
		t.Fatalf("timeout: %s\n%s", label, logs)
	}
	wait("authenticated service heartbeat", func() bool { return connections.Load() > 0 })
	identity, err := os.ReadFile(probe)
	if err != nil || !strings.Contains(strings.ToLower(string(identity)), "local service") {
		t.Fatalf("tool did not run as LocalService: %s (%v)", identity, err)
	}
	environment, err := os.ReadFile(envProbe)
	if err != nil || string(environment) != home {
		t.Fatal("tool did not use original user profile", err)
	}
	servicePID := func() uint32 {
		t.Helper()
		service, e := openWindowsService(windows.SERVICE_QUERY_STATUS)
		if e != nil {
			t.Fatal(e)
		}
		defer service.Close()
		state, e := service.Query()
		if e != nil || state.State != svc.Running {
			t.Fatal("service is not running", e)
		}
		return state.ProcessId
	}
	var sessionID uint32
	if err = windows.ProcessIdToSessionId(servicePID(), &sessionID); err != nil || sessionID != 0 {
		t.Fatal("daemon is not in noninteractive service session 0", err)
	}
	assertRetained()
	t.Log("automatic LocalService/SID, session 0, real child process, original environment and authenticated heartbeat passed")

	before := connections.Load()
	offline.Store(true)
	mu.Lock()
	for conn := range active {
		conn.Close()
	}
	mu.Unlock()
	time.Sleep(2 * time.Second)
	_ = servicePID()
	offline.Store(false)
	wait("network reconnection", func() bool { return connections.Load() > before })
	t.Log("network reconnection passed")
	before = connections.Load()
	oldPID := servicePID()
	// The disposable elevated test actor needs SeDebugPrivilege to open a
	// LocalService process for termination. os.Process.Kill duplicates its
	// limited query handle and fails across identities on current Windows.
	// This privilege is never enabled in the product service.
	var actorToken windows.Token
	if err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY|windows.TOKEN_ADJUST_PRIVILEGES, &actorToken); err != nil {
		t.Fatal(err)
	}
	defer actorToken.Close()
	var debugLUID windows.LUID
	debugName, _ := windows.UTF16PtrFromString("SeDebugPrivilege")
	if err = windows.LookupPrivilegeValue(nil, debugName, &debugLUID); err != nil {
		t.Fatal(err)
	}
	privileges := windows.Tokenprivileges{PrivilegeCount: 1, Privileges: [1]windows.LUIDAndAttributes{{Luid: debugLUID, Attributes: windows.SE_PRIVILEGE_ENABLED}}}
	var previousPrivileges windows.Tokenprivileges
	var returnedLength uint32
	if err = windows.AdjustTokenPrivileges(actorToken, false, &privileges, uint32(unsafe.Sizeof(previousPrivileges)), &previousPrivileges, &returnedLength); err != nil {
		t.Fatal(err)
	}
	defer windows.AdjustTokenPrivileges(actorToken, false, &previousPrivileges, 0, nil, nil)
	process, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, oldPID)
	if err != nil {
		t.Fatal("open isolated service for crash injection", err)
	}
	err = windows.TerminateProcess(process, 1)
	windows.CloseHandle(process)
	if err != nil {
		t.Fatal(err)
	}
	wait("SCM crash recovery", func() bool { return connections.Load() > before })
	t.Log("SCM crash restart passed")
	if servicePID() == oldPID {
		t.Fatal("SCM did not replace failed process")
	}
	assertRetained()

	before = connections.Load()
	legacyBin := filepath.Join(home, ".remote-agent", "bin")
	if err = os.MkdirAll(legacyBin, 0700); err != nil {
		t.Fatal(err)
	}
	legacyProgram := filepath.Join(legacyBin, "remote-agent.exe")
	if err = os.WriteFile(legacyProgram, []byte("obsolete agent fixture; must not execute"), 0600); err != nil {
		t.Fatal(err)
	}
	retainedTool := filepath.Join(legacyBin, "codex.exe")
	if err = os.WriteFile(retainedTool, []byte("retained Codex fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	msi("/i", goodMSI, false)
	assertCommandPath(true)
	if _, err = os.Stat(legacyProgram); !os.IsNotExist(err) {
		t.Fatal("MSI commit did not clean legacy Agent program", err)
	}
	if b, e := os.ReadFile(retainedTool); e != nil || string(b) != "retained Codex fixture" {
		t.Fatal("MSI cleanup removed another tool")
	}
	wait("heartbeat after major upgrade", func() bool { return connections.Load() > before })
	t.Log("major upgrade restored the authenticated connection")
	if version, e := exec.Command(exe, "version").CombinedOutput(); e != nil || !strings.Contains(string(version), "acceptance-upgrade") {
		t.Fatalf("major upgrade did not replace the executable: %v %s", e, version)
	}
	goodDigest, err := fileDigest(exe)
	if err != nil {
		t.Fatal(err)
	}
	assertRetained()
	before = connections.Load()
	msi("/i", badMSI, true)
	if service, e := openWindowsService(windows.SERVICE_QUERY_CONFIG | windows.SERVICE_QUERY_STATUS); e == nil {
		configuration, _ := service.Config()
		state, _ := service.Query()
		service.Close()
		t.Logf("rollback service state=%d SID type=%d start type=%d exit=%d", state.State, configuration.SidType, configuration.StartType, state.Win32ExitCode)
	}
	wait("service restored after failed upgrade", func() bool { return connections.Load() > before })
	t.Log("failed upgrade restored the authenticated connection")
	if got, e := fileDigest(exe); e != nil || got != goodDigest {
		t.Fatal("failed upgrade did not restore previous program", e)
	}
	assertRetained()
	before = connections.Load()
	msi("/famus", goodMSI, false)
	assertCommandPath(true)
	wait("service after MSI repair", func() bool { return connections.Load() > before })
	t.Log("MSI repair restored the authenticated connection")
	assertRetained()
	if output, e := exec.Command(exe, "update", "--config", c.Path).CombinedOutput(); e == nil || !strings.Contains(string(output), "managed by Windows Installer") {
		t.Fatal("EXE update bypassed MSI ownership", e)
	}
	msi("/x", goodMSI, false)
	assertCommandPath(false)
	installed = false
	if service, e := openWindowsService(windows.SERVICE_QUERY_STATUS); e == nil {
		service.Close()
		t.Fatal("uninstall retained the service")
	}
	assertRetained()
	retained, err := history.get(conversation.ID)
	if err != nil || retained.NativeID != "retained-native-session" {
		t.Fatal("native session identity lost", err)
	}
	profile, err := ReadServiceProfile(serviceProfilePath())
	if err != nil || profile.Config != c.Path {
		t.Fatal("uninstall lost retained service profile", err)
	}
	metadata, _ := json.Marshal(profile)
	if bytes.Contains(metadata, []byte(c.Credential)) {
		t.Fatal("service metadata leaked device credential")
	}
	t.Log("network recovery, SCM crash restart, major upgrade, injected-failure rollback, repair, uninstall and retained data passed")
}
