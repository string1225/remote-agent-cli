package main

// Kept separate from the ordinary execution tools: only verified manager
// services use the durable topic registry. Executors keep their existing flow.
func executionTopicTools() []any {
	str := func(description string) any { return map[string]any{"type": "string", "description": description} }
	return []any{
		map[string]any{"name": "list_execution_topics", "description": "CEO/管家恢复当前管理会话的 topic→client/project/session/task 持久映射，只返回元数据。taskId 是上次接收值，实时状态用 read_session 核对。", "annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false}, "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"cursor": map[string]any{"type": "integer", "minimum": 0, "description": "返回的 nextCursor；每页最多 50 条"}}, "additionalProperties": false}},
		map[string]any{"name": "bind_execution_topic", "description": "CEO/管家将已有普通执行 session 绑定到 topic；仅核对当前管理会话的直接子会话元数据，不读正文、不执行、不覆盖已有映射。用于升级前的任务恢复。", "annotations": map[string]bool{"readOnlyHint": false, "destructiveHint": false}, "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"topic": str("稳定的 topic/方向，1–320 字节"), "clientId": str("执行客户端 ID；默认当前客户端"), "projectId": str("原普通项目 ID"), "sessionId": str("已有直接子执行会话 ID")}, "required": []string{"topic", "projectId", "sessionId"}, "additionalProperties": false}},
	}
}
