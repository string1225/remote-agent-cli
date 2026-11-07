# 2026-11-07 Web 代理直连隔离与 Git 联动验收

代码提交 `24e1bba30a2ff2729779582b600817fe75d06961`，客户端版本 `20261107-proxy-direct-git`，发布序号 `2026110702`。使用已提交源码归档构建并发布六平台 runtime；本机 Windows 与 139 Linux 客户端均更新完成，现有服务正常运行。原有未提交改动未纳入发布。

## 根因与行为

旧版本虽已在本机直连非目标网址，却把这些 HTTP/CONNECT 请求也计入本地每设备 32 条的中转额度。测试复现：32 条普通连接占用后，新的普通 HTTP/CONNECT 均返回 429；中转额度被占满时也会错误阻断普通网站。

修复保留本机 HTTP 代理架构：只有既有 GitHub/附件、ChatGPT、OpenAI、Google 主域和子域的 80/443 经 139，其他目标由本机直接连接，不经过 139，也不占用中转额度。中转目标仍受本地和 139 的配额保护，失败不回退直连。普通连接保留期限及关闭清理。未使用 PAC、扩大域名列表或更改 TLS 验证。

`web-proxy on/off` 同步当前用户的全局 Git `http.proxy` 和实际监听端口，保留其余 Git 配置。原值私有备份，关闭恢复，外部后续修改保留。包含文件或 URL 专用配置导致有效 GitHub 代理未生效时，返回错误并回滚；重复开关和失败换端口的恢复点均有回归测试。命令 `web-proxy run` 仍仅临时运行，不修改系统/Git 设置。

## 验证

- 全量 `go test ./...`、`go vet ./...`、gofmt、六平台构建通过；独立复核发现的 Git 恢复点和有效配置问题已修复并补测。
- 在真实安装的 Windows 客户端上开启代理，保持 **64 条普通 CONNECT 连接**，再通过 WinINET 当前用户系统代理访问百度、必应、微软 robots.txt、Google Play，全部 HTTP 200。
- 同时运行普通 `git clone --depth 1 https://github.com/octocat/Hello-World.git`，未指定代理参数，成功。完整 clone 日志仅保存在私有验收目录。
- 关闭后，Windows 原代理设置及原 Git 代理值恢复；再开启、重复开启均使用同一实际端口。其他全局 Git 键值与私有基线一致。
- 原有 Windows 25 个、Linux 8 个会话的 ID、原生 ID、历史摘要保持；本机 4 个、139 2 个设备配置/SaySo 文件摘要保持。两端经认证目录确认在线，CLI 均确认已是当前发布版本，remote-agent skill 已备份同步。
- 139 更新通道的六份清单、下载包及权限检查通过；中转服务 healthz 正常，后端代码未改动。

Windows 最终状态为已开启，监听 `127.0.0.1:17891`，系统和 Git 代理均由本客户端管理。可用 `remote-agent web-proxy off` 恢复原设置。

本轮未执行 MSI 安装、整机冷启动、真实机器人收发、macOS/ARM64 原生运行验收；未升级模型 CLI 或更改模型凭据。

## 备份与回滚

- 更新通道：`/var/backups/remote-agent-clients/runtime-24e1bba30a2ff2729779582b600817fe75d06961-1794064284/rollback.sh`，root 执行；保留上一通道，不静默降低客户端序号。
- 139 客户端：`/home/junte/.remote-agent/backups/20261107-proxy-direct-git/`，含原程序、配置/历史/SaySo 快照及回滚脚本。
- 本机私有验收目录：`C:\Users\junte\AppData\Local\Temp\remote-agent-direct-git-20261107-42180bfb7bac4caf841f928b28b7b9fc\acceptance\`，含两次更新前快照、原 Git 配置、并发和开关测试结果；访问限制为本用户和 SYSTEM。
- 当前代理恢复点：`~/.remote-agent/web-proxy-system.json` 与 `~/.remote-agent/web-proxy-git.json`。需要回退到旧客户端时先关闭代理以恢复 Git 设置，再用现有 Windows updater 显式回滚，避免旧版本不认识新的 Git 恢复点。
