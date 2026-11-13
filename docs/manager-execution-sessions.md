# CEO 与管家按主题管理执行会话

CEO、客户端管家、项目管家理解目标、拆题、调度和验收汇总。按2026-11-13用户调整，预计5步以内可以完成的工作默认由管理者直接完成；超过5步的工程、调研、测试和部署交给原普通项目的执行 session。以可核对的实际工作步骤估计，不按工具调用次数计步，不把复杂工作合为一步。执行中发现超过5步时，保留已完成动作和证据，交同 topic 执行 session接续；执行者忙碌时不并行接手。普通执行 session不继承5步界限，不递归要求派生。CEO/客户端管家保留 `delegate_work` 分层派工，下游管理者也遵守同一界限。小任务仍需遵守原项目授权、验证、提交与发布规则。

五步例外的已发布版本、真实激活范围与恢复命令见 [2026-11-13发布记录](deployments/2026-11-13-manager-five-step.md)。

## 持久来源与生效范围

- `internal/agent/roles/session-manager.md` 为三个管理角色共用的内嵌指令。MCP 每次启动按配置的 Service.Scope 注入，覆盖已有管理会话的下轮和新角色；CLI `--scope` 不能改变角色。管理提示用明确的5步界限替换无条件执行/无条件派工措辞，普通执行角色保留执行自主性。
- CEO/客户端管家的独立 workspace AGENTS.md 在 resolve 或下轮 MCP 启动时同步托管指令块。迁移只移除原版本精确生成的旧指令，保留自定义内容，先写 `.role-backup-*`，再原子替换；损坏标记/非普通文件/写入失败保留原文件并报错。回滚前停止角色轮次，再恢复对应备份；新客户端下次运行会重新同步其内嵌版本。
- 项目管家与执行 session 共用原目录，因此不改项目 AGENTS.md，通过每轮 MCP 注入区分角色。客户端安装的 `remote-agent` skill 同样按 scope 区分管理与执行规则，激活新版后沿用现有技能备份/同步机制。
- 初版主题调度在客户端实现。后续事件回报使用 139 的既有 work/report 路由，需新版服务端与参与客户端；详见 [执行事件](execution-events.md)。源码提交不会更新已安装客户端，激活新版 core 后才覆盖角色后续轮次，已运行模型保留原轮次上下文。

## 主题映射与恢复

管理角色用 `list_execution_topics` 按 cursor 分页恢复元数据；映射存于来源客户端 `history.db/execution_topics`，按来源 client/project/session 和规范化 topic（空白归一、英文大小写折叠）隔离，包含目标 client/project/session 及上次接收 taskId。不同管理会话不会接管彼此的子会话。默认派工后结束管理轮次，由持久关键事件唤醒验收；`read_session` / `wait_session` 仅用于排查、用户问状态或恢复时按需核对。

`start_web_session` 在管理角色要求 topic 和原普通项目 ID。首次先事务保存目标、原请求与稳定 spawn ID，再经 139 派工；同 topic 复用原 session，原生执行上下文由既有会话机制恢复。并发首次派工与断线重试复用同一 spawn ID；忙碌续办由现有 RequireIdle 拒绝，不覆盖或另起任务。目标冲突、离线、删除或失败保留映射并报告。`recovered=true` 表示恢复首次未确认的派工，返回的是原任务；先核对原任务，再明确发送本次后续要求。不得将恢复确认当成本次后续工作已执行。

升级前创建的直接子会话用 `list_child_sessions` 分页核对后，调用 `bind_execution_topic` 绑定：只读会话元数据，不执行、不读历史、不覆盖已有映射。备份/恢复时保留整个来源 history.db 和客户端原会话数据；新增表不迁移、删除或改写原会话及录音，旧程序忽略新增表。

调用示例（MCP tools/call arguments）：

```json
{"topic":"CEO与管家按主题管理执行会话","clientId":"CLIENT","projectId":"ORIGINAL_PROJECT","title":"角色与主题调度","prompt":"任务范围、约束与验收"}
```

该参数用于 `start_web_session`。同 topic 在原 session 续办并订阅完成/失败/必要检查点；确认后管理会话结束本轮，等待事件排队唤醒，验收后向当前用户入口汇报。普通进度留在执行页，不循环等待或转述无变化状态。结果卡保留事件ID、topic、client/project/session/task、commit、push、可用产物/命令、验证和未验收项；未经明确授权不发 IM 或广播。未订阅的普通执行者派生不会递归强制唤醒。

## 本次父会话已有映射

以下 ID 来自任务授权和本执行环境的 MCP 元数据；未读取其他会话正文，也未代写父会话 history.db。用户在事件驱动续办任务中已确认父会话 `A9hYbszyzWisgbEzZxS7GJ3upYQUn5JD8fUc_nJo564` 成功完成 topic 登记；两条都保留并复用，新版启动按已有映射恢复事件订阅。

| topic | clientId | 原 projectId | sessionId | 已知 taskId |
| --- | --- | --- | --- | --- |
| Windows 服务授权命令 | T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg | s8YfNotejEtGbT5Y2miNymh83awpwU7KlG6jSXTeBxk | ESNg4d1C-2hb9nlQu8TRZg0oWwRnlRk5jgv3MV1XJ2E | 8KAdYn-5XrZJq07QqjSdtvDnNJqeeQCYjEDVkmnCQWI |
| CEO与管家按主题管理执行会话 | T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg | s8YfNotejEtGbT5Y2miNymh83awpwU7KlG6jSXTeBxk | b5ohN8aoI8N3cwpF5cCUE54l6WtFUMxBv6CJBrah7rU | Pzg5slFZIIrb93Kx8Y97XN5OAZFge2n7mGPPV-a3IzM |

## 验收边界

临时目录与模拟执行器覆盖持久恢复、首次并发去重、丢失确认、同 topic 原生上下文续办、不同 topic 隔离、忙碌拒绝、目标冲突、删除不重建、旧子会话登记、父会话隔离、分页、角色下轮注入、旧角色备份迁移与普通执行不递归。真实客户端激活、模型遵循提示的完整多轮行为、Windows SCM/MSI 生命周期及真实消息渠道需另行验收；单元测试不能替代。

本轮在仅包含本任务改动的归档源码上运行 `gofmt`、相关测试、`go test ./...`、`go vet ./...` 和 Windows/Linux/macOS amd64/arm64 六平台构建。全量测试的唯一失败是既有 `TestUserSkillServiceAccessIsLimitedAndInherited`：执行身份本身为 LocalService，该测试用 LocalService SID 验证临时目录权限隔离，父目录已因当前用户身份拥有写权限；未改动此测试或真实 ACL。跳过这一项的 `go test ./... -skip '^TestUserSkillServiceAccessIsLimitedAndInherited$'` 通过。内置 skill 的 `quick_validate.py` 通过。

首次交付时尚未激活客户端，SSH 默认路径预检曾返回 `No ED25519 host key is known ... Host key verification failed`。此判断未使用项目已授权的服务专用连接，不能据此认定缺少凭据或可信主机密钥。后续复用 `docs/service-credentials.md` 记录的 `C:/Users/junte/.remote-agent/git-auth/ssh-139/config`、别名 `deploy-139`，正常严格主机校验、登录和 `sudo -n true` 均成功；未导入凭据、改 ACL 或读取桌面 `.ssh`。

2026-11-10 已将包含管理规则 `beccf8a` 和 Windows 授权 `dd93783` 的已提交 main 发布为 **`20261110-manager-access-dd93783`，序号 `2026111002`**。六平台 runtime、两种 Windows 宿主的清单、公开下载、大小/摘要和 HTTPS healthz 均通过。确认激活 **1/5 台**（139 Linux）；另 4 台待升级或实际版本核验。本机服务承载父会话及两个执行主题，本轮未触发更新或重启，保持等待正常空闲自动检查。

139 初版已安装程序的 CEO、客户端管家、项目管家各两次 MCP 初始化验证通过，普通执行者不注入管理强制规则；共享 skill 和专用 AGENTS 已同步且有备份。这不代表所有安装端或正在运行模型即时切换。初版渠道/程序恢复点与覆盖见 [合并发布记录](deployments/2026-11-10-manager-access-runtime.md)。两个旧 topic 后续已由父会话完成登记；本执行 session 未代绑，事件能力仍须单独发布激活验收。
