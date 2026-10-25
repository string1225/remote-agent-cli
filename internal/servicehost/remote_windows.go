package servicehost

import (
	"context"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows/registry"
	"io"
)

func RemoteUpdateSupported() bool {
	k, e := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\RemoteAgent`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if e != nil {
		return false
	}
	defer k.Close()
	v, _, e := k.GetIntegerValue("HostVersion")
	return e == nil && v >= 2
}
func RemoteRuntimeState() (version string, pending bool, err error) {
	s, e := runtimeStore()
	if e != nil {
		return "", false, e
	}
	state, e := s.Load()
	if e != nil {
		return "", false, e
	}
	if state.Current == nil {
		return "seed", state.Pending, nil
	}
	return state.Current.Version, state.Pending, nil
}
func RemoteUpdate(ctx context.Context, out io.Writer) error {
	return requestUpdate(ctx, Request{Action: "apply"}, out)
}
func RuntimeRelease(ctx context.Context, server string) (clientupdate.Release, error) {
	s, e := runtimeStore()
	if e != nil {
		return clientupdate.Release{}, e
	}
	return clientupdate.Fetch(ctx, clientupdate.HTTPClient(), server, "windows", s.Arch, s.Keys)
}
