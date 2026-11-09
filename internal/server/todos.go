package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

// The server persists only the host pointer. Calls are serialized with host
// migration; todo bodies and reply receipts remain on the designated client.
func (s *Server) lockTodos(userID string) func() {
	value, _ := s.todoLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (s *Server) defaultHost(w http.ResponseWriter, r *http.Request) {
	var account struct{ UserID string }
	sourceID := ""
	if r.URL.Path == "/api/agent/default-host" {
		device, e := s.authenticateAgent(r)
		if e != nil {
			fail(w, 401, "device credentials denied")
			return
		}
		account.UserID, sourceID = device.UserID, device.ID
	} else {
		user, ok := s.require(w, r)
		if !ok {
			return
		}
		account.UserID = user.UserID
	}
	unlock := s.lockTodos(account.UserID)
	defer unlock()
	before, err := s.Store.DefaultHost(account.UserID)
	if err != nil {
		fail(w, 500, "无法读取默认主机")
		return
	}
	if r.Method == "GET" {
		s.hostView(w, before)
		return
	}
	check := s.todoAuth(r, account.UserID, sourceID)
	before, err = s.recoverTodoMigration(r.Context(), account.UserID, before, check)
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	var in struct {
		AgentID string `json:"agentId"`
	}
	if !decode(w, r, &in) {
		return
	}
	if before.AgentID == in.AgentID {
		s.hostView(w, before)
		return
	}
	target, err := s.Store.Agent(in.AgentID)
	if err != nil || target.UserID != account.UserID || target.Revoked || !target.Enrolled {
		fail(w, 404, "请选择本账号已绑定的客户端")
		return
	}
	next := store.DefaultHost{AgentID: target.ID, Generation: store.Token()}
	call := protocol.TodoStorageCall{AccountID: account.UserID, Generation: next.Generation, TodoCommand: protocol.TodoCommand{Operation: "probe"}}
	if _, err = s.todoCall(r.Context(), next.AgentID, call, check); err != nil {
		fail(w, 409, err.Error())
		return
	}
	if before.AgentID != "" {
		_, e := s.coordinationCall(r.Context(), before.AgentID, protocol.CoordinationCommand{Operation: "host-move-probe", AccountID: account.UserID}, check)
		if e != nil && !strings.Contains(e.Error(), "尚不支持分层 Agent") {
			fail(w, 409, e.Error())
			return
		}
		original := before
		before.MigrationID = store.Token()
		if err = s.Store.SetDefaultHost(account.UserID, original, before); err != nil {
			fail(w, 409, err.Error())
			return
		}
		committed := false
		defer func() {
			if !committed {
				_, _ = s.recoverTodoMigration(context.WithoutCancel(r.Context()), account.UserID, before, func() bool { return true })
			}
		}()
		if _, err = s.todoCall(r.Context(), before.AgentID, protocol.TodoStorageCall{AccountID: account.UserID, Generation: before.Generation, MigrationID: before.MigrationID, TodoCommand: protocol.TodoCommand{Operation: "freeze"}}, check); err != nil {
			fail(w, 409, "迁移冻结失败，原清单保留："+err.Error())
			return
		}
		// Each destination generation is new. Failed or interrupted copies are
		// never visible; the complete former generation remains recoverable.
		cursor := ""
		for {
			raw, e := s.todoCall(r.Context(), before.AgentID, protocol.TodoStorageCall{AccountID: account.UserID, Generation: before.Generation, TodoCommand: protocol.TodoCommand{Operation: "export", Cursor: cursor}}, check)
			if e != nil {
				fail(w, 409, "迁移失败，仍保留原默认主机："+e.Error())
				return
			}
			var page protocol.TodoExport
			if json.Unmarshal(raw, &page) != nil {
				fail(w, 502, "原主机返回无效备份")
				return
			}
			call.Operation, call.Records = "import", page.Records
			if _, e = s.todoCall(r.Context(), next.AgentID, call, check); e != nil {
				fail(w, 409, "迁移失败，仍保留原默认主机："+e.Error())
				return
			}
			if page.NextCursor == "" {
				break
			}
			if page.NextCursor == cursor {
				fail(w, 502, "备份游标没有推进")
				return
			}
			cursor = page.NextCursor
		}
		if !check() {
			fail(w, 401, "登录状态已失效")
			return
		}
		if err = s.Store.SetDefaultHost(account.UserID, before, next); err != nil {
			fail(w, 409, err.Error())
			return
		}
		committed = true
		s.hostView(w, next)
		return
	}
	if !check() {
		fail(w, 401, "登录状态已失效")
		return
	}
	if err = s.Store.SetDefaultHost(account.UserID, before, next); err != nil {
		fail(w, 409, err.Error())
		return
	}
	s.hostView(w, next)
}

// A persisted marker lets a restarted site thaw the former canonical host.
// The host records resolved migration IDs, so a delayed freeze cannot undo it.
func (s *Server) recoverTodoMigration(ctx context.Context, userID string, host store.DefaultHost, check func() bool) (store.DefaultHost, error) {
	if host.MigrationID == "" {
		return host, nil
	}
	_, err := s.todoCall(ctx, host.AgentID, protocol.TodoStorageCall{AccountID: userID, Generation: host.Generation, MigrationID: host.MigrationID, TodoCommand: protocol.TodoCommand{Operation: "unfreeze"}}, check)
	if err != nil {
		return host, fmt.Errorf("上次主机迁移未结束，原数据已保留；请连接原默认主机以恢复：%w", err)
	}
	next := host
	next.MigrationID = ""
	if err = s.Store.SetDefaultHost(userID, host, next); err != nil {
		return host, err
	}
	return next, nil
}

func (s *Server) hostView(w http.ResponseWriter, host store.DefaultHost) {
	name := ""
	if a, err := s.Store.Agent(host.AgentID); err == nil {
		name = a.Name
	}
	reply(w, 200, map[string]any{"agentId": host.AgentID, "name": name, "online": s.hub.online(host.AgentID)})
}

func (s *Server) todoAuth(r *http.Request, userID, sourceID string) func() bool {
	return func() bool {
		if sourceID != "" {
			a, e := s.authenticateAgent(r)
			return e == nil && a.ID == sourceID && a.UserID == userID
		}
		v, _, e := s.session(r)
		return e == nil && v.UserID == userID
	}
}

func (s *Server) todos(w http.ResponseWriter, r *http.Request) {
	userID, sourceID := "", ""
	if r.URL.Path == "/api/agent/todos" {
		a, e := s.authenticateAgent(r)
		if e != nil {
			fail(w, 401, "device credentials denied")
			return
		}
		userID, sourceID = a.UserID, a.ID
	} else {
		v, ok := s.require(w, r)
		if !ok {
			return
		}
		userID = v.UserID
	}
	var in protocol.TodoCommand
	if r.Method == "GET" {
		in.Operation = "list"
		in.Status = r.URL.Query().Get("status")
		in.Search = r.URL.Query().Get("search")
		in.Cursor = r.URL.Query().Get("cursor")
	} else if !decode(w, r, &in) {
		return
	}
	switch in.Operation {
	case "delete":
		// Keep the record and delivery receipts for recovery; never send a task.
		in.Operation = "cancel"
	case "list", "get", "create", "update", "complete", "cancel", "reopen", "reply":
	default:
		fail(w, 400, "不支持的待办操作")
		return
	}
	if len(in.RequestID) > 80 {
		fail(w, 400, "无效回复请求 ID")
		return
	}
	unlock := s.lockTodos(userID)
	defer unlock()
	check := s.todoAuth(r, userID, sourceID)
	if !check() {
		fail(w, 401, "身份已失效")
		return
	}
	host, err := s.Store.DefaultHost(userID)
	if err != nil {
		fail(w, 500, "无法读取默认主机")
		return
	}
	if host.AgentID == "" {
		fail(w, 409, "请先在客户端菜单选择默认主机，用于保存唯一待办清单")
		return
	}
	host, err = s.recoverTodoMigration(r.Context(), userID, host, check)
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	if in.Operation == "create" {
		if sourceID != "" {
			in.ClientID = sourceID
		}
		if in.ClientID != "" || in.ProjectID != "" || in.SessionID != "" {
			target, e := s.Store.Agent(in.ClientID)
			if e != nil || target.UserID != userID || target.Revoked {
				fail(w, 404, "待办来源客户端不存在")
				return
			}
			body, _ := json.Marshal(in)
			if _, e = s.accountRelay(r.Context(), target.ID, userID, protocol.Request{Type: "todos.validate", Body: body}, check); e != nil {
				fail(w, 409, e.Error())
				return
			}
		}
		// Prefix request identity with the authenticated caller, preventing one
		// client from claiming another client's idempotency key.
		if in.RequestID != "" {
			in.RequestID = store.Hash(sourceID + "\x00" + in.RequestID)
		}
	}
	call := protocol.TodoStorageCall{AccountID: userID, Generation: host.Generation, TodoCommand: in}
	raw, err := s.todoCall(r.Context(), host.AgentID, call, check)
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	if in.Operation == "reply" {
		var row protocol.Todo
		if json.Unmarshal(raw, &row) != nil || row.Reply == nil {
			fail(w, 502, "默认主机返回无效回复")
			return
		}
		if row.Reply.State != "delivered" {
			target, e := s.Store.Agent(row.ClientID)
			if e != nil || target.Revoked || target.UserID != userID {
				fail(w, 409, "回复已保存，来源客户端已移除；修复后可重试")
				return
			}
			body, _ := json.Marshal(row)
			if _, e = s.accountRelay(r.Context(), target.ID, userID, protocol.Request{Type: "todos.deliver", Body: body}, check); e != nil {
				fail(w, 409, "回复已保存但尚未确认送达，请重试同一回复："+e.Error())
				return
			}
			call.Operation = "ack"
			raw, err = s.todoCall(r.Context(), host.AgentID, call, check)
			if err != nil {
				fail(w, 409, "会话已接收回复，待办确认失败；重试不会重复执行："+err.Error())
				return
			}
		}
		if json.Unmarshal(raw, &row) != nil {
			fail(w, 502, "默认主机返回无效待办")
			return
		}
		if row.Status == "open" {
			call.Operation, call.Version = "complete", row.Version
			raw, err = s.todoCall(r.Context(), host.AgentID, call, check)
			if err != nil {
				fail(w, 409, "回复已送入会话，待办状态尚未保存；重试同一回复即可："+err.Error())
				return
			}
		}
	}
	reply(w, 200, json.RawMessage(raw))
}

func (s *Server) todoCall(ctx context.Context, target string, in protocol.TodoStorageCall, check func() bool) (json.RawMessage, error) {
	if in.Operation == "update" && in.Options == nil {
		lookup := in
		lookup.Operation = "get"
		raw, err := s.todoCall(ctx, target, lookup, check)
		if err != nil {
			return nil, err
		}
		var current protocol.Todo
		if err = json.Unmarshal(raw, &current); err != nil {
			return nil, err
		}
		in.Options = current.Options
	}
	if in.Operation == "create" || in.Operation == "update" {
		body, err := encodeTodoOptions(in.Body, in.Options)
		if err != nil {
			return nil, err
		}
		in.Body, in.Options = body, nil
	}
	body, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	raw, err := s.accountRelay(ctx, target, in.AccountID, protocol.Request{Type: "todos.call", Body: body}, check)
	if err != nil {
		return nil, err
	}
	if in.Operation == "list" {
		var page protocol.TodoPage
		if err = json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}
		for i := range page.Todos {
			decodeTodoOptions(&page.Todos[i])
		}
		return json.Marshal(page)
	}
	switch in.Operation {
	case "get", "create", "update", "complete", "cancel", "reopen", "reply", "ack":
		var row protocol.Todo
		if err = json.Unmarshal(raw, &row); err != nil {
			return nil, err
		}
		decodeTodoOptions(&row)
		return json.Marshal(row)
	}
	return raw, nil
}

func (s *Server) accountRelay(parent context.Context, target, userID string, req protocol.Request, check func() bool) (json.RawMessage, error) {
	return s.relayRequest(parent, target, userID, req, check, true)
}

func (s *Server) relayRequest(parent context.Context, target, userID string, req protocol.Request, check func() bool, accountTools bool) (json.RawMessage, error) {
	req.ID = store.Token()
	data, _ := json.Marshal(req)
	if len(data) > protocol.MaxRequestBytes {
		return nil, errors.New("请求超出传输限制")
	}
	ctx, cancel := context.WithTimeout(parent, 25*time.Second)
	defer cancel()
	call := &relayCall{agentID: target, userID: userID, messages: make(chan protocol.Signal, 16), cancel: cancel}
	peerID := store.Token()
	s.hub.mu.Lock()
	agent := s.hub.agents[target]
	count := 0
	for _, b := range s.hub.browsers {
		if b.agentID == target {
			count++
		}
	}
	for _, c := range s.hub.calls {
		if c.agentID == target {
			count++
		}
	}
	if agent == nil || count >= 8 {
		s.hub.mu.Unlock()
		return nil, errors.New("客户端离线或连接已满，请稍后重试")
	}
	s.hub.calls[peerID] = call
	s.hub.mu.Unlock()
	defer func() {
		s.hub.mu.Lock()
		delete(s.hub.calls, peerID)
		s.hub.mu.Unlock()
		_ = agent.send(protocol.Signal{Type: "close", PeerID: peerID})
	}()
	if e := agent.send(protocol.Signal{Type: "open", PeerID: peerID, LeaseSeconds: 30, AccountTools: accountTools}); e != nil {
		return nil, errors.New("客户端连接已断开")
	}
	sent := false
	var chunks []byte
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("客户端响应超时或连接中断")
		case msg := <-call.messages:
			switch msg.Type {
			case "close":
				return nil, errors.New("客户端关闭了请求")
			case "ready":
				if sent || msg.Transport != protocol.RelayTransport {
					continue
				}
				sent = true
				if !check() {
					return nil, store.ErrDenied
				}
				if e := agent.send(protocol.Signal{Type: "request", PeerID: peerID, Request: &req}); e != nil {
					return nil, e
				}
			case "event":
				if msg.Event == nil || msg.Event.ID != req.ID {
					continue
				}
				if !check() {
					return nil, store.ErrDenied
				}
				switch msg.Event.Type {
				case "error":
					if msg.Event.Error == "unknown request type" && req.Type == "sessions.create-external" {
						return nil, errors.New("目标客户端不支持外部创建会话，请先升级客户端后重试")
					}
					if msg.Event.Error == "unknown request type" && strings.HasPrefix(req.Type, "todos.") {
						return nil, fmt.Errorf("目标客户端版本不支持账号待办（%s），请先升级该客户端后重试；若检查更新显示已是最新版本，需要先发布包含待办协议的新客户端安装包", req.Type)
					}
					return nil, fmt.Errorf("客户端操作失败：%s", msg.Event.Error)
				case "chunk":
					var encoded string
					raw, _ := json.Marshal(msg.Event.Data)
					if json.Unmarshal(raw, &encoded) != nil {
						return nil, errors.New("无效结果分片")
					}
					part, e := base64.StdEncoding.DecodeString(encoded)
					if e != nil || len(chunks)+len(part) > 2<<20 {
						return nil, errors.New("结果分片超出限制")
					}
					chunks = append(chunks, part...)
				case "result":
					if len(chunks) > 0 {
						if !json.Valid(chunks) {
							return nil, errors.New("无效结果")
						}
						return json.RawMessage(chunks), nil
					}
					raw, e := json.Marshal(msg.Event.Data)
					return raw, e
				}
			}
		}
	}
}
