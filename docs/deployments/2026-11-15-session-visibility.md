# 七天旧会话隐藏发布

2026-11-15，源码`0166a606f0e7e563194f6afa6d9f3dca21a21ce9`已提交并推送origin/main；从该提交归档构建、部署网页/服务端并发布六平台runtime。共享工作区原有脏改全部保留，app.js仅暂存本题渲染改动，提交后原未提交变更仍为79行新增/69行删除。没有创建分支、worktree、PR或新执行会话，没有读取其他会话正文或发送额外消息。

## 行为和生效边界

详见[会话显示规则](../session-visibility.md)。末条持久消息/运行事件满7×24小时后，默认折叠到“7天前的会话”；可展开、搜索、按原session继续，新消息使其回到主列表。运行和排队任务保留显示，元数据带待处理项时也保留；手动归档仍独立。不存在归档/删除写入或历史迁移，也不为隐藏去读取其他会话正文。离线缓存有单一本地到期定时器，不新增轮询、模型或管理唤醒。

新版客户端返回独立`lastMessageAt`，旧历史直接读取最后一条记录时间，无迁移；重命名和取消归档不会把旧消息变新。旧客户端尚无该字段时兼容使用updatedAt，因此旧客户端仍存在“改名/恢复暂时按近期活动显示”的边界；严格按最后消息区分需正常空闲升级。网页刷新即加载新的折叠规则，不代表客户端均已升级。

- 版本：`20261115-session-recency-0166a60`；序号`2026111503`，六平台均由`2026111502`递增；Windows x64/ARM64宿主3、MSI `1.2.1117`。
- 当前账号目录只查询客户端/项目元数据，6台中3台在线、3台离线；目录不报告运行版本。本轮主动激活0台，确认已激活新runtime 0台；其余仍需正常空闲更新或核实，不能把未知状态算作已升级。
- 本机RemoteAgent PID43020、Updater PID4872发布前后Running且不变。进一步检查进程路径，服务host的core子进程PID54204仍在`a93e255e0f4d3fa9447f819e272b9bfc5d8b1b9b6c869ec96c1e797bef534294`目录，匹配此前`20261115-file-sharing-d3671ea`清单；未强制中断或升级当前父子进程树。其他5台版本未核实（其中3台离线）。

## 验证

- gofmt、相关history/session测试及`go vet ./...`通过。`go test ./...`仅失败于既有`TestUserSkillServiceAccessIsLimitedAndInherited`：LocalService临时目录的继承ACL使测试认为服务已具权限。未修改真实ACL；`go test ./... -skip '^TestUserSkillServiceAccessIsLimitedAndInherited$'`通过。所有测试使用临时数据库/模拟凭据，不安装真实SCM/MSI或触碰真实绑定。
- app/platform及新测试语法通过；session-recency、workspace-actions、workspace-navigation、compact-ui通过。导航fixture改用当前消息时间；actions fixture按稳定ID定位设置项并对未模拟API返回404，避免旧菜单位置断言和把API当静态文件读取。
- 新浏览器测试覆盖精确七天边界、重命名隔离、旧协议/空/无效时间、忙碌与队列例外、折叠键盘操作、搜索、分页、直接访问旧会话、新消息回到主列表、离线到期、账号/客户端范围切换、1440/390宽度。线上下载资源覆盖到模拟页面后，session-recency和五档compact-ui再次通过；不使用真实账号会话内容进行功能演示。
- 额外运行的独立旧`test-session-cache.cjs`在第11行失败：其断言仍要求内存历史小于250条，而现有实现保留完整内存历史、仅限制磁盘缓存。测试及session-cache.js均与HEAD一致，不加载本题app/history变更；本轮未修改此项，也未宣称它通过。导航回归中的完整内存历史/磁盘有界检查通过。
- 六平台Go构建、两个Windows宿主和两个MSI打包通过，清单及摘要验证通过。MSI采用PackageOnly，未执行ICE/安装会话/真实SCM生命周期验收；已有LGHT1103无版本文件警告保留，没有以打包结果代替安装验收。
- HTTPS healthz返回ok，remote-agent/nginx active。公网8项网页资源（index、app/platform、workspace.css、workspace-actions及coordination.js/css、theme）200，SHA256逐项匹配已提交部署源。发布脚本通过站点TLS检查六平台清单、六个完整core下载、Windows宿主及MSI完整下载摘要；本机公网另核对六份manifest/sha256及10个core/宿主/MSI的Range下载（206、1024字节）。
- 原生Safari、真实触控、生产账号操作及正常空闲激活后`sessions.list/get`消息时间端到端验收未完成；未执行实际回滚生命周期。单元测试和模拟网页不能替代这些验收。

## 发布与恢复

准备产物（含源码归档、六平台包、Windows宿主/MSI及清单）：

`C:/Users/junte/Code/remote-agent-cli/.local/runtime-releases/0166a606f0e7e563194f6afa6d9f3dca21a21ce9-20261115200303-defc5f43`

实际使用正常SERVICE SSH，未读取/复制凭据或更改主机信任：

```powershell
./tools/publish-runtime.ps1 -Version '20261115-session-recency-0166a60' -Sequence 2026111503 -PackageVersion '1.2.1117' -PrepareOnly
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
./tools/publish-runtime.ps1 -PreparedPath 'C:/Users/junte/Code/remote-agent-cli/.local/runtime-releases/0166a606f0e7e563194f6afa6d9f3dca21a21ce9-20261115200303-defc5f43'
```

Release：`/opt/remote-agent/releases/0166a606f0e7e563194f6afa6d9f3dca21a21ce9-20261115200441`。发布前备份：`/var/backups/remote-agent/0166a606f0e7e563194f6afa6d9f3dca21a21ce9-20261115200441`，包含停止写入后的数据、配置、服务及Nginx；部署自动失败恢复已准备。手动程序回滚恢复上一`5871f750eb7ceac43797e0e052be1d51e935f079-20261115193955`程序/网页，保留当前数据库及当前选择的runtime渠道，拒绝覆盖后续其他release。

runtime备份及渠道回滚：`/var/backups/remote-agent-clients/runtime-0166a606f0e7e563194f6afa6d9f3dca21a21ce9-1794744382/rollback.sh`，恢复此前渠道。两份手动脚本均root:root/0700，bash -n通过，均未执行。

```powershell
# 如需同时恢复渠道及网页，先恢复渠道，再执行程序回滚；单独恢复网页会保留所选渠道。
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent-clients/runtime-0166a606f0e7e563194f6afa6d9f3dca21a21ce9-1794744382/rollback.sh'
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent/0166a606f0e7e563194f6afa6d9f3dca21a21ce9-20261115200441/rollback-program.sh'
```

渠道恢复不强制降级已安装客户端；客户端程序回退按原正常维护流程单独安排。新字段向后兼容，历史数据继续保留。

证据在ignored的`.local/session-recency/`：Go完整/排除ACL日志、vet日志、runtime-prepare/runtime-publish/server-publish、live-assets.json、public-download-checks.json、live-test/live-compact-test、服务及core进程快照、模拟浅深主题截图等。
