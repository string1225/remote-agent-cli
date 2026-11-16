package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

type Provider struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error,omitempty"`
}

func Detect() map[string]Provider {
	providers := map[string]Provider{}
	for _, name := range []string{"codex", "qoder"} {
		p := Provider{Name: name}
		candidates := []string{name}
		if name == "codex" {
			home, _ := os.UserHomeDir()
			executable := "codex"
			if runtime.GOOS == "windows" {
				executable += ".exe"
			}
			candidates = []string{filepath.Join(home, ".remote-agent", "bin", executable), name}
		}
		if name == "qoder" {
			candidates = []string{"qodercli", "qoder"}
		}
		if path := os.Getenv("RA_" + strings.ToUpper(name) + "_PATH"); path != "" {
			candidates = []string{path}
		}
		for _, candidate := range candidates {
			path, err := exec.LookPath(candidate)
			if err != nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			cmd := command(ctx, path, []string{"--help"})
			var output boundedBuffer
			cmd.Stdout = &output
			cmd.Stderr = &output
			err = cmd.Run()
			cancel()
			help := output.String()
			expected := "exec"
			if name == "qoder" {
				expected = "--print"
			}
			if err == nil && strings.Contains(help, expected) {
				p.Available = true
				p.Path = path
				break
			}
			p.Error = "command found, but it is not a compatible CLI (the Qoder editor launcher is not Qoder CLI)"
		}
		if !p.Available && p.Error == "" {
			p.Error = providerSetupError(name).Error()
		} else if !p.Available {
			p.Error += "; " + providerSetupError(name).Error()
		}
		providers[name] = p
	}
	return providers
}

func providerSetupError(name string) error {
	if name == "codex" {
		return errors.New("Codex CLI 未安装或不可用。请在目标客户端所属用户的终端运行 remote-agent codex setup，自动安装并配置 Codex；仅安装可运行 remote-agent codex install（无需 --server）。完成后重启客户端服务并重新连接网页；Windows 重启服务需要管理员权限。已有项目和会话不会被删除。")
	}
	if name == "qoder" {
		return errors.New("Qoder CLI 未安装或不可用。请在目标客户端电脑上安装并登录 Qoder CLI（不是 Qoder 编辑器启动器），重启客户端服务后重新连接网页。")
	}
	return fmt.Errorf("unknown provider %q; select Codex or Qoder CLI", name)
}

type boundedBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := min(len(p), 128*1024-len(b.data))
	b.data = append(b.data, p[:n]...)
	return len(p), nil
}
func (b *boundedBuffer) String() string { b.mu.Lock(); defer b.mu.Unlock(); return string(b.data) }
func providerArgs(name string, write bool) []string {
	return providerSessionArgs(name, write, "")
}
func providerSessionArgs(name string, write bool, resumeID string) []string {
	if name == "codex" {
		sandbox := "danger-full-access"
		if resumeID != "" {
			return []string{"exec", "resume", "--json", "--skip-git-repo-check", "-c", "sandbox_mode=\"" + sandbox + "\"", "-c", "approval_policy=\"never\"", resumeID, "-"}
		}
		return []string{"exec", "--json", "--color", "never", "--sandbox", sandbox, "-c", "approval_policy=\"never\"", "--skip-git-repo-check", "-"}
	}
	// Full authorization is explicit across Web, IM and SaySo sessions.
	mode := "bypass_permissions"
	args := []string{"--print", "--output-format", "stream-json", "--permission-mode", mode, "--max-turns", "30"}
	if resumeID != "" {
		args = append(args, "--resume", resumeID)
	}
	return args
}

type outputWriter struct {
	mu         sync.Mutex
	pending    []byte
	total      int
	send       func(protocol.Event) error
	id, stream string
	cancel     context.CancelFunc
}

func (w *outputWriter) Write(p []byte) (written int, err error) {
	defer func() {
		if err != nil && w.cancel != nil {
			w.cancel()
		}
	}()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.total+len(p) > 16*1024*1024 {
		return 0, errors.New("output exceeds 16 MiB limit")
	}
	w.total += len(p)
	w.pending = append(w.pending, p...)
	for len(w.pending) > 0 {
		n := 0
		for n < len(w.pending) && n < 8192 {
			if !utf8.FullRune(w.pending[n:]) {
				break
			}
			_, size := utf8.DecodeRune(w.pending[n:])
			n += size
		}
		if n == 0 {
			return len(p), nil
		}
		chunk := strings.ToValidUTF8(string(w.pending[:n]), "�")
		if err := w.send(protocol.Event{ID: w.id, Type: "output", Data: map[string]string{"stream": w.stream, "text": chunk}}); err != nil {
			return 0, err
		}
		w.pending = w.pending[n:]
	}
	return len(p), nil
}
func runProcess(ctx context.Context, p Provider, s protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
	if !p.Available {
		return providerSetupError(s.Provider)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	args := providerSessionArgs(s.Provider, write, s.ResumeID)
	if s.Provider == "codex" && len(s.Images) > 0 {
		args = args[:len(args)-1]
		for _, path := range s.Images {
			args = append(args, "--image", path)
		}
		args = append(args, "-")
	}
	if s.Provider == "codex" && s.ResumeID == "" {
		args = args[:len(args)-1]
		if s.Model != "" && s.Model != "default" {
			args = append(args, "--model", s.Model)
		}
		if s.OutputSchema != "" {
			args = append(args, "--ephemeral", "--output-schema", s.OutputSchema)
		}
		args = append(args, "-")
	}
	args, cleanup, err := providerTools(s, args)
	if err != nil {
		return err
	}
	defer cleanup()
	cmd := command(ctx, p.Path, args)
	cmd.Dir = s.Workspace
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = &outputWriter{send: send, id: id, stream: "stdout", cancel: cancel}
	cmd.Stderr = &outputWriter{send: send, id: id, stream: "stderr", cancel: cancel}
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("task stopped: %w", ctx.Err())
		}
		return err
	}
	return nil
}
