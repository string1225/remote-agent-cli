# Web 代理与 Codex 客户端技能

## 开关与使用

Web 代理默认关闭。升级或读取旧配置不会自动启用；原来明确开启过 GitHub 代理的配置会继续沿用。绑定后的客户端提供以下命令；自定义配置加 `--config PATH`：

```text
remote-agent web-proxy status
remote-agent web-proxy on
remote-agent web-proxy off
```

`on` 保存监听地址，默认 `127.0.0.1:17891`；可用 `--listen 127.0.0.1:PORT` 改端口。随后等待本机健康检查核对客户端身份，并检查经 139 的 GitHub CONNECT。检查通过后，Windows 自动启用当前用户的系统代理；失败时返回明确错误，系统代理保持原设置。常驻客户端未运行或版本过旧时须先升级并启动服务。

Windows 系统代理（包含 PAC、自动检测、服务器及例外）先备份至 `~/.remote-agent/web-proxy-system.json`，仅当前用户与 SYSTEM 可读。`off` 先恢复原系统设置，再清除代理开关，避免浏览器仍指向已关闭端口；其他程序/用户后来修改的代理设置予以保留。恢复快照另存 `.restored`。重复 `on` 不覆盖原始恢复点。此联动不需要管理员权限，不新建服务或计划任务。

常驻客户端每秒读取开关；监听失败会重试，切换时停止原监听与活动隧道，不重启远程任务服务。`status` 分别返回配置启用、经身份核对的监听就绪、Windows 系统代理启用及是否指向本客户端。macOS/Linux 仍需在应用中选择代理。端口占用和中转失败写入脱敏日志，不记录凭证。

开关单独存于配置旁的 `config.json.web-proxy.json`，不改设备配置，避免项目设置的旧快照覆盖代理开关。更新前保留精确备份 `config.json.web-proxy.json.backup-<时间>`，保留未知代理设置。写入失败保留原文件和备份；恢复前停止客户端及其他配置操作，再用备份覆盖代理设置后重启。该文件不含凭证，原设备、渠道、项目、会话和录音均保留。

没有新设置文件时读取旧 `config.json.github-proxy.json`。首次修改从旧文件复制设置并备份，旧文件保留不动；新文件始终优先，因此 `off` 不会被旧 `on` 重新启用。旧客户端的 `/api/agent/github-proxy` 入口仍由服务器保留供升级过渡。

前台临时运行使用 `remote-agent web-proxy run [--listen 127.0.0.1:PORT]`，Ctrl+C 停止。此模式不改变保存的开关或系统代理。`on` 在常驻客户端未就绪时只保留开关并报错，不接管系统网络；按原安装方式启动客户端后重试。

开启后供单次 Git/curl 请求使用：

```text
git -c http.proxy=http://127.0.0.1:17891 ls-remote https://github.com/OWNER/REPO.git
git -c http.proxy=http://127.0.0.1:17891 clone https://github.com/OWNER/REPO.git
curl --proxy http://127.0.0.1:17891 https://api.github.com/
curl --proxy http://127.0.0.1:17891 https://chatgpt.com/
curl --proxy http://127.0.0.1:17891 https://example.com/
```

Windows 的 `on/off` 管理当前用户系统代理；不改全局 Git 或环境变量。其他平台可在应用中手工选择代理。支持普通 HTTP 请求转发和 HTTPS CONNECT，逐次按请求目标路由；重定向后的新请求重新选择路线。

| 目标 | 路线 |
| --- | --- |
| `github.com`、`githubusercontent.com`、`githubassets.com` 及其子域 | 经 139，覆盖 API、raw、codeload、发布附件和 LFS |
| `chatgpt.com` 及 `*.chatgpt.com` | 经 139，包含 CONNECT 内的 HTTPS/WSS |
| `openai.com`、`google.com` 及其子域 | 经 139，包含 API 和网页路径 |
| 其他目标 | 客户端直接连接，正常放行 HTTP/HTTPS；不要求 139 在线 |

139 中转仅接受规则内域名的 80/443；中转目标其他端口明确报错，不回退直连。其他站点的 CONNECT 允许有效目标端口。该服务是显式 HTTP 代理，不会接管未指定代理的应用流量。

## 传输与权限

中转流量：本机应用 → 回环代理 → 已绑定 139 站点的 `/api/agent/web-proxy?id=…&target=…` WebSocket → 规则内目标 TCP 80/443。支持 `/agents` 等站点子路径，设备凭证只放中转 Authorization 请求头，不转发本机 Proxy-Authorization。规则外流量从客户端直接连接，不发送设备凭证，也不走环境代理，避免递归代理。

CONNECT 中应用与目标的 TLS 保持原样，正常校验证书，139 不替换证书或解密 TLS。普通 HTTP 无目标 TLS，可信 139 可见中转内容。设备到公网站点使用 WSS。服务端先鉴权，再验证域名和 DNS 公网地址，直接连接已检查 IP，拒绝内网、链路本地及特殊地址；这些服务端限制不用于规则外的客户端直连。设备撤销、凭证轮换或账号停用在新建中转连接时立即生效，已有中转隧道最长约 20 秒内重新核对并关闭。

若 139 本机 DNS 失败或返回 `198.18.0.0/15` 等非公网地址，使用固定公网 IP 的 Cloudflare/Google DNS-over-HTTPS 查询真实 IPv4，并校验 HTTPS 证书、禁用环境代理和重定向，成功结果缓存 30 秒。回退结果仍需通过公网地址校验；绝不把伪 IP/内网地址直接加入白名单。这样兼容 139 上已有透明代理的 Fake-IP DNS，同时保留 SSRF 防护。

每设备最多 32 条连接，全服务器最多 256 条，额度包含尚未完成的连接建立；单隧道最长 30 分钟。二进制帧上限 32 KiB，有心跳、读写超时及双向关闭；关闭客户端/服务器会取消活动隧道。139 的原 WebSocket Nginx 配置已覆盖该路由，无需开放额外公网端口。

## Codex 自动安装技能

`remote-agent codex install/update` 成功安装官方 CLI 后安装 `remote-agent` skill；`codex setup` 和设备绑定的 Codex 配置流程也安装它，因此已安装 CLI 可通过 setup 补齐。位置为 `$CODEX_HOME/skills/remote-agent/SKILL.md`，未设置时为 `~/.codex/skills/remote-agent/SKILL.md`，新 Codex 会话中可发现。安装失败返回具体错误，CLI 本体保留，可重试 setup。

技能源为 [SKILL.md](../internal/codexsetup/skills/remote-agent/SKILL.md)，编译进客户端，介绍当前设备查询、项目/会话关系、三个 MCP 工具、任务与通知语义、客户端维护及默认关闭的 Web 代理。技能只保存说明，不包含真实身份或凭证，也不把安装视作联网、绑定、渠道或 MCP 已可用的证明。

安装相同内容不重复写入。更新前保留原同名技能文件 `SKILL.md.backup-<时间>`；拒绝链接、路径冲突及并发安装，失败保留原文件。其他技能和 Codex 配置不会由技能安装器改动。

## 验收边界

模拟上游测试覆盖 HTTPS 证书验证、双向大数据、CONNECT 缓冲字节、HTTP/HTTPS 直连、ChatGPT 中转、跨路由重定向、中转失败不回退、鉴权、目标限制、连接额度、关闭清理、开关和旧配置恢复。真实 139 需部署同版服务器和客户端后验证 Git clone/push、发布附件/LFS、ChatGPT HTTPS/WSS、直连、设备撤销和 Windows Service 开关；单元测试不能替代这些现场验收。
