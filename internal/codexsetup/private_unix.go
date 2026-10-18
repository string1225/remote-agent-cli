//go:build !windows

package codexsetup

import "os"

func protect(path string) error { return os.Chmod(path, 0600) }
