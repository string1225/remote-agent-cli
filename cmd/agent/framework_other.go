//go:build !windows

package main

import (
	"context"
	"io"
)

func frameworkUpdate(context.Context, []string, io.Writer) (bool, error) { return false, nil }
