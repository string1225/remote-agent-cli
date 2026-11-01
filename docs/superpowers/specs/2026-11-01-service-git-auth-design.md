# Windows 后台 GitHub 授权

用户已确认保留 LocalService，为服务启动的 Codex/Qoder 单独授权 GitHub，并要求客户端初始化提醒授权。

## 行为

新增 `remote-agent git login/status/logout --config PATH`。login 通过本机命名管道让业务 worker 运行 Git Credential Manager 的 `github login --device --no-ui`；用户在浏览器完成设备码授权。status 只列出后台保存的账号；logout 必须指定 `--username`，只移除对应后台账号。支持 login 的 `--force` 以更新失效授权。仅支持 github.com；不增加任意命令、URL、Token 输入或远程认证 API。

服务保持 LocalService + RemoteAgent SID。服务任务通过进程环境使用独立的 GCM DPAPI 存储和 GitHub credential helper；保留原 Git 配置、作者信息及其他主机配置。任务禁止弹窗或等待终端密码，缺少凭据时明确提示本机授权命令。服务管理命令只在本地输出设备码、账号和错误，不经过 139、历史或持久日志，不读取或复制桌面凭据。

## 边界与生命周期

命名管道只接收已配置的设备属主、管理员及系统的本机连接；客户端仅允许 SecurityIdentification，服务端读取有界请求后在锁定线程核对识别令牌，恢复服务身份后才运行 GCM。客户端通过 SCM 核对服务的 LocalService 身份、SID 类型和安装程序，核对管道 PID 是服务本身或其直接 core 子进程，避免依赖跨用户读取进程令牌。限制请求格式、大小、输出、并发和时长，断开终端或停止 worker 取消登录进程。认证随业务 worker 启停，固定 MSI 宿主和更新协议不变。

凭据目录位于原数据目录的 `git-auth`，拒绝重解析目录，以受保护 DACL 仅授 RemoteAgent SID、SYSTEM、管理员访问。OWNER RIGHTS 限制共享 LocalService 属主的隐式修改权限。GCM 使用 CurrentUser DPAPI，独立路径保证不会覆盖桌面凭据。初始化不清理任何旧文件；登录失败保留现有账号，退出登录只操作用户指定的后台账号。升级继续使用原路径。

## 提醒与验收

Windows connect（含重试/修复/恢复）、enroll、service configure 成功后提示 `remote-agent git login` 和 `remote-agent git status`；自定义配置保留 `--config` 参数，`--no-start` 提示先启动服务。网页初始化帮助说明桌面 Git 登录不自动供后台服务使用。

普通测试使用临时目录、模拟 GCM 和凭据，不触碰真实绑定或服务。覆盖独立环境、原 Git 配置保留、账号选择、失败/取消、协议限制、管道双向身份校验及 ACL；可用时在临时目录验证已安装 GCM 的 DPAPI 存储。完成 gofmt、go test ./...、go vet ./...、Windows x64/ARM64 构建及网页 JS 语法检查。真实设备码授权、LocalService 推送、退出登录、冷启动须在目标机器单独验收。
