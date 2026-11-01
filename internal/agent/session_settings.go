package agent

import (
	"errors"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// Deletion is a durable tombstone, retaining the original history, native
// context, attachment references and task records instead of unlinking files.
func (a *Agent) updateSession(req protocol.Request) (protocol.Conversation, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	result, err := a.historyStore().get(req.SessionID)
	if err != nil {
		return result, err
	}
	if req.Type == "sessions.rename" && (strings.TrimSpace(req.Title) == "" || len([]rune(strings.TrimSpace(req.Title))) > 80) {
		return result, errors.New("会话名称需为 1–80 字")
	}
	if result.ServiceID != req.ServiceID || (result.DeletedAt != 0 && req.Type != "sessions.delete") {
		return result, errors.New("会话不存在或不属于当前项目")
	}
	if req.Type == "sessions.delete" {
		if a.runs[result.ID] != nil {
			return result, errors.New("请先停止此会话的任务并处理排队消息，再删除会话")
		}
		for _, task := range a.taskRecords {
			if task.result.Session.ID == result.ID {
				return result, errors.New("请先停止此会话的任务并处理排队消息，再删除会话")
			}
		}
	}
	err = a.historyStore().update(result.ID, func(s *protocol.Conversation) {
		switch req.Type {
		case "sessions.rename":
			s.Title, s.CustomTitle = strings.TrimSpace(req.Title), true
		case "sessions.archive":
			s.ArchivedAt = 0
			if req.Archived {
				s.ArchivedAt = time.Now().UnixMilli()
			}
		case "sessions.delete":
			s.DeletedAt = time.Now().UnixMilli()
		}
		s.UpdatedAt = time.Now().UnixMilli()
		result = *s
	})
	return result, err
}
