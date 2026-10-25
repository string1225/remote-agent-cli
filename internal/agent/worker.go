package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/workerproto"
	bolt "go.etcd.io/bbolt"
)

// RunWorker is an internal pipe protocol. Preparing never connects to the site,
// recovers tasks, or rewrites business data; those start only after activation.
func RunWorker(ctx context.Context, profile string, input io.Reader, output io.Writer) error {
	d := json.NewDecoder(input)
	e := json.NewEncoder(output)
	reply := func(kind string, err error) error {
		m := workerproto.Message{Protocol: workerproto.Version, Type: kind, Version: buildinfo.Current()}
		if err != nil {
			m.Error = err.Error()
		}
		return e.Encode(m)
	}
	var first workerproto.Message
	if err := d.Decode(&first); err != nil {
		return err
	}
	if first.Protocol != workerproto.Version || first.Type != "prepare" {
		return errors.New("worker protocol mismatch")
	}
	p, err := ReadServiceProfile(profile)
	unbound := os.IsNotExist(err)
	if err != nil && !unbound {
		reply("ready", err)
		return err
	}
	var a *Agent
	if !unbound {
		for k, v := range p.Environment {
			if err = os.Setenv(k, v); err != nil {
				return err
			}
		}
		c, err := Load(p.Config)
		if err != nil {
			reply("ready", err)
			return err
		}
		unlock, err := LockRun(c.Path)
		if err != nil {
			reply("ready", err)
			return err
		}
		defer unlock()
		if err = validateWorkerStorage(c.Path); err != nil {
			reply("ready", err)
			return err
		}
		file, err := os.OpenFile(filepath.Join(filepath.Dir(c.Path), "agent.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			reply("ready", err)
			return err
		}
		defer file.Close()
		old := log.Writer()
		log.SetOutput(file)
		defer log.SetOutput(old)
		a = New(c)
	}
	if err = reply("ready", nil); err != nil {
		return err
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	active := false
	commands := make(chan workerproto.Message)
	go func() {
		defer close(commands)
		for {
			var m workerproto.Message
			if d.Decode(&m) != nil {
				return
			}
			select {
			case commands <- m:
			case <-workerCtx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			cancel()
			if active {
				select {
				case <-done:
				case <-time.After(12 * time.Second):
				}
			}
			return nil
		case err := <-done:
			return err
		case m, ok := <-commands:
			if !ok {
				cancel()
				if active {
					select {
					case <-done:
					case <-time.After(12 * time.Second):
					}
				}
				return nil
			}
			if m.Protocol != workerproto.Version {
				return errors.New("worker protocol mismatch")
			}
			switch m.Type {
			case "activate":
				if active {
					return errors.New("worker already active")
				}
				active = true
				go func() {
					if a == nil {
						<-workerCtx.Done()
						done <- nil
						return
					}
					log.Printf("worker starting version=%s permissions=full", buildinfo.Current())
					done <- a.Run(workerCtx)
				}()
				err = reply("active", nil)
			case "pause":
				var pauseErr error
				if a != nil {
					pauseCtx, stop := context.WithTimeout(workerCtx, 15*time.Second)
					pauseErr = a.PauseUpdates(pauseCtx, m.Force)
					stop()
				}
				err = reply("paused", pauseErr)
			case "resume":
				if a != nil {
					a.ResumeUpdates()
				}
				err = reply("active", nil)
			case "stop":
				cancel()
				if active {
					select {
					case <-done:
					case <-time.After(12 * time.Second):
						return errors.New("worker shutdown timed out")
					}
				}
				return reply("stopped", nil)
			default:
				return fmt.Errorf("unsupported worker command")
			}
			if err != nil {
				return err
			}
		}
	}
}

func validateWorkerStorage(config string) error {
	path := filepath.Join(filepath.Dir(config), "history.db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: true, Timeout: 2 * time.Second})
	if err != nil {
		return fmt.Errorf("prepare history: %w", err)
	}
	defer db.Close()
	return db.View(func(tx *bolt.Tx) error {
		for err := range tx.Check() {
			if err != nil {
				return fmt.Errorf("prepare history: %w", err)
			}
		}
		return nil
	})
}
