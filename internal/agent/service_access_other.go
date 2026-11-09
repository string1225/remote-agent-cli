//go:build !windows

package agent

import (
	"context"
	"io"

	"github.com/string1225/remote-agent-cli/internal/serviceaccess"
)

func ServiceAccess(_ context.Context, o serviceaccess.Options, output io.Writer) error {
	return serviceaccess.Execute(o, output)
}
