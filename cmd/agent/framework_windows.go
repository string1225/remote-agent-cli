package main

import (
	"context"
	"github.com/string1225/remote-agent-cli/internal/servicehost"
	"io"
)

func frameworkUpdate(ctx context.Context, args []string, out io.Writer) (bool, error) {
	return true, servicehost.ClientUpdate(ctx, args, out)
}
