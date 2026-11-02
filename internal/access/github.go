package access

import (
	"bytes"
	"context"
	"errors"
	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"github.com/string1225/remote-agent-cli/internal/store"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GitHubState struct {
	ID        string   `json:"id,omitempty"`
	Status    string   `json:"status"`
	URL       string   `json:"url,omitempty"`
	Code      string   `json:"code,omitempty"`
	ExpiresAt int64    `json:"expiresAt,omitempty"`
	Message   string   `json:"message,omitempty"`
	Accounts  []string `json:"accounts,omitempty"`
}

var deviceCode = regexp.MustCompile(`\b[A-Z0-9]{4}-[A-Z0-9]{4}\b`)

type codeWriter struct {
	m        *Manager
	id       string
	raw      string
	ready    chan struct{}
	notified bool
}

func (w *codeWriter) Write(p []byte) (int, error) {
	w.m.mu.Lock()
	defer w.m.mu.Unlock()
	if len(w.raw)+len(p) > 64<<10 {
		return 0, errors.New("authorization output limit exceeded")
	}
	w.raw += string(p)
	if code := deviceCode.FindString(w.raw); code != "" && w.m.github.ID == w.id && w.m.github.Code == "" {
		w.m.github.Code = code
		w.m.github.Status = "waiting"
		w.m.github.Message = "请在 GitHub 页面输入设备码并确认授权"
		_ = w.m.saveGitHub()
		if !w.notified {
			close(w.ready)
			w.notified = true
		}
	}
	return len(p), nil
}
func (m *Manager) saveGitHub() error {
	return writeJSON(filepath.Join(m.root, "github.json"), m.github)
}
func (m *Manager) startGitHub(r Request) (GitHubState, error) {
	request := gitauth.Request{Action: "login", Username: r.Username, Force: r.Force}
	if err := request.Validate(); err != nil {
		return GitHubState{}, err
	}
	m.mu.Lock()
	if m.cancel != nil {
		state := m.github
		m.mu.Unlock()
		return state, nil
	}
	ctx, cancel := context.WithTimeout(m.ctx, 15*time.Minute)
	m.cancel = cancel
	m.github = GitHubState{ID: store.Token(), Status: "starting", URL: "https://github.com/login/device", ExpiresAt: time.Now().Add(15 * time.Minute).UnixMilli(), Message: "正在申请 GitHub 设备码"}
	if err := m.saveGitHub(); err != nil {
		cancel()
		m.cancel = nil
		m.mu.Unlock()
		return GitHubState{}, err
	}
	id := m.github.ID
	ready := make(chan struct{})
	finished := make(chan struct{})
	m.mu.Unlock()
	go func() {
		defer close(finished)
		defer cancel()
		writer := &codeWriter{m: m, id: id, ready: ready}
		err := m.runGit(ctx, request, writer)
		var accounts []string
		if err == nil {
			accounts, err = m.accounts(ctx)
			if err == nil && len(accounts) == 0 {
				err = errors.New("authorization not saved")
			}
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.github.ID != id {
			return
		}
		m.cancel = nil
		m.github.Code = ""
		if err == nil {
			m.github.Status = "authorized"
			m.github.Message = "GitHub 授权已保存，可用于客户端后台任务"
			m.github.Accounts = accounts
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			m.github.Status = "expired"
			m.github.Message = "设备码已过期，请重新发起授权"
		} else if ctx.Err() != nil {
			m.github.Status = "cancelled"
			m.github.Message = "授权已取消，原有账号保留"
		} else {
			m.github.Status = "failed"
			m.github.Message = "GitHub 授权未保存，请检查 Git Credential Manager、网络或系统凭据存储后重试"
		}
		_ = m.saveGitHub()
	}()
	select {
	case <-ready:
	case <-finished:
	case <-time.After(5 * time.Second):
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.github, nil
}
func (m *Manager) accounts(ctx context.Context) ([]string, error) {
	var out boundedBuffer
	if err := m.runGit(ctx, gitauth.Request{Action: "status"}, &out); err != nil {
		return nil, errors.New("无法查询 GitHub 账号，请安装 Git Credential Manager 并检查系统凭据存储")
	}
	accounts := []string{}
	for _, line := range strings.Split(out.String(), "\n") {
		name := strings.TrimSpace(line)
		if (gitauth.Request{Action: "logout", Username: name}).Validate() == nil {
			accounts = append(accounts, name)
		}
	}
	return accounts, nil
}

type boundedBuffer struct{ bytes.Buffer }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 64<<10 {
		return 0, io.ErrShortBuffer
	}
	return b.Buffer.Write(p)
}
