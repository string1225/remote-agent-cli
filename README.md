# Remote Agent

把 Windows / macOS 电脑上的 Codex、Qoder CLI 接到一个私人网页工作台。

**服务器负责账号、设备目录、保活和 WebRTC 信令；任务请求和输出通过浏览器 ↔ Agent 的加密 DataChannel 传输。** ICE 优先尝试直连，无法穿透时可使用自建 TURN 中继。中继仍然转发端到端加密的数据，但会占用 TURN 服务器带宽。

这是可运行的单节点 MVP：Go 服务端、Go CLI、内嵌中文网页；无需 Node.js、npm 或数据库服务。

## 已实现

- 用户名 / 密码注册、bcrypt 密码哈希、30 天 HttpOnly Cookie 登录态、退出登录。
- 创建设备并生成 Windows PowerShell / macOS Shell 安装命令；一次性令牌 15 分钟过期，事务保证只能绑定一次。
- 下载独立 CLI、SHA-256 校验、设备凭证落盘、用户登录后自启动。
- Windows Scheduled Task / macOS LaunchAgent；心跳、掉线重连、在线状态、撤销设备。
- 在网页通过直连注册 Codex / Qoder 服务及项目目录；配置保存在目标电脑。
- 添加项目时可按目录名搜索、输入部分路径自动补全，或逐层浏览；支持键盘和手机触摸，目录候选通过 P2P 从目标电脑读取。
- WebRTC DTLS/SCTP 加密、Trickle ICE、STUN / TURN、短期 TURN 凭证、会话归属校验。
- 桌面三栏工作台：设备 → Agent → 项目目录、可搜索会话列表、占 50% 的会话区域；手机全屏会话与左右抽屉切换。
- 会话历史保存在目标电脑 `~/.remote-agent/history.db`，通过加密直连分页读取；支持新建、搜索、刷新后查看，以及按 Codex / Qoder 原生会话 ID 继续对话。
- 流式输出与打字效果、停止任务、断线取消、设备级单任务互斥、20 分钟任务超时、30 分钟连接上限。
- 工作目录白名单、默认只读策略、本机手动开启文件编辑权限。

## 本地启动

需要 **Go 1.26.0 或更高版本**。从仓库根目录运行：

```sh
go mod download
go run ./tools/build
go run ./cmd/server
```

打开 [http://localhost:8080](http://localhost:8080)。第一条构建命令会生成六个平台的 Agent 和校验文件到 `dist/`；服务器将从该目录提供安装包。首次交叉编译需要一些时间。

1. 注册账号并登录。
2. 点击「添加设备」，输入名称。
3. 在目标电脑执行生成的安装命令。
4. 等待设备上线，点击「连接设备」。
5. 可选：点击「注册服务」，选择已安装的 Codex / Qoder CLI，填写目标电脑上的项目绝对路径。
6. 左侧选择 Agent 下的项目文件夹，中间新建或选择会话，右侧输入请求；实时输出通过数据通道返回。手机左上角切换设备 / 项目，右上角切换会话。

Agent 可以独立安装、上线和连接，Codex 和 Qoder 均不是必需依赖；两者都未安装时仍可管理设备。只有执行对应的 AI 任务才需要安装并注册该服务。

连接有效期由服务端控制。新版 Agent 使用服务端签发的相对有效期，避免两台机器的时钟偏差导致连接协商失败；服务端仍执行到期和撤销检查。此修复需服务端与 Agent 均更新，旧版本仍应保持系统时间同步。

`localhost` 只适合在同一台电脑开发验证。连接其他电脑前，必须配置一个双方可访问的 **HTTPS 域名**；浏览器页和 Agent 必须使用同一 `PUBLIC_URL`。

## 部署服务端

```sh
go build -trimpath -o bin/remote-server ./cmd/server
PUBLIC_URL=https://agents.example.com LISTEN_ADDR=127.0.0.1:8080 ./bin/remote-server
```

PowerShell：

```powershell
go build -trimpath -o bin/remote-server.exe ./cmd/server
$env:PUBLIC_URL = 'https://agents.example.com'
$env:LISTEN_ADDR = '127.0.0.1:8080'
.\bin\remote-server.exe
```

在前方放置 Caddy / Nginx 终止 TLS，并代理 WebSocket Upgrade。示例见 [deploy/Caddyfile](deploy/Caddyfile)。不要缓存 `/api/`、`/install/` 响应；访问日志应去掉安装 URL 的查询参数，避免记录尚未使用的绑定令牌。

也可使用 Docker：

```sh
docker compose up --build -d
```

默认映射 `127.0.0.1:8080`，数据库保存在 volume 中。生产部署设置 `PUBLIC_URL` 并在宿主机配置 HTTPS 反向代理。镜像构建会同时打包 Windows / macOS / Linux Agent。

| 环境变量 | 默认值 | 用途 |
| --- | --- | --- |
| `PUBLIC_URL` | `http://localhost:8080` | 对外地址；除 loopback 外必须 HTTPS，支持 `/agents` 等子路径 |
| `LISTEN_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `DATABASE_PATH` | `data/remote-agent.db` | bbolt 数据库路径 |
| `DOWNLOADS_DIR` | `dist` | 已构建 Agent 及校验文件目录 |
| `ALLOW_SIGNUP` | `true` | 创建个人账号后可设为 `false` 关闭新注册 |
| `STUN_URLS` | `stun:stun.l.google.com:19302` | 逗号分隔；空值禁用，可替换为自己的 STUN |
| `TURN_URLS` | 空 | 例如 `turn:turn.example.com:3478?transport=udp,turn:turn.example.com:3478?transport=tcp` |
| `TURN_SECRET` | 空 | 与 coturn 相同的共享认证密钥；不要提交到 Git |

服务器使用 socket 来源 IP 做认证限流，不信任任意 `X-Forwarded-For`。反向代理部署时，后端会把同一代理视作一个来源；大规模部署前需在可信代理层做更细粒度限流。bbolt 只支持一个写入进程；不要给同一数据库启动多个实例。备份时停止服务器后复制数据库文件。

## NAT 穿透与 TURN

不是所有网络都能 P2P 直连：对称 NAT、企业防火墙或禁用 UDP 的网络通常需要中继。**未配置 TURN 时，这类网络会明确连接失败，不会自动把任务放到应用服务器转发。**

Linux 上的 coturn 示例见 [deploy/compose.turn.yaml](deploy/compose.turn.yaml)：

```sh
export TURN_PUBLIC_IP=203.0.113.10  # 替换为真实公网 IP
export TURN_SECRET=replace-with-a-long-random-secret
docker compose -f deploy/compose.turn.yaml up -d
```

放行 TCP/UDP 3478 和 UDP 49160–49200。配置应用服务的 `TURN_URLS` / `TURN_SECRET`。服务端按用户签发 1 小时有效的 TURN REST 凭证，不向浏览器或 Agent 下发共享密钥。需要应对只允许 HTTPS 的网络时，应另行配置带有效证书的 TURN/TLS 443 或 5349，以及相应 `turns:` URL；示例未配置 TURN/TLS。

## CLI 与本机权限

安装位置：

- Windows：`%USERPROFILE%\.remote-agent\bin\remote-agent.exe`
- macOS：`~/.remote-agent/bin/remote-agent`
- 配置：`~/.remote-agent/config.json`
- 会话数据库：`~/.remote-agent/history.db`，保存网页创建的会话、请求、回复和原生 CLI 会话 ID；不会上传到控制服务器或写入浏览器持久存储。
- 日志：`~/.remote-agent/agent.log`（仅运行状态，不记录任务正文；启动时对超过 10 MiB 的旧日志做一次轮换）

自启动指 **用户登录后** 启动，使用当前用户权限。不是登录前的系统级服务。电脑睡眠 / 关机时离线。macOS 在登录图形会话后安装 LaunchAgent；Windows 组策略若禁止创建任务，需要管理员调整策略或手动运行。

```sh
remote-agent status
remote-agent run
remote-agent autostart install
remote-agent autostart remove
```

以上命令使用完整安装路径，或自行将 `~/.remote-agent/bin` 加入 PATH。安装脚本不会修改 shell 配置。

手动绑定可指定根目录和写入授权：

```sh
printf '%s' 'ONE_TIME_TOKEN' | remote-agent enroll \
  --server https://agents.example.com --token-stdin \
  --roots /Users/you/projects --allow-write
```

`--roots` 多个目录使用操作系统的路径列表分隔符（Windows `;`，macOS `:`）。默认根目录是当前用户的 home。已绑定设备可先停止 Agent，编辑本地配置的 `allowedRoots` / `allowWrite`，再启动；网页无权修改这两个字段。不要把凭证文件上传或分享。Windows 安装器为目录设置用户 / SYSTEM ACL，macOS 使用用户私有目录和 `0600` 配置文件。

- Codex 默认 `exec --json --sandbox read-only`，写入授权后使用 `workspace-write`。非交互审批策略为 `never`，受限操作失败，不自动越权。
- Qoder 默认 `--print --output-format stream-json --permission-mode plan`，写入授权后使用 `accept_edits`。需要交互批准的操作不在本版本提供网页审批流。
- 各提供方 CLI 均为可选。使用某个服务时，在目标电脑安装并登录对应 CLI 即可。Agent 复用该用户本地的 CLI 登录态；不把模型 API Key / OAuth 凭证交给服务器。
- 检测 `codex`、`qodercli` 或 `qoder` 的帮助信息。Qoder 桌面编辑器的同名启动器不会被当成 CLI。可设置本机 `RA_CODEX_PATH` / `RA_QODER_PATH` 为 CLI 路径；自启动需要能继承这些环境变量。
- 若安装 CLI 后未检测到，重新启动 Agent。macOS LaunchAgent 保存安装时的 PATH，可重新执行 `autostart install` 更新。
- `allowedRoots` 检查规范化的真实目录及符号链接，限制**启动工作目录**；它不是独立 OS 沙箱，不能限制 CLI/MCP/本地配置的一切文件访问。提供方各自的权限机制仍然生效。不要把不受信任项目当成隔离环境。

接口依据：[Codex 非交互模式](https://learn.chatgpt.com/docs/non-interactive-mode)、[Qoder 脚本模式](https://docs.qoder.com/cli/run-in-scripts)、[Qoder Plan](https://docs.qoder.com/cli/plan-mode)。

## SaySo：讨论、录音与 Action

连接设备后，点击设备下的 **♫ SaySo · 讨论转 Action**。首次打开时会带入已经注册的项目目录，也可添加允许根目录内的其他项目。无需在目标电脑安装 Node.js 或另开 SaySo 服务，Codex/Qoder 都不是录音和本地规则功能的前提。

- 项目与讨论会话管理，默认暂停；支持继续录音、暂停、每 10 分钟切段、逐段播放和文字补充。
- 浏览器提供实时语音转写；不可用时仍保存录音。浏览器的识别服务可能联网处理音频，此行为不属于 Remote Agent 的 P2P 通道。
- 录音期间每 2 分钟整理最近 5 分钟讨论，暂停时处理剩余内容；可手动完整分析。分析模型默认跟随本机 Codex，也可选其他账号支持的模型，或完全使用本地关键词规则。
- Codex 分析会创建或补充已有 Action，保存分析输入、过程、结果及讨论原话；修改后的需求回到待完善，避免沿用旧的 Ready 状态。
- 编辑标题、详情、优先级，确认“讨论完善”后派发给 Codex。执行结果实时回传，可查看或停止任务；始终遵守 Agent 本地 `allowWrite`，默认只读。分析始终只读。
- 手机全屏展示讨论或 Action，顶部切换项目/会话；录音、文字、Action 和执行输出全部保存在目标电脑的 `~/.remote-agent/sayso/`，刷新或换浏览器后可恢复。

数据文件为 `sayso/state.json`（原子替换），录音文件在 `sayso/recordings/`。录音经已认证的 DataChannel 分片上传和读取，不通过控制服务器 HTTP 上传。每片 16 KB，每段最多 64 MiB；一个连接同时上传一段，上传期间不完整的文件在断连或重启后清理。上传失败时网页提供录音备份下载，请在关闭页面前保存；浏览器崩溃或关闭前尚未停止的录音不能保证恢复。

讨论快照超过单个 SCTP 消息大小时自动分帧。当前每台设备最多保存 500 个讨论会话、3000 条 Action、1000 条近期执行记录；单会话最多 10000 条发言、1000 段录音，状态文件最多 24 MiB。删除讨论会话只删除索引，目标电脑的录音文件保留供本地恢复。设备断线、撤销或 30 分钟连接租约到期会停止该连接的分析/执行并暂停讨论状态，需要重新连接后继续。

界面源代码及构建方法见 [sayso-ui/README.md](sayso-ui/README.md)。原 `say-so` 项目保持独立，其已有数据不会自动迁移。

## 撤销与卸载

网页「撤销设备」立即禁用设备凭证及未使用的绑定令牌，并关闭相关会话。退出登录只关闭当前浏览器登录态所建立的会话。控制连接断开时 Agent 取消该连接下的任务；静默网络故障最多需要约 65 秒被心跳检测。

在目标电脑运行 `remote-agent autostart remove`，再手动删除 `~/.remote-agent`。Windows 请等进程停止再删除可执行文件。CLI 不提供远程任意删除本地文件的接口。

## 开发与验证

```sh
go test ./...
go vet ./...
node --check web/app.js     # 只验证工作台原生 JS 语法
(cd sayso-ui && npm ci && npm run build) # 修改 SaySo 界面时需要（PowerShell 可进入目录后运行）
go run ./tools/build       # 交叉编译所有支持的平台
```

测试覆盖绑定令牌并发兑换、账号隔离、CSRF 来源校验、登录态持久化、工作目录边界、UTF-8 流、stdin 传递，以及真实 Pion ↔ Pion WebRTC 建连 / 数据传输 / 撤销中止任务。集成测试使用模拟 CLI，不调用付费模型或改动真实项目。GitHub Actions 在 Linux / Windows / macOS 上运行测试，另执行 Linux race detector 和交叉编译。

## 当前范围

- 单实例服务端和本地数据库；尚未实现多节点路由、组织共享、密码找回或 MFA。
- 每个请求开启一次 CLI 进程，同一会话后续请求使用原生会话 ID 恢复上下文。尚未实现交互终端、文件上传、网页审批或导入此前直接在 CLI 中创建的历史会话。
- 任务和输出不会落在应用服务器，也没有中央任务历史；关闭连接会取消任务，已收到的会话输出保存在目标电脑。
- 设备元数据、服务名称及项目路径会随心跳保存在服务器。服务端仍是登录、信令和网页代码的信任根，不宣称能抵抗已被攻陷的控制服务器。
- 每台设备最多 8 个连接 / 1 个任务；每次任务最长 20 分钟，连接最长 30 分钟；单次提示最多 32 KiB，stdout/stderr 各最多 16 MiB，会话历史按页读取；SaySo 使用前述独立数据限额。
- 已提供跨平台安装和构建，但真正跨公网 NAT / TURN、不同 Windows 策略以及 macOS 自启动仍需在目标部署环境验收。分发二进制尚未做代码签名 / Apple 公证。

详细协议见 [docs/architecture.md](docs/architecture.md)。

## 当前服务器部署

访问地址：`https://codex.sunny-string.cn/agents/`（不带末尾斜杠也会自动跳转）。

- `139.224.12.141` 上由独立 `remote-agent.service` 托管，监听 `127.0.0.1:8086`。
- Nginx 独立站点代理 `/agents/`，保留原路径，并支持 WebSocket Upgrade。
- TLS 使用独立的 `codex.sunny-string.cn` 证书，由服务器已有 Certbot 定时任务续期。
- 发布目录 `/opt/remote-agent/releases/`，`/opt/remote-agent/current` 指向当前版本。
- 数据库 `/var/lib/remote-agent/remote-agent.db`，服务使用 systemd DynamicUser；数据库与发布目录分离。
- 配置 `/etc/remote-agent/server.env`；日志使用 `journalctl -u remote-agent`。代理访问日志不记录查询参数。

配置模板：[systemd 服务](deploy/remote-agent.service)、[Nginx 站点](deploy/nginx-codex.conf)、[环境变量](deploy/server.env.example)。

子路径部署时必须把完整地址（包含 `/agents`）设为 `PUBLIC_URL`，代理不要移除此前缀。CLI 安装、Cookie 路径、API 与信令会自动跟随此前缀。`/agents/healthz` 可用于健康检查。当前部署未配置 TURN，中继配置方法见前文。
