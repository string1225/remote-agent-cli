package servicehost

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type controller struct{ s *mgr.Service }

func openController() (*controller, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	defer m.Disconnect()
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return nil, err
	}
	cfg, err := s.Config()
	if err != nil {
		s.Close()
		return nil, err
	}
	expected, err := installedHost()
	if err != nil {
		s.Close()
		return nil, err
	}
	args, err := windows.DecomposeCommandLine(cfg.BinaryPathName)
	if err != nil || len(args) != 3 || !strings.EqualFold(args[0], expected) || args[1] != "service" || args[2] != "run" || !strings.EqualFold(cfg.ServiceStartName, `NT AUTHORITY\LocalService`) {
		s.Close()
		return nil, errors.New("service identity or host path mismatch")
	}
	return &controller{s}, nil
}
func (c *controller) Running() (bool, error) {
	s, err := c.s.Query()
	if err != nil {
		return false, err
	}
	if s.State != svc.Running && s.State != svc.Stopped {
		return false, errors.New("service is changing state; retry shortly")
	}
	return s.State == svc.Running, nil
}
func (c *controller) wait(state svc.State) error {
	until := time.Now().Add(90 * time.Second)
	for time.Now().Before(until) {
		s, err := c.s.Query()
		if err != nil {
			return err
		}
		if s.State == state {
			return nil
		}
		if state == svc.Running && s.State == svc.Stopped {
			return errors.New("host failed local startup")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("service lifecycle timed out")
}
func (c *controller) Stop() error {
	s, err := c.s.Query()
	if err != nil {
		return err
	}
	if s.State == svc.Stopped {
		return nil
	}
	if _, err = c.s.Control(svc.Stop); err != nil {
		return err
	}
	return c.wait(svc.Stopped)
}
func (c *controller) Start() error {
	if err := c.s.Start(); err != nil {
		return err
	}
	return c.wait(svc.Running)
}
func (c *controller) Health() (clientupdate.Health, error) {
	var h clientupdate.Health
	root, err := machineRoot()
	if err != nil {
		return h, err
	}
	f, err := os.Open(filepath.Join(root, "health", "status.json"))
	if err != nil {
		return h, err
	}
	defer f.Close()
	if err = json.NewDecoder(io.LimitReader(f, 32769)).Decode(&h); err != nil {
		return h, err
	}
	status, err := c.s.Query()
	if err != nil {
		return h, err
	}
	if status.State != svc.Running || uint32(h.HostPID) != status.ProcessId {
		return h, errors.New("health belongs to a different host process")
	}
	return h, nil
}
func (c *controller) Pause(force bool) error {
	before, err := c.Health()
	if err != nil {
		return err
	}
	cmd := pauseControl
	if force {
		cmd = forceControl
	}
	if _, err = c.s.Control(cmd); err != nil {
		return err
	}
	until := time.Now().Add(30 * time.Second)
	for time.Now().Before(until) {
		h, err := c.Health()
		if err != nil {
			return err
		}
		if h.Generation != before.Generation {
			if h.Paused && h.Ready {
				return nil
			}
			if h.Error != "" {
				return errors.New(h.Error)
			}
			return errors.New("client could not pause safely")
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("client did not acknowledge update pause")
}
func (c *controller) Resume() error { _, err := c.s.Control(resumeControl); return err }
