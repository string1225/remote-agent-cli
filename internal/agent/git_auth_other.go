//go:build !windows

package agent

import (
	"context"
	"errors"
	"io"

	"github.com/string1225/remote-agent-cli/internal/gitauth"
)

func (a *Agent) startGitAuth(context.Context) func() { return func() {} }

func GitAuthorization(context.Context, string, gitauth.Request, io.Writer) error {
	return errors.New("service Git authorization is only needed on Windows; authenticate Git as the client user on this system")
}
