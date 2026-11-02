// Package access handles human-entered credentials outside model conversations.
package access

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/string1225/remote-agent-cli/internal/gitauth"
)

type Request struct {
	Action   string `json:"action"`
	Name     string `json:"name,omitempty"`
	Host     string `json:"host,omitempty"`
	User     string `json:"user,omitempty"`
	Port     int    `json:"port,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

type Manager struct {
	root   string
	ctx    context.Context
	mu     sync.Mutex
	sshMu  sync.Mutex
	github GitHubState
	cancel context.CancelFunc
	runGit func(context.Context, gitauth.Request, io.Writer) error
}

// SetGitRunner is only used during daemon initialization, before serving calls.
func (m *Manager) SetGitRunner(run func(context.Context, gitauth.Request, io.Writer) error) {
	m.runGit = run
}

func New(ctx context.Context, root string, git gitauth.Manager) (*Manager, error) {
	if !filepath.IsAbs(root) {
		return nil, errors.New("credential directory must be absolute")
	}
	if err := secureDir(root); err != nil {
		return nil, err
	}
	if git.Store != "" {
		if err := secureDir(git.Store); err != nil {
			return nil, err
		}
	}
	m := &Manager{root: root, ctx: ctx, runGit: git.Run}
	if raw, err := os.ReadFile(filepath.Join(root, "github.json")); err == nil {
		if err = json.Unmarshal(raw, &m.github); err != nil {
			return nil, errors.New("GitHub authorization state is damaged; original file retained")
		}
		if m.github.Status == "starting" || m.github.Status == "waiting" {
			m.github.Status = "interrupted"
			m.github.Code = ""
			m.github.Message = "客户端已重启，请重新发起授权"
		}
	}
	return m, nil
}

func (m *Manager) Call(ctx context.Context, raw json.RawMessage) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var r Request
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if len(raw) > 16<<10 || d.Decode(&r) != nil {
		return nil, errors.New("无效的凭据管理请求")
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("无效的凭据管理请求")
	}
	if r.Password != "" && r.Action != "servers.add" {
		return nil, errors.New("该操作不接受密码")
	}
	switch r.Action {
	case "servers.list":
		return m.ListServers()
	case "servers.add":
		return m.AddServer(ctx, r)
	case "servers.test":
		return m.TestServer(ctx, r.Name)
	case "servers.remove":
		return m.RemoveServer(r.Name)
	case "github.start":
		return m.startGitHub(r)
	case "github.status":
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.github, nil
	case "github.cancel":
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.cancel != nil {
			m.cancel()
		}
		return true, nil
	case "github.accounts":
		return m.accounts(ctx)
	case "github.logout":
		m.mu.Lock()
		busy := m.cancel != nil
		m.mu.Unlock()
		if busy {
			return nil, errors.New("请先完成或取消正在进行的授权")
		}
		if err := m.runGit(ctx, gitauth.Request{Action: "logout", Username: r.Username}, io.Discard); err != nil {
			return nil, errors.New("移除 GitHub 授权失败")
		}
		return true, nil
	default:
		return nil, errors.New("未知的凭据管理操作")
	}
}

func writeFile(path string, data []byte) error {
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("credential destination must be a regular file")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".access-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = secureFile(f.Name()); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(raw, '\n'))
}

func realDirectoryPath(path string) error {
	for cursor := filepath.Clean(path); ; cursor = filepath.Dir(cursor) {
		if info, err := os.Lstat(cursor); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return errors.New("credential directory contains a link or non-directory")
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if filepath.Dir(cursor) == cursor {
			return nil
		}
	}
}
