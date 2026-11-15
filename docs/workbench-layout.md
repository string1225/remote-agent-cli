# 紧凑工作台与回答状态

客户端管家与项目管家采用同样的行内入口：客户端行最右侧为“管家”，不再占用项目列表的一行。设备、项目、在线/离线权限及原有管家路由不变；点击管家不会同时展开或折叠客户端，键盘也可操作。

桌面默认三栏比例从21% / 23% / 56%调整为15.75% / 17.25% / 67%，即前两栏各收窄约四分之一，把空间交给对话。窄屏保留180px导航最小宽度；手机继续使用抽屉。已有手动保存的列宽保留，拖拽/方向键仍可调整，双击任一分隔线恢复新的紧凑默认值。

导航标题、设备与项目行、会话卡片、消息段落和输入区减少纵向留白；保持原有浅色/深色主题和正文大小，手机管家及设置按钮保留44px触控目标。

运行状态从顶部检查栏移到对应agent回答底部左侧，完成时间紧随状态。检查工具和连接/同步提示仍在顶部，运行详情继续可展开。

- 执行中显示变色跑马灯；仅已连接的实时状态播放动画，离线快照标明“上次状态”。开启减少动态效果时停用动画。
- 已完成显示绿色，执行失败显示红色；停止、排队、等待继续等状态有独立文字，不仅依赖颜色。
- 完成时间和耗时沿用持久轮次时间；当前轮可使用服务端taskCompletedAt，不以浏览器当前时间伪造完成时间。
- 新轮尚无输出时单独显示该轮状态，避免把上一条回答标成执行中；首条输出到达后复用为正常回答。
- 已结束轮次的状态/时间随对应回答保留，翻页、刷新和离线缓存仍按轮次区分。

`web/app.js` 本题仅修改默认比例一行。其余实现放在现有 coordination/workbench/workspace 模块，保留共享工作区的其他未提交改动，发布归档不包含这些脏改。

本机通过 skill-installer 安装了 [UI/UX Pro Max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill)，来源commit为 `d643997c82e94742dd3bee43210b0198e8bbba3d`，路径 `C:/Users/junte/.codex/skills/ui-ux-pro-max`。本轮读取并采用其Data-Dense Dashboard、语义颜色、减少动效及键盘/触控建议，沿用原页面主题；没有引入新的前端框架或生产依赖。后续轮次可直接调用该skill，不代表其他已安装客户端也安装了它。

验证使用模拟账号、客户端和会话数据，不读取真实会话正文。新增浏览器测试覆盖1440/1024/390宽度、行内入口与原路由、状态位置/颜色/时间、无输出与首输出交接、离线与减少动效、列宽保存/恢复及无页面横向溢出；既有聊天、workbench和分层协作浏览器测试通过。验证命令：

```powershell
node --check web/app.js
node --check web/platform.js
node --check web/workbench.js
node --check web/coordination.js
$env:PLAYWRIGHT_MODULE=(Resolve-Path .local/compact-ui/node_modules/playwright).Path
node tools/test-compact-ui.cjs
node tools/test-chat-ui.cjs
node tools/test-workbench.cjs
node tools/test-coordination.cjs
```

部署后已打开的工作台需要刷新才加载新资源；本轮不升级或重启客户端进程，不改项目绑定、会话或草稿数据。真实登录页面中的业务操作、原生Safari/移动设备触控仍需单独验收。

实际部署、线上资源与恢复点见 [2026-11-15发布记录](deployments/2026-11-15-compact-workbench.md)。
