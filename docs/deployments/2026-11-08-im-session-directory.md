# IM 收件人发现发布验收

代码：`d4e7fa8abd4bb27ff4f17751f2e3ca88073eb253`，已推送 `origin/main`。运行版本 `20261108-im-session-directory`，发布序号 `2026110802`；六平台客户端及 x64/ARM64 MSI `1.2.1107` 已发布。构建从提交归档取源码，未包含工作区其他任务的未提交改动。

## 对外行为

MCP 新增只读 `list_im_sessions`，通过同账号的 139 中转请求 `channels.sessions`。默认查询当前客户端全部项目，可按 clientId/projectId/search/cursor 筛选。返回外层 clientId，以及每条会话的 projectId/projectName、provider、channelName、chatType、peerName、nameKnown、title、sessionId。仅限当前机器人绑定中已收到消息的活动 IM 会话，不返回聊天正文、平台 Chat ID、凭据、工作目录或原生会话 ID，也不是 IM 通讯录。

姓名/群名已知且在用户指定或上下文确定的范围内唯一时，Agent 使用对应真实 ID；需读完相关分页，不能把第一页的单个命中当成全部结果。未说对象、重名、范围不明确或名称未知时，列出简短候选询问。机器人名与收件人姓名分别表示，未知名称不推测。正常 IM 回答继续从当前入口投递，不额外推送。Qoder MCP 工具白名单及全局 skill 已同步。

查询使用 bbolt ReadOnly + View，缺失数据库返回空列表，不初始化历史。分页同时限制 50 条与 24KB 条目预算，避免大名单触发中转分块丢失数据。

## 验证

- 新功能测试先复现 unknown request/tool；回归测试复现长中文名称页面 61,794 字节、查询改写数据库和创建缺失数据库，再验证修复通过。
- 干净源码 `go test ./...`、`go vet ./...` 通过；六平台构建、两套 MSI 构建和 MSI 表静态检查通过。
- Linux 原生执行 IM 目录测试二进制，当前绑定/名称映射/不返回正文、数据库不创建/不改写、条数及字节分页测试全部通过。
- 独立审查完成；MCP 无来源会话查询、Qoder 工具发现、跨账号拒绝及服务端中转测试通过。
- 本机与 139 正常升级后，客户端和全局 skill 的版本均匹配，skill status 的 current=true。
- 本机生产目录返回 1 条名称未知的活动 IM 私聊；139 返回 0 条。经本机 MCP 跨客户端查询 139 与 139 本机查询结果一致，不把未知名称伪装成真实姓名。未将实际收件人标识或名称写入本记录。
- 本机原有 26 个、139 原有 8 个会话 ID/原生 ID/历史摘要保留，4 个/2 个客户端配置及 SaySo 文件摘要保留。
- 本轮开始时本机系统/Git Web 代理已关闭，升级后保持关闭；Git 全局 http.proxy 仍未设置。两台客户端在线、Windows 服务运行、139 HTTPS healthz 正常。

未发送真实 IM、未上传文件、未做真实 MSI 安装/修复、原生 ARM64/macOS 或整机冷启动/退出登录验收。发送目标确认规则由 skill/MCP 工具说明提供；发送接口继续使用明确 ID，不按模糊名称静默选取收件人。

## 恢复点

- 服务端：`/opt/remote-agent/releases/d4e7fa8abd4bb27ff4f17751f2e3ca88073eb253-20261108130156`。
- 服务端备份：`/var/backups/remote-agent/d4e7fa8abd4bb27ff4f17751f2e3ca88073eb253-20261108130156`，保留原程序、数据库、密钥和配置，发布失败自动回滚。
- 当前更新渠道回滚脚本：`/var/backups/remote-agent-clients/runtime-d4e7fa8abd4bb27ff4f17751f2e3ca88073eb253-1794114204/rollback.sh`。客户端 runtime 回滚为独立操作。
- Windows 保留上一 runtime 记录；Linux 保留 `remote-agent.previous`。本轮不重装 MSI、不更换服务身份。
