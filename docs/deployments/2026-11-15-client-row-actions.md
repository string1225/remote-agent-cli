# 客户端行操作与暗色图标发布

2026-11-15已部署139，刷新工作台后加载。源码实现`217fcf519d992a67bcc7ce15457a78756487174f`，截图稳定性验证提交及实际发布源码`be69d77e970391bf863b84faec29a93dee8af21a`，均已推送origin/main。

## 当前行为

- 宽度低于1280px时，客户端行隐藏新增项目加号，在该客户端三点菜单中提供“新增项目”；1280px及以上保留行内加号。项目菜单不提供客户端级新增入口；离线项禁用，菜单默认聚焦可用操作。
- 客户端管家及项目管家均为同一轮廓机器人SVG，保留原路由、title及“打开…管家”的aria-label。图形本身aria-hidden，鼠标点击不折叠客户端，键盘可操作；手机触控目标继续为44px。
- 暗色主题使用浅色Apple图形及简化Linux轮廓，浅色主题保留原图；主题切换无需重绘客户端列表。来源为Devicon v2.17.0，许可与上游blob见web/icons/README.md。

详情见[工作台布局](../workbench-layout.md)。

## 验证与范围

在隔离归档中排除原共享脏改测试，再基于已合并的main复核。过程中另一个执行会话提交了`3e81311`聊天文件路径功能，文件与本题不重叠；当前发布保留它，并通过其markdown-files测试，没有读取其他会话正文或覆盖其文件。

- app.js、platform.js、device-settings.js、coordination.js语法检查通过；本题不修改已有脏改的app.js、index.html、platform.js。
- compact-ui覆盖1440/1280/1279/1024/390五档宽度、浅深主题、断点前后加号/菜单、正确客户端新增表单、离线禁用、项目菜单隔离、管家点击/Enter路由、装饰SVG与可访问名称、列宽保存及无横向页面溢出。浅色保留原图，深色两个新资源200并完成截图复核。
- coordination、project-preferences、access、binding-code通过；合并后的markdown-files通过。使用线上下载资源再次运行compact-ui通过，全部数据来自模拟账号/客户端/会话。
- 额外运行的device-maintenance独立测试在第83行等待“重试会核对同一请求”状态超时。只导出未修改HEAD中的测试和其加载的maintenance.js/style.css/maintenance.css再次运行，同一位置失败；这是既有基线边界，本轮未改它，也未声称远程维护全流程已验收。
- 未修改Go或原生平台代码，不重复Go长测；部署从已提交归档成功构建Linux服务端和Windows上传辅助程序。真实生产账号新增项目、原生Safari、真实手机触控及实际回滚生命周期待单独验收。

## 部署与回滚

使用原`tools/deploy-139.ps1`及SERVICE SSH配置，保持严格主机校验：

```powershell
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
```

- Release：`/opt/remote-agent/releases/be69d77e970391bf863b84faec29a93dee8af21a-20261115180419`。
- 备份：`/var/backups/remote-agent/be69d77e970391bf863b84faec29a93dee8af21a-20261115180419`，包括停止写入后的数据库、配置、systemd及Nginx；自动发布失败恢复已准备。
- HTTPS healthz返回ok，remote-agent/nginx active；公网下载index/app/platform、coordination.js/css、device-settings、theme、两个暗色SVG及合并的markdown/workspace共11项资源，SHA256与发布源码逐项一致。
- 新Release沿用原客户端dist，未发布或主动激活runtime；Windows渠道仍为`20261115-file-sharing-d3671ea`、序号`2026111502`且摘要未变。本机RemoteAgent PID40396、Updater PID4872发布前后Running且不变，没有中断活跃父子会话树。
- 手动`rollback-program.sh`位于上述备份目录，root:root/0700、bash -n通过，本轮未执行。它恢复原`3e81311f6f8e31155028d3b1d36aca4e2f8f4e2b-20261115175034`程序/网页，保留当前数据库及runtime渠道。

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent/be69d77e970391bf863b84faec29a93dee8af21a-20261115180419/rollback-program.sh'
```

ignored证据位于`.local/client-row-ui/`，含隔离/合并/线上资源截图、摘要、PID、发布和旧维护测试基线日志。共享脏改和他人的未跟踪测试全部保留，无额外IM、凭据读取、新会话或委派。

## 后续：管家改为无框人像图标

2026-11-15追加发布源码`5871f750eb7ceac43797e0e052be1d51e935f079`，已推送origin/main。客户端管家和项目管家均改为简单的人像线条SVG，去掉机器人头部的框形轮廓；按钮明确使用无边框、无原生按钮外观。修正暗色主题通用面板规则优先级过高导致管家按钮出现块状底色的问题，空闲时背景透明，悬停反馈和键盘焦点提示保留。

- 从已提交归档验证，排除共享脏改；app/platform/coordination语法检查、coordination浏览器测试及compact-ui通过。后者覆盖1440/1280/1279/1024/390宽度、原点击/Enter路由、离线禁用和44px移动端目标等。线上下载的9项资源（index、app/platform、coordination.js/css、theme、device-settings及两个暗色平台图标）均200，SHA256与发布源码一致；覆盖下载资源的模拟页面再次通过compact-ui。
- 额外等待主题切换过渡结束后，在1440/390下检查浅深主题的计算样式并复核截图：管家按钮背景均为透明、边框0、appearance:none、无阴影。所有业务数据均为模拟数据，未读取真实会话正文；原生Safari、真实触控设备、生产账号业务变更及实际回滚生命周期未验收。本题未修改Go或原生平台代码，不重复长测；发布脚本从提交归档构建Linux服务端和Windows上传辅助程序成功。
- Release：`/opt/remote-agent/releases/5871f750eb7ceac43797e0e052be1d51e935f079-20261115193955`。正常SERVICE SSH部署，部署脚本先备份并设自动失败恢复；备份目录为`/var/backups/remote-agent/5871f750eb7ceac43797e0e052be1d51e935f079-20261115193955`。HTTPS healthz为ok，remote-agent/nginx active。
- 同目录的`rollback-program.sh`为root:root/0700，bash -n通过，未执行；恢复上一`be69d77e970391bf863b84faec29a93dee8af21a-20261115180419`程序/网页，保留数据库及runtime渠道，并拒绝覆盖后续其他Release。
- 本次只更新服务端内嵌网页，已打开工作台需刷新。未发布或主动激活客户端runtime；Windows amd64渠道发布前后摘要相同，仍为`20261115-file-sharing-d3671ea`、序号`2026111502`。本机RemoteAgent PID43020、Updater PID4872发布前后Running且不变，没有重启客户端活跃进程树。共享脏改及他人未跟踪文件全部保留。

手动程序回滚命令（本轮未执行）：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 'sudo -n bash /var/backups/remote-agent/5871f750eb7ceac43797e0e052be1d51e935f079-20261115193955/rollback-program.sh'
```

本次证据位于ignored目录`.local/steward-icon/`：publish.log、live-assets.json、healthz.json、live-compact-test.log、live-style-probe.log、live-settled-screens和服务/渠道前后快照。
