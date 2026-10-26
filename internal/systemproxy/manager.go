// Package systemproxy manages a user proxy with a recoverable settings journal.
package systemproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	Direct uint32 = 1
	Proxy  uint32 = 2
)

type Settings struct {
	Flags         uint32 `json:"flags"`
	Server        string `json:"server"`
	Bypass        string `json:"bypass"`
	AutoConfigURL string `json:"autoConfigUrl"`
}
type Backend interface {
	Read() (Settings, error)
	Write(Settings) error
}
type Manager struct {
	Backend Backend
	Path    string
}
type journal struct {
	Original Settings `json:"original"`
	Applied  Settings `json:"applied"`
}

func (m Manager) Enable(address string) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := m.Backend.Read()
	if err != nil {
		return err
	}
	state, exists, err := m.load()
	if err != nil {
		return err
	}
	if !exists || current != state.Applied {
		state.Original = current
		// When the user manually selected this same proxy before the command,
		// off must still disable it rather than restore a dead local endpoint.
		if current.Server == address && current.Flags&Proxy != 0 {
			state.Original.Flags = current.Flags&^Proxy | Direct
		}
	}
	state.Applied = Settings{Flags: Direct | Proxy, Server: address, Bypass: current.Bypass}
	if err := m.save(state); err != nil {
		return err
	}
	if err := m.Backend.Write(state.Applied); err != nil {
		rollback := m.Backend.Write(current)
		return errors.Join(fmt.Errorf("cannot enable Windows system proxy: %w", err), rollback)
	}
	actual, err := m.Backend.Read()
	if err != nil || actual != state.Applied {
		return errors.Join(errors.New("system proxy verification failed"), err, m.Backend.Write(current))
	}
	return nil
}

func (m Manager) Disable(address string) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := m.Backend.Read()
	if err != nil {
		return err
	}
	state, exists, err := m.load()
	if err != nil {
		return err
	}
	if exists {
		if current == state.Applied {
			if err := m.Backend.Write(state.Original); err != nil {
				return err
			}
		} // Preserve settings changed by the user or another proxy application.
		if err := os.Rename(m.Path, m.Path+".restored"); err != nil {
			return err
		}
		return nil
	}
	if current.Flags&Proxy != 0 && current.Server == address {
		state = journal{Original: current, Applied: current}
		if err := m.save(state); err != nil {
			return err
		}
		current.Flags = current.Flags&^Proxy | Direct
		if err := m.Backend.Write(current); err != nil {
			return err
		}
		return os.Rename(m.Path, m.Path+".restored")
	}
	return nil
}

func (m Manager) lock() (func(), error) {
	if m.Backend == nil {
		return nil, errors.New("system proxy backend is unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(m.Path), 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(m.Path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("another system proxy operation is active; preserve %s and retry: %w", m.Path, err)
	}
	return func() { lock.Close(); os.Remove(m.Path + ".lock") }, nil
}
func (m Manager) load() (journal, bool, error) {
	var state journal
	data, err := os.ReadFile(m.Path)
	if os.IsNotExist(err) {
		return state, false, nil
	}
	if err != nil {
		return state, false, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, false, errors.New("system proxy backup is invalid; preserve it for recovery")
	}
	if state.Applied.Server == "" || state.Applied.Flags&Proxy == 0 {
		return state, false, errors.New("system proxy backup is incomplete; preserve it for recovery")
	}
	return state, true, nil
}
func (m Manager) save(state journal) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(m.Path), ".system-proxy-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = protectFile(f.Name()); err == nil {
		_, err = f.Write(data)
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
