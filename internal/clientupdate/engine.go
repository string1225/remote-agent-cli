package clientupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type Health struct {
	HostPID    int    `json:"hostPid"`
	Digest     string `json:"digest"`
	Version    string `json:"version"`
	Ready      bool   `json:"ready"`
	Paused     bool   `json:"paused"`
	Generation uint64 `json:"generation"`
	Error      string `json:"error,omitempty"`
}

// Controller never changes SCM configuration, identity, recovery or ACLs.
type Controller interface {
	Running() (bool, error)
	Pause(force bool) error
	Resume() error
	Stop() error
	Start() error
	Health() (Health, error)
}

type Engine struct {
	Store   Store
	Service Controller
	Output  io.Writer
}

func (e Engine) say(s string) {
	if e.Output != nil {
		fmt.Fprintln(e.Output, s)
	}
}

func (e Engine) Reconcile() error {
	s, err := e.Store.Load()
	if err != nil || !s.Pending {
		return err
	}
	running, err := e.Service.Running()
	if err != nil {
		return err
	}
	if !running {
		if !s.WasRunning {
			return nil
		}
		if err = e.Service.Start(); err != nil {
			return err
		}
	} // A stopped/disabled service is never enabled by updates.
	h, err := e.Service.Health()
	if err != nil {
		return err
	}
	expected := "seed"
	if s.Current != nil {
		expected = s.Current.SHA256
	}
	if !h.Ready {
		return errors.New("pending runtime has not completed local startup")
	}
	if h.Digest != expected {
		previous := "seed"
		if s.Previous != nil {
			previous = s.Previous.SHA256
		}
		if !s.HasPrevious || h.Digest != previous {
			return errors.New("runtime health does not match an approved version")
		}
		s.Current = s.Previous
		s.Previous = s.RollbackPrevious
		s.HasPrevious = s.RollbackHasPrevious
		if h.Paused {
			if err = e.Service.Resume(); err != nil {
				return err
			}
		}
		e.say("Candidate startup failed; previous runtime remains active.")
	}
	s.Pending = false
	s.RollbackPrevious = nil
	s.RollbackHasPrevious = false
	return e.Store.Save(s)
}

func (e Engine) Apply(ctx context.Context, release Release, download func(io.Writer) error, force bool) error {
	if err := release.Verify(e.Store.Keys, "windows", e.Store.Arch); err != nil {
		return err
	}
	if err := Compatible(release); err != nil {
		return err
	}
	if err := e.Reconcile(); err != nil {
		return err
	}
	s, err := e.Store.Load()
	if err != nil {
		return err
	}
	if s.Pending {
		return errors.New("a staged runtime is waiting for the stopped service to start")
	}
	if release.Sequence < s.HighestSequence || release.Sequence == s.HighestSequence && release.SHA256 != s.HighestSHA256 {
		return errors.New("stale runtime release rejected; use explicit rollback for the retained previous version")
	}
	if s.Current != nil && s.Current.SHA256 == release.SHA256 {
		s.Current = &release
		s.HighestSequence = release.Sequence
		s.HighestSHA256 = release.SHA256
		if err = e.Store.Save(s); err != nil {
			return err
		}
		e.say("Runtime is up to date.")
		return nil
	}
	if err := e.Store.Stage(release, download); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return e.switchTo(s, &release, force)
}

func (e Engine) Rollback(force bool) error {
	if err := e.Reconcile(); err != nil {
		return err
	}
	s, err := e.Store.Load()
	if err != nil {
		return err
	}
	if s.Pending || !s.HasPrevious {
		return errors.New("no stable previous runtime is available")
	}
	if s.Previous != nil {
		if _, err = e.Store.Binary(s.Previous, ""); err != nil {
			return err
		}
	}
	return e.switchTo(s, s.Previous, force)
}

func (e Engine) switchTo(old State, next *Release, force bool) error {
	running, err := e.Service.Running()
	if err != nil {
		return err
	}
	if running {
		e.say("Waiting for local tasks to become idle...")
		if err = e.Service.Pause(force); err != nil {
			return err
		}
	}
	s := old
	s.Current = next
	s.Previous = old.Current
	s.RollbackPrevious = old.Previous
	s.RollbackHasPrevious = old.HasPrevious
	s.HasPrevious = true
	s.Pending = true
	s.WasRunning = running
	if next != nil {
		if next.Sequence >= s.HighestSequence {
			s.HighestSequence = next.Sequence
			s.HighestSHA256 = next.SHA256
		}
	}
	if err = e.Store.Save(s); err != nil {
		if running {
			return errors.Join(err, e.Service.Resume())
		}
		return err
	}
	if !running {
		e.say("Runtime staged. The stopped/disabled service was left unchanged; startup will validate the candidate.")
		return nil
	}
	if err = e.Service.Stop(); err != nil {
		old.HighestSequence = s.HighestSequence
		old.HighestSHA256 = s.HighestSHA256
		return errors.Join(err, e.Store.Save(old), e.Service.Resume())
	}
	e.say("Starting candidate runtime and checking local readiness...")
	if err = e.Service.Start(); err == nil {
		err = e.Reconcile()
	}
	if err != nil {
		if stopErr := e.Service.Stop(); stopErr != nil {
			return errors.Join(err, stopErr)
		}
		old.HighestSequence = s.HighestSequence
		old.HighestSHA256 = s.HighestSHA256
		if saveErr := e.Store.Save(old); saveErr != nil {
			return errors.Join(err, saveErr)
		}
		return errors.Join(fmt.Errorf("candidate failed; restored previous runtime: %w", err), e.Service.Start())
	}
	committed, err := e.Store.Load()
	if err != nil {
		return err
	}
	if (committed.Current == nil) != (next == nil) || committed.Current != nil && next != nil && committed.Current.SHA256 != next.SHA256 {
		return errors.New("candidate failed local startup; host safely selected the previous runtime")
	}
	if err = e.Store.Clean(committed); err != nil {
		e.say("Runtime active; unused version cleanup deferred.")
	}
	e.say("Runtime updated. MSI registration, service configuration and user data were retained.")
	return nil
}
