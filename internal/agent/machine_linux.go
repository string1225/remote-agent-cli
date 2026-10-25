package agent

import "os"

func systemMachineIdentity() (string, error) {
	b, err := os.ReadFile("/etc/machine-id")
	return string(b), err
}
