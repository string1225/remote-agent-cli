# Windows Service 与 MSI

Windows 正式客户端由 MSI 安装到 `%ProgramFiles%\Remote Agent`，默认注册自动启动的 `RemoteAgent` Windows Service；支持激活页的新版本可选择手动启动。SCM 负责服务启动和进程恢复，用户退出登录不会停止正在运行的服务。正式 MSI 路径不创建计划任务。

## 安装与原安装迁移

网页新增 Windows 设备后复制绑定码，再下载 x64 或 ARM64 MSI。支持激活页的新版本在原向导的完成页接受绑定码，并提供随系统启动、网页代理、配置 Codex、安装 Codex CLI 四个选项；除网页代理外默认勾选。复制安装指令可自动下载对应架构 MSI 并预填绑定码。MSI 将 `%ProgramFiles%\Remote Agent` 加到系统 PATH 首位，保留其他 PATH 项。流程、恢复和验收限制见 [MSI 设备激活](msi-activation.md)。旧 `windows.ps1`、`cli-windows.ps1`、`codex-windows.ps1` 返回 410 和迁移说明。

安装器通过 Windows Installer 的相关产品检测选择页面：首次显示“安装”，找到旧版显示“升级”，当前产品已安装时显示“修复”。升级/修复说明明确保留原绑定和业务数据；修复设置 `REINSTALL=ALL`。磁盘判断必须比较 `OutOfDiskSpace = 1`；属性值为字符串 `0` 时表示空间充足，不能仅检查属性是否存在。

在 MSI 内完成激活后，如需使用终端，新开 PowerShell（Windows Terminal 请完全退出后重新打开）运行：

```powershell
remote-agent version
codex
```

绑定码确定设备所属账号；选择配置 Codex 时才申请设备关联 Token、备份并更新配置。CLI 的 connect 绑定命令继续兼容，旧版 MSI 仍可使用该路径；remote-agent codex setup 使用同一绑定刷新配置并补装缺少的 CLI。手动 --token-prompt 仍兼容，先验证并显示账号信息。模型 Token 不拼入命令行；服务配置在原用户环境捕获后请求管理员授权，只有路径信息进入 UAC 参数。

MSI 按 Windows 系统语言设置 `HKLM\SOFTWARE\RemoteAgent\Language`：中文（含繁体）为 `zh`，其他语言为 `en`。CLI 的帮助、输入提示、配置结果和已知操作错误统一读取该设置，JSON/MCP 协议不翻译。`REMOTE_AGENT_LANGUAGE=zh/en` 可临时覆盖用于诊断；尚未安装语言设置时使用系统语言。MSI 界面的语言与 CLI 设置相互独立。

设备右键或更多菜单的“修复绑定”提供安装、更新和恢复指引。普通修复复用原本地配置；配置丢失时优先恢复备份，也可生成同设备 ID 的恢复命令（--rebind <ID>）。生成恢复凭证不影响旧连接，成功兑换才替换设备凭证。CLI 兑换前保存私有备份及 pending journal，响应丢失/写文件失败可重复运行继续恢复；原项目、历史和录音保留。全部配置备份都丢失时，IM Secret 等仅存本机的数据需重新填写。

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

服务使用 **LocalService + RemoteAgent 服务 SID**，不以 LocalSystem 执行项目任务。配置时为服务 SID 添加原 Agent 数据目录、Codex/Qoder 配置目录及允许项目根目录的访问权限。当前所有会话为完整授权；旧安装若仅授 RX，升级后运行 connect 完成管理员 ACL 重配。原文件属主、凭证和业务内容不改变；不同服务不能仅凭同样的 LocalService 身份使用这些 SID 授权。刷新 config.toml 时仅继承原有 RemoteAgent SID 授权，避免 Token 更新后后台服务失去读取权限。

PATH、用户目录、`CODEX_HOME`、已检测到的 Codex/Qoder 路径从原用户捕获，不依赖服务的默认 `systemprofile` 或临时终端环境。更换工具路径、项目根目录或本机权限后，重新执行 `service profile` 和 `service configure`。其他用户通过 Windows 凭证管理器/DPAPI 保存的登录凭证不等同于文件授权；依赖交互用户登录的工具仍需为服务身份单独准备认证。

服务进程位于非交互的 Session 0，不接受登录/退出事件作为停止条件。Stop/Shutdown 会取消并等待 Agent 退出；进程意外终止时 Job Object 清理其 CLI 子进程，避免孤儿任务与恢复后的 daemon 重叠。SCM 配置为失败 5 秒后重启，网络恢复由 Agent 自行重连。

Codex App 商店页面等桌面操作仍应在登录用户会话中进行，不能通过 Session 0 服务直接显示。

## 后台 GitHub 授权

Git 登录修复及 `remote-agent ssh configure` 的一次性本机授权、回滚与验证见 [后台 Git 与 SSH](service-credentials.md)。

客户端初始化（connect，含已有绑定、修复和恢复）、enroll、service configure 完成后会提示后台 GitHub 授权。桌面 Git Credential Manager 的登录凭据不会自动供 LocalService 使用；后台任务读取私有仓库或推送前，在设备属主的普通 PowerShell 运行：

```powershell
remote-agent git login
remote-agent git status
```

需要先安装包含 Git Credential Manager（GCM）的 Git for Windows，并启动已配置的 RemoteAgent 服务。login 由服务执行 GCM 的设备码授权，在终端显示官方地址和验证码；请自行在浏览器确认。无需管理员授权，不会在 Session 0 打开浏览器。登录成功后，服务启动的 Codex/Qoder 使用独立的 DPAPI 凭据存储，退出 Windows 登录或重启后继续使用。status 只列出后台已保存的账号，不验证 Token 是否已在 GitHub 撤销，也不代表已拥有某个仓库的推送权限。

```powershell
# 指定账号或刷新失效授权
remote-agent git login --username YOUR_GITHUB_ACCOUNT --force
# 仅移除选定后台账号；桌面 Git 登录保留
remote-agent git logout --username YOUR_GITHUB_ACCOUNT
# 自定义设备配置必须与绑定和服务一致
remote-agent git login --config 'C:\Users\you\custom-agent\config.json'
```

`--no-start` 初始化后，应先在管理员终端用 `remote-agent start --config ...` 启用服务，再回到设备属主的终端授权。服务未升级或未启动时，git 命令明确返回不可用，不能把当前桌面的登录结果当作后台授权成功。macOS/Linux 的客户端继续以实际运行用户完成普通 Git 认证，无需此 Windows 命令。

授权管理只经本机命名管道，双方核对进程身份；仅设备属主、管理员或 SYSTEM 可管理。凭据位于原数据目录 `git-auth`，使用 GCM CurrentUser DPAPI 和仅 RemoteAgent SID/SYSTEM/管理员可访问的受保护目录权限；OWNER RIGHTS 限制共享 LocalService 属主的隐式改权限能力。设备码、Token 不经 139、不写聊天历史或服务日志。服务进程环境选择独立 GitHub helper，保留原 `.gitconfig`、桌面凭据、Git 作者信息和其他主机配置。任务禁用交互密码提示，缺少授权时不会无限等待；用上述 login/status 处理。

普通测试在临时目录使用模拟凭据验证 GCM DPAPI 存取、重启 helper 后读取、管道、取消和权限，不执行真实授权或修改真实服务。目标环境仍须验收真实 LocalService 登录、私有仓库读取、`git push --dry-run`，以及经过授权的实际推送、退出登录和冷启动。公开仓库的 `ls-remote` 成功不能代替认证验收。

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

MSI 安装固定宿主和 RemoteAgentUpdater 服务。宿主 3 信任绑定的 HTTPS 站点，不要求发布签名；日常使用 `remote-agent update` 校验摘要、大小及发布序号并更新独立业务版本，无需重新运行 MSI 或 UAC；`update --check` 只读检查，`update --rollback` 恢复保留的上一稳定版。更新不改 SCM、PATH 或目录权限。业务服务仍是 LocalService；SYSTEM 更新服务只管理受保护程序目录，不执行业务代码。

当前协议、权限边界和恢复见 [客户端维护](client-maintenance.md)。旧版更新框架文档保留供追溯。宿主 1/2 需一次管理员替换，可运行 `tools/upgrade-host.ps1` 而不重装 MSI；脚本保留原服务与用户数据，拒绝中断任务或录音并提供失败恢复。首次安装和服务注册变更仍由 MSI 处理。

卸载从 Windows 设置中的应用列表或 `msiexec /x` 进行，停止并删除服务、程序及 MSI 自己添加的 PATH 项，保留业务数据与机器服务路径配置。停止/卸载会中断正在执行的任务，应先结束任务再维护。

## 构建与验证

```powershell
./tools/msi/build.ps1 -PackageVersion 1.0.1 -ReleaseVersion <Git提交号>
```

脚本下载固定版本 WiX 3.14.1 并校验固定 SHA-256，也支持 `-WixDirectory` 指定预先准备的工具。输出：

在原生 x64 Windows 上，构建会自动运行 `tools/msi/verify-ui.ps1`：通过只读 MSI 数据库/会话验证磁盘 0/1 条件，以及安装、升级、修复模式的实际格式化文本。只运行产品检测、磁盘估算和属性设置动作，不执行安装、服务控制或自定义 EXE。也可单独传入 `-PackagePath` 检查已构建包。

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
