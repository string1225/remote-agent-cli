package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/windows/svc"
)

func TestWindowsServiceSurvivesLogoffAndStopsOnSCMShutdown(t *testing.T) {
	for _, stop := range []svc.Cmd{svc.Stop, svc.Shutdown} {
		requests := make(chan svc.ChangeRequest, 8)
		statuses := make(chan svc.Status, 16)
		done := make(chan uint32, 1)
		h := windowsServiceHandler{run: func(ctx context.Context, ready chan<- error) error {
			ready <- nil
			<-ctx.Done()
			return nil
		}}
		go func() { _, code := h.Execute(nil, requests, statuses); done <- code }()
		for {
			select {
			case status := <-statuses:
				if status.State == svc.Running {
					if status.Accepts&svc.AcceptSessionChange != 0 {
						t.Fatal("service must not depend on interactive session events")
					}
					goto running
				}
			case <-time.After(3 * time.Second):
				t.Fatal("service did not become ready")
			}
		}
	running:
		// WTS_SESSION_LOGOFF = 6. Even an unexpected session notification must
		// leave the worker alive; only Stop/Shutdown cancels its lifetime.
		requests <- svc.ChangeRequest{Cmd: svc.SessionChange, EventType: 6}
		requests <- svc.ChangeRequest{Cmd: svc.Interrogate}
		select {
		case status := <-statuses:
			if status.State != svc.Running {
				t.Fatal("logoff stopped service")
			}
		case <-done:
			t.Fatal("logoff cancelled worker")
		case <-time.After(3 * time.Second):
			t.Fatal("interrogate did not respond")
		}
		requests <- svc.ChangeRequest{Cmd: stop}
		select {
		case code := <-done:
			if code != 0 {
				t.Fatal("graceful stop failed")
			}
		case <-time.After(3 * time.Second):
			t.Fatal("SCM stop did not cancel and join worker")
		}
	}
}

func TestWindowsServiceReportsStartupFailure(t *testing.T) {
	h := windowsServiceHandler{run: func(_ context.Context, ready chan<- error) error {
		err := errors.New("invalid fixture configuration")
		ready <- err
		return err
	}}
	specific, code := h.Execute(nil, make(chan svc.ChangeRequest), make(chan svc.Status, 8))
	if !specific || code == 0 {
		t.Fatal("failed startup was reported as success")
	}
}
