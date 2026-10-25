//go:build !windows

package codexsetup

import "os"

func protect(path string) error { return os.Chmod(path, 0600) }

func preserveServiceAccess(original, replacement string) error { return nil }
