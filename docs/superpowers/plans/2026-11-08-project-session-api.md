# 项目 Session API 实施计划

目标：其他系统持项目独立 token，经 139 在指定项目创建 Web session。

设计：复用持久 WebSocket、客户端历史和后台任务，新增受限 `sessions.create-external` 请求；139 保存 token 的 SHA-256 验证值及使用既有 token-key 加密的可复制密文。Token 仅授权创建所属项目的会话，可传标题、消息和 execute；false 将消息原子持久保存为 paused，true 后台执行，不授予其他设备、模型或账号权限。中转使用 base64 消息以保证 HTML/控制字符也支持完整 32000 字节。原项目、会话和凭证不迁移。

- [x] 存储：新增项目 token 元数据、生成/自定义替换、属主读取/复制/撤销；在事务中验证有效账号、设备、普通项目及 token。测试轮换、撤销、跨账号/项目、删除项目和重启；密钥恢复沿用已有加密存储规则及测试。
- [x] HTTP：属主使用 Cookie 维护 `/api/agents/{id}/projects/{project}/session-token`；外部使用 Bearer 调用同级 `/sessions`，请求接受 `title`、`message`、`execute`，响应包含 session ID、对象与工作台链接；ready 和结果时复验授权，限制传输与连接数。测试子路径、非法输入、离线、真实中转及原生客户端持久化。
- [x] 网页：项目设置动态加载独立 token 模块，显示接口、生成/自定义保存、复制/查看、撤销及无明文 token 的调用示例；防止切换项目、关闭弹窗后的异步响应泄漏。
- [x] 验证：Go tests/vet、Linux 构建、JS 检查、桌面/手机 Playwright 设置交互，使用模拟凭证和临时数据。
- [x] 交付：仅提交本次文件到 main 并推送 origin/main；使用已提交源码部署 139，停止服务备份数据库及 token-key、保留回滚，验证线上健康及鉴权/创建功能。保留所有其他未提交改动。

## 2026-11-08 发布与验收

- 代码提交 `6ba5bbbe802fa3d0cbd3b906a72df9772bc3a1f4` 已推送 `origin/main`。从仅含 HEAD 和本次变更的源码快照执行 `go test ./...`、`go vet ./...` 全通过，JS 语法及桌面 1440/手机 390 Playwright 交互通过；六平台编译和更新包摘要验证通过。
- 独立审查发现 JSON 转义膨胀问题并修复。真实 HTTP → WebSocket → 客户端回归覆盖 31998 字节 HTML、32000 字节控制字符、保存后继续执行和握手前撤销 token；复核通过。
- 139 服务端发布目录 `/opt/remote-agent/releases/6ba5bbbe802fa3d0cbd3b906a72df9772bc3a1f4-20261108132150`，停服数据/密钥/配置备份位于 `/var/backups/remote-agent/6ba5bbbe802fa3d0cbd3b906a72df9772bc3a1f4-20261108132150`。原版本、数据库及密钥保留，部署失败自动成套回滚。
- 六平台客户端更新版本 `20261108-project-session-api`、序号 `2026110803`，与服务端来自同一代码提交。公开更新清单、六平台下载及 Windows 宿主摘要逐一验证通过。更新渠道回滚脚本 `/var/backups/remote-agent-clients/runtime-6ba5bbbe802fa3d0cbd3b906a72df9772bc3a1f4-1794115474/rollback.sh`。
- 公网 HTTPS 线上验收覆盖：生成 token、错误/撤销 token 拒绝、空会话、保存消息不执行、立即执行、会话链接；使用临时账号、临时目录与模拟执行器，经真实持久 WebSocket，未调用真实模型或发送 IM。临时测试账号、设备、登录态及 token 已按精确身份清理；清理前另留数据库/密钥私有备份。未修改已有设备绑定。
- 发布和清理后 `/agents/healthz` 返回 200、`remote-agent`/Nginx active，新增 UI 资源 200，匿名创建接口 401。客户端需升级后使用新协议；未执行真实 Windows MSI/SCM、原生 ARM64 或真实模型任务验收。
- GitHub 直连暂时失败，通过临时 SSH SOCKS 中转推送成功，未更改系统或 Git 的持久代理配置。
