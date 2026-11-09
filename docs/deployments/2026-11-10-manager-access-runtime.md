# 管理角色与 Windows 授权合并 runtime 发布

2026-11-10（北京时间）发布客户端渠道，确认激活 139 Linux 客户端；本机 Windows 服务继续承载父会话与执行任务，未触发其更新、重启或强制暂停。

| 项目 | 实际值 |
| --- | --- |
| 版本 | `20261110-manager-access-dd93783` |
| 发布序号 | `2026111002` |
| 源码 commit | `dd93783afd421e4ba0c0dbdd85181ade504687f4`，准备前已确认等于 main 和 origin/main |
| 包含的管理规则 | `beccf8a3cb980f5ed026ab9d7ef682e572ef40d3` |
| 包含的 Windows 授权 | `dd93783`：`feat: add reversible Windows service path authorization` |
| 原六平台渠道 | `20261109-todo-decisions`，序号 `2026110902` |
| 构建来源 | 现有发布脚本 `git archive` 已提交 main；共享脏改未入包 |
| 产物 | Windows/macOS/Linux × amd64/arm64 runtime，Windows amd64/arm64 宿主；未发布新 MSI，保留已有 MSI |

本次仅发布客户端 dist，没有替换 139 服务端程序。只含 beccf8a 的旧准备包未用于发布。后续文档提交不改变上述源码或二进制版本。

## 可信连接与可复用发布命令

原先“SSH 缺可信主机密钥”结论来自未使用显式服务配置的预检；既有连接本来可用。下列正常认证返回 uid=1001(junte)、退出码 0，`sudo -n true` 成功，可信主机校验保持开启：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'id; sudo -n true'
```

本次执行的准备与发布等价命令如下。PrepareOnly 必须在预期的已提交源码 HEAD 上运行；再次发布必须重新核对六平台线上序号并选择更大值，不能原样重发当前序号。

```powershell
./tools/publish-runtime.ps1 -Version 20261110-manager-access-dd93783 -Sequence 2026111002 -PrepareOnly
$prepared = (Get-Content .local/manager-release/prepared-path.txt -Raw).Trim()
./tools/publish-runtime.ps1 -PreparedPath $prepared `
  -SSHHost deploy-139 `
  -SSHConfig 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
```

`prepared-path.txt` 是本次结果记录，非 PrepareOnly 自动生成；本次实际目录：

```text
.local/runtime-releases/dd93783afd421e4ba0c0dbdd85181ade504687f4-20261110001054-4eb10852
```

未读取/打印凭据、私钥或真实环境文件内容，未使用桌面 `.ssh`，未初始化新连接或修改真实绑定。

## 渠道、健康与下载证据

- 新 dist：`/opt/remote-agent/client-releases/runtime-dd93783afd421e4ba0c0dbdd85181ade504687f4-1794240794`。
- 保留旧 dist：`/opt/remote-agent/client-releases/runtime-b66045987cd1b7a91a83a4d6d0b8fe6c4bb483e8-1794240002`。
- 备份/渠道回滚：`/var/backups/remote-agent-clients/runtime-dd93783afd421e4ba0c0dbdd85181ade504687f4-1794240794/rollback.sh`；root:root、0700，`bash -n` 通过，未执行回滚。
- 发布脚本验证单调递增序号、六平台 provenance/清单/大小/SHA256，原子切换并准备失败恢复。
- 发布脚本从 HTTPS 站点检查六平台清单和下载；另从开发机独立公网下载六个 runtime、两个 Windows 宿主，校验大小、SHA256 和公开 `.sha256`，8/8 通过。
- `https://codex.sunny-string.cn/agents/healthz` 返回 `ok=true`；00:27:59 +08:00 再查仍通过。未以内部文件存在替代公开下载检查。

公开渠道路径为 `/agents/updates/v1/{windows,darwin,linux}/{amd64,arm64}.json`；下载路径为 `/agents/downloads/文件名`。

| 文件名 | SHA256 |
| --- | --- |
| remote-agent-core-windows-amd64.exe | `fa99926f0cbf525c9c614ab4d7ef09be46f9641ddefe72df5df6b9886520e66a` |
| remote-agent-windows-amd64.exe | `010dc93ad226cd32dbee92832d265857e2553b79bbd38e3f21f0137e54458ad1` |
| remote-agent-core-windows-arm64.exe | `26e8491c50add8a1e9d73dadcae106dd842efabea7e0001cc47d38478c522b47` |
| remote-agent-windows-arm64.exe | `8d4b0019ff3fb036cf234ce6bae9d30942179600817dfe0a2c357aba9f7c6982` |
| remote-agent-darwin-amd64 | `28b0030d7e682179645445b2f74b0fb627ba0fba9a97de870eea08ec78aa7cc7` |
| remote-agent-darwin-arm64 | `980ce13d2f4fc4f426d68fb68167ae2111df1f84449698ac109af271ce6e4ede` |
| remote-agent-linux-amd64 | `671556e873915f0dd5bbb4700b8a1a49538d5bb6d36b6bfdd515ccb918cf35af` |
| remote-agent-linux-arm64 | `02baf2c194d043ead8e3f3e7ddbdbf1508de8c1a549f76e0001c9bf3d86f038a` |

渠道回滚命令（准备完毕，本轮未执行）：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 `
  'sudo -n bash /var/backups/remote-agent-clients/runtime-dd93783afd421e4ba0c0dbdd85181ade504687f4-1794240794/rollback.sh'
```

该脚本要求服务端 current 仍为发布时 release，否则安全拒绝；渠道回滚不改变已激活客户端。Windows 更新器以序号防止渠道降级覆盖，必要的程序回滚要独立处理。

## 实际激活与安全空闲路径

账号目录元数据列出 5 台客户端，139 更新后已重新在线；没有读取其他会话正文。

| 客户端 | ID | 本轮实际覆盖 |
| --- | --- | --- |
| 阿里云 / 139 Linux | `uf6Ii4zH0b4Y8afHjvkIV6fzkIL7mDPTM3s1ZtIBVRw` | 已正常更新并激活，运行中版本与 skill 核验通过 |
| 央玺服务器 / 本机 Windows | `T25tGjazATq_lN_pTWinURLOOA2xlhIYJwN7g3J3vAg` | 在线忙碌；未主动升级，保留原进程树 |
| hz-windows | `QwlHtSalXiHY2Bu8ovOQgG-Roh_HQzlWkhVoN8BaueY` | 在线；未主动更新、未核验新版激活 |
| macbook-2026 | `SsRJ4d1BBCS5iue_ohGx_eOlgzZrJjL0go3Ac9eWTuQ` | 离线，待升级/验收 |
| surface laptop | `oPwytmiDRx2p29I8Z587NT15rBSDyL75x7WKkcSyG-s` | 离线，待升级/验收 |

确认覆盖 **1/5 台**；其余 **4 台待升级或实际版本核验**。不把在线状态或渠道发布当成全部生效。

139 激活前两次检查 systemd MainPID 无执行子进程，SaySo recording/analysisBusy/actionBusy 均为 0，无 `.upload-*`。先备份程序与数据，再调用既有 `/home/junte/.remote-agent/bin/remote-agent update`；普通 updater 重新检查服务和执行任务。未使用 force，也未直接改服务配置或设备绑定。

旧运行版本 `20261108-im-session-directory`，旧 PID 1385627；新运行 PID 1448143，`/proc/1448143/exe version` 返回 `20261110-manager-access-dd93783 (linux/amd64)`，systemd `active`。`skill status` 返回新版且 `exists=true,current=true`，不是仅检查磁盘候选包。

本机服务只读复核：RemoteAgent PID 32592（LocalService）、RemoteAgentUpdater PID 4872（LocalSystem），均 Running，与发布前相同。本轮没有手动检查并激活本机版本、排队强制升级或重启。现有 Windows 后台更新器启动一分钟后、以后每六小时检查，采用 `Force=false` 的空闲门禁；任务/录音忙碌时拒绝更新。后续全部相关任务退出后等待正常检查，不能承诺“刚空闲即刻升级”。本轮没有重配计划周期。

## 三管理角色与普通执行者验收

在 139 实际安装的新程序上，对已有 CEO(account)、客户端管家(client)、项目管家(project)、普通 home 项目分别执行两次 `mcp --config ... --project ...`，仅发送 `initialize`、`tools/list`。省略 session 参数，不编造会话 ID，不调用 tools/call，不新建执行 session。

- 三管理角色两次返回一致，包含“你是会话管理员”、topic 恢复/复用、忙碌等待与普通执行边界；不含“你可以直接完成任务”“小事可以直接处理”。
- `list_execution_topics`、`bind_execution_topic`、带 topic 参数的 `start_web_session` 可发现，`delegate_work` / `reply_work` 保留。
- 普通项目不含强制管理指令，保留“你可以直接完成任务”，没有递归强制派生。
- CEO/客户端管家专用 AGENTS 有托管块与 `.role-backup-*`；项目管家走 MCP 注入，不把规则写入共用原项目 AGENTS。
- 全局 `~/.agents/skills/remote-agent/SKILL.md` 已包含 topic 与 scope 规则；SHA256 `8823ca1a5d35170c8941b2550630892ad59bd3e31ef974fdb46a76c91dd991ac`。

以上验证实际安装程序的后续 MCP 初始化及持久规则同步路径，没有启动模型完成多轮派工；已运行模型保留原轮次上下文，不能宣称其即时切换。

## 客户端程序与数据恢复点

139 私有备份目录：`/home/junte/.remote-agent/release-backups/20261110-manager-access-dd93783/`，保留旧程序及 config.json、history.db、SaySo 副本。普通 updater 另保留 `bin/remote-agent.previous`；二者旧程序 SHA256 均为 `9d867ecc4a743a070c6a3926af1a8d015239b6aac5add75f0cb0deafe2d4ac4b`。

config.json 与备份逐字节一致。history.db 启动后字节有变化，本轮未读取会话正文或逐会话对比，不能声称历史逐字节不变；完整备份仍在，不以恢复旧数据库覆盖升级后的新记录。

程序回滚脚本已准备于该目录 `rollback-client.sh`，junte:junte、0700，`bash -n` 通过；本轮未执行。脚本校验旧程序摘要、目标版本与服务、检查执行子进程/录音/上传；**仍需管理员保持整个回滚窗口不再派工**，检查无法排除之后的新任务竞态。脚本仅恢复旧程序，失败恢复新程序；不回写配置、history 或录音。需要时先回滚渠道，再在空闲维护窗口运行：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 `
  'bash /home/junte/.remote-agent/release-backups/20261110-manager-access-dd93783/rollback-client.sh'
```

角色 AGENTS 回滚另需在停止角色轮次后恢复对应 `.role-backup-*`；程序回滚不等于撤销持久角色指令。共享 skill 保留其安装备份，旧程序启动会按旧内嵌版本同步；旧角色行为/回滚生命周期未在真实环境演练。本机 Windows 未发生版本切换，不需要本轮程序恢复；未来必要时由管理员使用既有 `remote-agent update --rollback` 空闲回滚，服务身份不能发起该维护操作，不使用 `--force`。

## 结果卡与剩余验收

- topic：CEO与管家按主题管理执行会话；执行 client/project/session 及父会话两个旧 topic 映射见 [主题记录](../manager-execution-sessions.md)。后续本主题继续同一 session。
- 发布证据（本地 ignored）：`.local/manager-release/publish.log`、`channel-before.json`、`download-checks.json`、`activation-139.log`、`role-checks.json`；结果卡 `.local/manager-release/RESULT.md`。
- 本轮未改源码，不重复已有 Go 测试/六平台长测；本轮检查发布、真实安装程序、公开下载和文档。原代码验证及已知 LocalService 测试边界见主题记录。
- 本机及其余客户端新版激活、三角色完整模型多轮行为、真实 Windows 授权/UAC/SCM/MSI、原生 macOS/ARM64 运行、真实消息渠道与实际回滚生命周期尚未验收；本轮无 IM/广播。
- 父会话 `A9hYbszyzWisgbEzZxS7GJ3upYQUn5JD8fUc_nJo564` 在本机正常升级后自行用 `bind_execution_topic` 登记“Windows 服务授权命令”与“CEO与管家按主题管理执行会话”。本执行者未代绑、未改父会话 history.db。
