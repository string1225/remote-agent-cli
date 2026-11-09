# Windows 后台目录授权与 Git 仓库信任

`remote-agent access` 为已安装的 **RemoteAgent 服务 SID** 授权所选本机路径。它保留 LocalService 运行身份，不更换服务账号、不重启服务，也不读取或复制文件正文、GitHub Token、用户凭据。

| 问题或报错 | 对应能力 |
| --- | --- |
| 文件/目录 `Access is denied` | NTFS ACL：`access grant --permission read` 或 `modify` |
| Git `detected dubious ownership` | 仓库信任：明确仓库路径的 `--git-trust` |
| GitHub 认证失败或需要私库/push 权限 | 现有 `remote-agent git login` / `remote-agent github login`；先用对应的 `status` / `accounts` 检查已有登录 |
| 桌面用户的 DPAPI 凭据无法解密，或需要交互桌面 | ACL/UAC 无法解决；需要目标后台身份自己的登录或支持后台运行的工具 |

Windows 的管理员同意不等于提供用户密码，也不会把桌面登录、Windows Hello、桌面 ssh-agent 或交互窗口交给 Session 0 服务。已经有效的后台 GitHub 登录可以继续使用，不必因目录授权而重新登录。详见 [服务凭据说明](service-credentials.md) 与 [GitHub/服务器连接管理](connection-management.md)。

## 常用命令

在用户本机终端运行。未提升的修改命令会按需请求 UAC；已提升时直接执行。`list` 先尝试读取，读取被拒绝时才请求 UAC。

```powershell
# 只读取指定文件，不读取或复制正文来完成授权
remote-agent access grant --path 'D:\Documents\report.txt' --permission read

# 允许后台修改此项目及其现有内容、新建后继承权限的内容，并信任这一个 Git 仓库
remote-agent access grant --path 'C:\Users\junte\Code\remote-agent-cli' --permission modify --recursive --git-trust

# 只授予整个资料目录的读取/执行权限
remote-agent access grant --path 'D:\Reference' --permission read --recursive

remote-agent access list
remote-agent access restore --backup <grant或list输出的32位ID>
```

`read` 增加读取与执行权限；`modify` 增加读取、执行、写入及删除权限，不增加更改 ACL 或接管属主的权限。授权是**增量添加**：选择 `read` 不会移除既有写权限，也不会覆盖已有拒绝规则；如已有拒绝规则仍阻止访问，需由目录管理员检查。父目录原有的遍历限制也可能影响访问，本命令不会自动授权其他路径。

目录默认只授权目录对象本身。使用 `--recursive` 才授权现有整个目录树，并向目录增加可继承的服务 SID 权限，包括原本关闭继承的子目录。开始前请停止对目标树的文件移动、删除和 ACL 编辑；命令期间会固定现有对象，阻止其被重命名/替换，但不阻止其他程序新建文件。

只增加 `NT SERVICE\RemoteAgent` 的 SID 权限，不给所有 LocalService 进程或 Everyone 增加权限；原有其他主体的访问规则保留。不会修改属主、服务身份、客户端/项目 ID、会话、录音或 Git 用户配置。

## Git 信任的范围与生效

`--git-trust` 是显式允许后台 Git 使用该仓库的配置和钩子，只能用于包含真实 `.git` 目录的仓库根目录；不自动信任父目录、其他仓库、子模块或通配符。当前命令拒绝 linked worktree 与裸仓库，避免把所选路径之外的 Git 目录隐式纳入授权。

信任保存在受保护的 `%ProgramData%\RemoteAgent\access\trust.gitconfig`，使用精确的 `safe.directory`。**升级到包含此功能的业务 core 后**，Windows 后台任务继承的 Git 命令作用域通过 `include.path` 读取该文件，每次调用 Git 都重新读取。新增/恢复信任不需要再重启服务。桌面 Git 不注入此环境；不会改写用户或系统的 `.gitconfig`，不使用 `safe.directory=*`，也不触碰 GitHub 登录。

独立运行新 core 可以给旧服务增加 NTFS 权限；旧服务还不认识后台 Git 信任配置，需安排空闲时升级 core 后再开始新任务。现有进行中会话及其进程树不会因为 `access` 命令而重启或重新加载。

## 备份、恢复与失败处理

每次 grant 在修改第一个目标之前，先把所有现有对象的 DACL、属主/组描述及对象标识保存为 `%ProgramData%\RemoteAgent\access\<ID>.json`，并打印恢复命令。备份没有文件正文或凭据。该专用目录由 Administrators 持有，SYSTEM/管理员可写，RemoteAgent 服务 SID 只读；不修改父级 RemoteAgent 系统目录的 ACL。备份写入并刷盘后才修改权限，ACL 失败或 Git 信任发布失败时自动恢复已备份 DACL 和旧信任配置；失败备份也保留。

`access restore --backup ID` 恢复该次授权前的 DACL 和继承保护状态，撤销该次 Git 信任，备份仍保留；重复恢复安全。Windows 可能规范化自动继承的记账标志，不改变原 ACE、继承保护或属主。

恢复前检查整个备份：若后续有人修改了 ACL/属主，普通恢复拒绝覆盖。核对备份后，管理员可明确执行：

```powershell
remote-agent access restore --backup <ID> --force
```

`--force` 只允许覆盖之后的 **DACL** 修改，不改变当前属主/组，不绕过链接、对象替换或缺失检查。被移走、删除或替换的原对象需先恢复到原路径；命令不会把旧权限直接套到另一个新对象。发生中断或部分恢复失败时，备份显示 `prepared`，按原 ID 加 `--force` 可重试；恢复操作先撤销该次 Git 信任。自动恢复自身失败也会打印原备份 ID 和恢复命令。

恢复时，现有原对象按父目录到子对象的顺序恢复，清除本次增加的继承来源；之后创建且仍正常继承的文件也会失去对应权限。之后由其他程序另行保护、复制为显式 ACE 的新文件，需要目录管理员单独核对，不能把备份中不存在的对象视为已完成原 ACL 恢复。

拒绝活动授权的相同/父子路径重叠，避免恢复互相覆盖；需要调整 read/modify 时先恢复旧备份，再重新 grant。单次最多 20,000 个现有对象，超过时请选择更小目录。路径必须是明确的本机驱动器绝对路径，拒绝整个驱动器、UNC/设备路径、ADS、通配符、重解析点（包括 junction）、多硬链接文件和 RemoteAgent 系统状态目录；任何预检失败均不修改目标 ACL。不接管属主，不绕过拒绝权限，不启用备份/还原特权来强行访问对象。

## 验证边界

自动测试只使用临时目录、模拟凭据文件和虚构服务 SID，验证真实 NTFS ACL 的读取/修改、原属主及 DACL 保留、受保护子目录、新文件继承和恢复、部分失败回滚、冲突和链接拒绝。真实 Git 测试通过测试开关模拟属主不一致，检查明确仓库的信任及撤销；Git 用户配置指向临时空路径，不发起网络认证。

普通测试不请求真实 UAC，不安装/修改 SCM 或 MSI，不授权真实项目、不读取桌面凭据。用户桌面的 UAC 同意/取消、生产备份目录 ACL、真实服务令牌上的新目录访问、升级后后台 Git 继承及退出登录运行，仍需在目标 Windows 环境验收。现有服务的 GitHub 登录无需为这些测试重做。
