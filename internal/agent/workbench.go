package agent

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func (a *Agent) handleWorkbench(ctx context.Context, req protocol.Request, send func(protocol.Event) error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	var result any
	var err error
	if req.Type == "diagnostics.snapshot" {
		a.mu.Lock()
		providers := make(map[string]bool, len(a.Providers))
		for name, provider := range a.Providers {
			providers[name] = provider.Available
		}
		result = map[string]any{"version": buildinfo.Current(), "os": runtime.GOOS, "arch": runtime.GOARCH, "permissionMode": "full", "activeRuns": len(a.runs), "acceptedTasks": len(a.taskCancels), "providers": providers, "checkedAt": time.Now().UnixMilli()}
		a.mu.Unlock()
	} else {
		a.mu.Lock()
		config := a.Config
		config.AllowedRoots = append([]string{}, a.Config.AllowedRoots...)
		var project protocol.Service
		for _, service := range a.Config.Services {
			if service.ID == req.ServiceID {
				project = service
				break
			}
		}
		a.mu.Unlock()
		if project.ID == "" {
			err = errors.New("项目不存在")
		} else if req.Type == "sessions.children" {
			var session protocol.Conversation
			session, err = a.historyStore().get(req.SessionID)
			if err == nil && (session.DeletedAt != 0 || session.ServiceID != project.ID) {
				err = errors.New("会话不属于当前项目")
			}
			if err == nil {
				result, err = a.historyStore().childSessions(protocol.SessionControl{CallerAgentID: config.ID, CallerSessionID: session.ID, Cursor: req.Cursor})
			}
		} else {
			var workspace string
			workspace, err = config.Workspace(project.Workspace)
			if err == nil {
				if req.Type == "workspace.changes" {
					result, err = workspaceChanges(ctx, workspace)
				} else {
					result, err = workspaceDiff(ctx, workspace, req.Path)
				}
			}
		}
	}
	if err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()})
	} else {
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: result})
	}
}
