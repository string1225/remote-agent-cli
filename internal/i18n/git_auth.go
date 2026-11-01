package i18n

// These entries are exact messages without format patterns; adding them during
// initialization keeps this feature's catalog together.
func init() {
	for english, chinese := range map[string]string{
		"Windows service GitHub authorization: remote-agent git login | git status | git logout --username ACCOUNT":                                "Windows 后台 GitHub 授权：remote-agent git login | git status | git logout --username 账号",
		"Windows background GitHub authorization (separate from desktop Git login):":                                                               "Windows 后台 GitHub 授权（独立于桌面 Git 登录）：",
		"GitHub account (required for logout)":                                                                                                     "GitHub 账号（退出登录时必填）",
		"renew GitHub device authorization (login only)":                                                                                           "重新进行 GitHub 设备码授权（仅限 login）",
		"unexpected Git authorization arguments":                                                                                                   "存在多余的 Git 授权参数",
		"GitHub authorization for background tasks: desktop Git credentials are not shared with the Windows service.":                              "后台任务需要单独进行 GitHub 授权：桌面 Git 登录凭据不会自动供 Windows 服务使用。",
		"After configuring and starting the Windows service, authorize GitHub before using private repositories or pushing:":                       "配置并启动 Windows 服务后，使用私有仓库或推送前请先完成 GitHub 授权：",
		"If the service is disabled, start it from an administrator terminal first, then authorize from the device owner's terminal:":              "如果服务尚未启用，请先在管理员终端启动，再回到设备属主的终端执行授权：",
		"Login displays a device code to approve in your browser. Credentials stay on this client and remain available after sign-out or restart.": "登录命令会显示设备码，请在浏览器中确认授权。凭据仅保存在本机，退出 Windows 登录或重启后仍可供后台任务使用。",
		"No GitHub account is authorized for the Windows service. Run remote-agent git login.\n":                                                   "Windows 服务尚未授权任何 GitHub 账号，请运行 remote-agent git login。\n",
		"GitHub authorization saved for Windows service tasks.\n":                                                                                  "已保存供 Windows 服务任务使用的 GitHub 授权。\n",
		"Selected GitHub account removed from the Windows service credential store.\n":                                                             "已从 Windows 服务凭据库移除选定的 GitHub 账号。\n",
		"use git login/status/logout":                                                                              "请使用 git login/status/logout",
		"git status does not accept an account or --force":                                                         "git status 不接受账号或 --force 参数",
		"git logout requires --username; use git status to list accounts":                                          "git logout 必须指定 --username；可先用 git status 查看账号",
		"invalid GitHub username":                                                                                  "无效的 GitHub 用户名",
		"install Git for Windows with Git Credential Manager, then retry remote-agent git login":                   "请安装包含 Git Credential Manager 的 Git for Windows，然后重试 remote-agent git login",
		"service Git authorization is unavailable; start or update RemoteAgent, then retry remote-agent git login": "后台 Git 授权服务不可用；请启动或更新 RemoteAgent，再重试 remote-agent git login",
		"configure and start the Windows client first, then run remote-agent git login":                            "请先配置并启动 Windows 客户端，再运行 remote-agent git login",
		"Git authorization config does not match the installed Windows service":                                    "Git 授权指定的配置与已安装 Windows 服务不一致",
		"install the Remote Agent MSI and start its Windows service first":                                         "请先安装 Remote Agent MSI 并启动其 Windows 服务",
		"only the configured device user or an administrator may manage service Git authorization":                 "只有已配置的设备属主或管理员可以管理后台 Git 授权",
		"Git authorization connection ended; retry remote-agent git status":                                        "Git 授权连接已断开；请重试 remote-agent git status 核对结果",
		"another Git authorization operation is in progress":                                                       "另一个 Git 授权操作正在进行",
		"service Git authorization is only needed on Windows; authenticate Git as the client user on this system":  "仅 Windows 服务需要此 Git 授权；请在当前系统以客户端运行用户完成普通 Git 登录",
		"GitHub login failed; check the authorization output and retry":                                            "GitHub 登录失败；请检查授权输出后重试",
		"GitHub status failed; check the authorization output and retry":                                           "GitHub 状态查询失败；请检查输出后重试",
		"GitHub logout failed; check the authorization output and retry":                                           "GitHub 退出登录失败；请检查输出后重试",
	} {
		messages[english] = chinese
	}
}
