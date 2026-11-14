# 当前客户端会话只读排查

代码提交：`02ec5226eb2496b359e904fb1a1f79d4bdfd7a01`，已推送 `origin/main`。发布版本 `20261114-client-sessions-02ec522`，序号 `2026111401`。

`list_sessions` 通过现有 139 中转分页列出当前客户端全部项目的会话 ID、任务状态、时间与错误，包含 Web、IM、归档及已删除记录。可按项目、会话 ID 或标题筛选；列表不返回消息正文、原生恢复 ID、工作目录或排队消息。`read_session` / `wait_session` 可只读同客户端任意会话。跨客户端读取和所有发送、取消操作仍要求直接父子关系；服务端继续覆盖调用客户端身份并检查账号归属。

不迁移或删除原会话、录音、配置及凭据。共享工作区原有的 Codex 安装进度、网页轮询等未提交改动未纳入本次提交或发布。

## 验证

- gofmt、相关权限/分页/MCP/Qoder 测试、`go vet ./...` 通过；六平台客户端及两个 Windows 宿主构建通过。
- `go test ./...` 仅已有的 `TestUserSkillServiceAccessIsLimitedAndInherited` 在 LocalService 环境失败：临时目录父级本身已有该身份的写权限。排除这一项后全量测试通过；本次未改其实现或断言。
- 139 服务端已运行本提交，公网 `/agents/healthz` 返回 `ok:true`；六平台更新清单、下载大小和 SHA-256 均已校验。
- 139 Linux 客户端通过原更新器空闲升级，运行 PID 为 `1794838`，版本及用户 skill 均已确认。配置字节保持一致。
- 139 真实 MCP 经站点查询到 13 个会话，验证 ID 搜索、项目过滤、同客户端兄弟会话读取及敏感字段清除；只查询已有数据，没有创建模型任务或发送 IM。

## 备份、回滚与激活边界

服务端发布及停机一致性数据库备份：`/var/backups/remote-agent/02ec5226eb2496b359e904fb1a1f79d4bdfd7a01-20261114234733`。部署脚本已准备程序、数据库和 Nginx 的错误自动回滚，备份中的 `previous-release` 保留上一发布路径。

客户端下载渠道回滚：`/var/backups/remote-agent-clients/runtime-02ec5226eb2496b359e904fb1a1f79d4bdfd7a01-1794671376/rollback.sh`。139 客户端程序、配置、历史及 SaySo 备份：`/home/junte/.remote-agent/release-backups/20261114-client-sessions-02ec522`，原程序另保留为 `remote-agent.previous`。未实际执行回滚。

当前 Windows 执行客户端仍有本轮任务，未强制切换或重启。该设备及其他未核验设备需在任务结束后正常更新才能使用新查询能力；包已发布不代表全部设备已经激活。真实 Qoder、原生 ARM64/macOS、MSI/SCM 生命周期不由本次构建及 Linux MCP 验证替代。
