# 临时文件分享

已实现客户端上传到 139 后获取公开下载链接。CLI：

```powershell
remote-agent files upload --path "C:\path\remote-agent.apk"
```

可使用 `--config PATH` 选择已绑定客户端配置。MCP 提供 `upload_file`，参数 `path` 必须是当前机器上普通文件的绝对路径。两种入口均返回 `id/name/size/sha256/url/expiresAt`，不发送消息；Agent 可在回复或已经授权的消息里附上链接。该能力需要升级服务端和执行上传的客户端。

上传限 100 MiB，以设备 Bearer 凭据调用 `POST /api/agent/files?id=<deviceId>`，二进制 `application/octet-stream` 请求体，`X-File-Name` 为 URL query 编码后的文件名。凭据只发给所属站点，不跟随重定向。最多四个并发上传，流式计算 SHA-256，失败移除未完成文件；保存完整内容及元数据后才发布链接。

`GET/HEAD /files/<256-bit random ID>` 无需登录，支持 Range 续传和中文附件名称，强制下载并禁止缓存。文件名和原路径不在下载 URL 中。知道链接的人均可下载，直到下次北京时间（UTC+8）凌晨 4 点；这不是从上传时起保留 24 小时。

服务端 `SHARED_FILES_DIR` 默认为数据库旁的 `shared-files`，生产为 `/var/lib/remote-agent/shared-files`。每代目录以失效时间命名，下载只读取当前有效代；凌晨 4 点服务内置清理器删除所有已失效代（包括未完成上传），启动时补清理。无需 cron 或新增系统服务；不会清理客户端附件、项目、录音、数据库或发布 APK。若服务在清理时停机，下载仍在截止时间失效，文件在下次启动后删除。凌晨 4 点前开始、截止后完成的上传明确失败，需重试。

Nginx 必须使用仓库 `deploy/nginx-codex.conf` 中 `/agents/api/agent/files` 的专用配置，允许 100 MiB 和 15 分钟上传，关闭请求缓冲；其余 API 维持原限制。生产部署保留现有 Nginx 配置，只增加此 location，先 `nginx -t` 后 reload。

本地测试覆盖真实 HTTP 上传下载、跨站匿名下载、Range、中文文件名、SHA-256、空文件、认证失败、文件名穿越、长度及 chunked 限制、凌晨 4 点/跨年边界、到期立即拒绝与启动补清理。真实 139 上传和公网下载需部署后验证。

## 部署并上传 APK

在能使用 139 SSH 私钥的桌面用户终端运行（不在聊天里传递私钥或配置内容）：

```powershell
./tools/deploy-139.ps1 -IdentityFile "$env:USERPROFILE/.ssh/id_ed25519_139_224_12_141" -UploadFile ./dist/android/remote-agent.apk -ConfigPath "$env:USERPROFILE/.remote-agent/config.json"
```

脚本从当前已提交的 HEAD 导出干净源码，构建 Linux 服务器及本机独立上传 CLI，校验部署包后更新 139 服务和内嵌网页，加入 Nginx 上传专用配置。现有签名客户端更新包沿用上一版本，不替换其内容或签名；完整客户端发布仍需现有签名/MSI 流程，现有客户端需升级后才能获得新 MCP 工具。`-BuildOnly` 只准备产物。

服务停止后，脚本备份完整状态目录（包括数据库及密钥）、配置、Nginx 和原发布路径，保留旧程序；健康检查失败时恢复原程序和数据，并保存失败版本数据用于检查。成功后调用新 CLI 上传指定 APK，打印公网下载链接及失效时间。此操作会短暂重启服务；需要可用的 SSH、免交互 sudo 以及客户端设备配置，GitHub 推送凭据与 SSH 部署权限互相独立。
