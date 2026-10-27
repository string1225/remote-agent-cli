package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

const taskRunTimeout = 2 * time.Hour

var errTaskRunTimeout = errors.New("task execution limit reached")

func taskTimeoutError() error {
	return fmt.Errorf("任务执行超时：已达到 2 小时上限。已保留会话、已完成输出和原生上下文；检查已完成内容后，可在同一会话发送“继续”接续执行，避免重复操作。(%w)", context.DeadlineExceeded)
}

// Accept a turn in an existing Web session before acknowledging it. Its worker
// belongs to the daemon, so closing a browser/relay peer cannot cancel the turn.
func (a *Agent) startSessionTask(req protocol.Request) (protocol.TaskStarted, error) {
	var result protocol.TaskStarted
	if strings.TrimSpace(req.Prompt) == "" && len(req.Attachments) > 0 {
		req.Prompt = "请查看附件。"
	}
	if req.SessionID == "" || req.ID == "" || len(req.ID) > 100 || strings.TrimSpace(req.Prompt) == "" || len(req.Prompt) > 32000 {
		return result, errors.New("无效会话任务；内容需为 1–32000 字节")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.taskCtx == nil || a.taskCtx.Err() != nil {
		return result, errors.New("客户端正在停止")
	}
	if a.maintenance || a.toolInstalling {
		return result, errUpdating
	}
	if a.removingProjects[req.ServiceID] {
		return result, errors.New("项目正在删除")
	}
	var service protocol.Service
	for _, s := range a.Config.Services {
		if s.ID == req.ServiceID {
			service = s
			break
		}
	}
	if service.ID == "" || !a.Providers[service.Provider].Available {
		return result, errors.New("项目不存在或执行器不可用")
	}
	root, err := a.Config.Workspace(service.Workspace)
	if err != nil {
		return result, err
	}
	h := a.historyStore()
	session, err := h.get(req.SessionID)
	if err != nil {
		return result, err
	}
	sessionRoot, rootErr := a.Config.Workspace(session.Workspace)
	if rootErr != nil || session.ServiceID != service.ID || session.Provider != service.Provider || sessionRoot != root || session.Kind == "im" {
		return result, errors.New("session does not belong to this Web project")
	}
	// A lost acknowledgement can be retried with the same request ID, even
	// after completion, without replaying a turn that may have side effects.
	if session.TaskID == req.ID {
		return protocol.TaskStarted{TaskID: req.ID, Session: session}, nil
	}
	seen := false
	if err := h.transact(false, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("entries")).Bucket([]byte(session.ID))
		if b == nil {
			return errors.New("session history missing")
		}
		return b.ForEach(func(_, data []byte) error {
			var entry protocol.ChatEntry
			if json.Unmarshal(data, &entry) == nil && entry.TurnID == req.ID && entry.Role == "user" {
				seen = true
			}
			return nil
		})
	}); err != nil {
		return result, err
	}
	if seen {
		return protocol.TaskStarted{TaskID: req.ID, Session: session}, nil
	}
	if session.TaskStatus == "queued" || session.TaskStatus == "running" || a.runs[session.ID] != nil {
		return result, errSessionRunning
	}
	if len(a.taskCancels) >= maxBackgroundTasks {
		return result, errors.New("客户端任务队列已满，请稍后重试")
	}
	refs, _, _, err := a.resolveAttachments(session.ID, req.Attachments)
	if err != nil {
		return result, err
	}
	err = h.transact(true, func(tx *bolt.Tx) error {
		if _, e := appendHistory(tx, session.ID, protocol.ChatEntry{TurnID: req.ID, Role: "user", Kind: "message", Text: req.Prompt, Attachments: refs}); e != nil {
			return e
		}
		var e error
		session, e = sessionFrom(tx, session.ID)
		if e != nil {
			return e
		}
		session.TaskID, session.TaskStatus, session.TaskError = req.ID, "queued", ""
		return saveSession(tx, session)
	})
	if err != nil {
		return result, err
	}
	result = protocol.TaskStarted{TaskID: req.ID, Session: session}
	taskCtx, cancel := context.WithCancel(a.taskCtx)
	in := protocol.TaskStart{ServiceID: service.ID, Prompt: req.Prompt, Source: "web", Notify: session.NotifyAgentID != "", NotifyAgentID: session.NotifyAgentID, NotifyServiceID: session.NotifyServiceID, NotifySessionID: session.NotifySessionID}
	if a.enqueueTaskLocked(queuedTask{ctx: taskCtx, cancel: cancel, recorded: true, attachments: append([]string{}, req.Attachments...), start: in, result: result}) {
		return result, nil
	}
	cancel()
	err = errors.New("客户端任务队列已满，请稍后重试")
	_ = h.update(session.ID, func(v *protocol.Conversation) { v.TaskStatus, v.TaskError = "failed", err.Error() })
	_, _ = h.append(session.ID, protocol.ChatEntry{TurnID: req.ID, Role: "system", Kind: "error", Text: err.Error()})
	return result, err
}
