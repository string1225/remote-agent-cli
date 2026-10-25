package agent

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemoteMaintenanceMetadataAvailableDuringInstallAndSnapshotIsStable(t *testing.T) {
	a := &Agent{toolInstalling: true, Providers: map[string]Provider{"codex": {Name: "codex", Path: "before"}}}
	var event protocol.Event
	a.handle(nil, protocol.Request{ID: "metadata", Type: "services.list"}, func(v protocol.Event) error { event = v; return nil })
	if event.Type != "result" {
		t.Fatal("maintenance UI cannot reopen during installation", event)
	}
	a.mu.Lock()
	a.Providers["codex"] = Provider{Path: "after"}
	a.mu.Unlock()
	providers := event.Data.(map[string]any)["providers"].(map[string]Provider)
	if providers["codex"].Path != "before" {
		t.Fatal("response exposed mutable provider map")
	}
}

func TestRemoteMaintenanceDoesNotAdvanceWhenPersistenceFails(t *testing.T) {
	m := &maintenanceState{agentID: "fixture"}
	op, _, err := m.start(maintenanceInput{Kind: "client", RequestID: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(parent, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	m.path = filepath.Join(parent, "operations.json")
	if m.change(op.ID, func(v *MaintenanceOperation) { v.ExpectedVersion = "unsafe" }) == nil {
		t.Fatal("persistence failure ignored")
	}
	got, _ := m.get(op.ID)
	if got.ExpectedVersion != "" {
		t.Fatal("unsaved version exposed as durable")
	}
}

func TestRemoteMaintenanceIsDurableIdempotentAndDeviceScoped(t *testing.T) {
	root := t.TempDir()
	m := &maintenanceState{agentID: "client-a", path: filepath.Join(root, "operations.json")}
	first, start, e := m.start(maintenanceInput{Kind: "codex", RequestID: "request"})
	if e != nil || !start {
		t.Fatal(e)
	}
	same, start, e := m.start(maintenanceInput{Kind: "codex", RequestID: "request"})
	if e != nil || start || same.ID != first.ID {
		t.Fatal("retry duplicated install")
	}
	if _, _, e = m.start(maintenanceInput{Kind: "client", RequestID: "another"}); e == nil {
		t.Fatal("parallel maintenance admitted")
	}
	m.change(first.ID, func(v *MaintenanceOperation) { v.Status = "succeeded" })
	if _, _, e = m.start(maintenanceInput{Kind: "codex", RequestID: "../../escape"}); e == nil {
		t.Fatal("unsafe id accepted")
	}
	a := &Agent{Config: Config{ID: "client-b", Path: filepath.Join(root, "config.json")}}
	raw, _ := os.ReadFile(m.path)
	_ = os.WriteFile(filepath.Join(root, "maintenance.json"), raw, 0600)
	op, e := a.maintenanceState().get("")
	if e != nil || op.ID != "" {
		t.Fatal("another binding saw maintenance history")
	}
	a.toolInstalling = true
	if e = a.PauseUpdates(context.Background(), true); e == nil {
		t.Fatal("force interrupted a tool installation")
	}
}

func TestRemoteCodexInstallActivatesManagedToolWithoutRestart(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "fake.go")
	exe := filepath.Join(root, "fake")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	_ = os.WriteFile(source, []byte("package main\nimport(\"fmt\";\"os\")\nfunc main(){if len(os.Args)>1&&os.Args[1]==\"--version\"{fmt.Println(\"codex-cli 1.2.3\")}else{fmt.Println(\"exec\")}}"), 0600)
	if out, e := exec.Command("go", "build", "-o", exe, source).CombinedOutput(); e != nil {
		t.Fatal(e, string(out))
	}
	binary, _ := os.ReadFile(exe)
	name, e := codexsetup.AssetName(runtime.GOOS, runtime.GOARCH)
	if e != nil {
		t.Fatal(e)
	}
	entries := map[string][]byte{}
	if runtime.GOOS == "windows" {
		entries[strings.TrimSuffix(name, ".zip")] = binary
		for _, f := range []string{"codex-code-mode-host.exe", "codex-command-runner.exe", "codex-windows-sandbox-setup.exe", "codex-path/rg.exe"} {
			entries[f] = []byte("fixture")
		}
	} else {
		entries["bin/codex"] = binary
		entries["bin/codex-code-mode-host"] = []byte("fixture")
		entries["codex-path/rg"] = []byte("fixture")
		if runtime.GOOS == "linux" {
			entries["bwrap"] = []byte("fixture")
		}
	}
	var archive bytes.Buffer
	if runtime.GOOS == "windows" {
		w := zip.NewWriter(&archive)
		for n, b := range entries {
			f, _ := w.Create(n)
			_, _ = f.Write(b)
		}
		_ = w.Close()
	} else {
		gz := gzip.NewWriter(&archive)
		w := tar.NewWriter(gz)
		for n, b := range entries {
			_ = w.WriteHeader(&tar.Header{Name: n, Mode: 0700, Size: int64(len(b))})
			_, _ = w.Write(b)
		}
		_ = w.Close()
		_ = gz.Close()
	}
	sum := sha256.Sum256(archive.Bytes())
	var site *httptest.Server
	site = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("credentials leaked to download")
		}
		if r.URL.Path == "/install/codex-release" {
			_ = json.NewEncoder(w).Encode(codexsetup.Release{Assets: []codexsetup.Asset{{Name: name, Digest: "sha256:" + hex.EncodeToString(sum[:]), URL: site.URL + "/downloads/codex/fixture"}}})
			return
		}
		_, _ = w.Write(archive.Bytes())
	}))
	defer site.Close()
	config := Config{Path: filepath.Join(root, "config.json"), Server: site.URL, ID: "fixture", Credential: "fixture-private", AllowedRoots: []string{root}}
	if e = config.Save(); e != nil {
		t.Fatal(e)
	}
	a := &Agent{Config: config, Providers: map[string]Provider{}}
	m := a.maintenanceState()
	op, _, e := m.start(maintenanceInput{Kind: "codex", RequestID: "install"})
	if e != nil {
		t.Fatal(e)
	}
	a.runMaintenance(context.Background(), op)
	result, _ := m.get(op.ID)
	if result.Status != "succeeded" || result.AfterVersion != "1.2.3" || !a.Providers["codex"].Available || a.toolInstalling {
		t.Fatalf("remote installation failed: %+v", result)
	}
	saved, e := Load(config.Path)
	if e != nil || !saved.ManagedCodex || saved.Credential != config.Credential {
		t.Fatal("binding changed or managed choice not saved", e)
	}
	t.Setenv("RA_CODEX_PATH", filepath.Join(root, "missing"))
	t.Setenv("RA_QODER_PATH", filepath.Join(root, "missing-qoder"))
	if !New(saved).Providers["codex"].Available {
		t.Fatal("restart lost managed Codex selection")
	}
}
