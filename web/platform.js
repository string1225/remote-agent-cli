/* Account and token management stays separate from device/session transport. */
(() => {
  let user, gateway, tab = "setup", revision = 0;
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
    $("local-command-label").firstChild.textContent = "刷新已绑定设备的 Codex 配置";
    $("copy-local-cli").textContent = "复制配置命令";
    $("setup-install-description").textContent = "设备绑定会自动申请所属账号的 Token 并配置 Codex。再次刷新时，使用本机绑定识别账号，终端会显示账号名称。";
    $("remote-install-help").textContent = windows
      ? "首次安装 MSI（x64 / ARM64），再打开新的 PowerShell 执行当前账号的设备绑定命令。以后用 remote-agent update 更新业务程序，无需重复安装 MSI 或管理员授权。"
      : platform() === "macos" ? "macOS 使用 LaunchAgent，登录后自动启动。请在目标用户的终端运行绑定命令。" : "Linux 使用 systemd 用户服务；启用 linger 后可开机启动并在退出登录后继续在线。请勿在整条安装命令前加 sudo。";
    $("msi-downloads").replaceChildren();
    if (windows) for (const arch of ["amd64", "arm64"]) {
      const link = node("a", `下载 ${arch === "amd64" ? "x64" : "ARM64"} MSI`);
      link.href = server + `/downloads/remote-agent-windows-${arch}.msi?v=${Date.now()}`;
      $("msi-downloads").append(link, document.createTextNode("　"));
    }
    const help = [
      ["remote-agent --help", "查看本地 CLI 命令"],
      [`remote-agent codex install --server ${server}`, "安装官方 Codex CLI，无需 Node.js"],
      [`remote-agent codex update --server ${server}`, "更新托管的 Codex CLI"],
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
      else { await loadAdmin(); await loadCodexAccounts(); }
    } catch (e) { notice(e.message); }
  }
  for (const button of document.querySelectorAll("[data-platform]")) button.onclick = () => show(button.dataset.platform);
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
  async function loadAdmin() {
    const stamp = ++revision;
    const [users, tokens] = await Promise.all([api("/api/admin/users?search=" + encodeURIComponent($("admin-search-form").elements.search.value)), api("/api/admin/tokens")]);
    if (stamp !== revision || !dialog.open || !user?.admin) return;
    $("admin-user-list").replaceChildren(...users.map((account) => {
      const card = node("article", undefined, "platform-card admin-card"), heading = node("div");
      heading.append(node("strong", account.username), node("p", `${account.admin ? "管理员" : "普通用户"} · ${account.disabled ? "已禁用" : "正常"} · ${account.deviceCount} 台设备`, "small muted"));
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
    renderTokens($("admin-token-list"), tokens, true);
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
    if (codexLoading || !user?.admin || !dialog.open || tab !== "admin") return;
    codexLoading = true;
    clearTimeout(codexPoll);
    try {
      const data = await api("/api/admin/codex/accounts");
      if (!dialog.open || tab !== "admin" || !user?.admin) return;
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
