# 全局 Remote Agent skill 发布与验收

代码提交：`09c1c51c659e30f515236c2745c78eb9d03e51cb`，已推送 `origin/main`。运行版本 `20261108-shared-agent-skill`，发布序号 `2026110801`；Windows x64/ARM64 MSI 发布为 `1.2.1106`。仅发布本次提交的源码，不包含工作区其他任务的未提交改动。

## 能力与安装

本机其他支持 CLI/MCP 的 Agent 可使用 `remote-agent mcp` 的 `notify_im_session` 和 `upload_file`；上传也可用 `remote-agent files upload --path FILE`。凭据由客户端读取。独立 Agent 的 IM 推送必须明确已有目标项目/会话，并获得发送授权；上传返回公开随机链接，最多 100 MiB，下次北京时间凌晨 4 点失效。安装 skill 不自动注册长期 MCP，也不授予发送或分享文件的权限。

技能统一安装到用户级 `~/.agents/skills/remote-agent/SKILL.md`，支持 `remote-agent skill install/update/status`。Unix 安装成功后、Windows MSI 完成后，以及已激活的客户端启动时同步。Windows MSI 修复通过宿主选择当前 runtime，避免旧 seed 覆盖新 skill。准备/探测与仅暂存新版本不提前更新 skill。

旧 `.codex` 或自定义 `CODEX_HOME` 中的同名目录整体移动至对应 Codex home 的 `remote-agent-skill-backups/`，保留附带资源和旧备份；迁移期间持有旧安装锁。现有共享目录的自定义 SKILL.md 更新前另存备份。版本/来源摘要在 `.remote-agent-install.json`，不包含凭据。

Windows 安装/服务配置为本技能目录授予服务 SID 更新权限。旧安装若无新目录权限，所属用户运行一次 `remote-agent skill install` 补齐即可，后续服务激活自动更新。其他技能、模型配置和服务账号不由该过程改变。

## 验证结果

| 范围 | 结果 |
| --- | --- |
| Windows 隔离源码测试 | `go test ./...`、`go vet ./...` 通过；最终诊断文字调整后相关测试和全量 vet 再次通过 |
| Linux 模块测试 | 原生 `go test -p 1 ./internal/codexsetup`、对应 vet 通过 |
| 跨平台 | Windows/macOS/Linux 的 amd64/arm64 六平台构建通过 |
| MSI | x64/ARM64 构建、x64 MSI 表静态检查通过；检查安装动作在 InstallFinalize 之后，调用 AgentExe 当前 runtime |
| 独立 MCP | 无来源会话时可发现通知/上传工具；未指定通知目标会被拒绝；测试输出不含模拟凭据 |
| 本机 Windows 实测 | 正常 core 更新后自动迁移，无手工 skill install；status 的版本匹配且 current=true；完整旧目录 4 个文件与升级前副本逐一校验一致 |
| 139 Linux 实测 | 正常更新并由原 systemd 用户服务重启后自动迁移；版本/来源内容匹配且 current=true，旧目录保留在备份中 |
| 两台 MCP 实测 | 已安装 CLI 的 initialize/tools/list 均返回 upload_file 与 notify_im_session；未发送真实 IM，也未上传文件 |
| 真实 Unix 安装模板 | 在 139 临时 HOME 执行公开 cli-linux.sh，自动安装全局技能，current=true；未创建设备绑定或启用服务 |
| 原有数据 | 本机 25 个、139 上 8 个原会话 ID/原生 ID/历史摘要保留；4 个/2 个客户端配置及 SaySo 文件摘要保留 |
| 代理与在线状态 | 本机系统/Git 代理仍开启并归属原客户端，端口 17891；Git 配置值保持一致；鉴权目录确认两台客户端在线，139 HTTPS healthz 正常 |

139 上的全量 Go 测试尝试因并行 fixture/SDK 编译资源开销被停止，没有得到全量通过结论；后续采用上述原生模块测试、真实安装和升级验收。临时测试进程已清理。实际 MSI 安装/修复/卸载、原生 ARM64/macOS、冷启动/退出登录及真实机器人发送未在本轮验收。

## 发布与恢复点

- 服务端：`/opt/remote-agent/releases/09c1c51c659e30f515236c2745c78eb9d03e51cb-20261108000707`。
- 服务端程序/数据库/密钥/配置备份：`/var/backups/remote-agent/09c1c51c659e30f515236c2745c78eb9d03e51cb-20261108000707`。发布脚本包含失败自动恢复。
- 更新渠道备份：`/var/backups/remote-agent-clients/runtime-09c1c51c659e30f515236c2745c78eb9d03e51cb-1794067562`。其中 rollback.sh 针对发布时的旧服务端 release；服务端已随后切换，配套回滚时先恢复旧服务端，再恢复渠道。客户端 runtime 回滚为独立步骤。
- Linux 原二进制：`/home/junte/.remote-agent/bin/remote-agent.previous`；Windows 保留原 runtime 回退记录，未重装 MSI 或更改服务账号。

官方技能位置与发现说明：[Agent Skills](https://developers.openai.com/codex/skills/)。
