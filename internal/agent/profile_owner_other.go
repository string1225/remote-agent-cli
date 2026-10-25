//go:build !windows

package agent

func currentProfileOwner() (string, error) { return "", nil }
