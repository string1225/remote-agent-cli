package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func runMCP(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("mcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("config", agent.DefaultPath(), "local device configuration")
	project := flags.String("project", "", "current project")
	flags.String("notify-client", "", "legacy ignored option")
	flags.String("notify-project", "", "legacy ignored option")
	flags.String("notify-session", "", "legacy ignored option")
	source := flags.String("source", "web", "current session kind")
	turn := flags.String("turn", "", "current turn ID")
	parentClient := flags.String("parent-client", "", "parent client ID")
	parentSession := flags.String("parent-session", "", "parent session ID")
	session := flags.String("session", "", "current session")
	flags.String("scope", "", "assistant scope (role is verified from local configuration)")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("mcp requires --config, --project and --session")
	}
	if *turn == "" {
		*turn = store.Token()
	}
	c, err := agent.Load(*path)
	if err != nil {
		return err
	}
	role := ""
	for _, service := range c.Services {
		if service.ID == *project {
			role = agent.AssistantInstructions(service)
		}
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 128*1024)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			_ = encoder.Encode(map[string]any{"jsonrpc": "2.0", "id": nil, "error": map[string]any{"code": -32700, "message": "invalid JSON"}})
			continue
		}
		if len(request.ID) == 0 {
			continue
		}
		var result any
		var failure string
		switch request.Method {
		case "initialize":
			result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]bool{}}, "serverInfo": map[string]string{"name": "remote_agent", "version": "3.0"}, "instructions": fmt.Sprintf("当前客户端=%s，项目=%s，会话=%s，入口=%s；父客户端=%s，父会话=%s。%s\n%s", c.ID, *project, *session, *source, *parentClient, *parentSession, agent.SessionInstructions, role)}
		case "ping":
			result = map[string]bool{}
		case "tools/list":
			result = map[string]any{"tools": clientTools()}
		case "tools/call":
			var call struct {
				Name      string `json:"name"`
				Arguments struct {
					Path        string `json:"path"`
					AgentID     string `json:"clientId"`
					ProjectID   string `json:"projectId"`
					Prompt      string `json:"prompt"`
					Title       string `json:"title"`
					Message     string `json:"message"`
					TaskID      string `json:"taskId"`
					Cursor      uint64 `json:"cursor"`
					WaitSeconds int    `json:"waitSeconds"`
					SessionID   string `json:"sessionId"`
					TodoID      string `json:"todoId"`
					Body        string `json:"body"`
					Status      string `json:"status"`
					Operation   string `json:"operation"`
					Version     uint64 `json:"version"`
					PageCursor  string `json:"pageCursor"`
					Search      string `json:"search"`
					Scope       string `json:"scope"`
					WorkID      string `json:"workId"`
					ReplyHandle string `json:"replyHandle"`
				} `json:"arguments"`
			}
			if json.Unmarshal(request.Params, &call) != nil {
				failure = "invalid tool parameters"
				break
			}
			target, serviceID := call.Arguments.AgentID, call.Arguments.ProjectID
			if serviceID == "" {
				serviceID = *project
			}
			if target == "" {
				target = c.ID
			}
			var data json.RawMessage
			switch call.Name {
			case "list_assistants":
				data, err = agent.ClientAPI(ctx, c, "GET", "/api/agent/assistants", nil)
			case "delegate_work", "read_work", "reply_work", "cancel_work":
				if *project == "" || *session == "" {
					err = errors.New("分层协作需要真实的当前项目和会话 ID")
					break
				}
				operation := map[string]string{"delegate_work": "dispatch", "read_work": "read", "reply_work": "reply", "cancel_work": "cancel"}[call.Name]
				data, err = agent.ClientAPI(ctx, c, "POST", "/api/agent/coordination", protocol.CoordinationCommand{
					Operation: operation, RequestID: store.Hash(*session + "\x00" + *turn + "\x00" + string(request.ID)),
					Source: protocol.AssistantRef{ClientID: c.ID, ProjectID: *project, SessionID: *session},
					Target: protocol.AssistantRef{Scope: call.Arguments.Scope, ClientID: target, ProjectID: call.Arguments.ProjectID},
					WorkID: call.Arguments.WorkID, ReplyHandle: call.Arguments.ReplyHandle, Title: call.Arguments.Title, Prompt: call.Arguments.Prompt, TaskID: call.Arguments.TaskID,
				})
			case "upload_file":
				var file protocol.SharedFile
				file, err = agent.UploadFile(ctx, c, call.Arguments.Path)
				if err == nil {
					data, err = json.Marshal(file)
				}
			case "list_clients_projects":
				data, err = agent.ClientAPI(ctx, c, "GET", "/api/agent/catalog", nil)
			case "list_todos", "publish_todo", "update_todo":
				operation := map[string]string{"list_todos": "list", "publish_todo": "create", "update_todo": call.Arguments.Operation}[call.Name]
				if call.Name == "publish_todo" && (*session == "" || *project == "") {
					err = errors.New("发布待办需要真实的当前项目和会话 ID")
					break
				}
				if call.Name == "update_todo" && operation != "update" && operation != "complete" && operation != "cancel" && operation != "reopen" {
					err = errors.New("待办操作应为 update、complete、cancel 或 reopen")
					break
				}
				requestID := store.Hash(*session + "\x00" + *turn + "\x00" + string(request.ID))
				command := protocol.TodoCommand{Operation: operation, RequestID: requestID, ID: call.Arguments.TodoID, Title: call.Arguments.Title, Body: call.Arguments.Body, Status: call.Arguments.Status, Search: call.Arguments.Search, Cursor: call.Arguments.PageCursor, Version: call.Arguments.Version}
				if call.Name == "publish_todo" {
					command.ProjectID, command.SessionID = *project, *session
				}
				data, err = agent.ClientAPI(ctx, c, "POST", "/api/agent/todos", command)
			case "start_web_session":
				if *session == "" || *project == "" {
					err = errors.New("创建子会话需要真实的当前项目和会话 ID")
					break
				}
				requestID := store.Hash(*session + "\x00" + *turn + "\x00" + string(request.ID))
				data, err = agent.ClientTool(ctx, c, target, "sessions.spawn", protocol.TaskStart{RequestID: requestID, ServiceID: serviceID, Title: call.Arguments.Title, Prompt: call.Arguments.Prompt, Source: *source, ParentSessionID: *session})
			case "list_child_sessions", "read_session", "wait_session", "send_to_session", "cancel_session_task":
				operation := map[string]string{"list_child_sessions": "list", "read_session": "read", "wait_session": "wait", "send_to_session": "send", "cancel_session_task": "cancel"}[call.Name]
				requestID := store.Hash(*session + "\x00" + *turn + "\x00" + string(request.ID))
				data, err = agent.ClientTool(ctx, c, target, "sessions.control", protocol.SessionControl{Operation: operation, CallerSessionID: *session, SessionID: call.Arguments.SessionID, TaskID: call.Arguments.TaskID, RequestID: requestID, Prompt: call.Arguments.Prompt, Cursor: call.Arguments.Cursor, WaitSeconds: call.Arguments.WaitSeconds})
			case "notify_im_session":
				conversation := call.Arguments.SessionID
				if conversation == "" && *source == "im" && target == c.ID && serviceID == *project {
					conversation = *session
				}
				if conversation == "" {
					err = errors.New("主动发送 IM 消息需要明确的目标 IM 会话 ID")
					break
				}
				data, err = agent.ClientTool(ctx, c, target, "channels.notify", map[string]string{"serviceId": serviceID, "sessionId": conversation, "message": call.Arguments.Message})
			default:
				err = errors.New("unknown tool")
			}
			if err != nil {
				result = map[string]any{"isError": true, "content": []any{map[string]string{"type": "text", "text": err.Error()}}}
			} else {
				result = map[string]any{"content": []any{map[string]string{"type": "text", "text": string(data)}}}
			}
		default:
			failure = "method not found"
		}
		response := map[string]any{"jsonrpc": "2.0", "id": request.ID}
		if failure != "" {
			response["error"] = map[string]any{"code": -32601, "message": failure}
		} else {
			response["result"] = result
		}
		if err := encoder.Encode(response); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil && !strings.Contains(err.Error(), "closed") {
		return err
	}
	return nil
}
func clientTools() []any {
	str := func(description string) any { return map[string]any{"type": "string", "description": description} }
	number := func(description string) any {
		return map[string]any{"type": "integer", "minimum": 0, "description": description}
	}
	tool := func(name, description string, readOnly bool, properties map[string]any, required ...string) any {
		schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
		if len(required) > 0 {
			schema["required"] = required
		}
		return map[string]any{"name": name, "description": description, "annotations": map[string]bool{"readOnlyHint": readOnly, "destructiveHint": false}, "inputSchema": schema}
	}
	target := func() map[string]any {
		return map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "sessionId": str("直接子会话 ID；read_session 也可查询当前会话"), "taskId": str("预期任务 ID；防止操作已切换的任务")}
	}
	read := target()
	read["cursor"] = number("结果分页游标；使用返回的 nextCursor 继续读取")
	wait := target()
	wait["cursor"] = number("结果分页游标")
	wait["waitSeconds"] = map[string]any{"type": "integer", "minimum": 1, "maximum": 20, "description": "单次等待 1–20 秒，默认 20 秒；未结束可继续调用"}
	send := target()
	send["prompt"] = str("发送给子会话的新指令，最多 32000 字节；忙碌时需先等待")
	return []any{
		tool("list_assistants", "查询 CEO、客户端管家和项目管家目录。CEO 位于默认主机，各管家独立会话。", true, map[string]any{}),
		tool("delegate_work", "向指定管家委派任务，返回 workId、replyHandle 和执行会话。工程会持久化回报并唤醒当前会话；返回接收不表示完成。按需直接派项目管家。", false, map[string]any{"scope": str("account=CEO、client=客户端管家、project=项目管家"), "clientId": str("目标客户端 ID"), "projectId": str("项目管家的原项目 ID；其他层省略"), "title": str("任务标题"), "prompt": str("任务、范围与验收，最多 32000 字节")}, "scope", "clientId", "prompt"),
		tool("read_work", "读取当前会话派出的分层任务；保留 workId，不操作无关工作。", true, map[string]any{"workId": str("派工返回的 workId")}, "workId"),
		tool("reply_work", "在下游原执行会话继续一轮，复用上下文，结果仍可靠回报当前会话。", false, map[string]any{"workId": str("workId"), "prompt": str("后续指令")}, "workId", "prompt"),
		tool("cancel_work", "取消已核对的下游当前轮次，先读取任务 ID。", false, map[string]any{"workId": str("workId"), "taskId": str("当前 taskId")}, "workId", "taskId"),
		tool("upload_file", "上传当前机器上的文件到 139，返回无需登录即可下载的随机链接、SHA-256 和 expiresAt。最多 100 MiB，下次北京时间凌晨 4 点失效；可在回复或已获授权的消息中附带链接。上传会公开文件内容，先核对分享范围。", false, map[string]any{"path": str("当前机器上要分享的普通文件的绝对路径")}, "path"),
		tool("list_todos", "查询本账号唯一待办清单，存于默认主机；默认主机离线时返回明确错误。", true, map[string]any{"status": str("筛选 open、completed 或 cancelled；省略表示全部"), "search": str("搜索标题和正文"), "pageCursor": str("返回的 nextCursor，继续分页")}),
		tool("publish_todo", "工作需要人判断、选择或补充信息时发布待办。写清问题、可选方案和建议，然后结束本轮等待人工回复。待办自动关联当前客户端、项目和会话；人的回复将在原会话继续执行。", false, map[string]any{"title": str("问题标题，最多 480 字节"), "body": str("背景、问题、选项和建议，最多 8000 字节")}, "title", "body"),
		tool("update_todo", "维护本账号待办。先读取最新 version，避免覆盖其他人的修改；编辑时同时传 title 和 body。", false, map[string]any{"todoId": str("待办 ID"), "operation": str("update、complete、cancel 或 reopen"), "version": number("读取到的 version"), "title": str("编辑后的标题"), "body": str("编辑后的正文")}, "todoId", "operation", "version"),
		tool("list_clients_projects", "查询同账号客户端、在线状态和项目 ID；按需查询。", true, map[string]any{}),
		tool("start_web_session", "按需要在目标项目创建直接子 Web 会话并执行。返回仅代表已接收；可由当前会话用读取、等待、继续和取消工具管理，完成后不会自动发 IM 通知。", false, map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "projectId": str("目标项目 ID；默认当前项目"), "title": str("子会话标题，最多 80 字"), "prompt": str("任务内容，最多 32000 字节")}, "prompt"),
		tool("list_child_sessions", "列出当前会话在指定客户端直接创建的子会话；跨客户端分别查询。", true, map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "cursor": number("子会话列表分页游标")}),
		tool("read_session", "读取直接子会话或当前会话的状态、当前任务最终输出及分页游标。complete 与 taskError 表示任务状态；captured=false 表示仅有旧执行器文本，不能据此宣称完成。", true, read, "sessionId"),
		tool("wait_session", "等待直接子会话当前任务完成或等待期限到达，返回状态及结果；不会创建会话或向用户发消息。", true, wait, "sessionId"),
		tool("send_to_session", "在已有直接子会话继续一轮任务，复用原生上下文；当前任务忙碌时返回错误。", false, send, "sessionId", "prompt"),
		tool("cancel_session_task", "取消直接子会话的指定当前任务，不影响其他会话。", false, target(), "sessionId", "taskId"),
		tool("notify_im_session", "仅在决定主动发送额外 IM 消息时使用。当前会话的正常回答由入口投递，无需调用此工具；Web 会话必须指定目标 IM 会话。", false, map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "projectId": str("目标项目 ID；默认当前项目"), "sessionId": str("目标 IM 会话 ID；仅当前 IM 会话可省略"), "message": str("主动发送的消息，最多 16000 字节")}, "message"),
	}
}
