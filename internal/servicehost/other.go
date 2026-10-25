//go:build !windows

package servicehost

import (
	"context"
	"errors"
	"io"
)

func Run([]string) error { return errors.New("agent host requires Windows") }
func ClientUpdate(context.Context, []string, io.Writer) error {
	return errors.New("runtime updater requires Windows")
}
