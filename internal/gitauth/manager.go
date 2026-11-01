package gitauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Version = 1

// Request intentionally cannot carry passwords, arbitrary URLs or commands.
type Request struct {
	Version  int    `json:"version"`
	Action   string `json:"action"`
	Username string `json:"username,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

func (r Request) Validate() error {
	if r.Version != 0 && r.Version != Version {
		return errors.New("unsupported Git authorization protocol")
	}
	switch r.Action {
	case "login":
	case "status":
		if r.Username != "" || r.Force {
			return errors.New("git status does not accept an account or --force")
		}
	case "logout":
		if r.Username == "" || r.Force {
			return errors.New("git logout requires --username; use git status to list accounts")
		}
	default:
		return errors.New("use git login/status/logout")
	}
	if len(r.Username) > 39 {
		return errors.New("invalid GitHub username")
	}
	for i, c := range r.Username {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' && i > 0) {
			return errors.New("invalid GitHub username")
		}
	}
	return nil
}

type Manager struct {
	GCM   string
	Store string
	Env   []string
}

func findGCM(ctx context.Context) (string, error) {
	if path, err := exec.LookPath("git-credential-manager"); err == nil {
		return path, nil
	}
	git, err := exec.LookPath("git")
	if err != nil {
		return "", errors.New("install Git for Windows with Git Credential Manager, then retry remote-agent git login")
	}
	out, err := newCommand(ctx, git, "--exec-path").Output()
	if err == nil {
		root := strings.TrimSpace(string(out))
		for _, path := range []string{filepath.Join(root, "git-credential-manager.exe"), filepath.Join(root, "..", "..", "bin", "git-credential-manager.exe")} {
			if info, e := os.Stat(path); e == nil && !info.IsDir() {
				return filepath.Clean(path), nil
			}
		}
	}
	return "", errors.New("install Git for Windows with Git Credential Manager, then retry remote-agent git login")
}

func (m Manager) Run(ctx context.Context, request Request, output io.Writer) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if !filepath.IsAbs(m.Store) {
		return errors.New("Git credential store must be an absolute path")
	}
	path := m.GCM
	if path == "" {
		var err error
		path, err = findGCM(ctx)
		if err != nil {
			return err
		}
	}
	args := []string{"github"}
	switch request.Action {
	case "login":
		args = append(args, "login", "--device", "--no-ui")
		if request.Username != "" {
			args = append(args, "--username", request.Username)
		}
		if request.Force {
			args = append(args, "--force")
		}
	case "status":
		args = append(args, "list", "--no-ui")
	case "logout":
		args = append(args, "logout", request.Username, "--no-ui")
	}
	cmd := newCommand(ctx, path, args...)
	cmd.Dir = m.Store
	cmd.Env = Environment(m.Env, m.Store, request.Action == "login")
	cmd.Stdout, cmd.Stderr = output, output
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("GitHub %s failed; check the authorization output and retry", request.Action)
	}
	return nil
}
