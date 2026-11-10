# 普通执行 session 事件回报发布结果

2026-11-10 北京时间发布后端与 runtime。能力与恢复语义见 [执行事件](../execution-events.md)。本轮继续原执行 session，没有另建执行 session、委派或发送额外 IM/邮件/广播。

| 项目 | 实际结果 |
| --- | --- |
| 源码 commit | `fb38c87a3caeca4d357c264a54ed20a1fca56a5c`，已推送 origin/main |
| runtime 版本 / 序号 | `20261110-session-events-fb38c87` / `2026111003`，高于原 `2026111002` |
| 来源 | 已提交 main 归档；22个本题文件，未含共享未提交改动 |
| 产物 | 六平台 runtime 与 Windows amd64/arm64 宿主；未发布新 MSI |
| 本 topic | CEO与管家按主题管理执行会话 |
| 执行 client / project | `T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg` / `s8YfNotejEtGbT5Y2miNymh83awpwU7KlG6jSXTeBxk` |
| 执行 session | `b5ohN8aoI8N3cwpF5cCUE54l6WtFUMxBv6CJBrah7rU` |
| 来源管理 session | `A9hYbszyzWisgbEzZxS7GJ3upYQUn5JD8fUc_nJo564` |

## 能力与关键事件证据

普通 topic 的执行接入现有 coordination work/report/outbox/inbox；终态自动生成、必要检查点用 MCP `report_execution_event`，均以真实来源和任务身份生成稳定事件ID。管理角色派工后结束轮次，事件到达才排队审阅，运行日志留在执行页。旧 bind 映射在正常升级后恢复订阅，已结束轮次可回填；本执行者未代写父会话 history.db。

临时目录/模拟凭据、真实环回 relay 验证通过：

| 验证 | 已核对的结果 |
| --- | --- |
| 关闭网页、管理模型不运行 | 派工后等待阶段模型启动0次；配置检查点到达才审阅 |
| 事件与通知去重 | 同一检查点提交3次，正常模型审阅1次、结果卡1张 |
| 受阻/待审边界 | 运行中不发送停止事件，轮结束后只发1条受阻事件，不再重复自动完成；该测试合计2张卡/2次事件审阅 |
| 非阻塞阶段 | 未配置阶段不发送；配置的 checkpoint 不暂停执行 |
| 忙碌与用户消息竞态 | 原用户轮→事件审阅→后续用户消息，保持顺序，无并发模型或覆盖 |
| 崩溃/断线 | 失败任务执行1次，离线投递失败后重建 Agent 重试，确认成功后不重发；暂停审阅可恢复且按事件ID去重 |
| 同 topic 续办 | 原 session 的新轮仍自动登记 outbox，不重复新建 session |
| 身份与旧轮次 | 过期 task、伪造执行者、兄弟管理会话订阅被拒绝；已有账号隔离测试通过 |
| 旧绑定 bootstrap | 仅会话元数据恢复真实订阅，连续恢复2次只建立1次订阅 |

测试为 `TestExecutionEvents*`、`TestExecutionInbox*`、`TestExecutionSubscription*` 与既有 `TestCoordination*` / topic / MCP角色测试。它们验证真实队列/relay实现及模型执行器调用次数，未使用真实模型或真实消息渠道。正常关键事件去重不承诺外部模型副作用严格 exactly-once；中断恢复最多追加2次审阅尝试，超过上限保留卡片待排查。

在只含本题修改的归档源码上 gofmt、相关测试、`go vet ./...`、六平台/两宿主构建和共享 skill `quick_validate.py` 通过。`go test ./...` 唯一失败仍为既有 `TestUserSkillServiceAccessIsLimitedAndInherited`：运行身份自身是 LocalService，临时父目录已允许该 SID，无法验证“默认未授权”。未修改真实 ACL 或跳过其他测试；`go test ./... -skip '^TestUserSkillServiceAccessIsLimitedAndInherited$'` 全部通过。日志在 `.local/event-validation/`。

## 发布、下载与实际激活

使用既有 `tools/deploy-139.ps1`、`tools/publish-runtime.ps1`，显式 SERVICE SSH 配置 `C:/Users/junte/.remote-agent/git-auth/ssh-139/config`、别名 `deploy-139`。沿用严格主机校验和正常认证，未读取/复制 SSH 凭据内容或改主机信任。

- 服务端 release：`/opt/remote-agent/releases/fb38c87a3caeca4d357c264a54ed20a1fca56a5c-20261110101720`；运行中 `/proc/PID/exe` 指向该目录的 remote-server。
- 客户端 dist：`/opt/remote-agent/client-releases/runtime-fb38c87a3caeca4d357c264a54ed20a1fca56a5c-1794277094`。
- 发布脚本通过六平台清单、provenance、大小、SHA256、单调序号、公开下载与 healthz；开发机另独立下载8个产物及公开摘要，全部与准备目录一致。
- HTTPS `/agents/healthz` 返回 `ok=true`；remote-agent/nginx active；匿名 POST `/agents/api/agent/coordination` 返回401，验证新增能力仍需认证。
- 本地准备目录：`.local/runtime-releases/fb38c87a3caeca4d357c264a54ed20a1fca56a5c-20261110101615-6ae5a380`。

| 客户端 | 本轮实际范围 |
| --- | --- |
| 139 Linux / `uf6Ii4zH0b4Y8afHjvkIV6fzkIL7mDPTM3s1ZtIBVRw` | 已正常空闲更新并激活；运行 PID 1481130，版本匹配；skill current=true |
| 本机 Windows / `T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg` | 在线忙碌，未触发更新/重启；RemoteAgent PID 10252、Updater PID 4872 与本轮发布前一致 |
| hz-windows | 在线，未主动升级或核验新版 |
| macbook-2026、surface laptop | 离线，待升级/验收 |

确认新版激活 **1/5台**，其余 **4台待升级或实际核验**，不能把服务端或渠道发布当成全客户端生效。139 更新前检查无执行子进程、无录音/分析/Action忙碌、无音频上传，先备份再运行现有 `remote-agent update`，未 force。新 Linux 程序 SHA256：`4a26e1ae8bd90838e0ee0f1d33b3f6141347d559da947b9a5c3a747e16b36a09`。

139 安装后的 CEO、客户端管家、项目管家各两次 MCP initialize/tools/list 验证一致：事件驱动指令、topic恢复、报告工具及分层派工保留，旧主动循环等待要求已移除。普通执行者没有强制管理派生规则；CEO/客户端专用 AGENTS 托管块和备份通过；共享 skill 摘要为 `6d75e1e963af0b55a31c9018c446903b9f7548de1cd3f5f0dc742f3e3b2a0857`。这次安装端检查没有 tools/call 或创建生产执行 session。

## 本次完成回报的 bootstrap 边界

父会话已按用户确认成功登记两个旧 topic，保留原 session。当前本机执行端尚未激活事件版，不能为本轮普通任务生成事件回报；**本次真实父回报尚无 eventId/投递/模型唤醒证据，未声称已通知父会话**。当前网页正常返回本结果。

安全过渡已准备：新渠道已就绪，待本机全部任务/录音空闲，既有 Windows 更新器以 Force=false 正常检查激活；启动一分钟后、以后每6小时检查，并非刚空闲即刻更新。新版恢复已登记 topic→client/project/session 映射，以真实任务队列回填终态并排队唤醒父会话，带 recovered 标记核对既有验收。恢复失败仍保留结果、路由和数据，可在父原入口恢复一次排查；不建立常驻模型或高频轮询，不伪造 workId/父上下文，不强制中断进程树。

## 备份与回滚命令

- 服务端配置/服务定义/Nginx/数据备份：`/var/backups/remote-agent/fb38c87a3caeca4d357c264a54ed20a1fca56a5c-20261110101720`。既有脚本提供发布失败自动恢复；另准备 `rollback-program.sh`，root:700、bash -n通过，恢复旧程序及其dist，保留当前数据库/配置。
- 渠道备份：`/var/backups/remote-agent-clients/runtime-fb38c87a3caeca4d357c264a54ed20a1fca56a5c-1794277094`，原 dist 保留。rollback.sh 为 root:700、bash -n通过。
- 139 程序/config/history/SaySo恢复点：`/home/junte/.remote-agent/release-backups/20261110-session-events-fb38c87/`，另保留 bin/remote-agent.previous。旧程序两份摘要均为 `671556e873915f0dd5bbb4700b8a1a49538d5bb6d36b6bfdd515ccb918cf35af`；config逐字节一致；history启动后字节变化，未读正文或逐会话对比，不能宣称逐字节未变。
- 139 `rollback-client.sh` 为 junte:700、bash -n通过，需保持空闲维护窗口；只恢复程序，不覆盖新历史、配置或录音。角色 AGENTS/skill 有各自备份，指令回滚需停止角色轮次后单独核对/恢复。

下列为已准备、未执行的恢复命令。联合回滚先渠道、再后端；渠道脚本要求 current 仍为本次 release，不能先切旧后端再调用它。已激活客户端回滚另行在空闲窗口处理，Windows 管理维护命令不由服务身份强行调用。

```powershell
$sshConfig='C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
ssh -F $sshConfig deploy-139 'sudo -n bash /var/backups/remote-agent-clients/runtime-fb38c87a3caeca4d357c264a54ed20a1fca56a5c-1794277094/rollback.sh'
ssh -F $sshConfig deploy-139 'sudo -n bash /var/backups/remote-agent/fb38c87a3caeca4d357c264a54ed20a1fca56a5c-20261110101720/rollback-program.sh'
ssh -F $sshConfig deploy-139 'bash /home/junte/.remote-agent/release-backups/20261110-session-events-fb38c87/rollback-client.sh'
```

可复用发布命令如下；PrepareOnly按调用时已提交HEAD准备，重复发布必须先核对线上序号，选择新的更大序号，不能原样重发1003：

```powershell
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig $sshConfig
./tools/publish-runtime.ps1 -Version VERSION -Sequence NEW_SEQUENCE -PrepareOnly
./tools/publish-runtime.ps1 -PreparedPath PREPARED -SSHHost deploy-139 -SSHConfig $sshConfig
```

本地证据：`.local/events-release/server-publish.log`、`runtime-publish.log`、`download-checks.json`、`activation-139.log`、`role-checks.json`、本机服务前后快照及 `RESULT.md`。剩余验收为本机实际空闲升级与真实父事件回填、生产完整模型审核/续办行为、其余4台升级、Windows真机SCM/MSI、macOS/ARM64运行、真实渠道及实际回滚生命周期。本轮未用普通测试触碰真实 SCM/MSI。
