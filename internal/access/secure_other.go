//go:build !windows

package access

import (
	"errors"
	"os"
)

func secureDir(path string) error {
	if err := realDirectoryPath(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("credential directory must not be a link")
	}
	return os.Chmod(path, 0700)
}
func secureFile(path string) error { return os.Chmod(path, 0600) }
