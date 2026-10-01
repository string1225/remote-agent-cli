package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestUpdateDownloadVerification(t *testing.T) {
	data := []byte("a complete release binary")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	name := updateArtifact()
	var binaryRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("download leaked authentication")
		}
		switch r.URL.Path {
		case "/agents/downloads/" + name + ".sha256":
			fmt.Fprintf(w, "%s  %s\n", digest, name)
		case "/agents/downloads/" + name:
			binaryRequests.Add(1)
			_, _ = w.Write(data)
		case "/agents/wrong-name":
			fmt.Fprintf(w, "%s  other-file\n", digest)
		case "/agents/redirect":
			http.Redirect(w, r, "/agents/downloads/"+name, http.StatusFound)
		case "/agents/too-large":
			w.Header().Set("Content-Length", fmt.Sprint(maxUpdateSize+1))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ctx := context.Background()
	expected, err := updateChecksum(ctx, client, server.URL+"/agents", name)
	if err != nil || expected != digest || binaryRequests.Load() != 0 {
		t.Fatalf("checksum-only check: %s, %v, requests=%d", expected, err, binaryRequests.Load())
	}
	dir := t.TempDir()
	path, err := downloadUpdate(ctx, client, server.URL+"/agents/downloads/"+name, dir, expected)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, data) {
		t.Fatal("download did not preserve bytes")
	}
	_ = os.Remove(path)
	for _, test := range []struct{ url, checksum string }{
		{"/agents/downloads/" + name, strings.Repeat("0", 64)},
		{"/agents/redirect", digest},
		{"/agents/too-large", digest},
		{"/missing", digest},
	} {
		if _, err := downloadUpdate(ctx, client, server.URL+test.url, dir, test.checksum); err == nil {
			t.Errorf("accepted invalid download %s", test.url)
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Fatal("failed download left staged files")
		}
	}
	if _, err := updateChecksum(ctx, client, server.URL, "../wrong-name"); err == nil {
		t.Fatal("invalid manifest accepted")
	}
	for _, body := range []string{"", digest, digest + "  other-file", strings.Repeat("x", 1025)} {
		bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
		_, err := updateChecksum(ctx, client, bad.URL, name)
		bad.Close()
		if err == nil {
			t.Errorf("accepted malformed manifest %q", body[:min(len(body), 80)])
		}
	}
}

func TestUpdateReplacementAndRollback(t *testing.T) {
	for _, scenario := range []string{"success", "restart-failure", "replacement-failure", "busy"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			exe, staged := filepath.Join(dir, "agent"), filepath.Join(dir, "staged")
			for path, data := range map[string]string{exe: "old", staged: "new", exe + ".previous": "older", filepath.Join(dir, "config.json"): "binding", filepath.Join(dir, "history.db"): "history"} {
				if err := os.WriteFile(path, []byte(data), 0700); err != nil {
					t.Fatal(err)
				}
			}
			starts := 0
			service := serviceUpdate{managed: true, stop: func() error {
				if scenario == "busy" {
					return fmt.Errorf("task is active")
				}
				return nil
			}, start: func() error {
				starts++
				if scenario == "restart-failure" && starts == 1 {
					return fmt.Errorf("failed to start new agent")
				}
				return nil
			}}
			if scenario == "replacement-failure" {
				_ = os.Remove(staged)
			}
			err := replaceUpdate(exe, staged, service)
			if (scenario == "success") != (err == nil) {
				t.Fatalf("unexpected result: %v", err)
			}
			want := "old"
			if scenario == "success" {
				want = "new"
				backup, _ := os.ReadFile(exe + ".previous")
				if string(backup) != "old" {
					t.Fatal("missing previous binary backup")
				}
			}
			actual, _ := os.ReadFile(exe)
			if string(actual) != want {
				t.Fatalf("installed binary: %q, want %q", actual, want)
			}
			if scenario == "busy" {
				backup, _ := os.ReadFile(exe + ".previous")
				if string(backup) != "older" {
					t.Fatal("busy update changed previous backup")
				}
			}
			for path, want := range map[string]string{"config.json": "binding", "history.db": "history"} {
				data, _ := os.ReadFile(filepath.Join(dir, path))
				if string(data) != want {
					t.Fatalf("update touched %s", path)
				}
			}
		})
	}
}

func TestUpdateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update.lock")
	unlock, err := lockUpdate(path)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := lockUpdate(path); err == nil {
		second()
		t.Fatal("concurrent updater obtained the lock")
	}
	unlock()
	unlock, err = lockUpdate(path)
	if err != nil {
		t.Fatal("released lock cannot be obtained", err)
	}
	unlock()
}

// Exercise replacing a loaded executable, including Windows image-file locks.
func TestUpdateRunningExecutable(t *testing.T) {
	if os.Getenv("REMOTE_AGENT_UPDATE_HELPER") == "1" {
		exe, _ := os.Executable()
		service := serviceUpdate{stop: func() error { return nil }, start: func() error { return nil }}
		if err := replaceUpdate(exe, os.Getenv("REMOTE_AGENT_UPDATE_CANDIDATE"), service); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	dir := t.TempDir()
	exe, _ := os.Executable()
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	installed, staged := filepath.Join(dir, "agent.exe"), filepath.Join(dir, "staged.exe")
	if err := os.WriteFile(installed, data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("replacement"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(installed, "-test.run=^TestUpdateRunningExecutable$")
	cmd.Env = append(os.Environ(), "REMOTE_AGENT_UPDATE_HELPER=1", "REMOTE_AGENT_UPDATE_CANDIDATE="+staged)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("loaded executable update failed: %v\n%s", err, output)
	}
	got, _ := os.ReadFile(installed)
	if string(got) != "replacement" {
		t.Fatal("loaded executable was not replaced")
	}
	backup, _ := fileDigest(installed + ".previous")
	if backup != fmt.Sprintf("%x", sha256.Sum256(data)) {
		t.Fatal("loaded executable backup differs")
	}
}
