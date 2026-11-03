package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func (s *Server) coordination(w http.ResponseWriter, r *http.Request) {
	userID, clientID := "", ""
	if strings.HasPrefix(r.URL.Path, "/api/agent/") {
		a, e := s.authenticateAgent(r)
		if e != nil {
			fail(w, 401, "device credentials denied")
			return
		}
		userID, clientID = a.UserID, a.ID
	} else {
		a, ok := s.require(w, r)
		if !ok {
			return
		}
		userID = a.UserID
	}
	if r.Method == "GET" {
		rows, e := s.Store.Agents(userID)
		if e != nil {
			fail(w, 500, "无法读取管家目录")
			return
		}
		host, e := s.Store.DefaultHost(userID)
		if e != nil {
			fail(w, 500, "无法读取默认主机")
			return
		}
		out := []protocol.AssistantRef{}
		if host.AgentID != "" {
			out = append(out, protocol.AssistantRef{Scope: "account", ClientID: host.AgentID, Name: "CEO", Online: s.hub.online(host.AgentID)})
		}
		for _, a := range rows {
			if a.Revoked || !a.Enrolled {
				continue
			}
			out = append(out, protocol.AssistantRef{Scope: "client", ClientID: a.ID, Name: a.Name + " · 管家", Online: s.hub.online(a.ID)})
			for _, p := range a.Services {
				if p.Scope == "" {
					out = append(out, protocol.AssistantRef{Scope: "project", ClientID: a.ID, ProjectID: p.ID, Name: p.Name + " · 管家", Online: s.hub.online(a.ID)})
				}
			}
		}
		reply(w, 200, map[string]any{"assistants": out, "defaultHost": host.AgentID})
		return
	}
	var in protocol.CoordinationCommand
	if !decode(w, r, &in) {
		return
	}
	if len(in.Prompt) > 32000 || len(in.RequestID) > 160 || len(in.Path) > 6 || len(in.Title) > 320 {
		fail(w, 400, "无效协作请求")
		return
	}
	in.AccountID = userID
	if (in.Operation == "resolve" || in.Operation == "dispatch") && in.Target.Scope == "account" {
		unlock := s.lockTodos(userID)
		defer unlock()
	}
	check := s.todoAuth(r, userID, clientID)
	call := func(target string, command protocol.CoordinationCommand) (json.RawMessage, error) {
		return s.coordinationCall(r.Context(), target, command, check)
	}
	owned := func(id string) bool {
		a, e := s.Store.Agent(id)
		return e == nil && a.UserID == userID && a.Enrolled && !a.Revoked
	}
	if in.Operation == "resolve" {
		if in.Target.Scope == "account" {
			host, e := s.Store.DefaultHost(userID)
			if e != nil || host.AgentID == "" {
				fail(w, 409, "请先选择默认主机，CEO 的工作区和 SaySo 保存在该主机")
				return
			}
			in.Target.ClientID = host.AgentID
		}
		if !owned(in.Target.ClientID) {
			fail(w, 404, "目标客户端不存在")
			return
		}
		raw, e := call(in.Target.ClientID, in)
		if e != nil {
			fail(w, 409, e.Error())
			return
		}
		reply(w, 200, raw)
		return
	}
	if in.Operation == "dispatch" {
		if in.Target.Scope == "account" {
			host, e := s.Store.DefaultHost(userID)
			if e != nil || host.AgentID == "" {
				fail(w, 409, "请先选择默认主机")
				return
			}
			in.Target.ClientID = host.AgentID
		}
		if clientID != "" {
			in.Source.ClientID = clientID
		}
		if !owned(in.Source.ClientID) || !owned(in.Target.ClientID) || in.Source.SessionID == "" || in.RequestID == "" {
			fail(w, 400, "派工需要真实来源会话、目标和请求 ID")
			return
		}
		validate := in
		validate.Operation = "validate"
		raw, e := call(in.Source.ClientID, validate)
		if e != nil {
			fail(w, 409, e.Error())
			return
		}
		var origin protocol.AssistantRef
		if json.Unmarshal(raw, &origin) != nil {
			fail(w, 502, "来源会话无效")
			return
		}
		in.Source = origin
		in.Path = origin.Path
		resolve := in
		resolve.Operation = "resolve"
		raw, e = call(in.Target.ClientID, resolve)
		if e != nil {
			fail(w, 409, e.Error())
			return
		}
		var resolved protocol.AssistantResolved
		if json.Unmarshal(raw, &resolved) != nil {
			fail(w, 502, "目标管家无效")
			return
		}
		in.Target = resolved.AssistantRef
		if len(in.Path) == 0 {
			in.Path = []string{in.Source.ClientID + "/" + in.Source.Scope + "/" + in.Source.ProjectID}
		}
		id := store.Hash(userID + "\x00" + in.Source.ClientID + "\x00" + in.Source.SessionID + "\x00" + in.RequestID)
		key := in.Target.ClientID + "/" + in.Target.Scope + "/" + in.Target.ProjectID
		for _, p := range in.Path {
			if p == key {
				fail(w, 409, "委派路径出现循环")
				return
			}
		}
		if len(in.Path) >= 6 {
			fail(w, 409, "已达到最大委派深度")
			return
		}
		in.Path = append(in.Path, key)
		hash := store.Hash(in.Prompt + "\x00" + key)
		route, e := s.Store.CoordinationRoute(id, &store.WorkRoute{ID: id, AccountID: userID, ReplyHandle: store.Token(), RequestHash: hash, Source: in.Source, Target: in.Target, Path: in.Path, CreatedAt: time.Now().UnixMilli()})
		if e != nil || route.AccountID != userID || route.RequestHash != hash {
			fail(w, 409, "派工请求 ID 冲突")
			return
		}
		in.WorkID, in.ReplyHandle, in.Path = route.ID, route.ReplyHandle, route.Path
		raw, e = call(route.Target.ClientID, in)
		if e != nil {
			fail(w, 409, e.Error())
			return
		}
		var accepted protocol.WorkAccepted
		if json.Unmarshal(raw, &accepted) != nil || accepted.Target.SessionID == "" {
			fail(w, 502, "目标未返回执行会话")
			return
		}
		if e = s.Store.SaveCoordinationTarget(route.ID, accepted.Target); e != nil {
			fail(w, 500, "任务已接收，路由确认失败；请使用原请求 ID 重试")
			return
		}
		assignment := in
		assignment.Operation = "assigned"
		assignment.Target = accepted.Target
		if _, e = call(route.Source.ClientID, assignment); e != nil {
			fail(w, 409, "任务已接收，来源确认失败；请用原请求 ID 重试："+e.Error())
			return
		}
		reply(w, 200, accepted)
		return
	}
	lookup := in.WorkID
	if in.ReplyHandle != "" {
		lookup = in.ReplyHandle
	}
	route, e := s.Store.CoordinationRoute(lookup, nil)
	if e != nil || route.AccountID != userID || !owned(route.Target.ClientID) || !owned(route.Source.ClientID) {
		fail(w, 404, "任务或来源已不可用")
		return
	}
	if in.Operation == "report" {
		if clientID != route.Target.ClientID || in.Report == nil || in.Report.Origin.SessionID != route.Target.SessionID || in.Report.WorkID != route.ID || in.Report.ID != store.Hash(route.ID+"\x00"+in.Report.TaskID) || len(in.Report.TaskID) > 160 || len(in.Report.Summary) > 12000 || len(in.Report.Related) > 8 {
			fail(w, 403, "结果来源不匹配")
			return
		}
		in.Report.ReplyHandle = route.ReplyHandle
		in.Report.Origin = route.Target
		for i, link := range in.Report.Related {
			child, e := s.Store.CoordinationRoute(link.WorkID, nil)
			if e != nil || child.AccountID != userID || child.ReplyHandle != link.ReplyHandle || child.Target.SessionID != link.Origin.SessionID {
				fail(w, 403, "关联结果来源不匹配")
				return
			}
			// A reference must descend from this execution, never from a sibling.
			p := child
			descendant := false
			for depth := 0; depth < 6; depth++ {
				if p.Source.ClientID == route.Target.ClientID && p.Source.SessionID == route.Target.SessionID {
					descendant = true
					break
				}
				if len(p.Path) <= len(route.Path) {
					break
				}
				parentID, e := s.Store.CoordinationParent(userID, p.Source.ClientID, p.Source.SessionID)
				if e != nil {
					break
				}
				p, e = s.Store.CoordinationRoute(parentID, nil)
				if e != nil {
					break
				}
			}
			if !descendant {
				fail(w, 403, "关联结果不属于本工作")
				return
			}
			in.Report.Related[i].Origin = child.Target
		}
		in.Source = route.Source
		raw, e := call(route.Source.ClientID, in)
		if e != nil {
			fail(w, 409, e.Error())
			return
		}
		reply(w, 200, raw)
		return
	}
	if in.Operation != "reply" && in.Operation != "read" && in.Operation != "cancel" {
		fail(w, 400, "不支持的协作操作")
		return
	}
	if in.Operation == "reply" && (strings.TrimSpace(in.Prompt) == "" || in.RequestID == "") {
		fail(w, 400, "回复需要指令和稳定请求 ID")
		return
	}
	// Device tools can control their own delegated work, not arbitrary siblings.
	if clientID != "" && (clientID != route.Source.ClientID || in.Source.SessionID != route.Source.SessionID) {
		fail(w, 403, "只能管理当前会话派出的工作")
		return
	}
	if time.Since(time.UnixMilli(route.CreatedAt)) > 30*24*time.Hour {
		fail(w, 410, "回复引用已过期，请打开原会话继续")
		return
	}
	if route.Target.SessionID == "" {
		fail(w, 409, "任务尚未确认接收，请重试派工")
		return
	}
	in.Source, in.Target, in.WorkID, in.ReplyHandle = route.Source, route.Target, route.ID, route.ReplyHandle
	raw, e := call(route.Target.ClientID, in)
	if e != nil {
		fail(w, 409, e.Error())
		return
	}
	reply(w, 200, raw)
}

func (s *Server) coordinationCall(ctx context.Context, target string, in protocol.CoordinationCommand, check func() bool) (json.RawMessage, error) {
	body, e := json.Marshal(in)
	if e != nil {
		return nil, e
	}
	raw, e := s.accountRelay(ctx, target, in.AccountID, protocol.Request{Type: "coordination.call", Body: body}, check)
	if e != nil && strings.Contains(e.Error(), "unknown request type") {
		return nil, errors.New("目标客户端尚不支持分层 Agent，请安装新版 MSI 或更新客户端")
	}
	return raw, e
}
