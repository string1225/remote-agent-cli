//go:build !windows

package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/gitauth"
)

func (a *Agent) startGitAuth(ctx context.Context) func() {
	if a.Config.Path == "" {
		return func() {}
	}
	stop, err := gitauth.ListenLocal(ctx, a.Config.Path, gitauth.Manager{Store: filepath.Join(filepath.Dir(a.Config.Path), "git-auth"), Env: os.Environ()}, a.credentialManage)
	if err != nil {
		return func() {}
	}
	return stop
}

func GitAuthorization(ctx context.Context, config string, r gitauth.Request, output io.Writer) error {
	if r.Action != "manage" {
		return errors.New("use remote-agent github login/status/logout")
	}
	return gitauth.CallLocal(ctx, config, r, output)
}
