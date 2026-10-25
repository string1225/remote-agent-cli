package agent

import (
	"errors"
	"io"
	"os"
	"strconv"

	"golang.org/x/sys/windows"
)

func WorkerControlPipes() (io.Reader, io.Writer, func(), error) {
	read, write := os.Getenv("RA_WORKER_READ_HANDLE"), os.Getenv("RA_WORKER_WRITE_HANDLE")
	if read == "" && write == "" {
		return os.Stdin, os.Stdout, func() {}, nil
	}
	r, e1 := strconv.ParseUint(read, 10, 64)
	w, e2 := strconv.ParseUint(write, 10, 64)
	if e1 != nil || e2 != nil || r == 0 || w == 0 {
		return nil, nil, nil, errors.New("invalid inherited worker pipes")
	}
	for _, h := range []uint64{r, w} {
		kind, err := windows.GetFileType(windows.Handle(h))
		if err != nil || kind != windows.FILE_TYPE_PIPE {
			return nil, nil, nil, errors.New("worker control handle is not an inherited pipe")
		}
	}
	in, out := os.NewFile(uintptr(r), "worker-control-read"), os.NewFile(uintptr(w), "worker-control-write")
	return in, out, func() { in.Close(); out.Close() }, nil
}
