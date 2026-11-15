# 分类文件分享

普通代码包和生成产物默认作为临时文件直传storage.to；特别敏感文件先取得用户对具体文件及分享范围的明确确认，再沿用139。两种路径都不发消息、不搬迁已有文件、不删除本机原文件或会话附件，也不把网络失败当作换通道的授权。

| 分类 | 路径 | 链接与有效期 |
| --- | --- | --- |
| `temporary`（默认） | 客户端→storage.to API初始化→Cloudflare R2直传→API确认 | 下载页面，默认3天，以实际`expiresAt`为准；不是原始文件直链 |
| `sensitive` | 明确确认后，客户端→原139存储 | 持随机链接可下载，次日北京时间凌晨4点失效；不等于私密或加密存储 |

CLI：

```powershell
# 普通代码包：无需读取设备配置或创建storage.to账号
remote-agent files upload --path C:\out\source.zip
# 仅在用户明确确认该文件及139分享范围后使用
remote-agent files upload --path C:\out\sensitive.pdf --class sensitive --confirm-sensitive
# 提前删除本机有管理回执的storage.to副本，保留原文件
remote-agent files delete --id Abc123def
```

`--config PATH`指定管理回执所在的客户端配置路径；临时上传不读取其内容，敏感通道才正常加载设备绑定用于139认证。回执位于配置旁`shared-file-receipts`。MCP原`upload_file`保留，增加`class`与`userConfirmed`；默认temporary。`sensitive,userConfirmed=true`记录已经取得确认，不能替代询问或扩大用户授权。`delete_uploaded_file(fileId)`只能删除本机有对应管理回执的storage.to文件。

结果返回`provider/class/linkType/id/name/size/sha256/url/expiresAt`。storage.to的`linkType=page`只能作为下载页面交给用户，不构造未公开的下载接口或热链；用户在网页点击下载。139为`direct`，HTML/HTM继续提供原`previewUrl`沙箱静态预览，禁止脚本、表单、外部网络及读取工作台账号。临时通道不冒充139 HTML预览，也不会为预览便利自动回退。

## storage.to实测与限制

依据官方 [REST API Markdown](https://storage.to/docs/api.md)、[LLM文档](https://storage.to/llms.txt) 和[文件类型说明](https://storage.to/docs/file-types)。普通HTML文档可能触发Cloudflare验证，官网提供的`.md`文档可正常读取。

- 匿名API无需API Key，初始化、R2 PUT、确认三步已用纯合成文件实测成功。所有权证明只由客户端保存在本地回执，不进入分享URL、CLI/MCP结果或139请求。稳定visitor身份跨并发/重启复用，不靠更换身份规避额度。
- 官方匿名每天最多50文件；每visitor滚动24小时100GB、每IP500GB上传额度。该站支持25GB，但本产品首版仍限制1字节至100MiB，超过50MB按API返回分片上传。
- 默认3天到期；当前接入不修改免费/付费套餐、账号或到期设置，不承诺北京时间4点清理第三方文件。
- 官方拒收exe/scr/com/pif/apk/ipa；本地提前明确报错，不改名规避。ZIP、其他普通压缩包和MSI等按官方规则接受。
- 临时文件按文件名拦截明显凭据/密钥（如`.env`、`id_ed25519`、`auth.json`、pem/key），ZIP检查目录中的这些名称，不读取凭据内容。此检查不是内容扫描，不证明压缩包无秘密；嵌套/加密包和其他归档仍由打包者核对分享范围。
- API拒收、429、网络、安全验证或分片错误均明确失败，不切换至139。分片失败尝试abort；确认响应丢失可能留下第三方已创建但本机未收到回执的副本，不能声称已删除，应保留原文件并报告不确定性。成功后落盘回执失败会尝试删除刚上传的副本，失败时返回公开URL/实际到期时间供用户核查，不泄露管理证明。
- 当前自动化浏览器打开测试分享页遇到Cloudflare安全验证（403）；上传API成功不等于真实浏览器下载、吞吐或用户地区可达性已验收。不绕过验证码，也不虚构已验证的提速比例。

每次API请求最多45秒，R2单次PUT最多5分钟，全次上传15分钟；流式传输与SHA-256，不一次读入整个大文件。只接受固定storage.to API与HTTPS的R2账户域名；拒绝重定向、其他主机、异常端口和URL用户信息。平台设备Bearer不进入第三方接口；visitor/owner证明也不进入R2 PUT。配额错误保留HTTP状态及安全的Retry-After。

回执及稳定visitor身份按0700目录/0600文件创建并原子发布；Windows实际权限沿用所属数据目录ACL，不宣称POSIX模式位替代Windows ACL。删除仅访问本机对应普通回执和固定API；丢失回执不能代他人删除。提前删除和第三方到期仅影响分享副本。

## 139保留路径

原`POST /api/agent/files?id=<deviceId>`、设备鉴权、四路并发、100MiB、流式SHA-256、失败清理、随机256位链接和Range下载保持兼容。未升级的旧客户端仍保留旧行为，不能说所有安装端已采用新分类。

`GET/HEAD /files/<id>`无需登录；敏感确认必须说明这一性质。到下次北京时间UTC+8凌晨4点立即拒绝下载；服务内清理器及启动补清理只处理应用拥有的过期分享目录，保留项目、数据库、录音、附件和旧文件原始数据。

原`SHARED_FILES_DIR`及Nginx上传location保持不变，不为第三方路由修改真实配置或迁移存量文件。新的分类、MCP说明与共享skill需要执行上传的客户端正常空闲升级/激活后生效；活跃会话不强制更新。CEO、客户端管家、项目管家及普通执行者均从统一每轮指令获得此分类规则。

## 验证与发布

普通测试使用临时目录、模拟平台凭据和模拟storage.to/R2传输，涵盖默认路由、确认前拦截、秘密隔离、分片顺序/ETag/补充URL/中止、失败不换路由、所有权回执删除及并发身份恢复；不上传真实用户文件。真实连通测试只使用生成的无秘密样例。

遵守gofmt、相关测试、Go全量/vet与六平台构建；LocalService的既有ACL测试失败须单列。运行时发布复用`tools/publish-runtime.ps1`及SERVICE SSH配置，不读取凭据内容；保留原渠道与回滚。新增模块完全在客户端，139既有文件服务兼容，不以发布渠道代替所有客户端已激活的验收。
