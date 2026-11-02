# 默认主机、账号待办与 Android

## 使用方式

在客户端的更多菜单选择“设为默认主机”，或打开左下角“待办任务”选择主机。每个账号只有一个默认主机，必须是本账号已经绑定且在线的 Windows/macOS/Linux 客户端。服务器只保存主机和清单代次，待办正文、会话关联和回复记录保存在默认主机的 `history.db`。切换项目或客户端不改变清单，其他客户端通过现有 139 持久 WebSocket 中转查询和编辑。

待办支持新建、搜索、分页、编辑、完成、取消和重新打开。人工新建时可以关联当前会话，也可只保存个人事项。关联会话的待办显示来源客户端、项目与 Session ID；“回复”先在默认主机保存内容，再向原客户端、原项目、原 Web/IM 会话提交一轮消息，复用原生上下文。同一会话继续排队执行，其他会话不受影响；回复 IM 待办的结果仍投递到原 IM 会话。完成、取消只改变待办状态。

回复显示“等待送达确认”时可点“重试回复”。回复 ID 与待办 ID 一起形成固定任务 ID，断线、确认丢失、迁移后再次点击都不会重复执行。默认主机或来源客户端离线时明确报告失败；未确认回复保留，不宣称已执行。编辑使用版本号，旧版本不能覆盖新修改。回复草稿按站点、账号、待办隔离保存在浏览器/WebView 中。

也可以取消未确认回复，停止再次投递；确认丢失时来源会话可能已经接收，因此需到来源会话检查执行状态。取消待办不自动停止已接收的会话任务。

Agent MCP 新增 `list_todos`、`publish_todo` 和 `update_todo`。初始化说明要求 Agent 在需要人的判断、选择或补充信息时，用 `publish_todo` 写明背景、问题、选项和建议，然后结束本轮等待回复。来源使用当前真实客户端、项目和会话，不由模型自行选择他人的会话。人工回复到达后在原会话继续。

## 更换主机与恢复

更换主机需要新旧主机都在线。服务端按账号串行处理待办操作和迁移，将包括全部状态待办及回复去重记录的分页快照写入新主机的独立代次，全部成功后原子切换主机指向。复制、鉴权或落盘失败时原主机继续生效；未完成的目标代次不对外可见，可直接重新选择迁移。不同账号互不阻塞。

快照前会持久冻结原代次，阻止此前超时或断线的延迟写入。服务端先保存迁移标记，失败时解冻；服务端重启后下一次操作会根据标记恢复原主机。客户端记录已回滚的迁移 ID，延迟到达的旧冻结请求不能再次锁住清单。成功迁移后的旧代次保持只读备份。

原主机完整数据和原有会话、录音不删除。重新切回曾使用过的主机也会新建代次，避免复活旧待办。默认主机必须先成功迁移才能删除。备份时保留默认主机配置、完整 `history.db` 和服务器数据库/密钥；旧清单代次在数据库 `account_todos/<账号 ID>/<代次>` 中，仅供恢复，不自动合并。原主机数据永久损坏时，需要从备份恢复该主机数据再迁移；当前没有自动跨主机容灾。

接口：网页登录使用 `GET/PUT /api/default-host` 与 `GET/POST /api/todos`；客户端凭本机设备凭据使用 `GET/POST /api/agent/todos`。POST 的 `operation` 为 `list/get/create/update/complete/cancel/reopen/reply`，创建和回复带稳定 `requestId`，修改和首次回复带最新 `version`。内部的存储、导入和会话投递只允许服务端标记的账号工具连接调用，普通网页 WebSocket 不能绕过默认主机路由。

## Android 应用

`android/` 是原生 Java Android 宿主（Android 8.0/API 26 起），通过 HTTPS 加载已有响应式工作台，支持登录、会话/历史/附件、在已有客户端创建项目、待办和 SaySo 麦克风授权。手机不注册为执行客户端，添加客户端入口隐藏。站点默认 `https://codex.sunny-string.cn/agents/`，可在原生“设置 / 更新”中修改 HTTPS 站点；更换站点清除原站点登录状态和缓存。服务端和目标电脑需要升级到本提交才能使用新增能力。

更新默认在启动时自动检查，有新版则下载校验并提示安装；设置中可关闭自动检查或手动检查。更新清单和 APK 经过 `https://gh-proxy.com/`。APK 还必须通过 SHA-256、包名、版本号及当前已安装签名证书核验；系统仍要求用户确认安装，首次需允许本应用安装更新。普通 Android 应用不静默安装。更新保留应用数据，退出手机不影响电脑任务。调试 APK 的签名与正式版不同，不能用来验收正式版覆盖升级。

## Actions 发布

`Android APK Release` 工作流在 `main` 上的 `android/**`、`web/**`、`tools/android/**` 或自身修改时自动运行，也支持手动触发。锁定 Gradle 9.5.0（含分发包校验）和 Android Gradle Plugin 9.3.0；CI 使用 JDK 17、SDK 36，执行 `lintRelease assembleRelease` 并核验 APK 签名。

首次配置持久签名身份（在仓库根目录运行，需 JDK `keytool` 与有仓库管理权限的 Git/GH 授权）：

```powershell
$env:JAVA_HOME = 'C:\Program Files\Android\Android Studio\jbr'
node tools/android/configure-signing.cjs --upload
```

工具在被 Git 忽略的 `.local/android-signing/` 生成密钥及密码，上传四个 Actions secrets：`ANDROID_KEYSTORE_BASE64`、`ANDROID_KEYSTORE_PASSWORD`、`ANDROID_KEY_ALIAS`、`ANDROID_KEY_PASSWORD`。不要提交或输出这些文件；请在安全位置保留原始备份。已有签名 secrets 时默认拒绝覆盖；只有从原始备份恢复才能加 `--restore`。CI 缺少任何一项时拒绝发布，避免生成无法覆盖更新的随机签名 APK。

每次成功发布保留独立 `android-v<versionName>` Release，并更新 `android-latest` 的 APK、SHA-256 与 `android-update.json` 资产。版本号使用工作流序号与重试序号；该渠道是预发布标记，不占用桌面客户端的 latest Release。清单指向带版本的不可变 APK 下载地址。签名 secrets 配好后，推送代码或手动触发工作流即可发布。

本地构建：JDK 17+、Android SDK 36，`cd android` 后执行 `./gradlew lintDebug assembleDebug`。正式构建还需设置上述签名环境变量与 `ANDROID_KEYSTORE_PATH`，以及 `ANDROID_VERSION_CODE` / `ANDROID_VERSION_NAME`。

代码检查和模拟浏览器不能替代真机登录、附件选择、麦克风、安装来源授权及两次正式 Release 之间的覆盖更新验收。发布需单独确认 Actions 成功与 Release 资产可经 gh-proxy 下载。
