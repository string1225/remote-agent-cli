package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

func gitServiceSID() (string, error) {
	sid, _, _, err := windows.LookupSID("", `NT SERVICE\`+WindowsServiceName)
	if err != nil {
		return "", err
	}
	return sid.String(), nil
}

func (a *Agent) startGitAuth(ctx context.Context) func() {
	sid, err := gitServiceSID()
	if err != nil || !gitauth.IsService(sid) {
		return func() {}
	}
	p, err := ReadServiceProfile(serviceProfilePath())
	if err != nil || !strings.EqualFold(filepath.Clean(p.Config), filepath.Clean(a.Config.Path)) {
		log.Print("service Git authorization unavailable: service profile does not match this client")
		return func() {}
	}
	store := filepath.Join(filepath.Dir(a.Config.Path), "git-auth")
	if err = gitauth.PrepareStore(store, sid); err != nil {
		log.Printf("service Git authorization unavailable: cannot protect the credential directory: %v", err)
		return func() {}
	}
	base := os.Environ()
	stop, err := gitauth.Listen(ctx, gitauth.PipeName(p.Config), p.OwnerSID, sid, gitauth.Manager{Store: store, Env: base}, a.credentialManage)
	if err != nil {
		log.Printf("service Git authorization unavailable: cannot open the local pipe: %v", err)
		return func() {}
	}
	type previousValue struct {
		value  string
		exists bool
	}
	previous := map[string]previousValue{}
	managed := map[string]bool{"GCM_CREDENTIAL_STORE": true, "GCM_DPAPI_STORE_PATH": true, "GCM_INTERACTIVE": true, "GIT_TERMINAL_PROMPT": true, "GIT_CONFIG_PARAMETERS": true, "GIT_ASKPASS": true, "SSH_ASKPASS": true}
	restore := func() {
		for key, old := range previous {
			if old.exists {
				_ = os.Setenv(key, old.value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}
	for _, entry := range gitauth.Environment(base, store, false) {
		key, value, _ := strings.Cut(entry, "=")
		if !managed[key] {
			continue
		}
		old, exists := os.LookupEnv(key)
		previous[key] = previousValue{old, exists}
		if err = os.Setenv(key, value); err != nil {
			stop()
			restore()
			log.Print("service Git authorization unavailable: cannot configure the task environment")
			return func() {}
		}
	}
	log.Print("service Git authorization ready; use remote-agent git login in the device owner's terminal")
	return func() { stop(); restore() }
}

func GitAuthorization(ctx context.Context, config string, request gitauth.Request, output io.Writer) error {
	if err := request.Validate(); err != nil {
		return err
	}
	p, err := ReadServiceProfile(serviceProfilePath())
	if err != nil {
		return errors.New("configure and start the Windows client first, then run remote-agent git login")
	}
	if !strings.EqualFold(filepath.Clean(p.Config), filepath.Clean(config)) {
		return errors.New("Git authorization config does not match the installed Windows service")
	}
	if _, err := gitServiceSID(); err != nil {
		return errors.New("install the Remote Agent MSI and start its Windows service first")
	}
	service, err := openWindowsService(windows.SERVICE_QUERY_CONFIG | windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return err
	}
	defer service.Close()
	sc, err := checkServiceOwner(service)
	if err != nil {
		return err
	}
	status, err := service.Query()
	if err != nil || status.State != svc.Running || status.ProcessId == 0 || sc.SidType != windows.SERVICE_SID_TYPE_UNRESTRICTED {
		return errors.New("service Git authorization is unavailable; start or update RemoteAgent, then retry remote-agent git login")
	}
	return gitauth.Call(ctx, gitauth.PipeName(p.Config), status.ProcessId, request, output)
}
