# 紧凑工作台发布与验收

2026-11-15，网页源码版本 `e07bc9d62bd4a692f423d939ff6e6081cc6e0b37` 已推送 origin/main 并部署139。布局与交互说明见 [工作台布局](../workbench-layout.md)。

客户端管家移到客户端行最右侧；执行状态与完成时间移至对应回答底部。实时执行中显示变色跑马灯，完成绿色、失败红色，离线快照/减少动效不播放动画。默认左两栏各收窄约25%，窄屏保留最小宽度，标题、列表、消息与输入区减少留白；保留手动列宽，双击分隔线可恢复新默认。

## 实现与验证范围

- 本题只提交9个文件；`web/app.js` 仅暂存默认比例的一行修改，其余已有脏改没有提交。通过替换该行后与修改前逐字节对比，确认共享app.js其他内容保留；README、其他网页及Go共享改动也保留。
- `node --check` 覆盖app.js、platform.js、workbench.js、coordination.js。仅含本题变更的归档源码上，新增紧凑布局测试及既有chat-ui、workbench、coordination四组浏览器检查均通过。
- 新测试覆盖1440/1024/390宽度、客户端/项目管家路由与键盘操作、离线禁用、状态/时间位置、变色/绿红语义、无输出占位到首输出交接、停止与错误区分、深浅主题、减少动效、无页面横向溢出、键盘调整与列宽保存/恢复。
- 视觉复核使用模拟账号、客户端、会话及消息；没有读取真实会话正文。截图位于 `.local/compact-ui/before/`、`final-screens/`，线上资源复核截图在 `live-screens/`。
- 本轮没有修改Go源码，不重复无关Go长测；部署脚本从已提交归档成功构建Linux服务端及Windows上传辅助程序，没有运行真实SCM/MSI测试。

UI/UX Pro Max 已经通过 skill-installer 安装到 `C:/Users/junte/.codex/skills/ui-ux-pro-max`，上游为 [nextlevelbuilder/ui-ux-pro-max-skill](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill)，固定来源 `d643997c82e94742dd3bee43210b0198e8bbba3d`。本轮已读取并采用数据密集工作台、语义颜色、减少动效和交互目标建议；后续轮次可直接调用。该安装范围为本机Codex技能目录，没有声称其他客户端也安装了它。

## 139部署与线上检查

使用现有 `tools/deploy-139.ps1` 和显式服务专用SSH配置，保持正常主机校验，没有新增凭据或改信任：

```powershell
./tools/deploy-139.ps1 -SSHHost deploy-139 -SSHConfig 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
```

- Release：`/opt/remote-agent/releases/e07bc9d62bd4a692f423d939ff6e6081cc6e0b37-20261115111840`。
- 发布前配置、systemd定义、Nginx与数据库备份：`/var/backups/remote-agent/e07bc9d62bd4a692f423d939ff6e6081cc6e0b37-20261115111840`；旧程序保留，脚本有发布失败自动恢复。
- HTTPS `/agents/healthz` 返回ok=true，remote-agent/nginx active；发布脚本检查运行中服务端程序属于新release。
- 从公网下载index.html、app.js、coordination.js/css、workbench.js/css、workspace.css，7项HTTP200且SHA256与已提交构建源一致。再次用这些线上资源运行模拟浏览器检查通过，验证页面资源及交互实现已上线。
- 六平台客户端渠道的原版本、序号、摘要发布前后逐项一致；本轮没有发布/激活客户端runtime。本机RemoteAgent PID40396、Updater PID4872与发布前一致，均Running，没有中断当前父子进程树。

服务端网页更新已生效；已经打开的浏览器页面需刷新加载新资源。登录后的真实业务操作未用生产账号执行，浏览器功能验收使用模拟API与会话；原生Safari、真实移动设备触控和实际回滚生命周期未单独验收。

## 回滚与证据

额外准备的 `rollback-program.sh` 位于上述备份目录，root:root、0700，`bash -n`通过，本轮未实际回滚。该脚本恢复旧服务端程序及其页面资源，保留当前数据库/配置；本题没有数据结构迁移。需要时运行：

```powershell
ssh -F 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config' deploy-139 `
  'sudo -n bash /var/backups/remote-agent/e07bc9d62bd4a692f423d939ff6e6081cc6e0b37-20261115111840/rollback-program.sh'
```

复核与重现命令见布局说明；Playwright测试依赖装在ignored `.local/compact-ui/` 并生成package-lock，没有引入生产前端依赖。发布、测试、资源摘要、客户端渠道和本机服务前后快照保存在该目录的日志/JSON；`RESULT.md`为本次结果卡。

本轮仅通过原网页/会话返回，没有发额外IM/广播、另建执行session、操作其他会话正文或代写父会话history。
