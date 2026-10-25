package agent

import (
	"context"
	"errors"
	"time"
)

var errUpdating = errors.New("客户端正在更新，请稍后重试")

func (a *Agent) beginOperation() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.maintenance || a.toolInstalling {
		return errUpdating
	}
	a.operations++
	return nil
}
func (a *Agent) endOperation()  { a.mu.Lock(); a.operations--; a.mu.Unlock() }
func (a *Agent) ResumeUpdates() { a.mu.Lock(); a.maintenance = false; a.mu.Unlock() }

// PauseUpdates closes the admission gate before inspecting busy work. Forced
// cancellation uses existing run/task contexts and leaves recovery history intact.
func (a *Agent) PauseUpdates(ctx context.Context, force bool) error {
	a.mu.Lock()
	if a.toolInstalling {
		a.mu.Unlock()
		return errors.New("Codex CLI 正在安装或更新，完成后才能切换客户端版本")
	}
	if a.recordingActivity > 0 {
		a.mu.Unlock()
		return errors.New("recording or audio upload is active; finish it before updating")
	}
	a.maintenance = true
	busy := a.busy || len(a.taskCancels) > 0 || a.operations > 0
	if busy && !force {
		a.maintenance = false
		a.mu.Unlock()
		return errors.New("client has active tasks; retry after they finish or use --force")
	}
	if force {
		if a.activeCancel != nil {
			a.activeCancel()
		}
		for _, cancel := range a.taskCancels {
			cancel()
		}
	}
	a.mu.Unlock()
	for {
		a.mu.Lock()
		idle := !a.busy && len(a.taskCancels) == 0 && a.operations == 0
		a.mu.Unlock()
		if idle {
			return nil
		}
		select {
		case <-ctx.Done():
			a.ResumeUpdates()
			return errors.New("tasks have not stopped safely; update deferred")
		case <-time.After(50 * time.Millisecond):
		}
	}
}
