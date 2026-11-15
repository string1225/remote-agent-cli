# storage.to临时文件与经确认的139敏感分享

2026-11-15完成实现、测试与客户端渠道发布。源码`d3671ea11f24ae50248df272dd8c7f9a149606d0`已推送origin/main。功能、命令和边界见 [分类文件分享](../file-sharing.md)。

| 项目 | 实际结果 |
| --- | --- |
| Runtime版本 | `20261115-file-sharing-d3671ea` |
| 发布序号 | `2026111502`，高于原`2026111501` |
| Windows x64/ARM64 MSI | `1.2.1116`，高于原`1.2.1115` |
| 六平台业务core与两个Windows宿主 | 均由已提交的d3671ea归档构建 |
| 主动升级/激活客户端 | 0；本机等待正常空闲升级，不强制更新活跃会话 |
| 本机服务状态 | RemoteAgent PID40396、Updater PID4872，发布前后均Running、PID不变 |
| 139后端程序 | 仍为`9df51cb3166f9e77e965b6ccc31b55281e542be8-20261115120321`，本轮未重新部署或重启后端 |

客户端Go模块、CLI/MCP及共享skill更新；服务端原文件存储路径兼容保留，服务端代码只调整相应测试调用，没有修改真实配置、绑定或迁移存量文件。共享工作区的既有脏改和未跟踪测试保留，本次只提交19个题内文件。

## 分类与角色生效

普通代码包、生成产物默认`temporary`，客户端直接初始化storage.to、上传Cloudflare R2并确认，返回`linkType=page`及实际到期时间。敏感文件必须先取得针对具体文件和分享范围的用户确认，才允许`class=sensitive,userConfirmed=true`经139；确认参数不替代询问，139仍是持链接可下载并在次日北京时间凌晨4点失效。失败不自动换通道，不伪造直链或改名规避拒收。

统一`SessionInstructions`追加分类规则，CEO、客户端管家、项目管家和普通执行者均接收；MCP的upload_file schema/说明及用户级共享remote-agent skill同步更新。delete_uploaded_file及CLI的files delete使用本机管理回执删除storage.to副本，不发消息、不删除原文件。规则随升级后每轮MCP初始化及正常skill同步生效；已安装但未升级、活跃中的旧客户端不声称已经覆盖，未查询其他客户端/账号的实际升级数。

## 真实API证据

官方[REST Markdown文档](https://storage.to/docs/api.md)与[llms.txt](https://storage.to/llms.txt)为200。普通HTML文档可能出现Cloudflare安全验证；没有绕过验证或使用桌面凭据。

- Windows开发机：64字节合成文本初始化、R2 PUT、确认均200；1,048,872字节合成ZIP上传完整流程10.2秒；54,526,240字节ZIP按真实API分片上传成功。
- 已提交d3671ea构建的Windows core再次上传同一52MiB ZIP，完整分片流程18.94秒，返回分享记录及实际到期时间；正常CLI提前删除成功。本地SHA256与上传流程输出一致；这不是从下载端重新校验文件的证据。
- 提前删除实网验证采用官方推荐的`Authorization: Owner`，DELETE发送明确JSON请求体；不会转发平台设备Bearer。1MiB和52MiB测试副本均已清理；为用户普通浏览器点验保留的64字节公开小样例按第三方实际期限到期，无真实用户数据。
- 139：在`/tmp/remote-agent-storage-to-d3671ea`运行独立Linux测试core，只生成64字节说明，使用不存在的临时config路径（临时通道不读取设备配置），上传到storage.to并正常删除；没有更新或重启139客户端/服务，也没有读取真实配置或文件内容。
- 当前自动化浏览器访问storage.to分享页被Cloudflare安全验证拦截（403）。已通过当前网页/会话请求用户正常浏览器点验，尚未收到验收答复。因此不声称真实用户下载成功、下载提速已验证或给出节省比例，也没有139同条件吞吐对照。

官方匿名默认3天、每天最多50文件；拒收exe/apk/ipa等类型；下载链接为页面。本产品仍限制1字节至100MiB。管理证明留在客户端配置路径旁shared-file-receipts，不进入返回结果；visitor身份并发/重启稳定复用，不靠换身份绕额度。Windows回执继承原数据目录ACL，未以POSIX文件模式冒充已验证的Windows访问隔离。

## 工程验证

- Go已gofmt，相关测试与`go vet ./...`通过，最后从发布归档再次vet通过。全套`go test ./...`仅既有`TestUserSkillServiceAccessIsLimitedAndInherited`在LocalService临时目录继承权限下失败；排除此项后的全套通过。没有修改真实ACL绕过用例。
- 临时目录和模拟凭据/传输覆盖CLI/MCP默认路由、敏感确认前拦截、无绑定普通上传、平台凭据不外泄、R2不携带visitor/owner、URL/重定向限制、分片顺序/ETag/补充URL/失败中止、管理回执删除、配额错误脱敏及身份并发恢复。真实网络测试独立运行，只上传无秘密合成文件。
- 六平台构建完成；两个MSI只读表检查通过。PackageOnly不执行ICE/完整Installer session或真实SCM/MSI，保留既有LGHT1103无文件版本警告；原生ARM64与真实安装生命周期待目标环境验收。
- 发布脚本核对六平台版本、递增序号、摘要及完整下载，并检查139healthz；本机公网TLS再核对六平台manifest/sha256和10个core/宿主/MSI的Range下载（206、1024字节）。这些是本产品发布包下载验收，与storage.to用户下载待验收分别记录。
- 程序/网络故障不改变选定通道；确认响应丢失可能产生本机未知的第三方副本，应报告不确定性。ZIP文件名检查不是内容扫描，嵌套/加密归档和其他格式仍需打包者核对范围。

## 发布、产物与回滚

复用已有SERVICE SSH配置，保持严格主机校验：

```powershell
./tools/publish-runtime.ps1 -Version 20261115-file-sharing-d3671ea -Sequence 2026111502 -PackageVersion 1.2.1116 -PrepareOnly
./tools/publish-runtime.ps1 -PreparedPath .local/runtime-releases/d3671ea11f24ae50248df272dd8c7f9a149606d0-20261115131833-90e85a1a
```

上方序号已经发布，不可重复覆盖；后续版本需使用新序号。

- 139渠道目录：`/opt/remote-agent/client-releases/runtime-d3671ea11f24ae50248df272dd8c7f9a149606d0-1794720059`。
- 原渠道保留，备份与回滚脚本：`/var/backups/remote-agent-clients/runtime-d3671ea11f24ae50248df272dd8c7f9a149606d0-1794720059/rollback.sh`，root:root/0700、bash -n通过，本轮未执行。
- 服务器remote-agent/nginx active，HTTPS healthz为ok；本轮不修改服务器程序，无需程序切换回滚。已升级客户端回退程序需按原正常维护流程单独安排，不能用恢复渠道强制降级活跃客户端；回退旧core也会恢复旧版文件分享行为，应停止使用新版分类参数并核对分享授权。
- [Windows x64 MSI](https://codex.sunny-string.cn/agents/downloads/remote-agent-windows-amd64.msi)、[Windows ARM64 MSI](https://codex.sunny-string.cn/agents/downloads/remote-agent-windows-arm64.msi)为1.2.1116。六平台manifest路径为`/agents/updates/v1/{os}/{arch}.json`。

手动恢复渠道（未执行）：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent-clients/runtime-d3671ea11f24ae50248df272dd8c7f9a149606d0-1794720059/rollback.sh'
```

构建产物在prepared目录source/dist，测试/发布/摘要/PID证据在ignored `.local/storage-to/`。无额外IM/邮件/广播，无新建执行session，无读取其他会话正文或真实凭据内容。
