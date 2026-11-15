# 弹窗与无需网页登录的 Endpoint 安装

所有工作台 dialog 共用 `dialogs.js/css`：标题与关闭按钮放在固定的头部，正文独立滚动；后续创建的待办编辑、默认主机选择、下游回复、工作区检查等弹窗也适用。现有表单节点、监听器与未提交输入原样保留。浅色/深色主题与手机布局沿用现有样式。

SaySo 内嵌页面的通用 Modal 和“选择执行管家”使用独立固定头部及滚动正文，生产打包产物随源码提交；长配置、录音列表和派工说明滚动时关闭按钮仍可见。

账号管理“复制 Endpoint / Codex 指引”通过管理员专用 POST `/api/admin/users/{id}/endpoint-guide` 签发该账号的独立个人 Token。指引包含实际 Token、x64/ARM64 MSI 链接、安装末页填入 Token 的步骤以及 macOS 命令；收件人无需登录网站、创建 Token 或填写初始密码。网页版指引继续使用账号登录流程。

专用 Token 标记 `purpose=endpoint-guide`，使用现有加密存储及模型认证；同账号重复/并发复制、服务重启后复用同一个有效指引 Token，撤销后再复制会签发新的。管理员只能通过此接口取得指引专用 Token，不能读取该用户其他个人或设备 Token。接口核对管理员及目标账号有效性，沿用同源与账号隔离，响应禁止缓存。Token 只留在本次复制/手动指引区域，关闭、换页、退出及迟到响应均清理或丢弃；不写入浏览器历史缓存。

安装末页接受两种输入：设备绑定码用于远程设备激活；Endpoint Token 用于配置本地 Codex 的模型账号，不新增或替换远程设备。两者不混用。保留默认“安装 Codex CLI”和“配置 Codex”，填写 Token 后应用即可；Token 模式不修改原设备绑定、服务或系统代理。需要远程工作台时仍通过原设备绑定码流程。

新增“安装 Codex App”默认不勾选。选择后打开官方 Microsoft Store 产品页，由当前桌面用户安装；商店安装完成、App 启动及原生 ARM64 仍需目标机器验收。CLI 故障与安装器故障分开：安装末页异常时保留已安装程序和原绑定，可在新 PowerShell 运行以下命令，隐藏输入 Token，不进入命令历史：

```powershell
remote-agent codex setup --server https://codex.sunny-string.cn/agents --token-prompt
```

本轮移除了激活后跳转到自身 `AgentComplete` 的操作，改为原位显示结果；补齐子进程缺结果、非零退出、未知消息与原生启动错误的诊断。用户所述约“4283”缺少完整错误文字，尚未在全新 Windows 中复现，不能声称精确根因已确认。

验证使用临时数据库、模拟账号/Token、模拟步骤和原生子进程 fixture，不操作真实绑定、Codex 配置、系统代理、SCM 或商店。浏览器覆盖静态及动态弹窗在1440/390宽度、浅深主题的长正文滚动和关闭，以及指引复制、禁用账号、剪贴板失败、关闭清理与迟到响应；既有待办、分层协调、工作区、设备码、归档设置及目录创建测试继续通过。

```powershell
$env:PLAYWRIGHT_MODULE=(Resolve-Path .local/compact-ui/node_modules/playwright).Path
node tools/test-dialog-onboarding.cjs
./tools/msi/test-activation-protocol.ps1
```

真实 MSI 安装、UAC、服务生命周期、商店操作及用户所报错误的精确复现需要显式授权的一次性管理员 Windows 环境。发布记录将分别说明服务器网页和新下载 MSI 的覆盖范围，不把渠道发布当作所有存量客户端已升级。
