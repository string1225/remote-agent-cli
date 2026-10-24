# 139 服务器 Codex 账号管理

管理员在「账号管理」查看 139 上的 Codex 账号、当前代理账号、登录状态和凭据有效期。普通账号和模型 Token 无权调用管理接口。

支持最多 20 个独立账号，手动选择一个供模型代理使用，不自动轮换账号。添加账号不会切换当前账号；切换只影响新的模型调用。已有模型请求继续使用开始时选定的凭据。

「添加并授权」或「重新授权」通过原生 Codex app-server 的 `account/login/start`、`chatgptDeviceCode` 流程获取官方授权地址和设备码；管理员在浏览器完成登录。服务器最长等待 15 分钟，可取消。页面关闭后再次进入可继续查看进度。需要账号/工作区开启设备码登录，参见 [OpenAI 官方认证说明](https://learn.chatgpt.com/docs/auth)。

每次授权先写入独立 CODEX_HOME；成功后原子更新账号目录索引，失败、取消、超时或原账号身份不匹配时不覆盖旧凭据。原生 Codex 负责保存和续期 OAuth；管理接口不返回 access/refresh Token 或服务器路径。原有 `/home/junte/.codex` 以 `legacy` 账号接入，原文件保留。已成功授权的新增账号需要主动设为当前账号。

## 服务与部署

`cmd/personal-gateway` 将 139 原有 `/opt/codex-relay/internal/personalgateway` 网关及旧管理界面纳入本仓库维护，保留现有设备 Token 注册表、管理密码和旧 `/openai/personal/v1`、`/openai/admin` 接口。新增私有控制接口 `/internal/codex/{accounts,login,activate,cancel}`，服务仅绑定 loopback；该路径不添加公共 Nginx 映射。

编译：

```sh
go build -trimpath -o personal-gateway ./cmd/personal-gateway
```

在原网关 `serve` 参数上增加：

```sh
--accounts-root /home/junte/.config/sunny-personal-gateway/codex-accounts \
--legacy-codex-home /home/junte/.codex \
--control-token-file /home/junte/.config/sunny-personal-gateway/account-control.token
```

控制凭据使用单独的至少 32 字节随机秘密，文件模式 0600，仅网关用户与系统管理员可读。不得复用旧管理密码、设备 Token 或模型服务 Token。Remote Agent 服务通过 systemd `LoadCredential=codex-admin.token:/etc/remote-agent/codex-admin.token` 获得相同秘密；环境设置 `CODEX_ADMIN_TOKEN_FILE=/run/credentials/remote-agent.service/codex-admin.token`。现有 `CODEX_GATEWAY_URL` 和模型服务凭据保持原值。

部署前备份两份服务、原二进制和账号索引；等待旧网关无活动模型请求再重启，验证旧 Token、旧管理入口和新管理员接口。回滚恢复旧网关与 Remote Agent 二进制/服务配置；原 OAuth、设备注册表及业务数据保持不动。

## 验证范围

自动测试覆盖：管理员/普通用户隔离、固定 loopback 路由及专用凭据、拒绝任意 URL、多个凭据目录、明确切换、授权失败和错误账号保留旧数据、持久化当前账号、授权地址校验、接口不泄露凭据。实际设备授权成功需要管理员在官方页面输入设备码；自动测试不会替用户登录真实账号。
