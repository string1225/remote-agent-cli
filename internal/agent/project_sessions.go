package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func (a *Agent) createProjectSession(ctx context.Context, req protocol.Request) (protocol.TaskStarted, error) {
	var in protocol.ProjectSessionCreate
	var wire protocol.ProjectSessionRelay
	err := json.Unmarshal(req.Body, &wire)
	in = protocol.ProjectSessionCreate{Title: wire.Title, Message: string(wire.Message), Execute: wire.Execute}
	if err != nil || len(in.Message) > 32000 || !utf8.ValidString(in.Message) || utf8.RuneCountInString(strings.TrimSpace(in.Title)) > 80 || in.Execute && strings.TrimSpace(in.Message) == "" {
		return protocol.TaskStarted{}, errors.New("无效会话参数：标题最多 80 字、消息最多 32000 字节；立即执行需填写消息")
	}
	if err := a.beginOperation(); err != nil {
		return protocol.TaskStarted{}, err
	}
	defer a.endOperation()
	a.mu.Lock()
	var service protocol.Service
	for _, candidate := range a.Config.Services {
		if candidate.ID == req.ServiceID && candidate.Scope == "" && !a.removingProjects[candidate.ID] {
			service = candidate
			break
		}
	}
	a.mu.Unlock()
	if service.ID == "" {
		return protocol.TaskStarted{}, errors.New("项目不存在或正在删除")
	}
	if in.Execute {
		return a.startTask(ctx, protocol.TaskStart{ServiceID: service.ID, Title: in.Title, Prompt: in.Message, Source: "web"}, nil)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.removingProjects[service.ID] {
		return protocol.TaskStarted{}, errors.New("项目正在删除")
	}
	session, err := a.historyStore().createWithMessage(service, in.Title, in.Message)
	if err != nil {
		return protocol.TaskStarted{}, err
	}
	return protocol.TaskStarted{Session: publicSession(session), TaskID: session.TaskID}, nil
}
