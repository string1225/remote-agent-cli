//go:build !windows

package servicehost

import (
	"context"
	"errors"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"io"
)

func Run([]string) error { return errors.New("agent host requires Windows") }
func ClientUpdate(context.Context, []string, io.Writer) error {
	return errors.New("runtime updater requires Windows")
}
func RemoteUpdateSupported() bool { return false }
func RemoteRuntimeState() (string, bool, error) {
	return "", false, errors.New("remote client update requires Windows host")
}
func RemoteUpdate(context.Context, io.Writer) error {
	return errors.New("remote client update requires Windows host")
}
func RuntimeRelease(context.Context, string) (clientupdate.Release, error) {
	return clientupdate.Release{}, errors.New("remote client update requires Windows host")
}
