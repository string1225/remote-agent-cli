//go:build !windows

package agent

import "errors"

func CleanupLegacyWindows(string) error { return errors.New("Windows cleanup requires Windows") }

func RunWindowsService() error     { return errors.New("Windows Service requires Windows") }
func PrepareWindowsService() error { return errors.New("Windows Service requires Windows") }
func ConfigureWindowsService(ServiceProfile, bool) error {
	return errors.New("Windows Service requires Windows")
}
func SetupWindowsService(string, bool) error               { return errors.New("Windows Service requires Windows") }
func WindowsServiceStatus() (any, error)                   { return nil, errors.New("Windows Service requires Windows") }
func windowsServiceAutostart(string, string) (bool, error) { return false, nil }
func msiManagedExecutable(string) bool                     { return false }
