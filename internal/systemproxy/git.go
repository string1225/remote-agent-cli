package systemproxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// GitManager changes only the invoking user's global HTTP proxy. Originals
// may contain proxy credentials, so journals use the same private ACL as the
// Windows proxy backup. Git itself preserves unrelated configuration keys.
type GitManager struct {
	Git  string
	Env  []string
	Path string
}

type gitJournal struct {
	Original []string `json:"original"`
	Applied  string   `json:"applied"`
}

type GitStatus struct {
	Enabled bool
	Managed bool
}

func (m GitManager) command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, m.Git, append([]string{"config", "--global", "--no-includes"}, args...)...)
	cmd.Env = m.Env
	return cmd
}

func (m GitManager) read(ctx context.Context) ([]string, error) {
	output, err := m.command(ctx, "--null", "--get-all", "http.proxy").Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return nil, nil
		}
		return nil, errors.New("cannot read the current user's Git HTTP proxy")
	}
	return strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00"), nil
}

func (m GitManager) write(ctx context.Context, values []string) error {
	if len(values) == 1 {
		if err := m.command(ctx, "--replace-all", "http.proxy", values[0]).Run(); err != nil {
			return errors.New("cannot update the current user's Git HTTP proxy; original backup retained")
		}
		return nil
	}
	if err := m.command(ctx, "--unset-all", "http.proxy").Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 5 {
			return errors.New("cannot update the current user's Git HTTP proxy")
		}
	}
	for _, value := range values {
		if err := m.command(ctx, "--add", "http.proxy", value).Run(); err != nil {
			return errors.New("cannot restore the current user's Git HTTP proxy; original backup retained")
		}
	}
	return nil
}

func (m GitManager) lock() (func(), error) {
	if err := os.MkdirAll(filepath.Dir(m.Path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(m.Path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.New("another Git proxy operation is active; retry shortly")
	}
	return func() { f.Close(); os.Remove(m.Path + ".lock") }, nil
}

func (m GitManager) load() (gitJournal, bool, error) {
	var state gitJournal
	data, err := os.ReadFile(m.Path)
	if os.IsNotExist(err) {
		return state, false, nil
	}
	if err != nil {
		return state, false, err
	}
	if json.Unmarshal(data, &state) != nil || state.Applied == "" {
		return state, false, errors.New("Git proxy backup is invalid; preserve it for recovery")
	}
	return state, true, nil
}

func (m GitManager) save(state gitJournal) error {
	f, err := os.CreateTemp(filepath.Dir(m.Path), ".git-proxy-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = protectFile(f.Name()); err == nil {
		err = json.NewEncoder(f).Encode(state)
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), m.Path)
}

func (m GitManager) Status(ctx context.Context, address string) (GitStatus, error) {
	values, err := m.read(ctx)
	if err != nil {
		return GitStatus{}, err
	}
	effective, err := m.effective(ctx)
	if err != nil {
		return GitStatus{}, err
	}
	return GitStatus{Enabled: effective != "", Managed: address != "" && effective == "http://"+address && slices.Equal(values, []string{"http://" + address})}, nil
}

func (m GitManager) effective(ctx context.Context) (string, error) {
	output, err := m.command(ctx, "--includes", "--null", "--get-urlmatch", "http.proxy", "https://github.com/").Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return "", nil
		}
		return "", errors.New("cannot read the effective GitHub HTTP proxy")
	}
	return strings.TrimSuffix(string(output), "\x00"), nil
}

func (m GitManager) Enable(ctx context.Context, address string) error {
	endpoint, err := netip.ParseAddrPort(address)
	if err != nil || !endpoint.Addr().IsLoopback() || endpoint.Addr().Zone() != "" || endpoint.Port() == 0 {
		return errors.New("Git proxy requires a literal loopback IP and nonzero port")
	}
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := m.read(ctx)
	if err != nil {
		return err
	}
	state, exists, err := m.load()
	if err != nil {
		return err
	}
	previous := state
	applied := "http://" + address
	if !exists || !slices.Equal(current, []string{state.Applied}) {
		state.Original = current
		if slices.Equal(current, []string{applied}) {
			state.Original = nil
		}
	}
	state.Applied = applied
	if err = m.save(state); err != nil {
		return err
	}
	if err = m.write(ctx, []string{applied}); err == nil {
		var actual []string
		actual, err = m.read(ctx)
		if err == nil && !slices.Equal(actual, []string{applied}) {
			err = errors.New("Git HTTP proxy verification failed")
		}
		if err == nil {
			var effective string
			effective, err = m.effective(ctx)
			if err == nil && effective != applied {
				err = errors.New("GitHub proxy is overridden by included or URL-specific Git configuration; original settings retained")
			}
		}
	}
	if err != nil {
		rollback, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		restoreErr := m.write(rollback, current)
		if exists {
			restoreErr = errors.Join(restoreErr, m.save(previous))
		}
		return errors.Join(err, restoreErr)
	}
	return nil
}

func (m GitManager) Disable(ctx context.Context, address string) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := m.read(ctx)
	if err != nil {
		return err
	}
	state, exists, err := m.load()
	if err != nil {
		return err
	}
	if !exists {
		if address != "" && slices.Equal(current, []string{"http://" + address}) {
			return m.write(ctx, nil)
		}
		return nil
	}
	if slices.Equal(current, []string{state.Applied}) {
		if err = m.write(ctx, state.Original); err != nil {
			rollback, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return errors.Join(err, m.write(rollback, current))
		}
	}
	return os.Rename(m.Path, m.Path+".restored")
}
