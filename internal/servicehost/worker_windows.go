package servicehost

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/workerproto"
	"golang.org/x/sys/windows"
)

type worker struct {
	version string
	cmd     *exec.Cmd
	in      io.WriteCloser
	replies chan workerproto.Message
	done    chan error
}

func launchWorker(exe string, stderr io.Writer) (*worker, error) {
	root, err := machineRoot()
	if err != nil {
		return nil, err
	}
	return launchWorkerAt(exe, filepath.Join(root, "service.json"), stderr)
}

func launchWorkerAt(exe, profile string, stderr io.Writer) (*worker, error) {
	cmd := exec.Command(exe, "internal-worker", profile)
	cmd.Dir = filepath.Dir(exe)
	cmd.Stderr = stderr
	cmd.Stdout = stderr // SDK/CLI diagnostics can never corrupt the control stream.
	childIn, in, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer childIn.Close()
	out, childOut, err := os.Pipe()
	if err != nil {
		in.Close()
		return nil, err
	}
	defer childOut.Close()
	for _, file := range []*os.File{childIn, childOut} {
		if err = windows.SetHandleInformation(windows.Handle(file.Fd()), windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
			in.Close()
			out.Close()
			return nil, err
		}
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{AdditionalInheritedHandles: []syscall.Handle{syscall.Handle(childIn.Fd()), syscall.Handle(childOut.Fd())}}
	cmd.Env = append(os.Environ(), "RA_WORKER_READ_HANDLE="+strconv.FormatUint(uint64(childIn.Fd()), 10), "RA_WORKER_WRITE_HANDLE="+strconv.FormatUint(uint64(childOut.Fd()), 10))
	if err = cmd.Start(); err != nil {
		in.Close()
		out.Close()
		return nil, err
	}
	w := &worker{cmd: cmd, in: in, replies: make(chan workerproto.Message, 1), done: make(chan error, 1)}
	childIn.Close()
	childOut.Close()
	go func() {
		defer func() { close(w.replies); w.done <- cmd.Wait() }()
		d := json.NewDecoder(out)
		for {
			var m workerproto.Message
			if err := d.Decode(&m); err != nil {
				if err != io.EOF {
					fmt.Fprintf(stderr, "worker control decoding failed (%T)\n", err)
				}
				out.Close()
				return
			}
			w.replies <- m
		}
	}()
	if err = w.call("prepare", false); err != nil {
		w.kill()
		return nil, err
	}
	return w, nil
}
func (w *worker) call(kind string, force bool) error {
	if err := json.NewEncoder(w.in).Encode(workerproto.Message{Protocol: workerproto.Version, Type: kind, Force: force}); err != nil {
		return err
	}
	select {
	case msg, ok := <-w.replies:
		if !ok {
			return errors.New("worker control pipe closed")
		}
		expected := map[string]string{"prepare": "ready", "activate": "active", "pause": "paused", "resume": "active", "stop": "stopped"}[kind]
		if msg.Protocol != workerproto.Version || msg.Type != expected {
			return errors.New("worker handshake mismatch")
		}
		if msg.Error != "" {
			return errors.New(msg.Error)
		}
		w.version = msg.Version
		return nil
	case <-time.After(25 * time.Second):
		return errors.New("local worker readiness/control timeout")
	}
}
func (w *worker) kill() {
	_ = w.in.Close()
	if w.cmd.Process != nil {
		_ = w.cmd.Process.Kill()
	}
}
func (w *worker) stop() error {
	err := w.call("stop", false)
	_ = w.in.Close()
	select {
	case exit := <-w.done:
		return errors.Join(err, exit)
	case <-time.After(15 * time.Second):
		w.kill()
		return errors.New("worker shutdown timed out")
	}
}

func hostLog() (io.WriteCloser, error) {
	root, err := machineRoot()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, "health", "host.log")
	if st, e := os.Stat(path); e == nil && st.Size() > 10<<20 {
		_ = os.Rename(path, path+".old")
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
}
