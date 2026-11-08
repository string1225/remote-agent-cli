# 其他系统创建项目会话

在工作台的项目菜单中打开 **设置 → 外部系统 API**，生成项目 Token，或保存自定义 Token（32–256 位 ASCII 字母、数字、`-`、`_`）。可随时查看、复制、重置和撤销。仅项目所属账号可维护，管理员身份不授予他人项目的 Token 读取权限。Token 独立于账号模型 Token 和客户端凭据。

使用设置里显示的完整接口地址（包含客户端 ID 和项目 ID）：

```sh
curl -X POST 'https://codex.sunny-string.cn/agents/api/agents/CLIENT_ID/projects/PROJECT_ID/sessions' \
  -H 'Authorization: Bearer YOUR_PROJECT_TOKEN' \
  -H 'Content-Type: application/json' \
  --data '{"title":"外部系统任务","message":"请处理这条消息","execute":true}'
```

| 参数 | 规则 |
| --- | --- |
| `title` | 可省略；最多 80 字。空会话默认为“新会话”；立即执行且未指定标题时从消息生成标题。 |
| `message` | 可省略；最多 32000 UTF-8 字节。 |
| `execute` | 默认 `false`。`true` 必须填写非空消息，创建后交给客户端后台执行；HTTP 只确认接收，不等待模型完成。 |

`{}` 创建空 Web 会话。`{"title":"待处理","message":"先保存","execute":false}` 将消息持久保存为已暂停的待处理消息，在工作台可“插队”继续或“编辑”取回；不会自动执行，客户端重启后仍保留。消息不发送给 IM。

成功返回 **201**：

```json
{
  "agentId": "CLIENT_ID",
  "projectId": "PROJECT_ID",
  "sessionId": "NEW_SESSION_ID",
  "taskId": "TASK_ID_OR_EMPTY",
  "executed": true,
  "session": {"id":"NEW_SESSION_ID","serviceId":"PROJECT_ID","title":"外部系统任务"},
  "url": "https://codex.sunny-string.cn/agents/#client=CLIENT_ID&project=PROJECT_ID&session=NEW_SESSION_ID"
}
```

`session` 另包含状态、执行器和时间等元数据。`taskId` 在空会话时为空，有消息时为执行任务或已暂停消息的 ID。工作台链接仍需账号登录。接口只允许创建指定普通项目的会话，不接受请求中的客户端、项目、执行器、目录或任意 RPC 类型；不能列举/读取已有会话或管理账号。请把 Token 放在 Authorization 请求头，不放 URL。

| HTTP 状态 | 含义 |
| --- | --- |
| 400 | 参数无效、标题或消息超限、立即执行但缺少消息。 |
| 401 | 缺少/错误/已撤销 Token，或账号、设备、项目已不可用；Cookie 不能代替项目 Token。 |
| 502 | 客户端返回无效结果。 |
| 503 | 客户端离线、连接已满、超时、正在维护或版本不支持；根据错误详情处理。 |

响应丢失或超时时客户端可能已接受创建；本接口不提供重试去重键，重试前应在工作台检查，避免重复会话或重复执行。

139 仅经现有持久 WebSocket 转发 `sessions.create-external`，会话正文、任务和历史保存在目标客户端，不进入服务器数据库。客户端必须升级到支持此协议的版本；旧版本明确报错，不会丢弃消息或忽略 execute。Token 服务端保存 SHA-256 验证值和 AES-GCM 密文，复用数据库旁的 `.token-key`；部署备份和恢复必须包含数据库和该密钥。常规项目/设备目录不返回 Token。

账号设置维护接口：同一路径下 `GET /session-token` 获取配置状态；`PUT /session-token` 传 `{}` 生成/重置或 `{"token":"CUSTOM_TOKEN"}` 替换；`POST /session-token/reveal` 读取原 Token；`DELETE /session-token` 撤销。维护接口使用项目属主登录 Cookie 及现有同源校验，不接受项目 Token 代替登录。
