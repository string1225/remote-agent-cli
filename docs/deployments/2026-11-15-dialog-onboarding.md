# 固定弹窗标题与 Token 安装指引发布

2026-11-15 已发布至139。功能说明见 [安装与弹窗改进](../onboarding-workbench.md)。

| 项目 | 实际版本 |
| --- | --- |
| 弹窗、Token 指引、MSI 激活源码 | `186da7c51f2a3f2e8a0c0d26b00fac46b7c63f64` |
| SaySo 弹窗补齐及服务器/网页源码 | `9df51cb3166f9e77e965b6ccc31b55281e542be8`，包含前者 |
| 六平台 runtime | `20261115-onboarding-186da7c` |
| 渠道序号 | `2026111501`，高于原 `2026111401` |
| x64/ARM64 MSI | `1.2.1115`，高于线上原 `1.2.1107` |

两个源码提交均已推送 origin/main。SaySo 补齐只涉及网页资源，没有改变已发布 runtime 的激活实现；服务端构建从已提交的9df51cb归档完成。共享未提交改动全部保留；platform.js、index.html只暂存本题差异，没有带入其他人的安装进度/轮询改动。

## 可用入口与实际覆盖

- 工作台刷新后加载固定标题/关闭按钮，正文单独滚动；包含动态待办、工作区、回复及SaySo通用/派工弹窗。
- 账号管理复制Endpoint指引时签发/复用该账号专用Token，收件人下载安装后填写Token配置本地Codex，无需网页登录；此Token不登记远程设备。
- Windows安装末页提供默认不勾选的Codex App选项，选中后打开商店供用户完成安装；CLI安装与配置保持独立。
- 新下载MSI已更新；本轮主动激活存量客户端为0，不声称所有已安装客户端升级。普通空闲升级流程保持可用，不强制更新或重启本机父子会话树。
- 本机RemoteAgent PID40396、Updater PID4872在发布前后保持Running且未改变。其他客户端实际升级数未查询，避免读取无关账号/会话。

公开安装包：

- [Windows x64](https://codex.sunny-string.cn/agents/downloads/remote-agent-windows-amd64.msi)
- [Windows ARM64](https://codex.sunny-string.cn/agents/downloads/remote-agent-windows-arm64.msi)

安装末页异常时，可在新PowerShell中运行 `remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-prompt`，按提示隐藏输入指引中的Token。

## 验证与边界

- Go已gofmt，相关激活/Endpoint测试及 `go vet ./...` 通过。隔离归档设置 `GOFLAGS=-buildvcs=false`，完整 `go test ./...` 仍仅有既有 `TestUserSkillServiceAccessIsLimitedAndInherited` 在LocalService临时目录继承权限下失败；明确排除此项后的全套通过。未通过修改真实ACL绕过测试。
- 专用Token测试覆盖管理员与账号隔离、同源拒绝、并发去重、关闭重开数据库恢复、撤销/重新签发、其他Token保持有效及列表不泄露原文。激活使用临时文件/模拟步骤，覆盖Token模式不触碰设备绑定/服务/代理、App选择独立、失败停止后续步骤及秘密脱敏。
- `node --check`覆盖app.js、platform.js、dialogs.js；新的弹窗/指引测试及既有todos、coordination、workbench、binding-code、project-preferences、directory-create通过。新版SaySo已运行 `npm run build` 并提交web/sayso产物；通用Modal与派工Modal在1440/390宽度、浅深主题长正文滚动/关闭验收通过。
- 原生C#fixture覆盖UTF-8 stdin、进度、最终结果、大stderr排空、忽略未知/null消息、缺结果及非零退出码诊断，不运行真实MSI或SCM。
- 六平台业务core与两个Windows宿主构建成功；x64/ARM64 MSI均通过只读表检查，App默认关闭、隐藏Token属性、激活不重建自身弹窗符合预期。LocalService下使用PackageOnly，ICE/Installer session完整校验尚未运行；构建保留既有LGHT1103无文件版本警告。
- 139 healthz为200/ok，remote-agent/nginx active；公网正常TLS下载7项网页资源均与提交源码SHA256一致。以下载回来的真实资源、模拟账号/会话重新执行弹窗及SaySo测试通过。
- 从公网下载并核对六平台清单、业务包大小/SHA256及摘要文件，两个Windows宿主/MSI和摘要文件均通过；匿名POST新指引路由为401。没有用生产账号签发真实Token或进行设备绑定。

用户报告的约“4283”缺少完整错误文字，未在全新Windows中精确复现。已消除完成页跳转自身的操作并增强退出码/Windows原生启动错误诊断；不能把代码和fixture验证当作该错误已在新机解决。真实安装、UAC、SCM、商店完成安装、原生ARM64、Safari/真实手机触控及实际回滚生命周期待目标环境验收。

## 发布、备份与回滚

正常使用已授权SERVICE SSH配置 `C:/Users/junte/.remote-agent/git-auth/ssh-139/config`、别名deploy-139，严格主机校验保持开启；未读取/复制凭据或真实环境文件内容。

```powershell
./tools/publish-runtime.ps1 -Version 20261115-onboarding-186da7c -Sequence 2026111501 -PackageVersion 1.2.1115 -PrepareOnly
# 复用准备目录发布（不重启客户端）
./tools/publish-runtime.ps1 -PreparedPath .local/runtime-releases/186da7c51f2a3f2e8a0c0d26b00fac46b7c63f64-20261115115650-9ad10eba
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
```

上方版本/序号已发布，不可重复作为新版本覆盖。后续发布需使用新序号。

- Runtime目录：`/opt/remote-agent/client-releases/runtime-186da7c51f2a3f2e8a0c0d26b00fac46b7c63f64-1794715103`。
- 渠道备份/回滚：`/var/backups/remote-agent-clients/runtime-186da7c51f2a3f2e8a0c0d26b00fac46b7c63f64-1794715103/rollback.sh`，原dist保留。
- 服务器Release：`/opt/remote-agent/releases/9df51cb3166f9e77e965b6ccc31b55281e542be8-20261115120321`。
- 服务器配置、Nginx、systemd及停止写入后的数据库备份：`/var/backups/remote-agent/9df51cb3166f9e77e965b6ccc31b55281e542be8-20261115120321`；发布脚本自带失败恢复。
- 此目录的`rollback-program.sh`可恢复原e07bc9d程序/网页，保留当前数据库与渠道，root:root/0700。程序和渠道回滚脚本均已bash -n检查，本轮没有执行。

若需同时回滚，先程序后渠道：渠道脚本要求恢复到它记录的旧服务器release。回滚渠道不会自动降级已升级客户端，客户端runtime回退仍须按原正常维护流程单独安排。

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent/9df51cb3166f9e77e965b6ccc31b55281e542be8-20261115120321/rollback-program.sh'
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent-clients/runtime-186da7c51f2a3f2e8a0c0d26b00fac46b7c63f64-1794715103/rollback.sh'
```

测试、发布、公网摘要、服务PID和截图证据在ignored `.local/onboarding-ui/`；runtime可用构建在上述prepared目录的`source/dist`。本轮没有发额外IM、创建执行session或读取其他会话正文。
