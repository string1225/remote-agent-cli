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
- WebRTC DTLS/SCTP 加密、Trickle ICE、STUN / TURN、短期 TURN 凭证、会话归属校验。
- 流式输出、停止任务、断线取消、设备级单任务互斥、20 分钟任务超时、30 分钟连接上限。
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
5. 点击「注册服务」，选择已安装的 Codex / Qoder CLI，填写目标电脑上的项目绝对路径。
6. 选择服务、输入请求；实时输出通过数据通道返回。控制台显示实际使用 P2P 还是 TURN。

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
- 目标电脑需要自行安装、登录各提供方 CLI。Agent 复用该用户本地的 CLI 登录态；不把模型 API Key / OAuth 凭证交给服务器。
- 检测 `codex`、`qodercli` 或 `qoder` 的帮助信息。Qoder 桌面编辑器的同名启动器不会被当成 CLI。可设置本机 `RA_CODEX_PATH` / `RA_QODER_PATH` 为 CLI 路径；自启动需要能继承这些环境变量。
- 若安装 CLI 后未检测到，重新启动 Agent。macOS LaunchAgent 保存安装时的 PATH，可重新执行 `autostart install` 更新。
- `allowedRoots` 检查规范化的真实目录及符号链接，限制**启动工作目录**；它不是独立 OS 沙箱，不能限制 CLI/MCP/本地配置的一切文件访问。提供方各自的权限机制仍然生效。不要把不受信任项目当成隔离环境。

接口依据：[Codex 非交互模式](https://learn.chatgpt.com/docs/non-interactive-mode)、[Qoder 脚本模式](https://docs.qoder.com/cli/run-in-scripts)、[Qoder Plan](https://docs.qoder.com/cli/plan-mode)。

## 撤销与卸载

网页「撤销设备」立即禁用设备凭证及未使用的绑定令牌，并关闭相关会话。退出登录只关闭当前浏览器登录态所建立的会话。控制连接断开时 Agent 取消该连接下的任务；静默网络故障最多需要约 65 秒被心跳检测。

在目标电脑运行 `remote-agent autostart remove`，再手动删除 `~/.remote-agent`。Windows 请等进程停止再删除可执行文件。CLI 不提供远程任意删除本地文件的接口。

## 开发与验证

```sh
go test ./...
go vet ./...
node --check web/app.js     # 可选；只验证原生 JS 语法
go run ./tools/build       # 交叉编译所有支持的平台
```

测试覆盖绑定令牌并发兑换、账号隔离、CSRF 来源校验、登录态持久化、工作目录边界、UTF-8 流、stdin 传递，以及真实 Pion ↔ Pion WebRTC 建连 / 数据传输 / 撤销中止任务。集成测试使用模拟 CLI，不调用付费模型或改动真实项目。GitHub Actions 在 Linux / Windows / macOS 上运行测试，另执行 Linux race detector 和交叉编译。

## 当前范围

- 单实例服务端和本地数据库；尚未实现多节点路由、组织共享、密码找回或 MFA。
- 每个请求开启一次新的 CLI 执行；尚未实现 Codex/Qoder 原生多轮会话恢复、交互终端、文件上传或网页审批。
- 任务和输出不会落在应用服务器，也没有中央任务历史；关闭连接会取消任务，当前输出只保留在页面内存。
- 设备元数据、服务名称及项目路径会随心跳保存在服务器。服务端仍是登录、信令和网页代码的信任根，不宣称能抵抗已被攻陷的控制服务器。
- 每台设备最多 8 个连接 / 1 个任务；每次任务最长 20 分钟，连接最长 30 分钟；单次提示最多 32 KiB，stdout/stderr 各最多 16 MiB，网页只保留最近约 40 万字符。
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
