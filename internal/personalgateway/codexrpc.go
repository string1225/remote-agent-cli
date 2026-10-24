package personalgateway

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// The native app-server owns device authorization, polling and credential storage.
// No OpenAI credentials are returned by this control API.
type codexRPC struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Scanner
}

func startCodexRPC(ctx context.Context, binary, proxy, home string) (*codexRPC, error) {
	cmd := exec.CommandContext(ctx, binary, "app-server", "-c", `cli_auth_credentials_store="file"`)
	cmd.Env = os.Environ()
	if home != "" {
		cmd.Env = append(cmd.Env, "CODEX_HOME="+home)
	}
	if proxy != "" {
		cmd.Env = append(cmd.Env, "HTTP_PROXY="+proxy, "HTTPS_PROXY="+proxy, "http_proxy="+proxy, "https_proxy="+proxy)
	}
	cmd.Stderr = io.Discard
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		in.Close()
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		in.Close()
		return nil, err
	}
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	p := &codexRPC{cmd, in, scanner}
	_, err = p.call(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "remote_agent_admin", "version": "1.0"}, "capabilities": map[string]bool{"experimentalApi": true}})
	if err == nil {
		err = json.NewEncoder(in).Encode(map[string]string{"method": "initialized"})
	}
	if err != nil {
		p.close()
		return nil, err
	}
	return p, nil
}
func (p *codexRPC) close() { p.in.Close(); _ = p.cmd.Process.Kill(); _ = p.cmd.Wait() }
func (p *codexRPC) call(id int, method string, params any) (json.RawMessage, error) {
	if err := json.NewEncoder(p.in).Encode(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for p.out.Scan() {
		var v struct {
			ID     *int            `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  json.RawMessage `json:"error"`
		}
		if json.Unmarshal(p.out.Bytes(), &v) != nil || v.ID == nil || *v.ID != id {
			continue
		}
		if len(v.Error) > 0 && string(v.Error) != "null" {
			return nil, errors.New("Codex rejected the operation; check device authorization is enabled")
		}
		return v.Result, nil
	}
	return nil, errors.New("Codex authorization process stopped")
}

type deviceChallenge struct {
	LoginID         string `json:"loginId"`
	UserCode        string `json:"userCode"`
	VerificationURL string `json:"verificationUrl"`
}

func validChallenge(v deviceChallenge) bool {
	u, err := url.Parse(v.VerificationURL)
	return err == nil && u.Scheme == "https" && u.User == nil && u.Port() == "" && (u.Hostname() == "auth.openai.com" || u.Hostname() == "chatgpt.com") && len(v.VerificationURL) <= 2048 && len(v.LoginID) > 0 && len(v.LoginID) <= 200 && len(v.UserCode) > 0 && len(v.UserCode) <= 100 && !strings.ContainsAny(v.UserCode, "\r\n")
}
func (m *AccountManager) startLogin(ctx context.Context, home string) (deviceChallenge, func() (Credential, error), error) {
	p, err := startCodexRPC(ctx, m.Binary, m.Proxy, home)
	if err != nil {
		return deviceChallenge{}, nil, err
	}
	data, err := p.call(2, "account/login/start", map[string]string{"type": "chatgptDeviceCode"})
	var v deviceChallenge
	if err != nil || json.Unmarshal(data, &v) != nil || !validChallenge(v) {
		p.close()
		return deviceChallenge{}, nil, errors.New("无法开始设备授权；请确认账号已允许设备码登录及服务器网络正常")
	}
	return v, func() (Credential, error) {
		defer p.close()
		for p.out.Scan() {
			var notification struct {
				Method string `json:"method"`
				Params struct {
					LoginID string `json:"loginId"`
					Success bool   `json:"success"`
				} `json:"params"`
			}
			if json.Unmarshal(p.out.Bytes(), &notification) != nil || notification.Method != "account/login/completed" || notification.Params.LoginID != v.LoginID {
				continue
			}
			if !notification.Params.Success {
				return Credential{}, ErrLoginRequired
			}
			return (&CodexSource{Binary: m.Binary, Proxy: m.Proxy, Home: home}).Get(ctx, "")
		}
		return Credential{}, ErrLoginRequired
	}, nil
}
