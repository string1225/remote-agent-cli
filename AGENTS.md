# 项目协作规则

## Git

- 所有开发直接在 `main` 上进行，提交并推送到 `origin/main`。
- 不创建功能分支、额外 worktree 或 PR；用户明确更改此要求时再调整。
- 修改前检查工作区、分支和远端状态，保留其他人的未提交改动。只提交本次工作涉及的文件，不强推。
- 提交前完成相应验证；提交后确认推送成功，报告提交号及尚未验证的事项。

## 产品目标

- 当前目标和验收基线见 [docs/project-goals.md](docs/project-goals.md)，已实现协议见 [docs/architecture.md](docs/architecture.md)。旧设计和部署记录供追溯使用。
- `code` 在本项目需求中指 **Qoder**。工具检测区分 Codex CLI、Codex App、Qoder CLI 与 Qoder 编辑器启动器。
- Windows 正式客户端采用 MSI 和 Windows Service，要求开机启动、退出登录后持续运行；不得新建计划任务，旧任务仅保留识别、停止、迁移和已安装版本更新兼容。
- 消息渠道由客户端直接连接飞书/钉钉的长连接服务，AK/SK 只存客户端。每个客户端项目最多一个渠道，每个机器人在本平台最多绑定一个客户端项目，并映射固定 session。
- 项目是网页工作台、消息渠道和 SaySo 的共同上下文；目标实现需要统一项目身份，保留现有会话和录音。
- 文档与界面应区分已实现能力和待实现目标。涉及配置、会话、录音及凭证的数据迁移必须保留原数据并提供失败恢复。

## 代码与验证

- 保持现有 Go 服务端/客户端、内嵌网页和 `sayso-ui` 的结构，按职责添加模块。
- 修改 Go 代码后运行 `gofmt`、相关测试和 `go vet ./...`；涉及跨模块或合并代码时运行 `go test ./...`。
- 修改工作台 JS 后运行 `node --check web/app.js` 和 `node --check web/platform.js`。
- 修改 SaySo 界面后在 `sayso-ui` 运行 `npm run build`，同时提交 `web/sayso` 产物；安装依赖使用锁文件。
- 修改发布或平台代码时检查相应平台构建。MSI 安装、Windows Service 生命周期、真实消息渠道等需要在目标环境单独验收，不能以单元测试替代。
- 测试使用临时目录与模拟凭证，避免修改开发者真实 Codex 配置、真实设备绑定或发送真实机器人消息。
- 普通测试不得创建真实计划任务、LaunchAgent 或机器服务。MSI/SCM 安装测试只在显式授权的一次性管理员环境中运行，见 `tools/msi/acceptance.ps1`。
