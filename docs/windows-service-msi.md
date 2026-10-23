# Windows Service 与 MSI

Windows 正式客户端由 MSI 安装到 `%ProgramFiles%\Remote Agent`，注册自动启动的 `RemoteAgent` Windows Service。SCM 负责开机启动和进程恢复，用户退出登录不会停止服务。正式 MSI 路径不创建计划任务。

## 安装与原安装迁移

推荐在网页“添加设备”中复制 Windows 安装命令，以原项目用户运行。安装命令下载并校验 MSI，再请求管理员权限安装；原用户完成设备绑定、可选的 Codex 配置和工具环境捕获，管理员步骤配置服务。账号/模型 Token 不传给 MSI，不进入 MSI 属性或 UAC 命令行。

已有 `~/.remote-agent/config.json` 时，检查原服务器并复用设备绑定，既不重新兑换安装令牌，也不转移账号。服务继续使用原数据目录：

- `config.json`：原设备凭证、工作目录白名单、写入授权、项目服务。
- `history.db`：会话、消息、原生 Codex/Qoder session ID。
- `sayso/`：讨论状态、Action、执行记录和录音。
- Codex 配置仍在原 `CODEX_HOME` 或用户 `.codex` 中。

服务的路径与工具环境保存在 `%ProgramData%\RemoteAgent\service.json`。它只包含原配置路径和允许的环境变量，不复制设备或模型凭证。仅管理员和 SYSTEM 可以修改机器服务的目标。

迁移检查旧 `RemoteAgent` 计划任务的完整启动动作，仅移除指向这份配置和原安装程序的任务。发现其他配置的同名任务则拒绝迁移。新服务激活失败时恢复原服务配置及旧任务 XML/运行状态，原业务文件保留。

`-NoStart` 会完成安装与配置，并明确禁用服务；后续以管理员权限运行 `start` 启用。Windows 已移除新建计划任务的安装入口；旧任务仅保留识别、停止、迁移和旧程序更新兼容。独立 CLI/Codex 安装入口继续保留，但不会创建后台计划任务。

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

`start` / `stop` 在 MSI 安装中控制 SCM 服务并启用/禁用开机启动，需要管理员权限；不会退回到创建计划任务。传入其他用户配置时拒绝控制现有服务。

## 升级、修复和卸载

使用更高三段版本号的 MSI 升级。Windows Installer 在事务内停止服务、替换程序并重启，升级失败时回滚旧产品。MSI 只拥有程序、许可证、安装说明及安装目录登记，不拥有业务数据；修复、升级和卸载都不删除上述配置、项目、历史或录音。

MSI 管理的 EXE 拒绝原地 `update`，避免绕过 Windows Installer 文件和版本管理；`update --check` 可以继续检查服务器发布的二进制。旧用户目录内保留的独立 CLI 仍遵循其原有更新方式。

卸载从 Windows 设置中的应用列表或 `msiexec /x` 进行，停止并删除服务和程序文件，保留本机业务数据与机器服务路径配置以便重装。停止/卸载会中断正在执行的任务，应先结束任务再进行维护。

## 构建与验证

```powershell
./tools/msi/build.ps1 -PackageVersion 1.0.1 -ReleaseVersion <Git提交号>
```

脚本下载固定版本 WiX 3.14.1 并校验固定 SHA-256，也支持 `-WixDirectory` 指定预先准备的工具。输出：

- `dist/remote-agent-windows-amd64.msi` 与 `.sha256`。
- `dist/remote-agent-windows-arm64.msi` 与 `.sha256`。

MSI 版本采用 Windows Installer 三段版本号，发布时递增；程序版本单独使用 Git 提交号。安装包尚未做发布者代码签名。

普通 `go test ./...` 不安装 MSI、不注册机器服务或计划任务。Windows/macOS 旧更新测试只验证隔离目录中的替换和回滚，系统自启动脚本测试使用模拟接口。真实安装测试必须在一次性管理员 Windows 环境中显式执行：

```powershell
./tools/msi/acceptance.ps1
```

该测试拒绝接触已有 RemoteAgent 服务或服务配置，验证安装、自动启动/服务 SID/Session 0、服务身份下的实际工具子进程及原环境、设备鉴权心跳、网络重连、崩溃恢复、升级、注入失败后的产品回滚、修复、卸载，以及配置/项目/会话/原生 session ID/录音的保留。CI 的 `windows-msi` job 执行这套测试，再生成正式 x64/ARM64 安装包。

SCM 生命周期与登录事件由自动化覆盖；真实整机冷启动和退出当前桌面登录的人工验收不在共享开发机上执行。ARM64 包在构建阶段校验，原生 ARM64 安装运行仍需目标硬件验收。
