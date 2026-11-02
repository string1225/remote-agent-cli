package gitauth

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("RA_GIT_AUTH_TEST_HELPER") == "1" {
		if os.Getenv("GCM_CREDENTIAL_STORE") != "dpapi" || os.Getenv("GCM_DPAPI_STORE_PATH") == "" {
			os.Exit(9)
		}
		args := strings.Join(os.Args[1:], " ")
		switch {
		case strings.Contains(args, "login"):
			if os.Getenv("GCM_INTERACTIVE") != "always" || os.Getenv("GIT_TERMINAL_PROMPT") != "1" || !strings.Contains(args, "--device --no-ui") {
				os.Exit(8)
			}
			fmt.Println("Open https://github.com/login/device and enter TEST-CODE")
			if strings.Contains(args, "--username wait") {
				time.Sleep(time.Minute)
			}
			if strings.Contains(args, "--username fail") {
				fmt.Fprintln(os.Stderr, "simulated authorization declined")
				os.Exit(1)
			}
		case strings.Contains(args, "list"):
			fmt.Println("fixture-user")
		case strings.Contains(args, "logout fixture-user"):
			fmt.Println("Removed fixture-user")
		default:
			os.Exit(7)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fixtureManager(t *testing.T) Manager {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return Manager{GCM: exe, Store: t.TempDir(), Env: append(os.Environ(), "RA_GIT_AUTH_TEST_HELPER=1")}
}

func TestManagerRunsFixedCommandsAndStreamsDeviceCode(t *testing.T) {
	m := fixtureManager(t)
	// The service launches background jobs with prompting disabled. Explicit
	// login must override that inherited policy for the device-code flow.
	m.Env = Environment(m.Env, m.Store, false)
	for _, req := range []Request{{Action: "login"}, {Action: "status"}, {Action: "logout", Username: "fixture-user"}} {
		var output bytes.Buffer
		if err := m.Run(context.Background(), req, &output); err != nil {
			t.Fatal(err, output.String())
		}
		if output.Len() == 0 {
			t.Fatal("missing account or device-code output")
		}
		if req.Action == "login" && !strings.Contains(output.String(), "TEST-CODE") {
			t.Fatal(output.String())
		}
	}
}

func TestManagerRejectsInvalidRequestsBeforeStartingGCM(t *testing.T) {
	m := Manager{GCM: filepath.Join(t.TempDir(), "must-not-run"), Store: t.TempDir()}
	for _, req := range []Request{{Action: "get"}, {Action: "logout"}, {Action: "login", Username: "--help"}, {Action: "login", Username: "x\npassword=secret"}, {Action: "status", Force: true}, {Action: "status", Username: "someone"}} {
		if err := m.Run(context.Background(), req, &bytes.Buffer{}); err == nil || strings.Contains(err.Error(), "executable") {
			t.Fatalf("request was not rejected: %+v: %v", req, err)
		}
	}
}

func TestFailedAndCancelledLoginReturnErrors(t *testing.T) {
	m := fixtureManager(t)
	var output bytes.Buffer
	if err := m.Run(context.Background(), Request{Action: "login", Username: "fail"}, &output); err == nil {
		t.Fatal("declined authorization reported success")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := m.Run(ctx, Request{Action: "login", Username: "wait"}, &bytes.Buffer{}); err == nil {
		t.Fatal("cancelled login reported success")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("cancelled helper was not terminated promptly")
	}
}

func TestManagedEnvironmentIsPrivateAndPreservesGitSettings(t *testing.T) {
	root := t.TempDir()
	base := []string{"PATH=fixture", "GCM_CREDENTIAL_STORE=wincredman", "gcm_dpapi_store_path=desktop", "GCM_INTERACTIVE=always", "GIT_CONFIG_PARAMETERS='user.name=Fixture User'", "UNCHANGED=yes", "GCM_TRACE=fixture.log", "GCM_TRACE_SECRETS=1", "GIT_TRACE_CURL=1"}
	env := Environment(base, root, false)
	joined := strings.Join(env, "\n")
	for _, want := range []string{"GCM_CREDENTIAL_STORE=dpapi", "GCM_DPAPI_STORE_PATH=" + root, "GCM_INTERACTIVE=never", "GIT_TERMINAL_PROMPT=0", "'user.name=Fixture User'", "UNCHANGED=yes"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(joined, "wincredman") || strings.Contains(joined, "=desktop") {
		t.Fatal("desktop credential store leaked into service environment")
	}
	if strings.Contains(joined, "TRACE") {
		t.Fatal("authorization could be written to persistent trace logs")
	}
	if base[1] != "GCM_CREDENTIAL_STORE=wincredman" {
		t.Fatal("caller environment modified")
	}
}

func TestManagedEnvironmentKeepsRealGitConfiguration(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git is not installed")
	}
	root := t.TempDir()
	config := filepath.Join(root, "gitconfig")
	if err := os.WriteFile(config, []byte("[user]\n name = Fixture User\n email = fixture@example.test\n[credential]\n helper = desktop-fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	env := Environment(append(os.Environ(), "GIT_CONFIG_GLOBAL="+config, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_PARAMETERS="), filepath.Join(root, "store"), false)
	for _, test := range []struct {
		args []string
		want string
	}{
		{[]string{"config", "--get", "user.name"}, "Fixture User"},
		{[]string{"config", "--get-urlmatch", "credential.helper", "https://github.com/fixture/repo"}, "manager"},
		{[]string{"config", "--get-urlmatch", "credential.helper", "https://gitlab.com/fixture/repo"}, "desktop-fixture"},
	} {
		cmd := exec.Command(git, test.args...)
		cmd.Dir, cmd.Env = root, env
		out, err := cmd.CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) != test.want {
			t.Fatalf("%v: %s (%v)", test.args, out, err)
		}
	}
}

func TestManagedEnvironmentDisablesConfiguredAskpass(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git is not installed")
	}
	root := t.TempDir()
	marker := filepath.Join(root, "unexpected-askpass")
	askpass := filepath.Join(root, "askpass.sh")
	script := "#!/bin/sh\nprintf called > '" + filepath.ToSlash(marker) + "'\nprintf fixture-password\n"
	if err := os.WriteFile(askpass, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(root, "gitconfig")
	if err := os.WriteFile(config, []byte("[core]\n askPass = "+filepath.ToSlash(askpass)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, git, "-c", "credential.helper=", "credential", "fill")
	cmd.Dir = root
	cmd.Env = Environment(append(os.Environ(), "GIT_ASKPASS="+askpass, "SSH_ASKPASS="+askpass, "GIT_CONFIG_GLOBAL="+config, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_PARAMETERS="), filepath.Join(root, "store"), false)
	cmd.Stdin = strings.NewReader("protocol=https\nhost=fixture.invalid\nusername=fixture-user\n\n")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Git invoked an interactive askpass helper")
	}
	if !strings.Contains(string(out), "remote-agent git login") {
		t.Fatalf("missing task authorization guidance: %s (%v)", out, err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("configured askpass was started", err)
	}
}
