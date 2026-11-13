# 管理角色五步以内直接执行

按用户2026-11-13调整：CEO、客户端管家、项目管家预计5步以内可以完成的工作默认自己做，超过5步才派给原普通项目执行 session。步骤按实际工作估计，不按工具调用次数计，也不把复杂工程合为一步；中途发现超过5步，保留已完成动作与证据，交同 topic 原执行 session 接续。已有执行者忙碌时不并行接手。

该界限不扩大授权，不替代原项目验证、提交和发布要求；普通执行 session 不继承五步派工限制。分层派工、topic复用及事件驱动回报保留，只在需要派工时恢复映射。直接修改项目也须定位原项目目录及其 AGENTS.md。

## 提交、版本与验证

- 源码 commit：`c339dd5ee8999dc57e8d5c00084158fcb91cdc7e`，已推送 origin/main；仅7个本题文件，共享脏改保留。
- 持久来源：内嵌 `roles/session-manager.md`、按配置 scope 注入的 MCP 提示、自动同步的 CEO/客户端专用 AGENTS 与共享 skill。项目管家共用原项目目录，不把管理界限写入普通执行者共用 AGENTS。
- 发布版本：`20261113-manager-five-step-c339dd5`，序号 `2026111301`，高于原渠道 `2026111003`。仅从已提交 main 准备六平台 runtime 和两种 Windows 宿主，沿用既有 MSI。
- 仅含本题改动的归档源码经 gofmt、相关角色/MCP测试、`go vet ./...`、六平台构建与 skill校验通过。测试覆盖旧“即使任务很小也不亲自执行”托管块迁移、保留自定义内容及备份、三个管理 scope 的新界限、普通执行者不受该界限约束。
- `go test ./...` 唯一失败为既有 `TestUserSkillServiceAccessIsLimitedAndInherited`：执行身份是 LocalService，临时父目录已允许该 SID，无法验证默认未授权场景。未改真实ACL；跳过这一项的 `go test ./... -skip '^TestUserSkillServiceAccessIsLimitedAndInherited$'` 全部通过。

## 发布与实际生效

复用既有 `tools/publish-runtime.ps1` 与已授权 SERVICE SSH 配置 `C:/Users/junte/.remote-agent/git-auth/ssh-139/config`、别名 deploy-139；主机信任保持开启，没有读取/复制SSH凭据内容。六平台清单、序号、大小/SHA256及8个产物的HTTPS下载检查通过；开发机另核对公网六平台清单与准备包一致，healthz返回ok=true。本轮仅发布客户端渠道。

- dist：`/opt/remote-agent/client-releases/runtime-c339dd5ee8999dc57e8d5c00084158fcb91cdc7e-1794583043`。
- 准备目录：`.local/runtime-releases/c339dd5ee8999dc57e8d5c00084158fcb91cdc7e-20261113231603-6bb05f9d`；`source/dist` 含产物与摘要。
- 公共清单：`https://codex.sunny-string.cn/agents/updates/v1/{windows,darwin,linux}/{amd64,arm64}.json`；下载位于 `/agents/downloads/`。
- 139 在无执行子进程、录音/分析/Action忙碌及上传的检查通过后，先备份再用现有 `remote-agent update` 正常更新，未 force。运行 PID 1697354 的 `/proc/PID/exe version` 确认为新版，systemd active，skill current=true。
- 新 Linux 程序 SHA256：`1ec7ba53799a47e610b2cba691fd979dd956c2ef6a92a216867a2c01643e452b`。
- 139 三个管理角色各两次实际 MCP 初始化均含“预计5步以内”“超过5步”，没有旧无条件派工措辞；专用 AGENTS 的五步规则和备份通过，普通执行者未增加五步界限。没有调用 tools/call 或创建生产执行 session。
- 共享 skill 与已提交准备包逐字节摘要一致：`95da407d3dc1bc368cd35dee2af0e253d4bf226b69005777fd9e88155230932f`。

确认激活 **1/5台**（139 Linux）。本机与 Surface 在线但未主动更新/核验本次版本，hz-windows 与 Mac 离线；其余 **4台待正常升级或核验**。本机 RemoteAgent PID21040、Updater PID4872 与本轮发布前一致，均Running，未触发本机升级或重启。已有角色后续轮次和新角色需使用激活后的新core；不能把渠道发布说成全部安装端或正在运行的模型已生效。

## 备份、回滚与可用命令

渠道恢复点：`/var/backups/remote-agent-clients/runtime-c339dd5ee8999dc57e8d5c00084158fcb91cdc7e-1794583043/rollback.sh`，root:700、bash -n通过，原dist保留。139程序/config/history/SaySo备份：`/home/junte/.remote-agent/release-backups/20261113-manager-five-step-c339dd5/`；另有 bin/remote-agent.previous。两份旧程序摘要均为 `4a26e1ae8bd90838e0ee0f1d33b3f6141347d559da947b9a5c3a747e16b36a09`。

139 `rollback-client.sh` 已准备，junte:700、bash -n通过，需保持空闲维护窗口。config与备份逐字节一致；history启动后字节变化，备份保留，未读取正文或逐会话对比。角色AGENTS及skill有各自备份，指令回滚需停止角色轮次后单独恢复/核对。以下回滚均未执行，渠道回滚不替代客户端程序回滚：

```powershell
$sshConfig='C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
ssh -F $sshConfig deploy-139 'sudo -n bash /var/backups/remote-agent-clients/runtime-c339dd5ee8999dc57e8d5c00084158fcb91cdc7e-1794583043/rollback.sh'
ssh -F $sshConfig deploy-139 'bash /home/junte/.remote-agent/release-backups/20261113-manager-five-step-c339dd5/rollback-client.sh'
remote-agent version
remote-agent skill status
```

复用本次准备包的发布命令已执行成功；同序号不能再次发布，后续须从预期已提交main用新版本/更大序号PrepareOnly准备：

```powershell
$prepared=(Get-Content .local/five-step/prepared-path.txt -Raw).Trim()
./tools/publish-runtime.ps1 -PreparedPath $prepared -SSHHost deploy-139 -SSHConfig $sshConfig
```

证据与结果卡位于 `.local/five-step/`：测试/vet/构建日志、publish.log、public-manifests.json、activation-139.log、role-checks.json、本机服务前后快照、RESULT.md。角色初始化验证的是实际安装指令与持久同步，步数估计和真实模型多轮执行选择尚未单独验收；其他4台激活、原生macOS/ARM64、真实SCM/MSI及实际回滚生命周期待验收。本轮未发额外IM/广播，也未另建执行session。
