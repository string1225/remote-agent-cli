//go:build !windows

package systemproxy

import "os"

func Native() (Backend, bool)       { return nil, false }
func protectFile(path string) error { return os.Chmod(path, 0600) }
