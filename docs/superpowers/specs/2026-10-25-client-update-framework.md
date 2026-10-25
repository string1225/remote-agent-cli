# 客户端更新框架

状态：已实现签名发布、固定 Windows 宿主、常驻更新服务、worker 就绪协议和事务切换。实际 MSI/SCM 验收与原生 ARM64、冷启动/退出登录仍需在目标环境单独进行。

## 目标

首次安装本框架 MSI 时授权一次。日常运行 remote-agent update，由已安装的更新服务完成业务更新，不再运行 MSI、不再弹 UAC，不重新配置 SCM、PATH 或项目目录权限。不实现旧客户端更新兼容或旧计划任务迁移。

后台启动一分钟后及每六小时检查签名发布，客户端空闲时使用同一事务更新；忙碌时保留现有版本。首次绑定、更改目录权限、卸载及未来宿主/信任锚更新属于系统维护，仍可能需要管理员授权。

## 安装和权限边界

```text
%ProgramFiles%\Remote Agent\
  remote-agent.exe                 固定 CLI 启动器和 SCM 宿主；MSI 拥有
  seed\remote-agent-core.exe       MSI 提供的初始业务程序
%ProgramData%\RemoteAgent\
  service.json                     原用户路径、owner SID、更新站点
  runtime\state.json               current / previous / pending / 发布序号上限
  runtime\versions\<sha256>\
    remote-agent-core.exe
    release.json
  update-logs\updater.log
  health\status.json               运行版本、宿主 PID、就绪/暂停状态
  health\host.log
原用户目录\.remote-agent\          原绑定、项目、history.db、SaySo 和录音
```

MSI 注册两个使用同一固定宿主的服务：

- RemoteAgent：LocalService + 原服务 SID，执行业务 worker。开机运行，退出登录继续运行。
- RemoteAgentUpdater：LocalSystem，只验证和存放签名包、写受保护状态、停止/启动已有业务服务。不执行下载的业务代码，不执行用户命令，不重新授权目录。

runtime 和更新状态只有 SYSTEM/管理员可写；业务服务 SID 和普通用户只读。health 单独允许业务服务 SID 写入。安装拒绝运行时路径上的重解析点或非 SYSTEM/管理员属主；文件操作由固定路径和 os.Root 限制范围。

更新服务使用本机命名管道，拒绝远程连接。绑定后只允许原配置 owner SID 或管理员请求更新，未绑定时允许本机用户更新。IPC 仅接受更新/回滚及受限参数，不接受可执行路径、任意命令或配置写入。CLI 使用只读 SCM 权限核对管道服务 PID，无需管理员权限。

## 宿主和 worker

cmd/agent-host 和 internal/servicehost 持有 SCM 连接，处理生命周期、Job Object 子进程清理、本地就绪超时。普通 CLI 命令转发原参数、标准输入输出和退出码。version 分别显示宿主协议与实际业务版本。

core 内部 worker 使用继承的匿名管道，协议 v1 为 prepare / ready / activate / pause / resume / stop。prepare 只验证原配置和历史存储，不连接站点或恢复任务；activate 后才启动业务。未绑定时等待配置也算正常就绪，139 离线不判为版本故障。

更新先关闭新任务入口，检查运行任务、排队任务和业务操作。默认忙碌即拒绝；--force 通过原取消上下文结束任务并保留历史和恢复 ID。录音或录音上传必须先完成，强制更新也不丢弃未完成录音。

## 更新事务

1. 从受保护配置捕获的站点或显式 --server 获取发布，不携带设备/模型凭据。未绑定电脑也能显式指定服务器更新。
2. --check 只读比较，不下载二进制、不写状态、不调用更新服务。实际更新由固定服务取得进程内和文件单实例锁。
3. 验证平台、架构、单调发布序号、最低宿主、worker/数据协议、Ed25519 签名、大小和 SHA-256。不跟随重定向，只访问原站点的固定下载路径。
4. 在受保护目录下载校验，落盘后原子改名。损坏或中断下载不改变当前版本。
5. 运行服务先暂停接收任务，在停止前记录 pending、原选择、上一稳定版和启停状态。候选通过本地就绪检查后提交；失败恢复原选择。宿主也可直接回退受保护状态批准的旧程序。
6. 原先停止或禁用的服务只准备版本，不启动或启用；下次用户启动时验证候选，再由更新服务校对状态。
7. 中断后先校对 pending，不依赖服务器在线。原先运行的服务可恢复启动，原先停止的保持停止。健康状态必须来自当前 SCM 宿主 PID，只能确认受保护状态中的候选或旧版。
8. 保留当前版和上一稳定版，受限清理其他程序目录。失败候选不能成为上一稳定版。--rollback 仅选择已验证的上一稳定版；发布序号上限保留，允许重试同一已验证最高发布，拒绝过期发布或同序号不同摘要。

状态在同目录临时文件中写入、同步并原子替换。首版只接受数据版本 1，不覆写配置、数据库和录音，不支持不可逆数据迁移。失败回退只切程序，不能用旧数据库覆盖已有新消息。

## 发布和签名

清单为 /updates/v1/windows/{amd64|arm64}.json，文件为 /downloads/remote-agent-core-windows-{arch}.exe。internal/clientupdate 共用验证和事务；macOS/Linux 沿用原更新器，不改变启动注册。

清单包括 schemaVersion、sequence、version、os、arch、minHostVersion、workerProtocol、dataVersion、minRollbackDataVersion、size、sha256、keyId、signature。签名覆盖固定域、key ID 和规范化 Payload JSON。公钥编译进宿主，清单不能添加密钥或取消校验。宿主/协议不兼容时明确拒绝并提示升级宿主 MSI；首版不自动转为运行 MSI。

生产私钥只保存在发布环境，不提交、不进入 dist、不部署到业务服务器。构建发布：

```powershell
go run ./tools/build -version <提交号>
./tools/msi/build.ps1 -PackageVersion 1.1.0 -ReleaseVersion <提交号>
go run ./tools/release-sign -key <发布私钥绝对路径> -sequence <递增整数> -version <提交号>
```

一起发布对应 core 和 dist/updates/v1/windows/*.json。CI 一次性管理员验收使用独立测试密钥和测试宿主，禁止使用生产私钥。

## 修复、卸载和验收

MSI 修复只恢复宿主和 seed，不降低 runtime 选择。卸载成功提交后由包内固定维护程序清理 runtime/health；失败事务保留运行时供 MSI 回滚。不删除 service.json、原绑定、配置、项目、历史或录音。

临时目录/模拟凭据测试覆盖签名、损坏包、路径限制、任务忙碌、停止状态、启动失败、中断恢复和回滚。tools/msi/acceptance.ps1 只能在显式授权的一次性管理员 Windows runner 执行，覆盖普通用户令牌更新、MSI/SCM 注册不变、启动回退、MSI 修复保持 runtime 和卸载保留数据。原生 ARM64、真实断电、冷启动/退出登录另行验收。

日常更新通过标准：业务版本改变，MSI 产品代码/版本、RemoteAgent 路径/身份/SID/启动/恢复配置不变；绑定、项目、会话和录音仍可用，没有新 UAC。
