//go:build !windows

package agent

import (
	"io"
	"os"
)

func WorkerControlPipes() (io.Reader, io.Writer, func(), error) {
	return os.Stdin, os.Stdout, func() {}, nil
}
