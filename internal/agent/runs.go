package agent

import (
	"context"
	"errors"
)

var errSessionRunning = errors.New("this session is already running a task")

// The agent registry protects native session context and tracks every process
// for maintenance. Request IDs are scoped to a peer, not to the whole device.
// Call with a.mu held.
func (a *Agent) trackRunLocked(p *peer, requestID, sessionID string, cancel context.CancelFunc) {
	if a.runs == nil {
		a.runs = map[string]context.CancelFunc{}
	}
	a.runs[sessionID] = cancel
	p.mu.Lock()
	if p.runs == nil {
		p.runs = map[string]context.CancelFunc{}
	}
	p.runs[requestID] = cancel
	p.mu.Unlock()
}

func (a *Agent) untrackRun(p *peer, requestID, sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	p.mu.Lock()
	delete(p.runs, requestID)
	p.mu.Unlock()
	delete(a.runs, sessionID)
}
