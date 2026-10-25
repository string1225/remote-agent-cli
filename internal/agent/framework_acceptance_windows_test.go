package agent

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Called only by the existing explicit disposable-machine MSI acceptance test.
func runtimeAcceptance(t *testing.T, exe string, retained func(), setUpdateHandler func(http.Handler), remoteCall func(context.Context, string, any) (MaintenanceOperation, error)) {
	t.Helper()
	seedHex, err := os.ReadFile(os.Getenv("REMOTE_AGENT_UPDATE_TEST_KEY"))
	if err != nil {
		t.Fatal(err)
	}
	seed, err := hex.DecodeString(string(seedHex))
	if err != nil || len(seed) != 32 {
		t.Fatal("invalid fixture signing key")
	}
	key := ed25519.NewKeyFromSeed(seed)
	core, err := os.ReadFile(os.Getenv("REMOTE_AGENT_UPDATE_TEST_CORE"))
	if err != nil {
		t.Fatal(err)
	}
	var body = []byte(core)
	makeRelease := func(seq uint64, version string) clientupdate.Release {
		sum := sha256.Sum256(body)
		return clientupdate.Sign(clientupdate.Payload{SchemaVersion: 1, Sequence: seq, Version: version, OS: "windows", Arch: "amd64", MinHostVersion: 1, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:])}, "release-2026", key)
	}
	release := makeRelease(1, "acceptance-runtime")
	corrupt := false
	updates := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("download leaked credentials")
		}
		switch r.URL.Path {
		case "/updates/v1/windows/amd64.json":
			_ = json.NewEncoder(w).Encode(release)
		case "/downloads/remote-agent-core-windows-amd64.exe":
			if corrupt {
				_, _ = w.Write([]byte("corrupt"))
			} else {
				_, _ = w.Write(body)
			}
		default:
			http.NotFound(w, r)
		}
	})
	setUpdateHandler(updates)
	defer setUpdateHandler(nil)
	server := httptest.NewServer(updates)
	defer server.Close()
	configuration := func() any {
		s, err := openWindowsService(windows.SERVICE_QUERY_CONFIG)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		cfg, err := s.Config()
		if err != nil {
			t.Fatal(err)
		}
		return cfg
	}
	product := func() map[string]string {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.READ|registry.WOW64_64KEY)
		if err != nil {
			t.Fatal(err)
		}
		defer key.Close()
		names, err := key.ReadSubKeyNames(-1)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			k, e := registry.OpenKey(key, name, registry.QUERY_VALUE)
			if e != nil {
				continue
			}
			label, _, _ := k.GetStringValue("DisplayName")
			if label == "Remote Agent" {
				v, _, _ := k.GetStringValue("DisplayVersion")
				u, _, _ := k.GetStringValue("UninstallString")
				k.Close()
				return map[string]string{"product": name, "version": v, "uninstall": u}
			}
			k.Close()
		}
		t.Fatal("MSI product missing")
		return nil
	}
	beforeConfig, beforeProduct := configuration(), product()
	// A filtered token has no enabled Administrators group. The owner SID stays
	// the configured user's SID; this verifies the ordinary-user IPC path.
	token := acceptanceNormalToken(t)
	defer token.Close()
	run := func(args ...string) (string, error) {
		cmd := exec.Command(exe, args...)
		cmd.Env = append(os.Environ(), "REMOTE_AGENT_LANGUAGE=en")
		cmd.SysProcAttr = &syscall.SysProcAttr{Token: syscall.Token(token)}
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	if out, err := run("update", "--check", "--server", server.URL); err != nil || !strings.Contains(out, "available") {
		t.Fatal("unprivileged check", err, out)
	}
	if out, err := run("update", "--server", server.URL); err != nil {
		t.Fatal("unprivileged runtime update", err, out)
	}
	if out, err := run("version"); err != nil || !strings.Contains(out, "acceptance-runtime") {
		t.Fatal("runtime version did not change", err, out)
	}
	if !reflect.DeepEqual(beforeConfig, configuration()) || !reflect.DeepEqual(beforeProduct, product()) {
		t.Fatal("runtime update changed SCM/MSI registration")
	}
	retained()
	store := clientupdate.Store{Path: filepath.Join(filepath.Dir(serviceProfilePath()), "runtime"), Keys: map[string]ed25519.PublicKey{"release-2026": key.Public().(ed25519.PublicKey)}, Arch: "amd64"}
	state, err := store.Load()
	if err != nil || state.Pending || state.Current == nil {
		t.Fatal("runtime not committed", err)
	}
	stable := state.Current.SHA256
	// A signed but non-executable candidate must fall back before business work.
	body = []byte("signed invalid executable fixture")
	release = makeRelease(2, "bad-startup")
	if out, err := run("update", "--server", server.URL); err == nil {
		t.Fatal("bad startup accepted", out)
	}
	state, err = store.Load()
	if err != nil || state.Current == nil || state.Current.SHA256 != stable {
		t.Fatal("startup rollback did not retain stable runtime", err)
	}
	retained()
	body = append(append([]byte{}, core...), []byte("fixture distinct corrupt download")...)
	release = makeRelease(3, "corrupt-download")
	corrupt = true
	if out, err := run("update", "--server", server.URL); err == nil {
		t.Fatal("corrupt download accepted", out)
	}
	corrupt = false
	if out, err := run("update", "--rollback"); err != nil {
		t.Fatal("explicit retained rollback", err, out)
	}
	// Restore the good runtime as the highest accepted release for later MSI
	// upgrade/repair checks. They must preserve this runtime, not select seed.
	body = core
	release = makeRelease(4, "acceptance-runtime")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	op, err := remoteCall(ctx, "maintenance.start", maintenanceInput{Kind: "client", RequestID: "acceptance-remote-update"})
	// CLI rollback waits for local worker readiness, independently of the
	// network connection. Wait for its relay using the same idempotency key.
	for ctx.Err() == nil && (err != nil || op.ID == "") {
		time.Sleep(500 * time.Millisecond)
		op, err = remoteCall(ctx, "maintenance.start", maintenanceInput{Kind: "client", RequestID: "acceptance-remote-update"})
	}
	if err != nil || op.ID == "" {
		t.Fatal("LocalService remote update could not start", err, op)
	}
	for ctx.Err() == nil && (op.Status == "queued" || op.Status == "running") {
		time.Sleep(500 * time.Millisecond)
		next, e := remoteCall(ctx, "maintenance.status", maintenanceInput{OperationID: op.ID})
		if e == nil && next.ID == op.ID {
			op = next
		}
	}
	if op.Status != "succeeded" || op.AfterVersion != "acceptance-runtime" {
		t.Fatal("service SID update/restart/durable result failed", ctx.Err(), op)
	}
	if !reflect.DeepEqual(beforeConfig, configuration()) || !reflect.DeepEqual(beforeProduct, product()) {
		t.Fatal("recovery changed registration")
	}
	retained()
	t.Log("ordinary-user update/check/rollback, LocalService SID remote update/restart/result, signed invalid startup rollback, corrupt download rejection and unchanged MSI/SCM passed")
}

func acceptanceNormalToken(t *testing.T) windows.Token {
	t.Helper()
	var current windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY|windows.TOKEN_DUPLICATE|windows.TOKEN_ASSIGN_PRIMARY, &current); err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	if linked, err := current.GetLinkedToken(); err == nil {
		return linked
	}
	admin, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		t.Fatal(err)
	}
	disabled := windows.SIDAndAttributes{Sid: admin}
	var token windows.Token
	proc := windows.NewLazySystemDLL("advapi32.dll").NewProc("CreateRestrictedToken")
	ok, _, callErr := proc.Call(uintptr(current), 1, 1, uintptr(unsafe.Pointer(&disabled)), 0, 0, 0, 0, uintptr(unsafe.Pointer(&token)))
	if ok == 0 {
		t.Fatal("create fixture ordinary-user token", callErr)
	}
	groups, err := token.GetTokenGroups()
	member := false
	if err == nil {
		for _, group := range groups.AllGroups() {
			if group.Sid.Equals(admin) && group.Attributes&windows.SE_GROUP_ENABLED != 0 && group.Attributes&windows.SE_GROUP_USE_FOR_DENY_ONLY == 0 {
				member = true
			}
		}
	}
	if err != nil || member {
		token.Close()
		t.Fatal(fmt.Sprintf("fixture token still has administrative group: %v", err))
	}
	return token
}
