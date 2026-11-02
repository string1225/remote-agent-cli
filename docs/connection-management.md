# 服务器连接与 GitHub 授权

网页左下角“服务器与 GitHub”或客户端菜单中的同名入口，先选择执行客户端。服务器记录、私钥和 GitHub Token 属于该客户端，供它的后台任务使用；它不是在手机上添加新的执行客户端。旧客户端须先更新业务 core，网页会按 `credentialManagement` 能力提示升级。

## SSH 服务器

网页输入连接别名、IP、SSH 用户、端口和密码。命令行默认隐藏输入密码：

```powershell
remote-agent servers add --name deploy-139 --host 139.224.12.141 --user junte
remote-agent servers list
remote-agent servers test --name deploy-139
remote-agent servers remove --name deploy-139
```

自动化输入可显式使用 `--password-stdin`；不提供命令行密码参数。密码只在初始化请求的内存中使用，网页提交后立即清空，不写入聊天、模型提示、浏览器缓存、客户端配置或日志。网页经既有 139 TLS/WebSocket 中转，139 可见转发内容但不持久保存；这不是端到端加密。

客户端生成独立 Ed25519 密钥，用密码连接 SSH，在远端当前用户的 `~/.ssh/authorized_keys` 追加公钥并保留原文件备份，再使用新密钥实际验证登录。当前要求服务器提供 POSIX shell、公钥认证及常规 `sh/base64/mktemp` 工具。不会自动把 root 密钥转换成其他用户的授权；要使用哪个远端账号就填写哪个账号。

首次连接记录主机公钥指纹（TOFU），之后严格核验该主机身份。不会关闭主机验证。连接放在配置旁 `access/servers/<别名>/`，列表返回原生 `ssh -F CONFIG ALIAS` 所需的配置路径、目标及公钥指纹，不返回私钥。私钥不上传 139，也不同步到其他客户端。Windows 服务身份使用条件服务 SID ACL，避免共享 LocalService 身份读取或改写；普通用户安装使用自身权限，Unix 目录/私钥为 0700/0600。

公钥写入前保存不含密码的恢复记录与原密钥，响应丢失后用同一别名重试复用原密钥；已有连接只测试、不重复写入授权。同名不同目标被拒绝，原连接保留。移除本机登记会将配置和密钥归档至 `access/retired/`，远端公钥仍有效，可在远端删除对应 `remote-agent-<别名>` 公钥撤销访问。

## GitHub 设备授权

```powershell
remote-agent github login
remote-agent github login --username string1225 --wait
remote-agent github status
remote-agent github accounts
remote-agent github cancel
remote-agent github logout --username string1225
```

`git login/status/logout` 仍作为兼容入口。用户名可以留空，邮箱用于 GitHub 网页登录；`--username` 仅接受 GitHub 用户名。网页或 CLI 发起后，常驻客户端调用 Git Credential Manager 的设备授权，返回官方授权网址、设备码和状态。用户在 GitHub 官方页面确认，客户端继续轮询并检查账号确已保存；网页和 CLI 不接收 GitHub 密码或 Token。

关闭网页或终端不会取消已启动的授权；重新打开页面或运行 `github status` 查看同一操作。每个客户端同时只允许一个设备授权；再次发起复用进行中的授权。可明确取消，过期或失败后重新发起。客户端重启会将未完成操作标记中断，保留原账号。状态文件只包含授权码、到期时间、状态及账号名，完成后清除设备码，不保存 GCM 原始输出。

Windows Token 存在原有 `git-auth` DPAPI 存储，后台 Git 自动使用该存储。macOS/Linux 使用当前运行用户已配置的 GCM 系统凭据存储；需安装 Git Credential Manager 并配置可用的系统存储/credential helper。GitHub 已不接受账号密码用于 Git HTTPS 推送。启用客户端 Web 代理时，授权使用经健康核验的当前代理地址；不开启时沿用正常网络配置。

## 权限及升级验证

网页入口通过既有账号属主验证；管理请求不属于 MCP 工具白名单，也不生成会话或任务。Windows CLI 通过认证本机命名管道，限定设备属主、管理员和确含 RemoteAgent 服务 SID 的进程；其他 LocalService 进程不获授权。Unix CLI 使用配置目录内 0600 的本机 Unix socket。密码、原始 SSH 输出和 GCM 输出均不进入操作日志。

自动测试用临时密钥、环回 SSH 服务及模拟 GitHub 授权，验证原授权备份、原生 OpenSSH、幂等恢复、密码不落盘、关闭界面后等待、取消和重启状态；不连接真实服务器、不修改真实凭据或机器服务。网页上线与客户端 core 升级需分别验证；仅部署网页不会给旧 core 增加这些接口。Windows core 仍须按既有签名更新流程发布，不绕过签名检查。
