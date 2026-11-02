//go:build !windows

package agent

func ActivationConfigPath() (string, error) { return DefaultPath(), nil }
