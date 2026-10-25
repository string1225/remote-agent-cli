package agent

import (
	"context"
	"errors"
	"io"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"github.com/string1225/remote-agent-cli/internal/servicehost"
)

func remoteClientUpdateSupported() bool { return servicehost.RemoteUpdateSupported() }
func remoteClientRelease(ctx context.Context, server string) (clientupdate.Release, error) {
	return servicehost.RuntimeRelease(ctx, server)
}
func remoteRuntimeState(string) (string, bool, error) { return servicehost.RemoteRuntimeState() }
func (a *Agent) applyRemoteClientUpdate(ctx context.Context, op MaintenanceOperation, release clientupdate.Release) error {
	set := func(stage, message string) {
		_ = a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) { v.Stage, v.Message = stage, message })
	}
	return servicehost.RemoteUpdate(ctx, &maintenanceWriter{set: set})
}
func (a *Agent) commitRemoteRuntime() error                        { return nil }
func RunWithRemoteUpdateRecovery(_ string, run func() error) error { return run() }
func ProbeRemoteUpdate(string, io.Writer) error {
	return errors.New("Unix update probe is not supported on Windows")
}
func checkRemoteUpdateTransaction(string, string) error { return nil }
