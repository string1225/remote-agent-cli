# SaySo in Remote Agent

界面基于 `string1225/say-so` 的 `856311e`（MIT）整合，许可证保留在 `LICENSE`。

此目录只提供构建时依赖。React 界面挂载在工作台的 Shadow DOM 中，通过注入的 transport 调用已经认证的 WebRTC DataChannel。没有单独的 HTTP API、本机端口或 Node.js 后台进程；安装 Agent 即具备讨论、录音及本地规则能力。

```sh
cd sayso-ui
npm ci
npm run build
```

产物写入 `web/sayso/` 并纳入版本控制，由 Go 服务端内嵌分发。因此普通 Go 构建和目标电脑安装均无需 Node.js。修改此目录后必须同时更新构建产物。依赖许可证随产物输出为 `licenses.txt`。

后端实现位于 `internal/agent/sayso*.go`。不修改源 SaySo 仓库，也不自动导入其中已有的私有录音与数据。
