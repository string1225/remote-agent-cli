package servicehost

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func machineRoot() (string, error) {
	dir, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "RemoteAgent"), nil
}
func installedHost() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\RemoteAgent`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()
	dir, _, err := k.GetStringValue("InstallDir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dir) {
		return "", errors.New("invalid host install directory")
	}
	return filepath.Join(dir, "remote-agent.exe"), nil
}
func runtimeStore() (clientupdate.Store, error) {
	root, err := machineRoot()
	return clientupdate.Store{Path: filepath.Join(root, "runtime"), Keys: clientupdate.TrustKeys(), Arch: runtime.GOARCH}, err
}
func seedPath() (string, error) {
	host, err := installedHost()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(host), "seed", "remote-agent-core.exe"), nil
}
func readProfile() (Profile, error) {
	var p Profile
	root, err := machineRoot()
	if err != nil {
		return p, err
	}
	f, err := os.Open(filepath.Join(root, "service.json"))
	if err != nil {
		return p, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&p)
	return p, err
}
func sourceServer(explicit string) (string, error) {
	if explicit != "" {
		return clientupdate.NormalizeServer(explicit)
	}
	if p, err := readProfile(); err == nil && p.UpdateServer != "" {
		return clientupdate.NormalizeServer(p.UpdateServer)
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return clientupdate.DefaultServer, nil
}
func ownInstallation() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	installed, err := installedHost()
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Clean(exe), filepath.Clean(installed)) {
		return errors.New("service operation must use the MSI-installed host")
	}
	return nil
}
