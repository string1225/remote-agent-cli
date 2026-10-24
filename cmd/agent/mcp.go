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
)

func runMCP(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("mcp", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("config", agent.DefaultPath(), "local device configuration")
	project := flags.String("project", "", "current project")
	notifyClient := flags.String("notify-client", "", "origin notification client")
	notifyProject := flags.String("notify-project", "", "origin notification project")
	source := flags.String("source", "web", "origin session kind")
	session := flags.String("session", "", "current session")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("mcp requires --config, --project and --session")
	}
	c, err := agent.Load(*path)
	if err != nil {
		return err
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
			result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]bool{}}, "serverInfo": map[string]string{"name": "remote_agent", "version": "1.0"}, "instructions": "客户端包含项目，项目包含会话。先列出同账号客户端与项目，再用 start_web_session 在目标项目创建独立任务；立即返回排队结果。IM 会话将具体工作交给新 Web 会话并启用完成通知。notify_im_session 发送进度到指定项目已绑定的渠道。使用平台返回的 ID。"}
			if metadata, ok := result.(map[string]any); ok {
				metadata["instructions"] = fmt.Sprintf("当前客户端=%s，项目=%s，会话=%s；来源通知客户端=%s，项目=%s。%s", c.ID, *project, *session, *notifyClient, *notifyProject, metadata["instructions"])
			}
		case "ping":
			result = map[string]bool{}
		case "tools/list":
			result = map[string]any{"tools": clientTools()}
		case "tools/call":
			var call struct {
				Name      string `json:"name"`
				Arguments struct {
					AgentID   string `json:"clientId"`
					ProjectID string `json:"projectId"`
					Prompt    string `json:"prompt"`
					Title     string `json:"title"`
					Message   string `json:"message"`
					Notify    bool   `json:"notify"`
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
			case "list_clients_projects":
				data, err = agent.ClientAPI(ctx, c, "GET", "/api/agent/catalog", nil)
			case "start_web_session":
				data, err = agent.ClientTool(ctx, c, target, "tasks.start", protocol.TaskStart{ServiceID: serviceID, Title: call.Arguments.Title, Prompt: call.Arguments.Prompt, Source: *source, ParentSessionID: *session, Notify: call.Arguments.Notify || *source == "im", NotifyAgentID: notificationTarget(*notifyClient, c.ID), NotifyServiceID: notificationTarget(*notifyProject, *project)})
			case "notify_im_session":
				if call.Arguments.AgentID == "" && call.Arguments.ProjectID == "" && *notifyClient != "" && *notifyProject != "" {
					target = *notifyClient
					serviceID = *notifyProject
				}
				data, err = agent.ClientTool(ctx, c, target, "channels.notify", map[string]string{"serviceId": serviceID, "sessionId": *session, "message": call.Arguments.Message})
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
	str := func(description string) map[string]string {
		return map[string]string{"type": "string", "description": description}
	}
	return []any{
		map[string]any{"name": "list_clients_projects", "annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false}, "description": "列出当前账号的全部客户端与项目，包含客户端在线状态和项目 ID。跨设备操作先调用此工具确认目标。", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}},
		map[string]any{"name": "start_web_session", "annotations": map[string]bool{"readOnlyHint": false, "destructiveHint": false}, "description": "在同账号指定客户端/项目创建独立的新 Web 会话并排队执行任务。返回可跟踪的会话与任务 ID；遵守目标设备本机权限。", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "projectId": str("目标项目 ID；默认当前项目"), "title": str("会话标题，最多 80 字"), "prompt": str("任务内容，最多 32000 字节"), "notify": map[string]any{"type": "boolean", "description": "完成后通知发起会话项目已绑定的 IM 渠道"}}, "required": []string{"prompt"}, "additionalProperties": false}},
		map[string]any{"name": "notify_im_session", "annotations": map[string]bool{"readOnlyHint": false, "destructiveHint": false}, "description": "向同账号目标项目固定的 IM 渠道会话发送通知，并保存在渠道会话历史。需要项目已绑定飞书或钉钉。", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"clientId": str("目标客户端 ID；默认当前客户端"), "projectId": str("目标项目 ID；默认任务来源项目；没有来源时为当前项目"), "message": str("发送给用户的任务进度或完成通知")}, "required": []string{"message"}, "additionalProperties": false}},
	}
}

func notificationTarget(inherited, current string) string {
	if inherited != "" {
		return inherited
	}
	return current
}
