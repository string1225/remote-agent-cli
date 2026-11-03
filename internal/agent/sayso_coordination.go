package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

// Called with s.mu held. Only a local, durable outbox is created here; network
// dispatch occurs outside this lock and can be retried after disconnect/restart.
func (s *saysoStore) dispatchRemote(action *soAction, target protocol.AssistantRef) (any, error) {
	if action.Status != "ready" && action.Status != "failed" {
		return nil, errors.New("请先确认 Action 并标记为 Ready")
	}
	if (target.Scope != "client" && target.Scope != "project") || target.ClientID == "" || len(target.ClientID) > 160 || len(target.ProjectID) > 160 || len(target.Name) > 320 {
		return nil, errors.New("请选择客户端管家或项目管家")
	}
	project, discussion := s.project(action.ProjectID), s.session(action.SessionID)
	if project == nil || discussion == nil {
		return nil, errors.New("讨论不存在")
	}
	var service protocol.Service
	for _, v := range s.a.services() {
		if v.ID == project.ServiceID {
			service = v
			break
		}
	}
	if service.Scope != "account" || service.OwnerAccount == "" {
		return nil, errors.New("跨客户端 Action 请从 CEO 的 SaySo 发起；原项目讨论和录音仍保留")
	}
	h := s.a.historyStore()
	parent, e := h.get(project.ActionSessionID)
	if e != nil || parent.DeletedAt != 0 {
		parent, e = h.create(service, "SaySo · Action 回报")
		if e != nil {
			return nil, e
		}
		project.ActionSessionID = parent.ID
	}
	prompt := fmt.Sprintf("处理用户已确认的 SaySo Action，遵守目标项目规则，完成并验证。\nAction：%s\n详细说明：%s\n来源：%s\n讨论：%s", action.Title, action.Detail, action.Source, discussion.Title)
	run := &soRun{ID: store.Token(), ActionID: action.ID, Status: "running", StartedAt: soNow(), Target: &target}
	run.Command = &protocol.CoordinationCommand{Operation: "dispatch", RequestID: run.ID, Source: protocol.AssistantRef{ClientID: s.a.configSnapshot().ID, ProjectID: service.ID, SessionID: parent.ID}, Target: target, Title: soBound(action.Title, 80), Prompt: soBound(prompt, 32000)}
	s.state.Runs = append([]*soRun{run}, s.state.Runs...)
	action.RunID, action.Status, action.UpdatedAt = run.ID, "running", soNow()
	return run, nil
}

func (a *Agent) configSnapshot() Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	c := a.Config
	c.Services = append([]protocol.Service(nil), c.Services...)
	return c
}

func (a *Agent) flushSaysoCoordination(ctx context.Context) {
	s, e := a.saysoStore()
	if e != nil {
		return
	}
	s.mu.Lock()
	pending := []soRun{}
	for _, r := range s.state.Runs {
		if r.Command != nil && r.Status == "running" {
			c := *r
			command := *r.Command
			c.Command = &command
			pending = append(pending, c)
		}
	}
	s.mu.Unlock()
	for _, r := range pending {
		if ctx.Err() != nil {
			return
		}
		if r.WorkID == "" {
			// Even after cancellation, resolve an in-flight dispatch with the original
			// idempotency key before cancelling, so no accepted remote task is orphaned.
			raw, err := ClientAPI(ctx, a.configSnapshot(), "POST", "/api/agent/coordination", r.Command)
			var accepted protocol.WorkAccepted
			if err == nil {
				err = json.Unmarshal(raw, &accepted)
			}
			s.mu.Lock()
			for _, live := range s.state.Runs {
				if live.ID == r.ID {
					if err != nil {
						live.Error = soBound(err.Error(), 500)
					} else {
						live.WorkID, live.ReplyHandle, live.WebSessionID, live.TaskID = accepted.WorkID, accepted.ReplyHandle, accepted.Target.SessionID, accepted.TaskID
						live.Target = &accepted.Target
						live.Error = ""
					}
					_ = s.save()
					break
				}
			}
			s.mu.Unlock()
			continue
		}
		command := protocol.CoordinationCommand{Operation: "read", Source: r.Command.Source, WorkID: r.WorkID}
		raw, err := ClientAPI(ctx, a.configSnapshot(), "POST", "/api/agent/coordination", command)
		if err != nil {
			continue
		}
		var state protocol.SessionState
		if json.Unmarshal(raw, &state) != nil {
			continue
		}
		if !state.Complete {
			if r.CancelRequested && (state.Session.TaskStatus == "running" || state.Session.TaskStatus == "queued") {
				command.Operation = "cancel"
				command.TaskID = state.TaskID
				_, _ = ClientAPI(ctx, a.configSnapshot(), "POST", "/api/agent/coordination", command)
			}
			continue
		}
		s.mu.Lock()
		for _, live := range s.state.Runs {
			if live.ID == r.ID {
				live.Status = "done"
				if state.Session.TaskError != "" || !state.Captured {
					live.Status = "failed"
					live.Error = state.Session.TaskError
					if live.Error == "" {
						live.Error = "未捕获可确认的最终结果，请检查执行会话"
					}
				}
				live.Output = soBound(state.Result, 6000)
				live.FinishedAt = soNow()
				if action := s.action(live.ActionID); action != nil && action.RunID == live.ID {
					action.Status = live.Status
					action.UpdatedAt = soNow()
				}
				_ = s.save()
				break
			}
		}
		s.mu.Unlock()
	}
}
