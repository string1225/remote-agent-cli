# Windows Service 与 MSI

Windows 正式客户端由 MSI 安装到 `%ProgramFiles%\Remote Agent`，注册自动启动的 `RemoteAgent` Windows Service。SCM 负责开机启动和进程恢复，用户退出登录不会停止服务。正式 MSI 路径不创建计划任务。

## 安装与原安装迁移

先从网页“Codex 接入与帮助”下载 x64 或 ARM64 MSI，双击安装。MSI 安装自动 Windows Service，并将 `%ProgramFiles%\Remote Agent` 加到系统 PATH 首位，保留其他 PATH 项。完成页保持打开，显示可复制的 Token 配置与设备绑定命令，点击“完成”才关闭。无需 Windows 安装脚本；原 `windows.ps1`、`cli-windows.ps1`、`codex-windows.ps1` 返回 410 和迁移说明。

新开 PowerShell（Windows Terminal 请完全退出后重新打开），运行：

```powershell
remote-agent version
remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-prompt
```

Token 在终端隐藏输入，不出现在命令历史或 MSI 属性中。网页创建 Token 后也可复制 `'<Token>' | remote-agent codex setup --server '<站点>' --token-stdin`；这条含秘密的命令不要分享。设备绑定独立进行：网页添加客户端后运行生成的 `remote-agent connect --server ... --token-stdin`，或用 `--token-prompt` 隐藏输入设备凭证。`connect` 捕获原用户环境后请求管理员授权配置服务，只有路径信息进入 UAC 参数。

MSI 按 Windows 系统语言设置 `HKLM\SOFTWARE\RemoteAgent\Language`：中文（含繁体）为 `zh`，其他语言为 `en`。CLI 的帮助、输入提示、配置结果和已知操作错误统一读取该设置，JSON/MCP 协议不翻译。`REMOTE_AGENT_LANGUAGE=zh/en` 可临时覆盖用于诊断；尚未安装语言设置时使用系统语言。MSI 界面的语言与 CLI 设置相互独立。

已有设备旁的“修复”重新显示安装和绑定指引。待绑定设备可生成属于同一设备 ID 的新一次性凭证；已绑定设备在原电脑运行 `remote-agent connect --server ...`，复用原本地配置并重新配置后台服务，不更换凭证、不创建设备。配置文件丢失时先恢复备份，修复入口不承诺恢复丢失的本地凭证或业务数据。

已有 `~/.remote-agent/config.json` 时，检查原服务器并复用设备绑定，既不重新兑换安装令牌，也不转移账号。服务继续使用原数据目录：

- `config.json`：原设备凭证、工作目录白名单、写入授权、项目服务。
- `history.db`：会话、消息、原生 Codex/Qoder session ID。
- `sayso/`：讨论状态、Action、执行记录和录音。
- Codex 配置仍在原 `CODEX_HOME` 或用户 `.codex` 中。

服务的路径与工具环境保存在 `%ProgramData%\RemoteAgent\service.json`。它只包含原配置路径和允许的环境变量，不复制设备或模型凭证。仅管理员和 SYSTEM 可以修改机器服务的目标。

迁移检查旧 `RemoteAgent` 计划任务的完整启动动作，仅移除指向这份配置和原安装程序的任务。发现其他配置的同名任务则拒绝迁移。新服务激活失败时恢复原服务配置及旧任务 XML/运行状态，原业务文件保留。

`connect --no-start` 完成绑定与配置，并禁用服务；后续以管理员权限运行 `start` 启用。Windows 不提供新建计划任务或独立脚本安装入口。已有绑定执行 `connect` 会检查原服务器并保留配置，不兑换新凭证。

用户确认的清理范围是旧版程序与启动项，**保留绑定、配置、项目、历史和录音**。MSI 成功提交后，才移除原用户 `.remote-agent/bin/remote-agent.exe` 及旧副本，并检查、停止和删除动作完全匹配原配置的旧计划任务。外来同名任务、重定向目录或正在使用的文件不会被强删；错误记入安装日志。保留同目录中的 Codex/Qoder 等其他工具，不递归删除业务数据。

## 服务身份与访问权限

服务使用 **LocalService + RemoteAgent 服务 SID**，不以 LocalSystem 执行项目任务。配置时为服务 SID 添加原 Agent 数据目录和 Codex/Qoder 配置目录的访问权限，并按原 `allowedRoots` / `allowWrite` 授予项目读取或修改权限。原文件属主、凭证及业务内容不改变；不同服务不能仅凭同样的 LocalService 身份使用这些 SID 授权。

PATH、用户目录、`CODEX_HOME`、已检测到的 Codex/Qoder 路径从原用户捕获，不依赖服务的默认 `systemprofile` 或临时终端环境。更换工具路径、项目根目录或本机权限后，重新执行 `service profile` 和 `service configure`。其他用户通过 Windows 凭证管理器/DPAPI 保存的登录凭证不等同于文件授权；依赖交互用户登录的工具仍需为服务身份单独准备认证。

服务进程位于非交互的 Session 0，不接受登录/退出事件作为停止条件。Stop/Shutdown 会取消并等待 Agent 退出；进程意外终止时 Job Object 清理其 CLI 子进程，避免孤儿任务与恢复后的 daemon 重叠。SCM 配置为失败 5 秒后重启，网络恢复由 Agent 自行重连。

Codex App 商店页面等桌面操作仍应在登录用户会话中进行，不能通过 Session 0 服务直接显示。

## 手动配置与诊断

先安装对应 CPU 架构的 MSI，再在原用户 PowerShell 捕获环境：

```powershell
$exe = "$env:ProgramFiles\Remote Agent\remote-agent.exe"
$config = "$env:USERPROFILE\.remote-agent\config.json"
# 新设备先使用网页令牌完成 enroll --token-stdin；已有设备保留原配置。
& $exe service profile --config $config --output "$env:TEMP\remote-agent-profile.json"
```

以管理员身份运行 PowerShell，使用刚捕获文件的绝对路径：

```powershell
& "$env:ProgramFiles\Remote Agent\remote-agent.exe" service configure --profile 'C:\Users\you\AppData\Local\Temp\remote-agent-profile.json'
& "$env:ProgramFiles\Remote Agent\remote-agent.exe" service status
Get-Service RemoteAgent
sc.exe qc RemoteAgent
sc.exe qfailure RemoteAgent
```

首次安装但尚未配置的服务保持运行并等待配置，不会默认绑定或访问某个用户目录。`service status` 的 `configured` 表示是否存在有效服务路径配置。运行状态日志仍位于原数据目录 `agent.log`，配置加载等启动错误也写入 Windows 应用程序事件日志的 `RemoteAgent` 来源。

通过 `connect` 请求管理员授权时，子进程的配置输出和具体错误保存到原配置目录 `service-setup.log`，失败提示包含路径和退出状态；取消 UAC 与服务配置失败分别报告。授权目录继承可能需要数分钟，单目录超时为 10 分钟，授权启动器总超时为 20 分钟，超时会明确报告；不能因为等待时没有新输出就把它判为取消授权。诊断日志只包含阶段、路径、耗时及操作系统错误，不包含设备或模型 Token。

2026-10-25 实机排查：旧版本对整个用户主目录的 `icacls` 权限继承限定 60 秒，处理被中断后只返回无详情的退出码 1。相同目录实际耗时 3 分 31 秒；延长时限后原设备配置成功，`configured=true` 且 `agent.log` 出现 `agent connected`。这次是现有服务的用户故障修复，不属于普通测试的 MSI/SCM 安装验收。

139 服务端可用 `journalctl -u remote-agent` 查看服务日志，Nginx 访问日志为 `/var/log/nginx/remote-agent.access.log*`。本地服务配置发生在连接前，服务器无法记录其 UAC 或目录授权错误；应先查本地配置日志。旧客户端连接返回 HTTP 426 时应升级对应平台客户端。

`start` / `stop` 在 MSI 安装中控制 SCM 服务并启用/禁用开机启动，需要管理员权限；不会退回到创建计划任务。传入其他用户配置时拒绝控制现有服务。

## 升级、修复和卸载

使用更高三段版本号的 MSI 升级。组件 GUID 和关键路径保持稳定，Windows Installer 在事务内停止服务、替换程序并重启，成功执行新安装后才移除旧产品登记。升级复用现有服务对象，保留服务 SID 和恢复策略；失败时回滚程序文件并恢复服务。MSI 只拥有程序、许可证、安装说明及安装目录登记，不拥有业务数据；修复、升级和卸载都不删除上述配置、项目、历史或录音。

MSI 管理的 EXE 拒绝原地 `update`，升级应下载安装新版 MSI。`update --check` 可以继续检查发布版本。Windows 不再分发独立脚本安装器。

卸载从 Windows 设置中的应用列表或 `msiexec /x` 进行，停止并删除服务、程序及 MSI 自己添加的 PATH 项，保留业务数据与机器服务路径配置。停止/卸载会中断正在执行的任务，应先结束任务再维护。

## 构建与验证

```powershell
./tools/msi/build.ps1 -PackageVersion 1.0.1 -ReleaseVersion <Git提交号>
```

脚本下载固定版本 WiX 3.14.1 并校验固定 SHA-256，也支持 `-WixDirectory` 指定预先准备的工具。输出：

- `dist/remote-agent-windows-amd64.msi` 与 `.sha256`。
- `dist/remote-agent-windows-arm64.msi` 与 `.sha256`。

MSI 版本采用 Windows Installer 三段版本号，发布时递增；程序版本单独使用 Git 提交号。安装包尚未做发布者代码签名。

服务端发布时，把同一提交构建的两个 MSI 及其 SHA-256 文件放到 `DOWNLOADS_DIR`（默认 `dist`），网页 MSI 下载链接才可使用。`go run ./tools/build` 仅构建 EXE/Unix Agent，不生成 MSI；可从 CI 的 `remote-agent-windows-msi` artifact 取得安装包。Docker 构建上下文允许带入这些 MSI 文件，镜像会将它们与 Go 构建的其他分发包一起复制到下载目录；Linux 镜像构建本身不生成 Windows MSI。

普通 `go test ./...` 不安装 MSI、不注册机器服务或计划任务。Windows/macOS 旧更新测试只验证隔离目录中的替换和回滚，系统自启动脚本测试使用模拟接口。真实安装测试必须在一次性管理员 Windows 环境中显式执行：

```powershell
./tools/msi/acceptance.ps1
```

该测试拒绝接触已有 RemoteAgent 服务或服务配置，验证安装、自动启动/服务 SID/Session 0、服务身份下的实际工具子进程及原环境、设备鉴权心跳、网络重连、崩溃恢复、升级、注入失败后的产品回滚、修复、卸载，以及系统 PATH 安装/升级去重/卸载、新 PowerShell 直接调用、旧程序清理和配置/项目/会话/原生 session ID/录音的保留。CI 的 `windows-msi` job 执行这套测试，再生成正式 x64/ARM64 安装包。

SCM 生命周期与登录事件由自动化覆盖；真实整机冷启动和退出当前桌面登录的人工验收不在共享开发机上执行。ARM64 包在构建阶段校验，原生 ARM64 安装运行仍需目标硬件验收。
