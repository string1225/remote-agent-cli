# Windows 后台服务的 Git 与 SSH

Windows Service 的自动启动与运行账号是两个独立配置。本项目继续使用 LocalService + RemoteAgent 服务 SID，开机启动、退出登录继续运行。把普通用户设为服务账号通常需要提供其账号密码，以便 Windows 在无人登录时执行服务登录；UAC 同意只授权修改系统配置，不提供这种可供以后使用的登录凭据。使用 Windows Hello PIN 的账号不能把 PIN 当作服务登录密码。普通用户账号的服务仍位于 Session 0，不自动接入桌面的 ssh-agent 或交互窗口。

以下方案不用在 Remote Agent 中保存 Windows 用户密码，保留原设备、配置、历史与录音。

## GitHub HTTPS

先升级到包含本次修复的业务 core，再在设备属主的普通终端执行：

```powershell
remote-agent git login --username string1225
remote-agent git status
```

`--username` 中间没有空格。登录通过已认证本机管道请求服务执行 GCM 设备码授权，在发起命令的终端显示授权地址和验证码，由用户自行在浏览器确认。Token 存在后台服务的独立 DPAPI 存储，桌面 Git 配置及登录保留。

2026-11-02 修复：显式 login 此前仍携带 `GIT_TERMINAL_PROMPT=0`，GCM 即使收到 `GCM_INTERACTIVE=always --device --no-ui` 也拒绝显示设备码。本次仅在显式 login 子进程设置 `GIT_TERMINAL_PROMPT=1`；普通任务、status 和 logout 继续禁止交互。回归测试模拟继承后台禁交互环境后再发起登录，防止重现。

`status` 仅列出已保存账号；需要真实仓库的认证请求才能验证 Token 及仓库权限。GitHub 浏览器授权由用户完成，测试不访问真实账号。

## SSH：一次本机授权，后台持续使用

新增 `remote-agent ssh configure`，在持有原私钥的桌面终端执行：

```powershell
remote-agent ssh configure --alias deploy-139 --host 139.224.12.141 --user junte --identity "$env:USERPROFILE/.ssh/id_ed25519_139_224_12_141" --known-hosts "$env:USERPROFILE/.ssh/known_hosts"
```

该命令可以由本次构建的独立业务 core 执行，不要求正在运行的服务先升级。命令校验指定文件和 SSH 目标，自动请求一次 UAC（系统策略要求管理员凭据时，按 Windows 提示完成）。只导入选定私钥和已核验的 known_hosts，原 `.ssh` 文件保持不变。不会扫描或导入其他私钥，不会自动信任新主机，也不会发送材料到 139。

当前支持普通文件形式的无口令私钥。带口令或硬件密钥暂不支持后台导入，命令明确拒绝；可改用单独的部署密钥，先把公钥配置到目标主机。命令不保存密钥口令或 Windows 账户密码。

目标目录固定为 `%ProgramData%/RemoteAgent/ssh/<alias>/`。配置通过独立代次发布，新副本及内容摘要校验成功后原子替换 `config`，保留旧代次和 `previous-config`；失败不会删除原私钥或旧连接。生产副本由 Administrators 持有，使用条件 ACL 仅向带 RemoteAgent 服务 SID 的进程授予读取权限，SYSTEM/管理员保留恢复权限；不会向所有 LocalService 进程授予读取私钥的权限。拒绝目标重解析路径，不把私钥写入命令行、日志或生成的配置文件正文。

后台连接使用显式配置：

```powershell
ssh -F 'C:/ProgramData/RemoteAgent/ssh/deploy-139/config' deploy-139 id
```

配置固定 `BatchMode yes`、`StrictHostKeyChecking yes`、`IdentitiesOnly yes`、`IdentityAgent none`，只依赖已导入的指定密钥和主机验证记录。普通桌面终端不能代表服务身份验证该副本，应在真实后台任务中验证连接。该命令不重启服务。

配置完成后，后台可运行部署并上传 APK：

```powershell
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig 'C:/ProgramData/RemoteAgent/ssh/deploy-139/config' -UploadFile ./dist/android/remote-agent.apk -ConfigPath 'C:/Users/junte/.remote-agent/config.json'
```

恢复旧连接时，管理员将对应 `previous-config` 原子恢复到别名的 `config`，旧代次原样保留。撤销访问需要移除配置副本，并在目标服务器撤销对应部署公钥；删除本机副本不会撤销同一私钥的其他副本。

自动验证使用临时生成的密钥、环回 SSH 模拟服务器和临时 ACL。覆盖系统原生 OpenSSH 的实际认证、主机核验、条件读取权限、输入校验、源文件保留和 PowerShell 语法；不调用真实 UAC、不修改开发机服务或用户 `.ssh`，不代替 139 真连接、真实 GitHub 登录和升级后的验收。
