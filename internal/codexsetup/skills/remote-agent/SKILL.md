---
name: remote-agent
description: 使用本机 Remote Agent 客户端查询设备与项目、创建远程任务、通知已绑定的飞书或钉钉会话，或管理客户端、Codex CLI 和按域名分流的 Web 代理时使用。
---

# Remote Agent 客户端

Remote Agent 是本机常驻客户端，属于一个平台账号，管理多个“工作目录 + Codex/Qoder 执行器”项目。工作台、Web/IM 会话和 SaySo 共享项目身份。这里的 `code` 指 Qoder。

先执行 `remote-agent version`、`remote-agent status` 和 `remote-agent web-proxy status` 核对当前客户端。配置默认在 `~/.remote-agent/config.json`；自定义配置用 `--config PATH`。status 返回的项目、工具检测与代理状态才是当前能力，不要把安装了本技能当作已绑定、在线或已配置渠道的证据。不要读取、输出或复制配置中的 credential、模型 Token 或机器人 Secret。

## 任务与跨设备协作

工作台启动的 Codex/Qoder 进程会注入 `remote_agent` MCP。先检查当前会话的工具列表和 MCP 初始化返回的当前客户端、项目、来源会话，再按需使用：

| 工具 | 能力与参数 |
| --- | --- |
| `list_clients_projects` | 无参数，查询同账号未撤销的客户端、在线状态和项目 ID。跨设备操作先查询，使用返回的 ID。 |
| `start_web_session` | `prompt` 必填；可传 `clientId`、`projectId`、`title`、`notify`。在目标项目新建独立 Web 会话并排队执行。返回 session/task ID 仅表示已接收，不能宣称执行完成。 |
| `notify_im_session` | `message` 必填；可传 `clientId`、`projectId`、`sessionId`、`final`。向已绑定的飞书/钉钉会话发送消息，默认继承任务来源会话。仅在用户要求通知或已授权的任务流程中调用；`queued:true` 不代表平台已送达。 |

IM 来源任务默认开启完成通知。创建后立即简短确认已排队，不等待完成，不附网页会话地址。最终回答用中文直接给出实质结论、关键结果、验证/提交情况及待处理事项；系统提取最终文本并回到任务保存的原始私聊/群聊，不用网页地址替代结果。`notify_im_session` 通常只发进度；如确需主动发送最终实质结果，传 `final:true`，成功入队到继承来源后才会抑制重复自动结果；发往其他会话、发送失败或普通进度均不抑制。旧运行客户端需升级后才支持这一结果回传与 final 标记。

单独启动的本机 Codex 不一定有 MCP 注入。可使用 `remote-agent mcp --config PATH` 的 JSON-RPC stdio 接口；先 `initialize`、`tools/list`，再 `tools/call`，工具名和参数与上表一致。该进程从本机配置读取凭证。若需要当前项目/会话，使用真实 ID 传 `--project`、`--session`；不得编造来源会话。不要仅为查询长期改写用户 MCP 配置。

客户端在本机执行 CLI，设备同一时间执行一个模型任务，后台任务可排队。网页关闭不会取消已接收的后台任务；服务停止/更新可能中断它们。保留返回的任务和原生会话 ID，检查状态后再决定重试，避免重复执行。

飞书/钉钉长连接由客户端直接建立，凭证仅存本机；每项目最多一个渠道，私聊/群聊自动映射独立持久会话。SaySo 提供讨论、录音、分析和 Action；Action 经确认后派生同项目 Web 会话。录音需要用户主动开始及麦克风授权，不能通过本技能自动录音。

## Web 代理与 139 分流

Web 代理**默认关闭**，与模型代理分别配置。GitHub、ChatGPT、OpenAI 和 Google 经 139 中转，其他请求由客户端直接放行。仅在用户要求或已授权的流程需要时开启：

```text
remote-agent web-proxy status
remote-agent web-proxy on
remote-agent web-proxy off
```

`on` 默认监听 `127.0.0.1:17891`，可用 `--listen 127.0.0.1:PORT` 改端口。它先等待健康检查核对本机客户端，再验证 GitHub 中转 CONNECT；成功后自动开启 Windows 当前用户系统代理，无需管理员权限。未就绪/中转失败会报错且不改系统代理，先检查或升级客户端服务。`off` 先恢复系统代理原设置，再关闭监听；原设置私有备份至 `~/.remote-agent/web-proxy-system.json`，外部后来修改的代理会保留。`status` 区分配置、监听就绪、系统代理启用及归属。macOS/Linux 需在应用中选择代理。临时前台运行用 `remote-agent web-proxy run`，Ctrl+C 停止，此模式不自动改系统代理。

开启后按单次请求显式选择代理，例如：

```text
git -c http.proxy=http://127.0.0.1:17891 ls-remote https://github.com/OWNER/REPO.git
curl --proxy http://127.0.0.1:17891 https://api.github.com/
curl --proxy http://127.0.0.1:17891 https://chatgpt.com/
```

支持普通 HTTP 转发和 HTTPS CONNECT：`github.com`、`githubusercontent.com`、`githubassets.com`、`chatgpt.com`、`openai.com`、`google.com` 及其子域的 80/443 经已绑定站点的鉴权 WebSocket 到 139，其他目标直接放行。重定向后按新目标重新选择路由。中转失败报错，不回退直连；其他目标不依赖 139 在线。139 遇 Fake-IP DNS 会使用固定公网解析器的 HTTPS 查询并继续拒绝非公网地址。CONNECT 中目标 TLS/WSS 透传，由应用正常校验证书，不要关闭证书验证。旧 `.github-proxy.json` 开关保留，新 `.web-proxy.json` 优先。按用户要求使用 `on/off` 管理系统代理，不额外改全局 Git、环境变量或证书验证。

## 本机维护

- `remote-agent start/stop/status`：客户端启停与绑定/项目/工具状态；Windows 正式版本使用 MSI 安装的 Windows Service，开机运行且退出登录后持续在线，不创建计划任务。
- `remote-agent service status`：检查 Windows 服务配置与状态。启动前先确认绑定，避免重复实例。
- `remote-agent update --check`：检查发布版本；用户要求升级时再使用 `remote-agent update`，先处理运行任务和录音。
- `remote-agent codex install/update/setup`：安装/更新官方 CLI，或备份并配置所属账号的模型入口，同时安装本技能。`app install/upgrade` 在 Windows 打开商店页面，仍需用户完成商店操作。
- 设备绑定/恢复从已登录工作台复制命令；保留原项目、会话、录音和备份，不打印 Secret。不要自行撤销、重建或强制替换设备。

平台远程业务经 139 中转，无 P2P/WebRTC 回退；Web 代理按上述域名规则分流。当前跨设备接口仅有上述目录查询、派生任务和通知工具，不能声称存在独立通用远程 shell API。真实 MSI/服务生命周期、机器人收发和模型授权须在目标环境验证，单元测试通过不能替代这些验收。
