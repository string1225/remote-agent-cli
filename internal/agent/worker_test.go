package agent

import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/workerproto"
)

func TestWorkerPrepareUnboundActivateAndStop(t *testing.T) {
	toWorker, in := io.Pipe()
	out, fromWorker := io.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- RunWorker(ctx, filepath.Join(t.TempDir(), "missing-profile"), toWorker, fromWorker)
		fromWorker.Close()
	}()
	e, d := json.NewEncoder(in), json.NewDecoder(out)
	for _, step := range []struct{ send, want string }{{"prepare", "ready"}, {"activate", "active"}, {"pause", "paused"}, {"resume", "active"}, {"stop", "stopped"}} {
		if err := e.Encode(workerproto.Message{Protocol: 1, Type: step.send}); err != nil {
			t.Fatal(err)
		}
		var m workerproto.Message
		if err := d.Decode(&m); err != nil || m.Type != step.want || m.Error != "" {
			t.Fatal(step, m, err)
		}
	}
	in.Close()
	out.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestMaintenanceGateRetainsActiveWorkAndForceCancels(t *testing.T) {
	a := &Agent{busy: true}
	cancelled := make(chan struct{})
	a.activeCancel = func() { close(cancelled); go func() { a.mu.Lock(); a.busy = false; a.mu.Unlock() }() }
	if err := a.PauseUpdates(context.Background(), false); err == nil {
		t.Fatal("busy client paused without force")
	}
	select {
	case <-cancelled:
		t.Fatal("default update canceled a task")
	default:
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.PauseUpdates(ctx, true); err != nil {
		t.Fatal(err)
	}
	if err := a.beginOperation(); err == nil {
		t.Fatal("new work admitted while updating")
	}
	a.ResumeUpdates()
	if err := a.beginOperation(); err != nil {
		t.Fatal(err)
	}
	a.endOperation()
}

func TestUpdateNeverInterruptsRecording(t *testing.T) {
	a := &Agent{recordingActivity: 1}
	if err := a.PauseUpdates(context.Background(), true); err == nil {
		t.Fatal("forced update interrupted recording")
	}
	if err := a.beginOperation(); err != nil {
		t.Fatal("recording gate left client paused")
	}
	a.endOperation()
}
