# 项目 Session API 实施计划

目标：其他系统持项目独立 token，经 139 在指定项目创建 Web session。

设计：复用持久 WebSocket、客户端历史和后台任务，新增受限 `sessions.create-external` 请求；139 保存 token 的 SHA-256 验证值及使用既有 token-key 加密的可复制密文。Token 仅授权创建所属项目的会话，可传标题、消息和 execute；false 将消息原子持久保存为 paused，true 后台执行，不授予其他设备、模型或账号权限。中转使用 base64 消息以保证 HTML/控制字符也支持完整 32000 字节。原项目、会话和凭证不迁移。

- [x] 存储：新增项目 token 元数据、生成/自定义替换、属主读取/复制/撤销；在事务中验证有效账号、设备、普通项目及 token。测试轮换、撤销、跨账号/项目、删除项目、重启与密钥恢复。
- [x] HTTP：属主使用 Cookie 维护 `/api/agents/{id}/projects/{project}/session-token`；外部使用 Bearer 调用同级 `/sessions`，请求只接受 `title`，响应包含 session ID、对象与工作台链接；ready 和结果时复验授权，限制传输与连接数。测试子路径、非法输入、离线、真实中转及原生客户端持久化。
- [x] 网页：项目设置动态加载独立 token 模块，显示接口、生成/自定义保存、复制/查看、撤销及无明文 token 的调用示例；防止切换项目、关闭弹窗后的异步响应泄漏。
- [x] 验证：Go tests/vet、Linux 构建、JS 检查、桌面/手机 Playwright 设置交互，使用模拟凭证和临时数据。
- [ ] 交付：仅提交本次文件到 main 并推送 origin/main；使用已提交源码部署 139，停止服务备份数据库及 token-key、保留回滚，验证线上健康及鉴权/创建功能。保留所有其他未提交改动。
