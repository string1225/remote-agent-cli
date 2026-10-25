package clientupdate

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type State struct {
	SchemaVersion       int      `json:"schemaVersion"`
	Current             *Release `json:"current,omitempty"`
	Previous            *Release `json:"previous,omitempty"`
	HasPrevious         bool     `json:"hasPrevious"`
	HighestSequence     uint64   `json:"highestSequence"`
	HighestSHA256       string   `json:"highestSha256,omitempty"`
	Pending             bool     `json:"pending"`
	WasRunning          bool     `json:"wasRunning"`
	RollbackPrevious    *Release `json:"rollbackPrevious,omitempty"`
	RollbackHasPrevious bool     `json:"rollbackHasPrevious,omitempty"`
}

type Store struct {
	Path string
	Keys map[string]ed25519.PublicKey
	Arch string
}

func (s Store) Load() (State, error) {
	var state State
	r, err := os.OpenRoot(s.Path)
	if os.IsNotExist(err) {
		return State{SchemaVersion: 1}, nil
	}
	if err != nil {
		return state, err
	}
	defer r.Close()
	f, err := r.Open("state.json")
	if os.IsNotExist(err) {
		return State{SchemaVersion: 1}, nil
	}
	if err != nil {
		return state, err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 32769))
	d.DisallowUnknownFields()
	if err = d.Decode(&state); err != nil {
		return state, err
	}
	if d.Decode(&struct{}{}) != io.EOF || state.SchemaVersion != 1 {
		return state, errors.New("invalid runtime state")
	}
	for _, release := range []*Release{state.Current, state.Previous, state.RollbackPrevious} {
		if release != nil {
			if err := release.Verify(s.Keys, "windows", s.Arch); err != nil {
				return state, err
			}
			if err := Compatible(*release); err != nil {
				return state, err
			}
			if release.Sequence > state.HighestSequence {
				return state, errors.New("invalid runtime sequence floor")
			}
		}
	}
	return state, nil
}

func (s Store) Save(state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	r, err := os.OpenRoot(s.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	return atomicWrite(r, "state.json", data)
}

func randomName() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func atomicWrite(root *os.Root, name string, data []byte) error {
	tmp := ".transaction-" + randomName()
	f, err := root.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(tmp)
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return root.Rename(tmp, name)
}

func (s Store) Binary(release *Release, seed string) (string, error) {
	if release == nil {
		return seed, nil
	} // The seed belongs to MSI, not the writable data directory.
	if err := release.Verify(s.Keys, "windows", s.Arch); err != nil {
		return "", err
	}
	if err := Compatible(*release); err != nil {
		return "", err
	}
	r, err := os.OpenRoot(s.Path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	base := "versions/" + release.SHA256
	info, err := r.Lstat(base)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("runtime directory is a reparse point or non-directory")
	}
	f, err := r.Open(base + "/remote-agent-core.exe")
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, MaxSize+1))
	if err != nil {
		return "", err
	}
	if n != release.Size || hex.EncodeToString(h.Sum(nil)) != release.SHA256 {
		return "", errors.New("installed runtime verification failed")
	}
	return filepath.Join(s.Path, "versions", release.SHA256, "remote-agent-core.exe"), nil
}

func (s Store) Stage(release Release, download func(io.Writer) error) error {
	if err := release.Verify(s.Keys, "windows", s.Arch); err != nil {
		return err
	}
	if err := Compatible(release); err != nil {
		return err
	}
	if _, err := s.Binary(&release, ""); err == nil {
		return nil
	}
	r, err := os.OpenRoot(s.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	if err = r.MkdirAll("versions", 0755); err != nil {
		return err
	}
	stage := ".staging-" + randomName()
	if err = r.Mkdir(stage, 0700); err != nil {
		return err
	}
	defer r.RemoveAll(stage) // Generated relative path, confined by os.Root.
	f, err := r.OpenFile(stage+"/remote-agent-core.exe", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	err = download(f)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	// Independently rehash the staged bytes even when a caller supplies a downloader.
	f, err = r.Open(stage + "/remote-agent-core.exe")
	if err != nil {
		return err
	}
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, MaxSize+1))
	f.Close()
	if e != nil || n != release.Size || hex.EncodeToString(h.Sum(nil)) != release.SHA256 {
		return errors.New("staged runtime verification failed")
	}
	data, _ := json.Marshal(release)
	if err = atomicWriteIn(r, stage+"/release.json", data); err != nil {
		return err
	}
	return r.Rename(stage, "versions/"+release.SHA256)
}

func atomicWriteIn(root *os.Root, name string, data []byte) error {
	f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	return errors.Join(err, f.Close())
}

func (s Store) Clean(state State) error {
	r, err := os.OpenRoot(s.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	dir, err := r.Open("versions")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	entries, err := dir.ReadDir(-1)
	dir.Close()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !digestRE.MatchString(name) || (state.Current != nil && name == state.Current.SHA256) || (state.HasPrevious && state.Previous != nil && name == state.Previous.SHA256) {
			continue
		}
		info, err := r.Lstat("versions/" + name)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing runtime cleanup of non-directory %s", name)
		}
		if err = r.RemoveAll("versions/" + name); err != nil {
			return err
		}
	}
	return nil
}
