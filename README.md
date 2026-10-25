<!-- Current transport supersedes historical deployment records below. -->

> 2026-10-24：网页和客户端统一通过当前站点中转，客户端与服务器使用持久 WebSocket；不再使用 WebRTC、P2P、STUN 或 TURN。部署时同时更新服务器、客户端包和网页；旧客户端需要升级：Windows 下载并安装新版 MSI，其他系统使用安装命令；原数据保留。

# Remote Agent

把 Windows / macOS / Linux 电脑上的 Codex、Qoder CLI 接到一个私人网页工作台，并为 Codex App / CLI 提供账号管理、个人模型 Token 和本机配置工具。生产入口为 [codex.sunny-string.cn/agents](https://codex.sunny-string.cn/agents/)。

**网页与客户端都通过 139 站点传输。** 客户端保持到服务器的 WebSocket 长连接；服务器转发远程任务、会话与 SaySo 数据，并保留账号授权和本机执行权限。配置了 Codex 代理的模型调用继续走 HTTPS 入口。

这是可运行的单节点 MVP：Go 服务端、Go CLI、内嵌中文网页；无需 Node.js、npm 或数据库服务。

## 项目目标与开发约定

目标是搭建“服务器管理设备与网页入口，客户端常驻执行任务”的平台。Windows 正式客户端使用 **MSI + Windows Service**，开机启动，退出登录后仍在线；支持命令执行、Codex/Qoder 检测、Codex 代理配置与 CLI 安装更新、Codex App 商店入口。飞书/钉钉由客户端直接长连接，AK/SK 只存客户端；每个客户端项目最多一个渠道，每个机器人最多绑定一个项目，映射到固定 session。网页从指定客户端的项目进入 SaySo。需求中的 `code` 指 **Qoder**。

完整职责和现状见 **[docs/project-goals.md](docs/project-goals.md)**，会话命名、MCP 工具与跨设备通知见 **[docs/session-model.md](docs/session-model.md)**。Windows 正式安装入口使用 MSI + Windows Service，支持原数据目录迁移；安装、升级和服务管理见 **[docs/windows-service-msi.md](docs/windows-service-msi.md)**。独立通用命令 API 与完整网页工具管理仍待实现。

所有代码直接提交并推送到 `main`，不新建分支。旧 `codex/unified-codex-platform` 分支的代码已汇入主线；协作规范见 [AGENTS.md](AGENTS.md)。

## 已实现

- 仅管理员创建账号，公开注册始终关闭；bcrypt 密码哈希、30 天 HttpOnly Cookie 登录态、账号停用和最后一名有效管理员保护。
- 用户确认当前密码后可修改用户名或密码；用户名支持 1–64 个 Unicode 字符，去首尾空白并按 Unicode 大小写折叠保持唯一。修改后退出所有浏览器登录，账号 ID、设备归属和会话数据保持不变。
- 用户创建、复制和撤销自己的 Token，复制/撤销按钮右侧并列；管理员仅管理元数据和撤销。新 Token 加密保存，旧哈希 Token 继续有效但无法复制，不提供补存入口。设备绑定自动幂等申请账号关联 Token，并明确显示账号归属。
- 可选的 `/agents/codex/v1` 模型入口接入现有网关，支持 Responses / Compact / 模型目录、SSE 和账号停用取消；保留既有 `/openai/` 接口、旧 Token 及 OAuth 状态。
- 持久本地 CLI 可独立安装，无需绑定设备或启用远程服务；支持官方 Codex CLI 安装/更新、本机 TOML 配置，以及可选的 `start` / `stop`。Windows App 命令打开商店页面，由用户完成安装/更新。
- 创建设备并生成 Windows PowerShell / macOS Shell / Linux Bash 安装命令；一次性令牌 15 分钟过期，事务保证只能绑定一次。
- 下载独立 CLI、SHA-256 校验、设备凭证落盘、跨平台自启动。
- 非 MSI 安装的 `remote-agent update` 从服务器更新 CLI，保留绑定和会话；Windows MSI 仍通过安装新版 MSI 升级。宿主/业务核心分离的更新框架目前仅有设计，见 docs/superpowers/specs/2026-10-25-client-update-framework.md。
- Windows MSI / Windows Service，开机自动运行、退出登录后在线、失败恢复与原目录数据保留；macOS LaunchAgent / Linux systemd 用户服务；心跳、掉线重连、在线状态、撤销设备。旧 Windows 用户安装可迁移。
- 在网页通过站点中转注册 Codex / Qoder 服务及项目目录；配置保存在目标电脑。
- 添加项目时可按目录名搜索、输入部分路径自动补全，或逐层浏览；支持键盘和手机触摸，目录候选经站点从目标电脑读取。
- 网页与客户端统一通过 139 站点的 WebSocket 中转，客户端持续心跳与断线重连，保留账号和会话归属校验。
- 桌面三栏工作台：客户端 → 项目、可搜索会话列表、右侧工作区；系统图标及灰/绿在线点，客户端 + 添加项目，项目图标区分 Codex/Qoder；分隔线可拖动并保存列宽，手机使用左右抽屉。
- 每项目置顶 SaySo，飞书/钉钉默认所有人可聊，按来信自动隔离私聊/群聊会话并显示对方名称。支持 Markdown 和飞书 THINKING 表情。Action/IM/Web 派生任务关闭网页后继续，完成通知回到原渠道会话；Secret 只持久存本机，绑定全局唯一。
- 每次 Codex/Qoder 工作进程注入客户端 MCP：`list_clients_projects`、`start_web_session`、`notify_im_session`；支持同账号跨设备协作，遵守目标客户端允许目录和文件权限。
- 会话历史保存在目标电脑 `~/.remote-agent/history.db`，通过站点分页读取；支持新建、搜索、刷新后查看，以及按 Codex / Qoder 原生会话 ID 继续对话。
- 流式 Markdown、固定底部输入框和长消息滚动；Enter 发送，Shift+Enter 换行，兼容中文输入法；停止任务、断线取消、设备单任务互斥、20 分钟任务超时、30 分钟连接上限。
- 项目目录白名单；所有会话完整授权（Codex danger-full-access/never，Qoder bypass_permissions），旧配置读取时迁移。设备右键支持重命名、安装/更新指引及同 ID 重新绑定恢复。

## Windows 安装与 Token 配置

1. 在网页下载 x64/ARM64 MSI，双击安装；无需安装脚本。
2. MSI 写入系统 PATH，安装自动 Windows Service，清理旧程序和匹配启动项，保留原设备绑定、配置、项目、会话、录音及 Codex/Qoder 工具。
3. 安装完成窗口会保持打开并显示可复制的命令。打开新的 PowerShell（Windows Terminal 需完全退出重开）：

```powershell
remote-agent version
remote-agent codex setup
codex
```

先在已登录网页添加客户端，运行生成的 `remote-agent connect` 绑定命令。绑定会自动申请所属账号的设备 Token 并备份更新 Codex 配置；再运行上面的 setup 补装缺少的 Codex CLI 或刷新配置。手动 --token-prompt 仍可使用，但账号由粘贴的 Token 决定，终端会显示身份。秘密不进入 MSI 属性。

## 本地启动

需要 **Go 1.26.0 或更高版本**。从仓库根目录运行：

```bash
go mod download
go run ./tools/build
read -r -s -p 'Administrator password: ' ADMIN_PASSWORD
printf '\n'
printf '%s' "$ADMIN_PASSWORD" | go run ./cmd/server admin create --username admin --password-stdin
unset ADMIN_PASSWORD
go run ./cmd/server
```

打开 [http://localhost:8080](http://localhost:8080)。`go run ./tools/build` 会生成六个平台的 Agent 和校验文件到 `dist/`；服务器将从该目录提供安装包。首次交叉编译需要一些时间。

1. 使用刚创建的管理员账号登录；其他账号由管理员在网页的「账号管理」中创建。
2. 点击「添加设备」，输入名称。
3. Windows 先下载并安装 MSI，再执行生成的设备绑定命令；其他系统执行安装命令。
4. 等待客户端右侧状态点变绿，点击行上的 + 添加项目。
5. 选择已安装的 Codex / Qoder CLI，填写本机允许目录内的项目绝对路径。
6. 左侧选择项目自动连接，中间选择 Web、SaySo 或已绑定的 IM 会话，右侧工作；项目齿轮配置渠道。实时数据均经站点 WebSocket。手机左上角切换客户端 / 项目，右上角切换会话。

本机仅使用 Codex App / CLI 时，可从「Codex 接入与帮助」安装持久 CLI、创建个人 Token 并复制配置命令，无需添加设备。需要远程操作时再绑定设备并运行 `remote-agent start`。模型网关未启用时，设备安装仍可使用；安装器会跳过自动 Codex 配置，Codex 和 Qoder 仍是可选服务。

`remote-server admin create --username NAME --password-stdin` 创建管理员，`remote-server admin grant --username NAME` 将已有账号提升为管理员。命令读取 `DATABASE_PATH`，必须在服务停止、数据库无其他进程打开时执行；服务不提供匿名的首个管理员初始化接口。已有库升级后账号默认仍是普通用户，需明确提升管理员。

## 本机 Codex 与持久 CLI

Windows 从「Codex 接入与帮助」下载 MSI，安装后在新的 PowerShell 执行 `remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-prompt`，隐藏输入个人 Token。原 Windows 安装脚本已移除，旧 URL 返回 410。macOS/Linux 继续提供 `cli-macos.sh`、`cli-linux.sh`；这些命令不绑定设备或启用远程服务。

```sh
remote-agent --help
remote-agent codex install --server https://codex.sunny-string.cn/agents
remote-agent codex update --server https://codex.sunny-string.cn/agents
remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-stdin
remote-agent app install
remote-agent app upgrade
remote-agent start
remote-agent stop
```

`codex setup --token-stdin` 从标准输入读取网页创建的个人模型 Token，先验证入口，再配置本机；无需设备绑定。已绑定设备也可运行 `remote-agent codex setup`，使用本机设备凭证为该设备新签发专属模型 Token，保留已有 Token；安装或配置失败不会使原有连接失效。`codex install/update` 从服务器转发的官方发布包安装托管 CLI，强制校验官方 SHA-256，不依赖 Node.js；`setup` 会保留已检测到的兼容 CLI，并在缺失时安装。Windows 的 `app install/upgrade` 均打开对应 Microsoft Store 产品页面，由当前登录用户完成安装或更新；命令成功表示已发起打开页面，安装结果需在商店确认。macOS 调用 `brew install/upgrade --cask codex-app`；本工具不支持 Linux Codex App 安装，但支持 Linux Codex CLI。

CLI 安装保留官方完整包和辅助资源：Unix 使用 `codex-package-*`，保留 `bin/codex`、资源目录及相对符号链接命令入口；Windows 保留架构主程序，另以硬链接或复制提供 `codex.exe`。更新先发布辅助文件、最后替换主入口，失败时回滚。下载超时为 25 分钟；压缩包最多 256 MiB、单文件最多 512 MiB、解压总量最多 2 GiB、条目最多 10000 个，拒绝包内链接、特殊文件和越界路径。

网页在 Token 首次显示时提供「复制安装并配置命令」。该一条命令内含个人 Token，可能进入终端历史、剪贴板或屏幕录制；不要分享，使用后清理敏感历史，泄露时在网页撤销。更适合日常操作的方式是下载 `codex-*` 安装脚本并在隐藏输入提示中粘贴 Token，或使用不会把秘密写进命令文本的 stdin 管道。不要将个人 Token 写进安装下载 URL。

本机配置默认写入 `~/.codex/config.toml`，尊重 `CODEX_HOME`。工具先保存私有的 `config.toml.backup-*`，解析 TOML 后原子写入，保留既有模型、其他提供方、项目、通知、MCP 等设置及独立的 OAuth 文件。默认选中的 profile 也会切换到 `sunny_agents`，显式选择的其他 profile 保持原样；旧的 `extrahigh` 推理设置转为 `xhigh`。TOML 格式无效或配置被其他进程同时修改时会停止更新。

提供方使用 `wire_api = "responses"`、`requires_openai_auth = false`、`supports_websockets = false`，把个人 Token 放在私有 TOML 的 `http_headers.Authorization` 中。Windows 限制为当前用户/SYSTEM，Unix 使用私有文件权限；备份也包含凭证，应同样保护。这使 Codex App、终端和后台 Agent 使用同一份配置，无需依赖某个终端的临时环境变量。完全退出并重启 App 后新建本地对话；旧对话可能继续沿用原提供方。此模型入口不代替 ChatGPT 登录、云任务或插件服务。

远程设备安装器支持 Windows `-NoStart` / `-SkipCodex`、macOS/Linux `--no-start` / `--skip-codex`：前者完成安装绑定但保持远程服务停止，后者跳过自动 Codex 配置。Shell 管道形式可写为 `curl -fsSL '设备安装URL' | bash -s -- --no-start --skip-codex`；PowerShell 在生成的脚本块调用末尾附加相应开关。`start` 启用自启动并启动已绑定设备；`stop` 停止并移除对应自启动项，保留持久 CLI、本机 Codex 配置、设备绑定和历史。仅安装本地 CLI 的入口无需这些开关。

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

模型入口只代理 `GET /models`、`POST /responses`、`POST /responses/compact`，完整生产路径为 `https://codex.sunny-string.cn/agents/codex/v1`。客户端仅发送自己的 `Authorization: Bearer ra_codex_...`；适配器逐次检查账号和设备状态，再替换为服务器私有的网关服务凭证。Cookie、代理授权和上游账号选择覆盖字段不会转发，凭证查询参数、浏览器 Origin 请求、重定向和其他模型路由会被拒绝。上游必须为明确配置的 HTTP loopback IP 根地址，例如 `http://127.0.0.1:8088`，不能使用公网地址、`localhost` DNS 名或任意客户端代理目标。

模型请求全实例最多同时 **2** 个，超出返回 `429`；请求体最多 **32 MiB**，单次请求最长 **30 分钟**，SSE 到达即刷新。停用账号会取消该账号的活动模型请求，包括未完成上传和阻塞输出，并关闭其设备/浏览器连接；其他账号不受影响。撤销个人 Token 对后续请求生效，撤销设备同时撤销其专属模型 Token。网关错误使用 `{error:{message,type}}`，上游凭证不会出现在错误响应中。

也可使用 Docker：

```sh
docker compose up --build -d
```

默认映射 `127.0.0.1:8080`，数据库保存在 volume 中。生产部署设置 `PUBLIC_URL` 并在宿主机配置 HTTPS 反向代理。镜像构建会同时打包 Windows / macOS / Linux Agent。

Compose 默认不启用模型网关。容器内的 `127.0.0.1` 不指向宿主机；连接宿主机已有的 loopback 网关时，在支持的 Linux 部署上使用经过审查的 host network 配置，并将服务监听地址限制为 loopback、私有 Token 文件只读挂载到正确路径。普通 bridge 网络下不能把 `host.docker.internal` 当作替代，因为适配器只允许字面 loopback IP。不要为所有容器或环境统一打开网关配置；没有同一网络命名空间中的固定 loopback 网关时，保持 `CODEX_GATEWAY_URL` 为空。

| 环境变量 | 默认值 | 用途 |
| --- | --- | --- |
| `PUBLIC_URL` | `http://localhost:8080` | 对外地址；除 loopback 外必须 HTTPS，支持 `/agents` 等子路径 |
| `LISTEN_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `DATABASE_PATH` | `data/remote-agent.db` | bbolt 数据库路径 |
| `DOWNLOADS_DIR` | `dist` | 已构建 Agent 及校验文件目录 |
| `ALLOW_SIGNUP` | 不再使用 | 兼容旧部署配置；公开注册始终返回 `403`，设为 `true` 也不会开放 |
| `CODEX_GATEWAY_URL` | 空 | 可选固定 HTTP loopback 根地址；空值禁用模型转发 |
| `CODEX_GATEWAY_TOKEN_FILE` | 空 | 启用网关时必填，服务启动时读取的私有服务凭证文件；不要放个人 Token |
| `CODEX_MODEL` | `gpt-6.1-sol` | 网页和已绑定设备配置返回的默认模型；保留已有本机模型选择 |

服务器使用 socket 来源 IP 做认证限流，不信任任意 `X-Forwarded-For`。反向代理部署时，后端会把同一代理视作一个来源；大规模部署前需在可信代理层做更细粒度限流。bbolt 只支持一个写入进程；不要给同一数据库启动多个实例。备份时停止服务器后复制数据库文件。

### 生产升级、管理员和回滚

生产账号 `string` 已在本次 139 升级中明确提升为管理员，原密码保持不变；其他部署同样需要明确授权。升级前先停止 `remote-agent.service` 并确认数据库不再被打开，在私有备份目录保留数据库、旧服务二进制与 `dist/`、环境文件、Nginx 配置和 systemd unit，保留原属主和权限。新版首次打开数据库（包括执行管理员命令）会在一个事务中迁移到稳定用户 ID + Unicode 名称索引；保留原密码哈希、登录记录、设备凭证和归属，不会自动给任何账号加管理员权限。名称归一化冲突会中止迁移，原数据库事务不会被部分提交。

```sh
sudo systemctl stop remote-agent.service
# 在此完成停止状态检查和上述备份，再使用已验证的新版本：
sudo env DATABASE_PATH=/var/lib/remote-agent/remote-agent.db /opt/remote-agent/current/remote-server admin grant --username string
```

若账号尚不存在，使用 `admin create --username string --password-stdin`，通过私有 stdin 输入密码。离线命令应有原数据库访问权限；不要意外创建服务无法读取的 root-only 新数据库。使用 DynamicUser 时保留 StateDirectory 的既有属主/权限安排。管理员后续在网页创建普通用户、重置密码或停用账号；不能停用或降级最后一名有效管理员。

按 [deploy/server.env.example](deploy/server.env.example) 配置新进程，保持现有网关和 `/openai/personal/v1` 路由、旧 Token、OAuth 状态及其他站点配置。网关服务凭证源文件应为 `root:root 0600`，由 [deploy/remote-agent.service](deploy/remote-agent.service) 的 `LoadCredential` 安全提供给 DynamicUser 进程；`CODEX_GATEWAY_TOKEN_FILE` 指向 `/run/credentials/remote-agent.service/codex-gateway.token`，不要直接指向服务不可读的 root-only 源文件，也不要将秘密放进 env 文件。未启用网关时去掉对应 `LoadCredential`，保留 URL 空值。

服务通常拒绝任何组或其他用户可读的凭证文件。Linux systemd 的运行时 `0440` 文件仅在环境变量声明的规范化、无符号链接的 `CREDENTIALS_DIRECTORY` 内获得例外，目录必须为 `/run/credentials/<unit>`。systemd 负责保护该凭证命名空间；应用校验路径和权限，不检查文件系统挂载。普通组可读文件仍会被拒绝；源文件继续保持 `root:root 0600`。

仅为目标站点和 `/agents` 安装 [deploy/nginx-codex.conf](deploy/nginx-codex.conf) 所需规则，保留 URI 前缀；模型位置单独允许 `32m`、关闭响应与请求缓冲、支持 1800 秒流式请求，普通账号接口保留较小限制。校验证书、`nginx -t`、环境文件和 systemd unit 后再 reload/restart，验收旧接口、新登录、Token、模型流、设备连接和停用取消。2026-10-19 已完成 139 r2 线上部署和账号/模型烟雾验收：保留 4 个账号、4 台设备，通过离线管理会话验证管理员会话和查询，Token 创建、模型目录、实际模型 SSE 的 `PROXY_OK`、撤销后 `401` 及公开注册 `403` 均通过；管理员原密码、既有网关路由、旧 Token 和 OAuth 状态保持不变，旧 `/openai/docs` 返回 `200`、匿名旧模型目录返回 `401`。本地浏览器登录和密码修改已实际测试。

r2 已完成 25 分钟下载超时及官方完整包布局修复。官方 Windows Codex 0.163 的完整 167080091 字节 ZIP 已通过 SHA-256、本地实际安装和原生 `--version`；Linux/macOS 官方包也已在 Windows 宿主上实际提取并逐文件比对 SHA-256。Codex 0.163 原生提供方的鉴权 SSE 测试上游验证和 Chrome 账号/Token/配置/手机流程重跑通过。Linux 139 已在隔离 HOME 中经真实公网地址完成完整安装，独立复查原生命令返回 `codex-cli 0.163.0`，根命令别名指向 `bin/codex`，未残留安装临时目录、归档或锁。Windows 也已在隔离 USERPROFILE 中通过真实公网元数据与下载入口完成完整 ZIP 校验及全包安装，主程序 341270320 字节，原生 `--version` 返回 `codex-cli 0.163.0`，没有修改真实 Codex 配置。核心实现已以 [5b8107f](https://github.com/string1225/remote-agent-cli/commit/5b8107f) 推送至 `origin/codex/unified-codex-platform`。

回滚也先停服务，成套恢复旧数据库快照、二进制、环境、Nginx 和 systemd 配置并校验后启动；旧二进制不能直接使用已迁移的新账号索引。回滚到快照会丢失快照之后的账号/Token/设备变更，应在维护窗口执行。网关本身及原有 OAuth 数据不属于此次账号库迁移，不要用新平台备份覆盖它们。

## WebSocket 中转

客户端通过出站 WSS 连接 139，网页使用同一站点的 WSS 入口。只需开放站点 HTTPS 端口；反向代理须转发 Upgrade 并关闭缓冲。20 秒心跳维持连接，客户端掉线后持续重试。服务器升级时应同步发布同版本客户端包；旧客户端返回升级提示，重新执行原安装命令保留配置和历史。

## CLI 与本机权限

安装位置：

- Windows MSI：`%ProgramFiles%\Remote Agent\remote-agent.exe`；独立 CLI/旧用户安装：`%USERPROFILE%\.remote-agent\bin\remote-agent.exe`
- macOS / Linux：`~/.remote-agent/bin/remote-agent`
- 配置：`~/.remote-agent/config.json`
- 会话数据库：`~/.remote-agent/history.db`，保存网页创建的会话、请求、回复和原生 CLI 会话 ID；不会上传到控制服务器或写入浏览器持久存储。
- 日志：`~/.remote-agent/agent.log`（仅运行状态，不记录任务正文；启动时对超过 10 MiB 的旧日志做一次轮换）

Windows MSI 使用系统服务自动启动，退出登录后继续运行；程序位于 `%ProgramFiles%\Remote Agent`，业务数据仍在原用户目录。服务配置与启停需要管理员权限，任务使用 LocalService 与独立服务 SID 的本机授权。macOS LaunchAgent 和尚未迁移的历史 Windows 用户安装仍在用户登录后启动。新版 Windows 不再提供新建计划任务的入口。电脑睡眠 / 关机时离线。

Linux 安装命令适用于使用 systemd 的 amd64 / arm64 系统，网页「添加设备」会分别提供 Windows、macOS、Linux 命令。请在目标用户的正常登录会话里运行 Linux 命令，勿在整条安装命令前加 sudo。安装器预先检查 systemd 用户会话，下载并校验 Linux 程序，绑定账号，然后启用 `systemctl --user` 的 `remote-agent.service`；服务异常退出后 5 秒重启，无限重试。单元文件位于 `${XDG_CONFIG_HOME:-~/.config}/systemd/user/remote-agent.service`。

Linux 安装器还会尝试 `loginctl enable-linger`，必要时使用无需密码的 `sudo -n loginctl enable-linger`，使当前用户的服务在退出 SSH 后继续运行并随系统开机启动。如果没有相应权限，会给出需管理员执行的具体命令，已启动的服务仍可在当前登录会话中使用。可通过 `loginctl show-user "$USER" -p Linger` 检查，结果为 `yes` 才具备退出登录和开机恢复能力。

Windows MSI 的常驻进程由 SCM 管理，失败 5 秒后重启；服务通过 Job Object 管理其 CLI 子进程。迁移成功后移除匹配原安装的旧计划任务，正式服务不创建或依赖计划任务。旧任务仅保留识别、停止、迁移和已安装旧程序的更新兼容。macOS 使用 LaunchAgent `KeepAlive` 自动重启。每份配置有独立的进程锁，防止重复进程互相挤掉连接；进程退出或崩溃时锁由操作系统释放。

设备离线时，先查看 `agent.log` 和系统服务状态。网络断开后 Agent 会持续重试，服务器恢复、电脑唤醒或网络恢复后自动重连。Windows MSI 下载新版安装包升级，或以管理员权限执行 `start` 恢复服务、`stop` 禁用服务；退出登录保持在线。其他平台可使用 `update` 和 `autostart install`。睡眠、关机期间离线；macOS 和未迁移的旧 Windows 用户安装退出登录会离线，Linux 启用 linger 后可继续运行。

```sh
remote-agent status
remote-agent version
remote-agent update --check
remote-agent update
remote-agent run
remote-agent autostart install
remote-agent autostart remove
```

本地 CLI / Codex 安装模式会将 `~/.remote-agent/bin` 加入用户 PATH：Windows MSI 将 `%ProgramFiles%\Remote Agent` 加入系统 PATH；macOS 更新 `.profile` / `.zshrc`，Linux 更新 `.profile` / `.bashrc`，同一行不会重复添加。启用自动 Codex 配置的设备安装也会执行此步骤；新终端才会继承持久修改。未刷新 PATH 时可使用上面的完整安装路径。

Windows MSI 使用 MSI 包升级，拒绝通过 EXE `update` 原地替换；`update --check` 仍可检查。以下自更新说明适用于独立 CLI、历史 Windows 用户安装和其他平台。

更新命令从配置中的服务器下载对应系统和 CPU 架构的安装包及 SHA-256 文件，通过校验和版本启动检查后替换程序。`update --check` 只比较校验值，不下载程序、不重启服务；重复更新相同安装包会直接结束。Windows / macOS / Linux 会重启正在运行且与当前程序、配置匹配的 Scheduled Task / LaunchAgent / 安装器创建的 systemd 用户服务；检测到 CLI 子任务时拒绝更新，应先结束任务。自启动未运行时保持原状态；手动启动的 Agent 或自定义服务需自行重启。替换或重启失败时尝试恢复旧程序，成功后旧版本保留为 `remote-agent.exe.previous` / `remote-agent.previous`。设备凭证、权限配置、项目服务、会话和 SaySo 数据均保留。

Windows PowerShell 可直接使用安装路径：

```powershell
& "$env:USERPROFILE\.remote-agent\bin\remote-agent.exe" update
```

macOS / Linux：

```sh
~/.remote-agent/bin/remote-agent update
```

Linux 服务管理与诊断：

```sh
systemctl --user status remote-agent.service
systemctl --user restart remote-agent.service
journalctl --user -u remote-agent.service -n 50 --no-pager
```

Linux 的 `autostart remove` 只停止并移除该安装对应的用户单元，保留绑定和会话，也不会关闭可能被其他用户服务使用的 linger。

**Windows 下载安装新版 MSI 升级；其他系统可以重新执行原安装命令升级。** 以下脚本重装说明仅适用于 macOS/Linux 及历史安装器。 安装脚本发现配置已存在时，使用刚下载并校验的新版 CLI 更新现有安装，不调用账号绑定接口，也不再次兑换安装令牌，因此原链接的绑定令牌已使用或过期也能升级。设备仍绑定原账号，权限、项目和会话记录保留；不同服务器的安装脚本会拒绝升级。普通设备重装会启动或重启匹配的自启动项，即使程序已经是最新版本；指定 `--no-start` / `-NoStart` 会保持服务停止，仅本地 CLI/Codex 模式重装则保留原启停状态。这个入口也支持没有 `update` 命令的旧版 CLI。成功后可继续使用 `remote-agent update`；已有配置损坏或 CLI 文件丢失时应先修复本机安装。

构建发布安装包时可运行 `go run ./tools/build -version <Git提交号>`，尤其是从源码归档构建时，确保 `version` 命令显示发布版本。

手动绑定可指定根目录和写入授权：

```sh
printf '%s' 'ONE_TIME_TOKEN' | remote-agent enroll \
  --server https://agents.example.com --token-stdin \
  --roots /Users/you/projects --allow-write
```

`--roots` 多个目录使用操作系统的路径列表分隔符（Windows `;`，macOS / Linux `:`）。默认根目录是当前用户的 home。已绑定设备可先停止 Agent，编辑本地配置的 `allowedRoots` / `allowWrite`，再启动；网页无权修改这两个字段。不要把凭证文件上传或分享。Windows 服务配置为原用户目录添加服务 SID ACL，macOS / Linux 使用用户私有目录和 `0600` 配置文件。

- Codex 默认 `exec --json --sandbox read-only`，写入授权后使用 `workspace-write`。非交互审批策略为 `never`，受限操作失败，不自动越权。
- Qoder 默认 `--print --output-format stream-json --permission-mode plan`，写入授权后使用 `accept_edits`。需要交互批准的操作不在本版本提供网页审批流。
- 各提供方 CLI 均为可选。可继续使用本机已有登录；启用平台提供方的 Codex 使用本机私有 TOML 中的个人 Token，经平台适配器访问现有模型网关。Qoder 等其他提供方仍使用其本机凭证；配置工具不读取、上传或删除已有 OAuth 文件。
- 检测 `codex`、`qodercli` 或 `qoder` 的帮助信息。Qoder 桌面编辑器的同名启动器不会被当成 CLI。可设置本机 `RA_CODEX_PATH` / `RA_QODER_PATH` 为 CLI 路径；自启动需要能继承这些环境变量。
- 若安装 CLI 后未检测到，重新启动 Agent。macOS LaunchAgent / Linux systemd 用户服务保存安装时的 PATH，可重新执行 `autostart install` 更新。
- `allowedRoots` 检查规范化的真实目录及符号链接，限制**启动工作目录**；它不是独立 OS 沙箱，不能限制 CLI/MCP/本地配置的一切文件访问。提供方各自的权限机制仍然生效。不要把不受信任项目当成隔离环境。

接口依据：[Codex 非交互模式](https://learn.chatgpt.com/docs/non-interactive-mode)、[Qoder 脚本模式](https://docs.qoder.com/cli/run-in-scripts)、[Qoder Plan](https://docs.qoder.com/cli/plan-mode)。

## SaySo：讨论、录音与 Action

选择客户端的项目后，点击会话列表置顶的 **♫ SaySo**，右侧显示该项目的讨论、录音和 Action。旧讨论及录音保留，可在历史讨论中选择；确认派发 Action 会在同项目创建新 Web 会话工作。无需在目标电脑安装 Node.js 或另开 SaySo 服务，Codex/Qoder 都不是录音和本地规则功能的前提。

- 项目与讨论会话管理，默认暂停；支持继续录音、暂停、每 10 分钟切段、逐段播放和文字补充。
- 浏览器提供实时语音转写；不可用时仍保存录音。浏览器的识别服务可能联网处理音频，此行为不属于 Remote Agent 的中转通道。
- 录音期间每 2 分钟整理最近 5 分钟讨论，暂停时处理剩余内容；可手动完整分析。分析模型默认跟随本机 Codex，也可选其他账号支持的模型，或完全使用本地关键词规则。
- Codex 分析会创建或补充已有 Action，保存分析输入、过程、结果及讨论原话；修改后的需求回到待完善，避免沿用旧的 Ready 状态。
- 编辑标题、详情、优先级，确认“讨论完善”后派发给 Codex。执行结果实时回传，可查看或停止任务；始终遵守 Agent 本地 `allowWrite`，默认只读。分析始终只读。
- 手机全屏展示讨论或 Action，顶部切换项目/会话；录音、文字、Action 和执行输出全部保存在目标电脑的 `~/.remote-agent/sayso/`，刷新或换浏览器后可恢复。

数据文件为 `sayso/state.json`（原子替换），录音文件在 `sayso/recordings/`。录音经已认证的 139 WebSocket 分片中转至客户端保存。每片 16 KB，每段最多 64 MiB；一个连接同时上传一段，上传期间不完整的文件在断连或重启后清理。上传失败时网页提供录音备份下载，请在关闭页面前保存；浏览器崩溃或关闭前尚未停止的录音不能保证恢复。

讨论快照超过单个 SCTP 消息大小时自动分帧。当前每台设备最多保存 500 个讨论会话、3000 条 Action、1000 条近期执行记录；单会话最多 10000 条发言、1000 段录音，状态文件最多 24 MiB。删除讨论会话只删除索引，目标电脑的录音文件保留供本地恢复。设备断线、撤销或 30 分钟连接租约到期会停止该连接的分析/执行并暂停讨论状态，需要重新连接后继续。

界面源代码及构建方法见 [sayso-ui/README.md](sayso-ui/README.md)。原 `say-so` 项目保持独立，其已有数据不会自动迁移。

## 撤销与卸载

网页「撤销设备」立即禁用设备凭证、未使用的绑定令牌和设备专属模型 Token，并关闭相关会话。退出登录只关闭当前浏览器登录态所建立的会话；修改用户名/密码关闭该账号所有浏览器登录，停用账号还关闭设备连接并取消活动模型请求。控制连接断开时 Agent 取消该连接下的任务；静默网络故障最多需要约 65 秒被心跳检测。`remote-agent stop` 仅停止远程服务，不卸载持久 CLI 或取消独立个人 Token。

Windows MSI 从系统应用列表卸载，服务和程序被移除，配置、项目、会话和录音保留。其他安装可在目标电脑运行 `remote-agent autostart remove`，再按需手动删除 `~/.remote-agent`；删除业务数据需要用户明确选择。Windows 请等进程停止再删除旧的独立可执行文件。

## 开发与验证

```sh
go test ./...
go vet ./...
node --check web/app.js     # 只验证工作台原生 JS 语法
node --check web/platform.js # 账号、Token 和 Codex 接入界面
(cd sayso-ui && npm ci && npm run build) # 修改 SaySo 界面时需要（PowerShell 可进入目录后运行）
go run ./tools/build       # 交叉编译所有支持的平台
```

测试覆盖账号迁移、Unicode 唯一性/搜索、管理员权限与并发保护、密码和登录态失效、Token 归属/配置失败保留/撤销、网关固定凭证替换、SSE/部分上传取消及并发限额、本机 TOML 备份/profile/权限和安装校验，还包含绑定令牌并发兑换、CSRF、工作目录边界、UTF-8/stdin 与真实 WebSocket 中转数据传输及撤销中止。集成测试使用模拟 CLI/上游，不调用付费模型或改动真实项目。GitHub Actions 在 Linux / Windows / macOS 上运行测试，另执行 Linux race detector 和交叉编译；本地 CGO 关闭且无 C 编译器时无法运行 race detector。

`20261019-codex-platform-r2` 已通过完整 Go tests/vet、两份 JavaScript 语法检查和六平台构建；官方 Windows 0.163 本地与公网实际安装、Unix 官方完整包提取后的逐文件哈希、Linux 139 公网原生完整安装、新版原生提供方鉴权 SSE 测试上游和 Chrome 流程复验均通过。

## 当前范围

- 单实例服务端和本地数据库；尚未实现多节点路由、组织共享、密码找回或 MFA。
- 每个请求开启一次 CLI 进程，同一会话后续请求使用原生会话 ID 恢复上下文。尚未实现交互终端、文件上传、网页审批或导入此前直接在 CLI 中创建的历史会话。
- 远程任务和会话历史保存在目标电脑，服务端没有中央任务历史；配置平台提供方时模型请求/响应会经服务器转发，但适配器不写入账号数据库或应用日志。关闭远程连接会取消其任务，已有会话输出保存在目标电脑。
- 设备元数据、服务名称及项目路径会随心跳保存在服务器。服务端仍是登录、信令和网页代码的信任根，不宣称能抵抗已被攻陷的控制服务器。
- 每台设备最多 8 个连接 / 1 个任务；每次任务最长 20 分钟，连接最长 30 分钟；单次提示最多 32 KiB，stdout/stderr 各最多 16 MiB，会话历史按页读取；SaySo 使用前述独立数据限额。
- 已提供跨平台安装和构建，但真正跨公网长连接、不同 Windows 策略以及 macOS 自启动仍需在目标部署环境验收。分发二进制尚未做代码签名 / Apple 公证。
- macOS Codex CLI 原生运行，以及 Windows/macOS Codex App 安装、升级和重启生命周期尚未在真实目标硬件上验收；macOS 官方包已完成提取和逐文件哈希校验。

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

子路径部署时必须把完整地址（包含 `/agents`）设为 `PUBLIC_URL`，代理不要移除此前缀。CLI 安装、Cookie 路径、API 与 WebSocket会自动跟随此前缀。`/agents/healthz` 可用于健康检查。客户端和网页均使用当前站点，不需要 TURN。
