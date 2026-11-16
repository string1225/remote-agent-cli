/* Account and token management stays separate from device/session transport. */
(() => {
  let user, gateway, tab = "setup", adminTab = "users", revision = 0;
  let codexPoll, codexLoginID, codexLoading = false;
  const dialog = $("platform-dialog");
  const server = location.origin + basePath;
  const shellQuote = (s) => "'" + s.replaceAll("'", "'\"'\"'") + "'";
  const psQuote = (s) => "'" + s.replaceAll("'", "''") + "'";
  const platform = () => $("setup-platform").value;
  const date = (value) => value ? new Date(value * 1000).toLocaleString() : "尚未使用";
  function clearSecret() {
    $("token-secret").value = "";
    $("token-config-command").value = "";
    $("token-secret-box").hidden = true;
  }
  function clearGuide() {
    guideRevision++;
    $("admin-guide-text").value = "";
    $("admin-guide-box").hidden = true;
  }
  let guideRevision = 0;
  window.platformSignedIn = (account) => {
    user = account;
    $("admin-entry").hidden = !user.admin;
    $("admin-tab").hidden = !user.admin;
    $("account-form").elements.username.value = user.username;
  };
  window.platformSignedOut = () => {
    user = null;
    revision++;
    clearTimeout(codexPoll);
    $("codex-login").hidden = true;
    $("codex-login-code").value = "";
    clearSecret();
    clearGuide();
    selectAdminTab("users");
    for (const id of ["token-list", "admin-user-list", "admin-token-list"]) $(id).replaceChildren();
    for (const id of ["account-form", "admin-create-form", "token-form"]) $(id).reset();
    dialog.close();
  };
  function installCommand(mode = "cli") {
    if (mode === "cli") return platform() === "windows" ? "remote-agent codex setup" : '"$HOME/.remote-agent/bin/remote-agent" codex setup';
    const url = server + "/install/" + mode + "-" + ({macos:"macos.sh",linux:"linux.sh"})[platform()];
    return `curl -fsSL ${shellQuote(url)} | bash`;
  }
  function updateCommands() {
    $("local-cli-command").value = installCommand();
    const token = $("token-secret").value;
    if (token) $("token-config-command").value = platform() === "windows"
      ? `remote-agent codex setup --server ${psQuote(server)} --token-prompt`
      : `"$HOME/.remote-agent/bin/remote-agent" codex setup --server ${shellQuote(server)} --token-prompt`;
    const windows = platform() === "windows";
    $("setup-git-auth").hidden = !windows;
    $("local-command-label").firstChild.textContent = "刷新已绑定设备的 Codex 配置";
    $("copy-local-cli").textContent = "复制配置命令";
    $("setup-install-description").textContent = windows ? "MSI 激活页可分别选择安装 Codex CLI 和配置 Codex；配置使用绑定账号的模型入口，并备份原配置。之后可用下方命令刷新。" : "设备绑定会自动申请所属账号的 Token 并配置 Codex。再次刷新时，使用本机绑定识别账号，终端会显示账号名称。";
    $("remote-install-help").textContent = windows
      ? "首次安装 MSI（x64 / ARM64），在安装向导最后一页粘贴网页设备绑定码即可激活，可选择随系统启动、网页代理和 Codex 配置。以后用 remote-agent update 更新业务程序。"
      : platform() === "macos" ? "macOS 使用 LaunchAgent，登录后自动启动。请在目标用户的终端运行绑定命令。" : "Linux 使用 systemd 用户服务；启用 linger 后可开机启动并在退出登录后继续在线。请勿在整条安装命令前加 sudo。";
    $("msi-downloads").replaceChildren();
    if (windows) for (const arch of ["amd64", "arm64"]) {
      const link = node("a", `下载 ${arch === "amd64" ? "x64" : "ARM64"} MSI`);
      link.href = server + `/downloads/remote-agent-windows-${arch}.msi?v=${Date.now()}`;
      $("msi-downloads").append(link, document.createTextNode("　"));
    }
    const help = [
      ["remote-agent --help", "查看本地 CLI 命令"],
      [codexInstallCommand("install", server), "安装官方 Codex CLI，无需 Node.js；默认站点可省略 --server"],
      [codexInstallCommand("update", server), "更新托管的 Codex CLI；默认使用设备绑定站点"],
      ["remote-agent codex setup", "使用设备绑定的账号申请 Token，显示所属账号并更新配置"],
      ["codex", "配置成功后，在终端启动 Codex CLI"],
      ["remote-agent app install", windows ? "打开 Windows 商店安装 Codex App" : "通过 Homebrew 安装 Codex App"],
      ["remote-agent app upgrade", "通过 Microsoft Store / Homebrew 升级 Codex App"],
      ["remote-agent start", "启动远程 Agent 并启用自启动（需先绑定设备）"],
      ["remote-agent stop", "停止并禁用远程自启动，保留绑定（Windows MSI 需管理员）"],
      ["remote-agent run", "在当前终端运行远程 Agent"],
      ["remote-agent status", "查看设备绑定、权限和本机服务"],
      [`remote-agent update --server ${server}`, windows ? "通过固定更新服务校验签名并切换业务版本，无需 MSI / UAC，保留绑定和数据" : "更新本地 remote-agent CLI"],
    ];
    $("cli-help").replaceChildren(...help.map(([command, description]) => {
      const row = node("div"); row.append(node("code", command), node("span", description)); return row;
    }));
  }
  async function copy(id) {
    try { await navigator.clipboard.writeText($(id).value); notice("已复制"); }
    catch { $(id).select(); notice("请按 Ctrl / ⌘ + C 复制"); }
  }
  for (const [button, target] of [["copy-local-cli","local-cli-command"],["copy-token-secret","token-secret"],["copy-token-command","token-config-command"]]) $(button).onclick = () => copy(target);
  $("setup-platform").onchange = updateCommands;
  $("setup-add-device").onclick = () => { dialog.close(); showEnroll(); };
  if (/Mac/i.test(navigator.platform)) $("setup-platform").value = "macos";
  else if (/Linux/i.test(navigator.platform)) $("setup-platform").value = "linux";
  dialog.addEventListener("close", () => {
    revision++;
    clearTimeout(codexPoll);
    $("codex-login").hidden = true;
    $("codex-login-code").value = "";
    clearSecret();
    clearGuide();
    $("token-list").replaceChildren();
    $("account-form").elements.currentPassword.value = "";
    $("account-form").elements.password.value = "";
    $("admin-create-form").elements.password.value = "";
    $("admin-user-list").replaceChildren();
  });
  async function show(next) {
    if (!user) user = state.user;
    if (!user || (next === "admin" && !user.admin)) return;
    tab = next;
    revision++;
    clearTimeout(codexPoll);
    clearGuide();
    clearSecret();
    for (const input of $("token-list").querySelectorAll('input[type="password"]')) input.value = "";
    for (const name of ["setup","tokens","account","admin"]) $("platform-" + name).hidden = name !== tab;
    for (const button of dialog.querySelectorAll("[data-platform]")) button.setAttribute("aria-current", String(button.dataset.platform === tab));
    closeDrawers();
    if (!dialog.open) dialog.showModal();
    try {
      if (tab === "setup") {
        gateway = await api("/api/codex");
        $("gateway-state").textContent = gateway.enabled ? `代理入口：${gateway.baseUrl} · 默认模型 ${gateway.model}` : "管理员尚未启用模型代理；本地 CLI 和远程设备功能仍可使用。";
        updateCommands();
      } else if (tab === "tokens") { updateCommands(); await loadTokens(); }
      else if (tab === "account") $("account-form").elements.username.value = user.username;
      else await loadAdmin();
    } catch (e) { notice(e.message); }
  }
  for (const button of document.querySelectorAll("[data-platform]")) button.onclick = () => show(button.dataset.platform);
  function selectAdminTab(next) {
    adminTab = next;
    revision++;
    clearTimeout(codexPoll);
    clearGuide();
    for (const button of dialog.querySelectorAll("[data-admin]")) {
      const active = button.dataset.admin === next;
      button.setAttribute("aria-selected", String(active));
      button.setAttribute("aria-current", String(active));
      button.tabIndex = active ? 0 : -1;
      $(button.getAttribute("aria-controls")).hidden = !active;
    }
  }
  const adminTabs = [...dialog.querySelectorAll("[data-admin]")];
  for (const button of adminTabs) {
    button.onclick = () => {
      if (!user?.admin || tab !== "admin") return;
      selectAdminTab(button.dataset.admin);
      loadAdmin().catch(e => notice(e.message));
    };
    button.onkeydown = (event) => {
      const index = adminTabs.indexOf(button);
      const next = ({ArrowRight:(index + 1) % adminTabs.length, ArrowLeft:(index + adminTabs.length - 1) % adminTabs.length, Home:0, End:adminTabs.length - 1})[event.key];
      if (next === undefined) return;
      event.preventDefault(); adminTabs[next].focus(); adminTabs[next].click();
    };
  }
  async function loadTokens() {
    const stamp = ++revision, tokens = await api("/api/tokens");
    if (stamp !== revision || !dialog.open || !user) return;
    renderTokens($("token-list"), tokens, false);
  }
  function renderTokens(host, tokens, admin) {
    if (!tokens.length) { host.replaceChildren(node("p", "暂无 Token。", "muted")); return; }
    host.replaceChildren(...tokens.map((token) => {
      const card = node("article", undefined, "platform-card"), info = node("div"), actions = node("div", undefined, "token-actions");
      const title = node("strong", token.name); title.append(node("span", token.revoked ? "已撤销" : "可用", "token-state" + (token.revoked ? " revoked" : "")));
      info.append(title, node("p", `${token.prefix}…${token.agentId ? " · 设备自动申请" : ""}`, "small muted"), node("p", `创建 ${date(token.createdAt)}`, "small muted"), node("p", `最近使用 ${date(token.lastUsedAt)}`, "small muted"));
      if (admin) info.append(node("p", "账号 ID：" + token.userId, "small muted"));
      card.append(info);
      if (!admin && !token.revoked) {
        const reveal = node("button", "复制 Token", "secondary");
        reveal.disabled = !token.copyable;
        if (!token.copyable) reveal.title = "旧 Token 未保存原文，可创建新的 Token";
        reveal.onclick = async () => {
          const stamp = revision, account = user.id;
          reveal.disabled = true;
          try {
            const result = await api(`/api/tokens/${token.id}/reveal`, "POST", {});
            if (!dialog.open || !user || user.id !== account || revision !== stamp || tab !== "tokens") return;
            try { await navigator.clipboard.writeText(result.token); notice("Token 已复制"); }
            catch { if (dialog.open && user?.id === account && revision === stamp && tab === "tokens") { clearSecret(); $("token-secret").value = result.token; $("token-secret-box").hidden = false; updateCommands(); $("token-secret").select(); notice("请按 Ctrl / ⌘ + C 复制"); } }
          } catch (e) { notice(e.message); } finally { reveal.disabled = false; }
        };
        actions.append(reveal);
      }
      const revoke = node("button", token.revoked ? "已撤销" : "撤销", "secondary token-revoke"); revoke.disabled = token.revoked;
      revoke.onclick = async () => {
        if (!confirm(`撤销「${token.name}」？使用它的设备将无法发起新的模型请求。`)) return;
        revoke.disabled = true;
        try { await api((admin ? "/api/admin/tokens/" : "/api/tokens/") + token.id, "DELETE"); clearSecret(); if (admin) await loadAdmin(); else await loadTokens(); }
        catch (e) { revoke.disabled = false; notice(e.message); }
      };
      actions.append(revoke); card.append(actions); return card;
    }));
  }
  $("token-form").onsubmit = async (e) => {
    e.preventDefault(); const button = e.target.querySelector("button"); button.disabled = true;
    try {
      const stamp = revision, account = user.id;
      const result = await api("/api/tokens", "POST", {name:e.target.elements.name.value});
      if (!dialog.open || !user || user.id !== account || revision !== stamp) return;
      clearSecret(); $("token-secret").value = result.token; $("token-secret-box").hidden = false;
      updateCommands(); e.target.reset(); await loadTokens();
    } catch (e) { notice(e.message); } finally { button.disabled = false; }
  };
  $("account-form").onsubmit = async (e) => {
    e.preventDefault(); const button = e.target.querySelector("button"); button.disabled = true;
    try {
      const values = Object.fromEntries(new FormData(e.target)); if (!values.password) delete values.password;
      await api("/api/account", "PATCH", values); signedOut(); notice("账号已更新，请重新登录。");
    } catch (e) { notice(e.message); } finally { button.disabled = false; }
  };
  function accountGuide(account, mode, token) {
    const credentials = `登录地址：${server}/\n用户名：${account.username}\n初始密码：<请填写初始密码>\n首次登录后，请在「账号设置」修改密码。`;
    if (mode === "web") return `【网页版使用指引】\n${credentials}\n\n1. 用浏览器打开上述地址并登录。\n2. 在左侧「客户端」添加设备，选择实际使用的 Windows 或 macOS。Windows 先下载安装对应架构的 MSI，再打开新的 PowerShell；macOS 打开终端。将网页生成的绑定命令复制到这台电脑执行，按提示完成安装和绑定。\n3. 设备上线后，添加项目，选择这台电脑上的工作目录和 Codex / Qoder CLI。Codex 会自动配置；如提示缺少 CLI，按页面安装指引完成后重新连接。\n4. 打开项目下的 Web 会话，输入需求即可开始。Enter 发送，Shift+Enter 换行；可以创建多个会话。任务在绑定的电脑上执行，使用时请保持该电脑开机联网。`;
    return `【Endpoint / Codex 本地使用指引（无需网页登录）】\n账号：${account.username}\nAPI Token：${token}\n\n一、下载安装客户端并填写 Token\nWindows：\nx64：${server}/downloads/remote-agent-windows-amd64.msi\nARM64：${server}/downloads/remote-agent-windows-arm64.msi\n安装最后一页，在「设备绑定码 / Endpoint Token」填入上述 Token，保留「安装 Codex CLI」和「配置 Codex」勾选，点击「激活 / 应用选项」。可选「安装 Codex App」，默认不勾选；选择后会打开 Microsoft Store，由你完成安装。\nToken 仅配置本机 Codex 的模型账号，不新增远程设备，也不需要网页登录。已有远程绑定保持不变；开机启动和网页代理选项仅用于设备绑定码激活。\n\nmacOS：\ncurl -fsSL ${shellQuote(server + "/install/cli-macos.sh")} | bash\n安装后运行：\n"$HOME/.remote-agent/bin/remote-agent" codex setup --server ${shellQuote(server)} --token-prompt\n按提示粘贴上述 Token，等待 CLI 安装和配置完成。\n\n二、仅使用 Endpoint\nBase URL：${server}/codex/v1\n接口类型：OpenAI Responses API（/responses）\nAPI Key：${token}\n模型以 ${server}/codex/v1/models 返回的列表为准。\n\n三、开始使用与恢复\n打开新的终端，进入项目目录后运行 codex。Codex App 请完全退出并重开，新建本地会话。配置前会备份原 config.toml；配置失败可恢复备份后重试。\n安装末页无法打开时，可在新的 PowerShell 运行：\nremote-agent codex setup --server ${psQuote(server)} --token-prompt\n按提示粘贴上述 Token。\n此指引含个人 Token，请仅私下交给该账号使用者；可在「我的 Token」或管理员 Token 管理中撤销。`;
  }
  async function copyAccountGuide(account, mode, button) {
    const stamp = revision, actor = user?.id;
    if (!actor || !user.admin || account.disabled) return;
    if (button) button.disabled = true;
    clearGuide();
    const guideStamp = guideRevision;
    const current = () => dialog.open && tab === "admin" && adminTab === "users" && user?.id === actor && user.admin && revision === stamp && guideRevision === guideStamp;
    try {
      const result = mode === "endpoint" ? await api(`/api/admin/users/${account.id}/endpoint-guide`, "POST", {}) : null;
      if (!current()) return;
      if (mode === "endpoint" && !result?.token) throw Error("服务器未返回指引 Token，请刷新后重试");
      const text = accountGuide(account, mode, result?.token);
      try {
        await navigator.clipboard.writeText(text);
        if (current()) notice(mode === "web" ? "网页版指引已复制，请填写初始密码后发送" : "Endpoint 指引已复制，含专用 Token，请仅交给该账号使用者");
      } catch {
        if (!current()) return;
        $("admin-guide-text").value = text;
        $("admin-guide-box").hidden = false;
        $("admin-guide-text").focus(); $("admin-guide-text").select();
        notice(mode === "web" ? "请按 Ctrl / ⌘ + C 复制，并填写初始密码后发送" : "请按 Ctrl / ⌘ + C 复制，指引含专用 Token");
      }
    } catch (e) { if (current()) notice(e.message); }
    finally { if (button) button.disabled = account.disabled; }
  }
  $("admin-guide-close").onclick = clearGuide;

  async function loadAdmin() {
    if (adminTab === "codex") return loadCodexAccounts();
    if (adminTab === "tokens") {
      const stamp = ++revision, tokens = await api("/api/admin/tokens");
      if (stamp !== revision || !dialog.open || tab !== "admin" || !user?.admin) return;
      renderTokens($("admin-token-list"), tokens, true);
      return;
    }
    const stamp = ++revision;
    const users = await api("/api/admin/users?search=" + encodeURIComponent($("admin-search-form").elements.search.value));
    if (stamp !== revision || !dialog.open || tab !== "admin" || !user?.admin) return;
    $("admin-user-list").replaceChildren(...users.map((account) => {
      const card = node("article", undefined, "platform-card admin-card"), heading = node("div", undefined, "admin-account-heading"), info = node("div"), guides = node("div", undefined, "admin-guide-actions");
      info.append(node("strong", account.username), node("p", `${account.admin ? "管理员" : "普通用户"} · ${account.disabled ? "已禁用" : "正常"} · ${account.deviceCount} 台设备`, "small muted"));
      for (const [mode, label] of [["web","复制网页版指引"],["endpoint","复制 Endpoint / Codex 指引"]]) {
        const button = node("button", label, "secondary"); button.type = "button";
        button.setAttribute("aria-label", `为 ${account.username} ${label}`);
        button.disabled = account.disabled;
        button.onclick = () => copyAccountGuide(account, mode, button);
        guides.append(button);
      }
      heading.append(info, guides);
      card.append(heading);
      const form = node("form", undefined, "admin-edit");
      for (const [key, label] of [["admin","管理员"],["disabled","禁用账号"]]) {
        const row = node("label", undefined, "check-label"), checkbox = node("input"); checkbox.type = "checkbox"; checkbox.name = key; checkbox.checked = account[key]; row.append(checkbox, document.createTextNode(label)); form.append(row);
      }
      const password = node("input"); password.type = "password"; password.name = "password"; password.placeholder = "重置密码（留空不改）"; password.autocomplete = "new-password"; password.minLength = 6; password.maxLength = 72; password.setAttribute("aria-label", "重置 " + account.username + " 的密码");
      const submit = node("button", "保存", "secondary"); submit.type = "submit"; form.append(password, submit);
      form.onsubmit = async (e) => {
        e.preventDefault(); submit.disabled = true;
        const body = {admin:form.elements.admin.checked, disabled:form.elements.disabled.checked}; if (password.value) body.password = password.value;
        try { await api("/api/admin/users/" + account.id, "PATCH", body); password.value = ""; if (account.id === user.id && (!body.admin || body.disabled || body.password)) { signedOut(); notice("账号权限已更新，请重新登录。"); } else { notice("账号已更新"); await loadAdmin(); } }
        catch (e) { notice(e.message); } finally { submit.disabled = false; }
      };
      card.append(form); return card;
    }));
    if (!users.length) $("admin-user-list").append(node("p", "没有匹配的账号。", "muted"));
  }
  $("admin-search-form").onsubmit = (e) => { e.preventDefault(); loadAdmin().catch(e => notice(e.message)); };
  $("admin-create-form").onsubmit = async (e) => {
    e.preventDefault(); const button = e.target.querySelector("button"); button.disabled = true;
    try {
      await api("/api/admin/users", "POST", {username:e.target.elements.username.value,password:e.target.elements.password.value,admin:e.target.elements.admin.checked});
      e.target.reset(); notice("账号已创建"); await loadAdmin();
    } catch (e) { notice(e.message); } finally { button.disabled = false; }
  };
  async function loadCodexAccounts() {
    if (codexLoading || !user?.admin || !dialog.open || tab !== "admin" || adminTab !== "codex") return;
    codexLoading = true;
    clearTimeout(codexPoll);
    try {
      const data = await api("/api/admin/codex/accounts");
      if (!dialog.open || tab !== "admin" || adminTab !== "codex" || !user?.admin) return;
      $("codex-accounts-status").textContent = data.enabled ? "账号凭据仅存 139 服务器；授权使用 OpenAI 官方设备码流程。" : "尚未配置服务器账号管理，请部署账号管理网关及专用控制凭据。";
      $("codex-account-form").hidden = !data.enabled;
      $("codex-account-list").replaceChildren(...(data.accounts || []).map(account => {
        const card = node("article", undefined, "platform-card"), info = node("div");
        info.append(node("strong", account.name + (account.active ? " · 当前代理账号" : "")), node("p", ({ready:"已登录",login_required:"需要重新授权",unknown:"状态暂不可用"})[account.state] || "未知", "small"));
        if (account.accountId) info.append(node("p", "账号 ID：" + account.accountId, "small muted"));
        if (account.expiresAt > 0) info.append(node("p", "当前凭据有效期至 " + date(account.expiresAt) + " · 使用时自动续期", "small muted"));
        card.append(info);
        const reauth = node("button", "重新授权", "secondary");
        reauth.disabled = data.login?.state === "pending";
        reauth.onclick = () => startCodexLogin(account.id, account.name);
        card.append(reauth);
        if (!account.active) {
          const activate = node("button", "设为当前账号", "secondary");
          activate.disabled = account.state !== "ready";
          activate.onclick = async () => { activate.disabled = true; try { await api("/api/admin/codex/activate", "POST", {id:account.id}); await loadCodexAccounts(); } catch (e) { notice(e.message); activate.disabled = false; } };
          card.append(activate);
        }
        return card;
      }));
      const login = data.login;
      codexLoginID = login?.id;
      $("codex-login").hidden = !login;
      $("codex-login-code").value = login?.userCode || "";
      const url = $("codex-login-url"); url.removeAttribute("href"); url.textContent = "";
      if (login?.verificationUrl) { const target = new URL(login.verificationUrl); if (target.protocol !== "https:" || !["auth.openai.com","chatgpt.com"].includes(target.hostname)) throw Error("授权地址无效"); url.href = target.href; url.textContent = "打开官方授权页面 ↗"; }
      $("codex-login-state").textContent = login?.state === "pending" ? `等待授权 · 最迟 ${date(login.expiresAt)} 结束` : login?.state === "completed" ? "授权已完成。可在上方选择此账号供代理使用。" : login?.state === "cancelled" ? "本次授权已取消。" : login?.error || "";
      $("cancel-codex-login").hidden = login?.state !== "pending";
      $("copy-codex-code").hidden = !login?.userCode;
      if (login?.state === "pending") codexPoll = setTimeout(() => loadCodexAccounts().catch(e => notice(e.message)), 3000);
    } finally { codexLoading = false; }
  }
  async function startCodexLogin(id, name) {
    const button = $("codex-account-form").querySelector("button"); button.disabled = true;
    $("codex-accounts-status").textContent = "正在向 Codex 获取设备授权码…";
    try { await api("/api/admin/codex/login", "POST", {id,name}); await loadCodexAccounts(); }
    catch (e) { notice(e.message); $("codex-accounts-status").textContent = e.message; }
    finally { button.disabled = false; }
  }
  $("codex-account-form").onsubmit = e => { e.preventDefault(); startCodexLogin("", e.target.elements.name.value); };
  $("refresh-codex-accounts").onclick = () => loadCodexAccounts().catch(e => notice(e.message));
  $("copy-codex-code").onclick = () => copy("codex-login-code");
  $("cancel-codex-login").onclick = async () => { try { await api("/api/admin/codex/cancel", "POST", {id:codexLoginID}); await loadCodexAccounts(); } catch (e) { notice(e.message); } };
  if (state.user) window.platformSignedIn(state.user);
})();
