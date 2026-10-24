package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func TestUpdateCLI(t *testing.T) {
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	dir := t.TempDir()
	exe, release := filepath.Join(dir, "agent.exe"), filepath.Join(dir, "release.exe")
	for path, version := range map[string]string{exe: "fixture-old", release: "fixture-new"} {
		// Allow native tests on a cross-compiled target that has no Go toolchain.
		if fixture := os.Getenv("REMOTE_AGENT_TEST_" + strings.ToUpper(strings.TrimPrefix(version, "fixture-")) + "_CLI"); fixture != "" {
			data, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0700); err != nil {
				t.Fatal(err)
			}
			continue
		}
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-X github.com/string1225/remote-agent-cli/internal/buildinfo.Version="+version, "-o", path, ".")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build CLI fixture: %v\n%s", err, out)
		}
	}
	old, _ := os.ReadFile(exe)
	latest, _ := os.ReadFile(release)
	name := "remote-agent-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	var downloads atomic.Int32
	var enrollments atomic.Int32
	var allowEnrollment atomic.Bool
	var corrupt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/enroll") {
			enrollments.Add(1)
			if allowEnrollment.Load() {
				var fields map[string]string
				if err := json.NewDecoder(r.Body).Decode(&fields); err != nil || fields["token"] != strings.Repeat("A", 43) || fields["os"] != runtime.GOOS || fields["arch"] != runtime.GOARCH {
					t.Error("invalid first enrollment", err)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"id":"first-install","name":"Linux server","credential":"fixture-only"}`)
				return
			}
		}
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("download included credentials")
		}
		switch r.URL.Path {
		case "/agents/downloads/" + name + ".sha256":
			if corrupt.Load() {
				fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), name)
			} else {
				fmt.Fprintf(w, "%x  %s\n", sha256.Sum256(latest), name)
			}
		case "/agents/downloads/" + name:
			downloads.Add(1)
			_, _ = w.Write(latest)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	config := filepath.Join(dir, "config.json")
	binding, _ := json.Marshal(map[string]any{"server": server.URL + "/agents", "id": "fixture-device", "name": "央玺服务器", "credential": "never-send-this-credential", "allowedRoots": []string{dir}, "allowWrite": false})
	if err := os.WriteFile(config, binding, 0600); err != nil {
		t.Fatal(err)
	}
	history := filepath.Join(dir, "history.db")
	if err := os.WriteFile(history, []byte("saved conversations"), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) (string, error) {
		cmd := exec.Command(exe, args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	before, _ := os.ReadDir(dir)
	if out, err := run("update", "--check", "--config", config); err != nil || !strings.Contains(out, "available") {
		t.Fatalf("check update: %v\n%s", err, out)
	}
	after, _ := os.ReadDir(dir)
	unchanged, _ := os.ReadFile(exe)
	if downloads.Load() != 0 || len(before) != len(after) || sha256.Sum256(unchanged) != sha256.Sum256(old) {
		t.Fatal("--check downloaded a binary or modified local files")
	}
	if out, err := run("update", "--config", config); err != nil || !strings.Contains(out, "fixture-new") {
		t.Fatalf("install update: %v\n%s", err, out)
	}
	if downloads.Load() != 1 {
		t.Fatal("unexpected download count")
	}
	installed, _ := os.ReadFile(exe)
	backup, _ := os.ReadFile(exe + ".previous")
	if sha256.Sum256(installed) != sha256.Sum256(latest) || sha256.Sum256(backup) != sha256.Sum256(old) {
		t.Fatal("replacement or backup bytes differ")
	}
	if out, err := run("version"); err != nil || !strings.Contains(out, "fixture-new") {
		t.Fatalf("updated version: %v\n%s", err, out)
	}
	if out, err := run("update", "--config", config); err != nil || !strings.Contains(out, "Already up to date") || downloads.Load() != 1 {
		t.Fatalf("repeat update: %v\n%s", err, out)
	}
	corrupt.Store(true)
	if out, err := run("update", "--config", config); err == nil || !strings.Contains(out, "checksum mismatch") {
		t.Fatalf("corrupt update was not rejected: %v\n%s", err, out)
	}
	installed, _ = os.ReadFile(exe)
	retainedBinding, _ := os.ReadFile(config)
	retainedHistory, _ := os.ReadFile(history)
	if sha256.Sum256(installed) != sha256.Sum256(latest) || string(retainedBinding) != string(binding) || string(retainedHistory) != "saved conversations" {
		t.Fatal("update changed binding/history or damaged the installed binary")
	}
	// Reinstall must bootstrap a legacy CLI that has no update command, without
	// redeeming an old enrollment token or changing its existing account binding.
	corrupt.Store(false)
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	installedName := "remote-agent"
	if runtime.GOOS == "windows" {
		installedName += ".exe"
	}
	standard := filepath.Join(bin, installedName)
	legacy := []byte("legacy binary without an update command")
	if err := os.WriteFile(standard, legacy, 0700); err != nil {
		t.Fatal(err)
	}
	wrong := exec.Command(release, "update", "--installer", "--config", config, "--server", server.URL+"/another-server")
	if output, err := wrong.CombinedOutput(); err == nil || !strings.Contains(string(output), "different server") {
		t.Fatalf("installer accepted a different server: %v\n%s", err, output)
	}
	var installerCommand *exec.Cmd
	if runtime.GOOS == "linux" {
		scriptName := "linux.sh"
		body, err := os.ReadFile(filepath.Join("..", "..", "internal", "server", "installers", scriptName))
		if err != nil {
			t.Fatal(err)
		}
		script := strings.ReplaceAll(strings.ReplaceAll(string(body), "__SERVER__", server.URL+"/agents"), "__TOKEN__", strings.Repeat("A", 43))
		script = strings.Replace(script, `root="$HOME/.remote-agent"`, "root='"+strings.ReplaceAll(dir, "'", "'\"'\"'")+"'", 1)
		path := filepath.Join(dir, scriptName)
		if err := os.WriteFile(path, []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
		installerCommand = exec.Command("/bin/bash", path)
		// Exercise the script only with an isolated, mocked service manager.
		fakeBin := filepath.Join(dir, "fake-systemd")
		if err := os.Mkdir(fakeBin, 0700); err != nil {
			t.Fatal(err)
		}
		for name, body := range map[string]string{"systemctl": "#!/bin/sh\nexit 0\n", "loginctl": "#!/bin/sh\necho yes\n"} {
			if err := os.WriteFile(filepath.Join(fakeBin, name), []byte(body), 0700); err != nil {
				t.Fatal(err)
			}
		}
		installerCommand.Env = append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(dir, "xdg"), "PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	} else {
		// Real Windows/macOS installers alter the machine's Scheduled Tasks or
		// LaunchAgents. Test their shared updater here without enabling autostart;
		// actual MSI/SCM installation has its own opt-in disposable-runner suite.
		installerCommand = exec.Command(release, "update", "--installer", "--preserve-start", "--config", config, "--server", server.URL+"/agents")
	}
	beforeInstallerDownloads := downloads.Load()
	if output, err := installerCommand.CombinedOutput(); err != nil || !strings.Contains(string(output), "fixture-new") {
		t.Fatalf("reinstall failed to upgrade a legacy CLI: %v\n%s", err, output)
	}
	standardBytes, _ := os.ReadFile(standard)
	legacyBackup, _ := os.ReadFile(standard + ".previous")
	if sha256.Sum256(standardBytes) != sha256.Sum256(latest) || string(legacyBackup) != string(legacy) {
		t.Fatal("reinstall did not preserve the previous binary or install the release")
	}
	wantDownloads := beforeInstallerDownloads
	if runtime.GOOS == "linux" {
		wantDownloads++
	}
	if enrollments.Load() != 0 || downloads.Load() != wantDownloads {
		t.Fatal("reinstall redeemed an enrollment token or downloaded the binary twice")
	}
	retainedBinding, _ = os.ReadFile(config)
	retainedHistory, _ = os.ReadFile(history)
	if string(retainedBinding) != string(binding) || string(retainedHistory) != "saved conversations" {
		t.Fatal("reinstall changed device binding or conversations")
	}
	if runtime.GOOS == "linux" {
		allowEnrollment.Store(true)
		body, err := os.ReadFile(filepath.Join("..", "..", "internal", "server", "installers", "linux.sh"))
		if err != nil {
			t.Fatal(err)
		}
		installFresh := func(root string) error {
			script := strings.ReplaceAll(strings.ReplaceAll(string(body), "__SERVER__", server.URL+"/agents"), "__TOKEN__", strings.Repeat("A", 43))
			script = strings.Replace(script, `root="$HOME/.remote-agent"`, "root='"+strings.ReplaceAll(root, "'", "'\"'\"'")+"'", 1)
			path := filepath.Join(root, "linux.sh")
			if err := os.WriteFile(path, []byte(script), 0700); err != nil {
				return err
			}
			cmd := exec.Command("/bin/bash", path)
			cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+filepath.Join(root, "xdg"), "PATH="+filepath.Join(dir, "fake-systemd")+string(os.PathListSeparator)+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%w: %s", err, out)
			}
			return nil
		}
		fresh := t.TempDir()
		if err := installFresh(fresh); err != nil {
			t.Fatal("first Linux installation:", err)
		}
		data, err := os.ReadFile(filepath.Join(fresh, "config.json"))
		if err != nil || !strings.Contains(string(data), "first-install") || enrollments.Load() != 1 {
			t.Fatalf("first installation did not bind once: %v", err)
		}
		if _, err := os.Stat(filepath.Join(fresh, "xdg", "systemd", "user", "remote-agent.service")); err != nil {
			t.Fatal("no systemd user service:", err)
		}
		corrupt.Store(true)
		bad := t.TempDir()
		if err := installFresh(bad); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatal("bad installer checksum accepted", err)
		}
		if enrollments.Load() != 1 {
			t.Fatal("bad checksum consumed an enrollment token")
		}
		if _, err := os.Stat(filepath.Join(bad, "config.json")); !os.IsNotExist(err) {
			t.Fatal("bad download created an agent binding")
		}
	}
}
