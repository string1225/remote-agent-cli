# CEO 与管家按主题管理执行会话

CEO、客户端管家、项目管家理解目标、拆题、创建/复用执行 session、调度、读取结果、验收汇总。工程、调研、测试和部署由原普通项目的执行 session 完成，小任务也适用。普通执行 session 直接执行已授权任务，不递归要求派生。CEO/客户端管家保留 `delegate_work` 分层派工，保存 workId 并用 `reply_work` 续办下游管理会话；项目管家最终调度普通执行 session。

## 持久来源与生效范围

- `internal/agent/roles/session-manager.md` 为三个管理角色共用的内嵌指令。MCP 每次启动按配置的 Service.Scope 注入，覆盖已有管理会话的下轮和新角色；CLI `--scope` 不能改变角色。管理角色的初始化提示移除“可以直接完成任务”，普通执行角色保留执行自主性。
- CEO/客户端管家的独立 workspace AGENTS.md 在 resolve 或下轮 MCP 启动时同步托管指令块。迁移只移除原版本精确生成的旧指令，保留自定义内容，先写 `.role-backup-*`，再原子替换；损坏标记/非普通文件/写入失败保留原文件并报错。回滚前停止角色轮次，再恢复对应备份；新客户端下次运行会重新同步其内嵌版本。
- 项目管家与执行 session 共用原目录，因此不改项目 AGENTS.md，通过每轮 MCP 注入区分角色。客户端安装的 `remote-agent` skill 同样按 scope 区分管理与执行规则，激活新版后沿用现有技能备份/同步机制。
- 本次功能在客户端实现，沿用已有 139 `sessions.spawn/control` 中转，不修改服务端或网页。源码提交不会更新已安装客户端；须激活新版 core 后才覆盖该客户端管理角色的后续轮次。已运行的模型进程保留原轮次上下文，不能宣称已即时切换。

## 主题映射与恢复

管理角色用 `list_execution_topics` 按 cursor 分页恢复元数据；映射存于来源客户端 `history.db/execution_topics`，按来源 client/project/session 和规范化 topic（空白归一、英文大小写折叠）隔离，包含目标 client/project/session 及上次接收 taskId。不同管理会话不会接管彼此的子会话。实时状态和结果仍通过 `read_session` / `wait_session` 核对。

`start_web_session` 在管理角色要求 topic 和原普通项目 ID。首次先事务保存目标、原请求与稳定 spawn ID，再经 139 派工；同 topic 复用原 session，原生执行上下文由既有会话机制恢复。并发首次派工与断线重试复用同一 spawn ID；忙碌续办由现有 RequireIdle 拒绝，不覆盖或另起任务。目标冲突、离线、删除或失败保留映射并报告。`recovered=true` 表示恢复首次未确认的派工，返回的是原任务；先核对原任务，再明确发送本次后续要求。不得将恢复确认当成本次后续工作已执行。

升级前创建的直接子会话用 `list_child_sessions` 分页核对后，调用 `bind_execution_topic` 绑定：只读会话元数据，不执行、不读历史、不覆盖已有映射。备份/恢复时保留整个来源 history.db 和客户端原会话数据；新增表不迁移、删除或改写原会话及录音，旧程序忽略新增表。

调用示例（MCP tools/call arguments）：

```json
{"topic":"CEO与管家按主题管理执行会话","clientId":"CLIENT","projectId":"ORIGINAL_PROJECT","title":"角色与主题调度","prompt":"任务范围、约束与验收"}
```

该参数用于 `start_web_session`。同 topic 的后续调用会在原 session 续办。普通执行子会话不会自动唤醒父会话，管理角色须主动读取/等待并验收。结果卡保留 topic、client/project/session/task、commit、push、可用产物/命令、验证和未验收项；未经明确授权不发 IM 或广播。

## 本次父会话已有映射

以下 ID 来自任务授权和本执行环境的 MCP 元数据；未读取其他会话正文，未对真实 history.db 写入或迁移。父会话 `A9hYbszyzWisgbEzZxS7GJ3upYQUn5JD8fUc_nJo564` 激活新版后可用 `bind_execution_topic` 登记，两条都应继续保留并复用。

| topic | clientId | 原 projectId | sessionId | 已知 taskId |
| --- | --- | --- | --- | --- |
| Windows 服务授权命令 | T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg | s8YfNotejEtGbT5Y2miNymh83awpwU7KlG6jSXTeBxk | ESNg4d1C-2hb9nlQu8TRZg0oWwRnlRk5jgv3MV1XJ2E | 8KAdYn-5XrZJq07QqjSdtvDnNJqeeQCYjEDVkmnCQWI |
| CEO与管家按主题管理执行会话 | T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg | s8YfNotejEtGbT5Y2miNymh83awpwU7KlG6jSXTeBxk | b5ohN8aoI8N3cwpF5cCUE54l6WtFUMxBv6CJBrah7rU | Pzg5slFZIIrb93Kx8Y97XN5OAZFge2n7mGPPV-a3IzM |

## 验收边界

临时目录与模拟执行器覆盖持久恢复、首次并发去重、丢失确认、同 topic 原生上下文续办、不同 topic 隔离、忙碌拒绝、目标冲突、删除不重建、旧子会话登记、父会话隔离、分页、角色下轮注入、旧角色备份迁移与普通执行不递归。真实客户端激活、模型遵循提示的完整多轮行为、Windows SCM/MSI 生命周期及真实消息渠道需另行验收；单元测试不能替代。

本轮在仅包含本任务改动的归档源码上运行 `gofmt`、相关测试、`go test ./...`、`go vet ./...` 和 Windows/Linux/macOS amd64/arm64 六平台构建。全量测试的唯一失败是既有 `TestUserSkillServiceAccessIsLimitedAndInherited`：执行身份本身为 LocalService，该测试用 LocalService SID 验证临时目录权限隔离，父目录已因当前用户身份拥有写权限；未改动此测试或真实 ACL。跳过这一项的 `go test ./... -skip '^TestUserSkillServiceAccessIsLimitedAndInherited$'` 通过。内置 skill 的 `quick_validate.py` 通过。

实际已同步客户端数为 **0**，未激活任何真实角色。139 HTTPS `/agents/healthz` 返回 `ok=true`；公开 Windows 更新渠道仍为 `20261109-workbench`、序号 `2026110901`。服务身份的 SSH 预检因 `No ED25519 host key is known ... Host key verification failed` 失败；未关闭严格校验、导入信任记录或读取/复制桌面凭据。此为客户端包发布阻塞，不能以本地提交或构建宣称上线。

交付时用 `tools/publish-runtime.ps1 -Version VERSION -Sequence SEQUENCE -PrepareOnly` 从已提交源码准备六平台 runtime、Windows 宿主和校验清单（不安装、不修改服务）。父会话后续应通过已授权的服务端连接配置执行 `tools/publish-runtime.ps1 -PreparedPath PREPARED -SSHConfig SERVICE_SSH_CONFIG -SSHHost HOST`，发布前重新核对序号。既有发布脚本保留原 dist，并在 `/var/backups/remote-agent-clients/runtime-REVISION-TIMESTAMP/rollback.sh` 准备渠道回滚；客户端 runtime 回滚仍是独立步骤。发布后检查六平台渠道/下载摘要及 healthz，再在任务与录音空闲时按正常更新流程激活客户端，检查 `skill status` 和三个管理角色下轮的实际 MCP 指令，并由父会话登记上述两个旧 topic。角色 AGENTS 原文件备份及 history.db 应随客户端恢复点一并保留。
