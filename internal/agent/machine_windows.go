package agent

import "golang.org/x/sys/windows/registry"

func systemMachineIdentity() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()
	value, _, err := k.GetStringValue("MachineGuid")
	return value, err
}
