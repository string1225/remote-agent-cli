//go:build linux || darwin

package agent

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func unixReleaseKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if path := os.Getenv("REMOTE_AGENT_UNIX_TEST_KEY"); path != "" {
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		seed, e := hex.DecodeString(strings.TrimSpace(string(raw)))
		if e != nil || len(seed) != ed25519.SeedSize {
			t.Fatal("invalid isolated test key")
		}
		key = ed25519.NewKeyFromSeed(seed)
		pub = key.Public().(ed25519.PublicKey)
	}
	original := clientupdate.ReleasePublicKey
	clientupdate.ReleasePublicKey = hex.EncodeToString(pub)
	t.Cleanup(func() { clientupdate.ReleasePublicKey = original })
	return key
}

func signUnixFixture(key ed25519.PrivateKey, seq uint64, version string, data []byte) clientupdate.Release {
	sum := sha256.Sum256(data)
	return clientupdate.Sign(clientupdate.Payload{SchemaVersion: 1, Sequence: seq, Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH, MinHostVersion: 1, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])}, "release-2026", key)
}

func TestUnixRunWithoutJournalDoesNotNeedWritableExecutableDirectory(t *testing.T) {
	root := t.TempDir()
	exeDir := filepath.Join(root, "read-only")
	if err := os.Mkdir(exeDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(exeDir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(exeDir, 0700) })
	ran := false
	err := runWithUnixRecovery(filepath.Join(root, "config.json"), filepath.Join(exeDir, "remote-agent"), "existing", func() error { ran = true; return nil }, func(string, []string, []string) error { return errors.New("unexpected exec") })
	if err != nil || !ran {
		t.Fatal("ordinary startup unexpectedly required update permissions", err)
	}
	if _, err = os.Stat(filepath.Join(exeDir, "remote-agent.update.lock")); !os.IsNotExist(err) {
		t.Fatal("ordinary startup wrote to executable directory")
	}
}

func TestUnixStartupFailureAndInterruptedStartupRestorePrevious(t *testing.T) {
	key := unixReleaseKey(t)
	for _, phase := range []string{"switching", "booting"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			exe := filepath.Join(root, "remote-agent")
			config := filepath.Join(root, "config.json")
			c := Config{Path: config, ID: "fixture", Credential: "fixture-secret", Server: "https://example.test", AllowedRoots: []string{root}}
			if err := c.Save(); err != nil {
				t.Fatal(err)
			}
			old, new := []byte("previous native fixture"), []byte("candidate native fixture")
			if err := os.WriteFile(exe, new, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(exe+".previous", old, 0700); err != nil {
				t.Fatal(err)
			}
			a := &Agent{Config: c}
			op, _, err := a.maintenanceState().start(maintenanceInput{Kind: "client", RequestID: "fixture"})
			if err != nil {
				t.Fatal(err)
			}
			release := signUnixFixture(key, 5, "candidate", new)
			oldsum := sha256.Sum256(old)
			state := unixUpdateState{SchemaVersion: 1, AgentID: c.ID, Server: c.Server, Executable: exe, OperationID: op.ID, Phase: phase, PreviousSHA256: hex.EncodeToString(oldsum[:]), HighestSequence: 5, HighestSHA256: release.SHA256, Release: release}
			if err := saveUnixUpdate(config, state); err != nil {
				t.Fatal(err)
			}
			ran, executed := false, false
			err = runWithUnixRecovery(config, exe, "candidate", func() error { ran = true; return errors.New("fixture startup failure") }, func(path string, args, env []string) error {
				executed = true
				b, _ := os.ReadFile(path)
				if string(b) != string(old) {
					t.Error("rollback executed candidate")
				}
				return errors.New("fixture exec returned")
			})
			if err == nil || !executed || (phase == "booting" && ran) {
				t.Fatal("startup did not safely rollback", phase, err, ran, executed)
			}
			current, _ := os.ReadFile(exe)
			failed, _ := os.ReadFile(exe + ".failed")
			if string(current) != string(old) || string(failed) != string(new) {
				t.Fatal("previous or failed executable was not retained")
			}
			recovered, err := loadUnixUpdate(config)
			if err != nil || recovered.Phase != "rolled_back" || recovered.HighestSequence != 5 {
				t.Fatal("rollback lost release floor", err, recovered)
			}
			result, _ := (&Agent{Config: c}).maintenanceState().get(op.ID)
			if result.Status != "failed" || result.Stage != "rollback" {
				t.Fatal("rollback did not persist failure", result)
			}
		})
	}
}

func TestUnixRemoteUpdateRejectsCorruptionBusyAndOldRelease(t *testing.T) {
	key := unixReleaseKey(t)
	for _, scenario := range []string{"corrupt", "busy", "old-release", "probe-failure"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			exe := filepath.Join(root, "remote-agent")
			old := []byte("original executable")
			if err := os.WriteFile(exe, old, 0700); err != nil {
				t.Fatal(err)
			}
			body := []byte("not an executable")
			release := signUnixFixture(key, 5, "candidate", body)
			site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Error("download leaked credentials")
				}
				if scenario == "corrupt" {
					_, _ = w.Write([]byte("corrupt"))
				} else {
					_, _ = w.Write(body)
				}
			}))
			defer site.Close()
			c := Config{Path: filepath.Join(root, "config.json"), ID: "fixture", Server: site.URL, Credential: "fixture-secret", AllowedRoots: []string{root}}
			if err := c.Save(); err != nil {
				t.Fatal(err)
			}
			a := &Agent{Config: c}
			if scenario == "busy" {
				a.runs = map[string]context.CancelFunc{"active": func() {}}
			}
			op, _, err := a.maintenanceState().start(maintenanceInput{Kind: "client", RequestID: "fixture"})
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "old-release" {
				newer := signUnixFixture(key, 6, "later", []byte("later"))
				sum := sha256.Sum256(old)
				state := unixUpdateState{SchemaVersion: 1, AgentID: c.ID, Server: c.Server, Executable: exe, OperationID: op.ID, Phase: "committed", PreviousSHA256: hex.EncodeToString(sum[:]), HighestSequence: 6, HighestSHA256: newer.SHA256, Release: newer}
				if err := saveUnixUpdate(c.Path, state); err != nil {
					t.Fatal(err)
				}
			}
			executed := false
			err = a.applyUnixUpdate(context.Background(), op, release, exe, func(string, []string, []string) error { executed = true; return nil })
			if err == nil || executed {
				t.Fatal("unsafe update admitted", scenario, err)
			}
			b, _ := os.ReadFile(exe)
			if string(b) != string(old) || a.maintenance {
				t.Fatal("failure replaced original or left admission closed")
			}
		})
	}
}

// Real native daemon + WebSocket update + exec; no system service/LaunchAgent is
// registered. CI runs this on Linux and macOS with isolated profiles and keys.
func TestUnixRemoteUpdateNativeProcessReexec(t *testing.T) {
	key := unixReleaseKey(t)
	root := t.TempDir()
	exe := filepath.Join(root, "remote-agent")
	candidate := filepath.Join(root, "candidate")
	module, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, build := range []struct{ path, version string }{{exe, "native-old"}, {candidate, "native-new"}} {
		fixture := os.Getenv("REMOTE_AGENT_UNIX_TEST_OLD")
		if build.version == "native-new" {
			fixture = os.Getenv("REMOTE_AGENT_UNIX_TEST_NEW")
		}
		if fixture != "" {
			data, e := os.ReadFile(fixture)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(build.path, data, 0700); e != nil {
				t.Fatal(e)
			}
			continue
		}
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-X github.com/string1225/remote-agent-cli/internal/buildinfo.Version="+build.version+" -X github.com/string1225/remote-agent-cli/internal/clientupdate.ReleasePublicKey="+clientupdate.ReleasePublicKey, "-o", build.path, "./cmd/agent")
		cmd.Dir = module
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatal("build native fixture", err, string(out))
		}
	}
	body, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	release := signUnixFixture(key, 1, "native-new", body)
	result := make(chan MaintenanceOperation, 1)
	var initiated atomic.Bool
	var upgraded atomic.Bool
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/updates/v1/"+runtime.GOOS+"/"+runtime.GOARCH+".json" {
			if r.Header.Get("Authorization") != "" {
				t.Error("release leaked credential")
			}
			_ = json.NewEncoder(w).Encode(release)
			return
		}
		if r.URL.Path == "/downloads/"+release.Artifact() {
			if r.Header.Get("Authorization") != "" {
				t.Error("download leaked credential")
			}
			_, _ = w.Write(body)
			return
		}
		if r.URL.Path != "/api/agent/connect" || r.URL.Query().Get("id") != "native-device" || r.Header.Get("Authorization") != "Bearer fixture-private" {
			http.Error(w, "denied", 401)
			return
		}
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		request := func(kind string, body any) error {
			raw, _ := json.Marshal(body)
			return conn.WriteJSON(protocol.Signal{Type: "request", PeerID: "fixture-peer", Request: &protocol.Request{ID: kind, Type: kind, Body: raw}})
		}
		opened := false
		for {
			_ = conn.SetReadDeadline(time.Now().Add(40 * time.Second))
			var signal protocol.Signal
			if conn.ReadJSON(&signal) != nil {
				return
			}
			if signal.Type == "heartbeat" {
				if conn.WriteJSON(protocol.Signal{Type: "pong"}) != nil {
					return
				}
				if !opened {
					opened = true
					_ = conn.WriteJSON(protocol.Signal{Type: "open", PeerID: "fixture-peer", Transport: protocol.RelayTransport, LeaseSeconds: 120})
				}
				continue
			}
			if signal.Type == "ready" {
				if request("services.list", nil) != nil {
					return
				}
			}
			if signal.Event == nil {
				continue
			}
			event := signal.Event
			if event.Type == "error" {
				select {
				case result <- MaintenanceOperation{Status: "failed", Message: event.Error}:
				default:
				}
				return
			}
			if event.ID == "services.list" {
				data, _ := event.Data.(map[string]any)
				if data["clientVersion"] == "native-old" && initiated.CompareAndSwap(false, true) {
					if request("maintenance.start", maintenanceInput{Kind: "client", RequestID: "native-update"}) != nil {
						return
					}
				}
				if data["clientVersion"] == "native-new" {
					upgraded.Store(true)
					if request("maintenance.status", nil) != nil {
						return
					}
				}
			}
			if event.ID == "maintenance.status" {
				raw, _ := json.Marshal(event.Data)
				var op MaintenanceOperation
				_ = json.Unmarshal(raw, &op)
				if op.Status == "succeeded" || op.Status == "failed" {
					select {
					case result <- op:
					default:
					}
					return
				}
				time.Sleep(30 * time.Millisecond)
				if request("maintenance.status", nil) != nil {
					return
				}
			}
		}
	}))
	defer site.Close()
	workspace := filepath.Join(root, "workspace")
	if err = os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	c := Config{Path: filepath.Join(root, "config.json"), ID: "native-device", Server: site.URL, Credential: "fixture-private", AllowedRoots: []string{workspace}, Services: []protocol.Service{{ID: "project", Provider: "codex", Name: "Keep", Workspace: workspace}}, AllowWrite: true, PermissionMode: "full"}
	if err = c.Save(); err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(c.Path)
	h := (&Agent{Config: c}).historyStore()
	conversation, err := h.create(c.Services[0], "retained session")
	if err != nil {
		t.Fatal(err)
	}
	recording := filepath.Join(root, "recording.webm")
	_ = os.WriteFile(recording, []byte("retained recording"), 0600)
	logPath := filepath.Join(root, "agent.log")
	cmd := exec.Command(exe, "run", "--config", c.Path, "--log", logPath)
	cmd.Env = append(os.Environ(), "HOME="+root, "XDG_CONFIG_HOME="+root, "CODEX_HOME="+filepath.Join(root, ".codex"), "RA_CODEX_PATH="+filepath.Join(root, "missing"), "RA_QODER_PATH="+filepath.Join(root, "missing-qoder"))
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait(); close(done) }()
	t.Cleanup(func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(12 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	})
	select {
	case op := <-result:
		if op.Status != "succeeded" || op.BeforeVersion != "native-old" || op.AfterVersion != "native-new" || !upgraded.Load() {
			t.Fatalf("native remote update failed: %+v", op)
		}
	case err := <-done:
		t.Fatalf("native daemon exited before update: %v", err)
	case <-time.After(75 * time.Second):
		log, _ := os.ReadFile(logPath)
		t.Fatal("native update timed out", string(log))
	}
	log, _ := os.ReadFile(logPath)
	for _, version := range []string{"native-old", "native-new"} {
		pattern := regexp.MustCompile(`version ` + version + `, pid ([0-9]+)`)
		match := pattern.FindStringSubmatch(string(log))
		if len(match) != 2 || match[1] != fmt.Sprint(cmd.Process.Pid) {
			t.Fatal("update did not preserve supervisor PID", string(log))
		}
	}
	saved, _ := os.ReadFile(c.Path)
	if string(saved) != string(original) {
		t.Fatal("update changed binding/configuration")
	}
	if _, err = h.get(conversation.ID); err != nil {
		t.Fatal("update lost session", err)
	}
	if b, _ := os.ReadFile(recording); string(b) != "retained recording" {
		t.Fatal("update lost recording")
	}
	state, err := loadUnixUpdate(c.Path)
	if err != nil || state.Phase != "committed" {
		t.Fatal("native update did not commit", err, state)
	}
	if _, err = os.Stat(exe + ".previous"); err != nil {
		t.Fatal("previous binary not retained", err)
	}
	if strings.Contains(string(log), "fixture-private") {
		t.Fatal("credential leaked to logs")
	}
}
