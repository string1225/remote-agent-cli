/* Account and token management stays separate from device/session transport. */
(() => {
  let user, gateway, tab = "setup", revision = 0;
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
    clearSecret();
    for (const id of ["token-list", "admin-user-list", "admin-token-list"]) $(id).replaceChildren();
    for (const id of ["account-form", "admin-create-form", "token-form"]) $(id).reset();
    dialog.close();
  };
  function installCommand(mode = "cli") {
    const url = server + "/install/" + mode + "-" + ({windows:"windows.ps1",macos:"macos.sh",linux:"linux.sh"})[platform()];
    return platform() === "windows"
      ? `& ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing ${psQuote(url)}).Content))`
      : `curl -fsSL ${shellQuote(url)} | bash`;
  }
  function updateCommands() {
    $("local-cli-command").value = installCommand();
    const token = $("token-secret").value;
    if (token) $("token-config-command").value = platform() === "windows"
      ? `${installCommand()}; ${psQuote(token)} | & "$env:USERPROFILE\\.remote-agent\\bin\\remote-agent.exe" codex setup --server ${psQuote(server)} --token-stdin`
      : `${installCommand()} && printf '%s' ${shellQuote(token)} | "$HOME/.remote-agent/bin/remote-agent" codex setup --server ${shellQuote(server)} --token-stdin`;
    const help = [
      ["remote-agent --help", "查看本地 CLI 命令"],
      [`remote-agent codex install --server ${server}`, "安装官方 Codex CLI，无需 Node.js"],
      [`remote-agent codex update --server ${server}`, "更新托管的 Codex CLI"],
      ["remote-agent codex setup", "已绑定设备：刷新本机代理 Token 并合并 TOML"],
      ["remote-agent app install", "通过系统渠道安装 Codex App"],
      ["remote-agent app upgrade", "通过 Microsoft Store / Homebrew 升级 Codex App"],
      ["remote-agent start", "启动远程 Agent 并启用自启动（需先绑定设备）"],
      ["remote-agent stop", "停止远程 Agent 并移除自启动，保留绑定"],
      ["remote-agent run", "在当前终端运行远程 Agent"],
      ["remote-agent status", "查看设备绑定、权限和本机服务"],
      [`remote-agent update --server ${server}`, "更新本地 remote-agent CLI"],
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
  if (/Mac/i.test(navigator.platform)) $("setup-platform").value = "macos";
  else if (/Linux/i.test(navigator.platform)) $("setup-platform").value = "linux";
  dialog.addEventListener("close", () => {
    revision++;
    clearSecret();
    $("account-form").elements.currentPassword.value = "";
    $("account-form").elements.password.value = "";
    $("admin-create-form").elements.password.value = "";
    $("admin-user-list").replaceChildren();
  });
  async function show(next) {
    if (!user) user = state.user;
    if (!user || (next === "admin" && !user.admin)) return;
    tab = next;
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
  async function loadTokens() {
    const stamp = ++revision, tokens = await api("/api/tokens");
    if (stamp !== revision || !dialog.open || !user) return;
    renderTokens($("token-list"), tokens, false);
  }
  function renderTokens(host, tokens, admin) {
    if (!tokens.length) { host.replaceChildren(node("p", "暂无 Token。", "muted")); return; }
    host.replaceChildren(...tokens.map((token) => {
      const card = node("article", undefined, "platform-card"), info = node("div");
      info.append(node("strong", token.name), node("p", `${token.prefix}… · ${token.revoked ? "已撤销" : "可用"}`, "small"), node("p", `创建 ${date(token.createdAt)} · 最近使用 ${date(token.lastUsedAt)}`, "small muted"));
      if (admin) info.append(node("p", "账号 ID：" + token.userId, "small muted"));
      card.append(info);
      const revoke = node("button", token.revoked ? "已撤销" : "撤销", "secondary"); revoke.disabled = token.revoked;
      revoke.onclick = async () => {
        if (!confirm(`撤销「${token.name}」？使用它的设备将无法发起新的模型请求。`)) return;
        revoke.disabled = true;
        try { await api((admin ? "/api/admin/tokens/" : "/api/tokens/") + token.id, "DELETE"); if (admin) await loadAdmin(); else await loadTokens(); }
        catch (e) { revoke.disabled = false; notice(e.message); }
      };
      card.append(revoke); return card;
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
  if (state.user) window.platformSignedIn(state.user);
})();
