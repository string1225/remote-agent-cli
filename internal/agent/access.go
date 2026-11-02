package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/access"
	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// Credential operations are never converted to tasks or model prompts.
func (a *Agent) accessManager(ctx context.Context) (*access.Manager, error) {
	if a.Config.Path == "" {
		return nil, errors.New("客户端配置路径未设置")
	}
	a.accessOnce.Do(func() {
		root := filepath.Join(filepath.Dir(a.Config.Path), "access")
		a.access, a.accessErr = access.New(ctx, root, gitauth.Manager{Store: filepath.Join(filepath.Dir(a.Config.Path), "git-auth"), Env: os.Environ()})
		if a.accessErr == nil {
			a.access.SetGitRunner(func(ctx context.Context, r gitauth.Request, out io.Writer) error {
				env := os.Environ()
				address, err := WebProxyAddress(a.Config.Path)
				if err != nil {
					return err
				}
				if address != "" {
					if err = CheckWebProxy(ctx, a.Config, address); err != nil {
						return err
					}
					env = append(env, "HTTP_PROXY=http://"+address, "HTTPS_PROXY=http://"+address)
				}
				return (gitauth.Manager{Store: filepath.Join(filepath.Dir(a.Config.Path), "git-auth"), Env: env}).Run(ctx, r, out)
			})
		}
	})
	return a.access, a.accessErr
}
func (a *Agent) handleAccess(p *peer, req protocol.Request, send func(protocol.Event) error) {
	if p.accountTools {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: "凭据管理仅供客户端属主操作"})
		return
	}
	result, err := a.credentialManage(p.ctx, req.Body)
	if err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()})
		return
	}
	_ = send(protocol.Event{ID: req.ID, Type: "result", Data: result})
}
func (a *Agent) credentialManage(ctx context.Context, raw json.RawMessage) (any, error) {
	if err := a.beginOperation(); err != nil {
		return nil, err
	}
	defer a.endOperation()
	manager, err := a.accessManager(context.Background())
	if err != nil {
		return nil, err
	}
	return manager.Call(ctx, raw)
}
func CredentialCall(ctx context.Context, config string, r access.Request) (json.RawMessage, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	err = GitAuthorization(ctx, config, gitauth.Request{Version: gitauth.Version, Action: "manage", Body: raw}, &output)
	if err != nil {
		return nil, err
	}
	if !json.Valid(output.Bytes()) {
		return nil, errors.New("客户端返回了无效的凭据管理结果")
	}
	return output.Bytes(), nil
}
func ManagedGitAuthorization(ctx context.Context, config string, r gitauth.Request, out io.Writer) error {
	action := map[string]string{"login": "github.start", "status": "github.accounts", "logout": "github.logout"}[r.Action]
	raw, err := CredentialCall(ctx, config, access.Request{Action: action, Username: r.Username, Force: r.Force})
	if err != nil {
		return err
	}
	_, err = out.Write(append(raw, '\n'))
	return err
}
