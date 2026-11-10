package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func validExecutionReport(report *protocol.WorkReport) bool {
	if report.Event == "completed" || report.Event == "failed" {
		return report.Stage == ""
	}
	if report.Stage != "plan" && report.Stage != "validation" && report.Stage != "predeploy" {
		return false
	}
	return report.Event == "checkpoint" || report.Event == "review_required" || report.Event == "blocked" || report.Event == "needs_input"
}

// Validates only child metadata via the existing authenticated relay. Works with
// old clients too, but registering their outbox still requires a new runtime.
func (s *Server) executionChild(r *http.Request, account string, source, target protocol.AssistantRef, check func() bool) (protocol.Conversation, error) {
	for cursor := uint64(0); ; {
		body, _ := json.Marshal(protocol.SessionControl{Operation: "list", CallerAgentID: source.ClientID, CallerSessionID: source.SessionID, Cursor: cursor})
		raw, err := s.accountRelay(r.Context(), target.ClientID, account, protocol.Request{Type: "sessions.control", Body: body}, check)
		if err != nil {
			return protocol.Conversation{}, err
		}
		var page protocol.ConversationPage
		if err := json.Unmarshal(raw, &page); err != nil {
			return protocol.Conversation{}, err
		}
		for _, child := range page.Sessions {
			if child.ID == target.SessionID && child.ServiceID == target.ProjectID && child.DeletedAt == 0 && child.ParentAgentID == source.ClientID && child.ParentSessionID == source.SessionID {
				return child, nil
			}
		}
		if page.NextCursor == 0 || page.NextCursor <= cursor {
			break
		}
		cursor = page.NextCursor
	}
	return protocol.Conversation{}, errors.New("目标不是来源管理会话的直接执行子会话")
}

func (s *Server) subscribeExecution(w http.ResponseWriter, r *http.Request, in protocol.CoordinationCommand, client string, check func() bool) {
	if client == "" || in.Source.SessionID == "" || in.Target.SessionID == "" || strings.TrimSpace(in.Topic) == "" || len(in.Topic) > 320 || len(in.Checkpoints) > 3 {
		fail(w, 400, "订阅需要真实来源、目标和有界topic/checkpoints")
		return
	}
	in.Source.ClientID = client
	for _, stage := range in.Checkpoints {
		if stage != "plan" && stage != "validation" && stage != "predeploy" {
			fail(w, 400, "无效检查点")
			return
		}
	}
	device, err := s.Store.Agent(in.Target.ClientID)
	if err != nil || device.UserID != in.AccountID || device.Revoked || !device.Enrolled {
		fail(w, 404, "目标客户端不存在")
		return
	}
	ordinary := false
	for _, p := range device.Services {
		if p.ID == in.Target.ProjectID && p.Scope == "" {
			ordinary = true
		}
	}
	if !ordinary {
		fail(w, 403, "目标须为原普通项目")
		return
	}
	validate := in
	validate.Operation = "validate"
	raw, err := s.coordinationCall(r.Context(), client, validate, check)
	var origin protocol.AssistantRef
	if err != nil || json.Unmarshal(raw, &origin) != nil || (origin.Scope != "account" && origin.Scope != "client" && origin.Scope != "project") {
		fail(w, 403, "来源不是已核实的管理会话")
		return
	}
	if _, err = s.executionChild(r, in.AccountID, origin, in.Target, check); err != nil {
		fail(w, 403, err.Error())
		return
	}
	id := store.Hash("execution\x00" + in.AccountID + "\x00" + client + "\x00" + origin.SessionID + "\x00" + in.Target.ClientID + "\x00" + in.Target.SessionID)
	hash := store.Hash(strings.ToLower(strings.Join(strings.Fields(in.Topic), " ")) + "\x00" + in.Target.ProjectID)
	path := append(append([]string{}, origin.Path...), in.Target.ClientID+"/execution/"+in.Target.SessionID)
	if len(path) > 6 {
		fail(w, 409, "已达到最大来源深度")
		return
	}
	route, err := s.Store.CoordinationRoute(id, &store.WorkRoute{ID: id, Direct: true, Topic: in.Topic, Checkpoints: in.Checkpoints, AccountID: in.AccountID, ReplyHandle: store.Token(), RequestHash: hash, Source: origin, Target: in.Target, Path: path, CreatedAt: time.Now().UnixMilli()})
	if err != nil || route.RequestHash != hash || !route.Direct {
		fail(w, 409, "执行订阅冲突，保留原映射")
		return
	}
	if err = s.Store.SaveCoordinationTarget(route.ID, route.Target); err != nil {
		fail(w, 500, "无法保存执行路由")
		return
	}
	if err = s.Store.SaveCoordinationCheckpoints(route.ID, in.Checkpoints); err != nil {
		fail(w, 500, "无法保存检查点配置")
		return
	}
	in.Operation, in.WorkID, in.ReplyHandle, in.Source, in.Target, in.Path = "watch", route.ID, route.ReplyHandle, origin, route.Target, route.Path
	raw, err = s.coordinationCall(r.Context(), route.Target.ClientID, in, check)
	if err != nil {
		fail(w, 409, "路由已保存，执行端事件回报待正常升级/重试："+err.Error())
		return
	}
	assignment := in
	assignment.Operation = "assigned"
	var accepted protocol.WorkAccepted
	if json.Unmarshal(raw, &accepted) != nil {
		fail(w, 502, "无效执行订阅确认")
		return
	}
	assignment.TaskID = accepted.TaskID
	if _, err = s.coordinationCall(r.Context(), client, assignment, check); err != nil {
		fail(w, 409, "订阅已建立，来源确认待恢复："+err.Error())
		return
	}
	reply(w, 200, raw)
}
