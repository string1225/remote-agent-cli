//go:build !windows

package agent

import "os"

func protectRebindFile(path string) error { return os.Chmod(path, 0600) }
