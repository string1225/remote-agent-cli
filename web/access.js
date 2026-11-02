/* Human credential setup uses a separate relay and never creates a chat turn. */
(() => {
  const entry = node("button", "服务器与 GitHub"); entry.type = "button"; entry.role = "menuitem";
  $("device-menu").insertBefore(entry, $("device-menu-delete"));
  let menuDevice;
  const originalDeviceMenu = window.openDeviceMenu, originalProjectMenu = window.openProjectMenu;
  window.openDeviceMenu = (...args) => { menuDevice = args[1]; entry.hidden = false; entry.disabled = !menuDevice.online; originalDeviceMenu(...args); };
  window.openProjectMenu = (...args) => { entry.hidden = true; originalProjectMenu(...args); };
  const nav = node("button", "⌘ 服务器与 GitHub"); document.querySelector(".platform-nav").append(nav);
  const dialog = node("dialog"); dialog.id = "access-dialog";
  dialog.setAttribute("aria-labelledby", "access-title");
  dialog.innerHTML = `<div class="dialog-head"><h2 id="access-title">服务器与 GitHub</h2><button type="button" class="icon-button" aria-label="关闭">×</button></div>
    <label>执行客户端<select id="access-client"></select></label><p class="small muted">连接和授权保存在所选客户端，供它的后台任务使用。</p>
    <p id="access-message" role="status"></p><div id="access-content" hidden>
    <section aria-labelledby="access-servers-title"><h3 id="access-servers-title">SSH 服务器</h3><div id="access-servers"></div>
    <form id="access-server-form" autocomplete="off"><div class="access-fields">
    <label>连接别名<input name="name" placeholder="deploy-139" required maxlength="64" pattern="[A-Za-z0-9][A-Za-z0-9_.-]*"></label>
    <label>服务器 IP<input name="host" placeholder="192.0.2.1" required maxlength="64"></label>
    <label>SSH 账号<input name="user" placeholder="root" required maxlength="64" autocomplete="off"></label>
    <label>端口<input name="port" type="number" min="1" max="65535" value="22" required></label>
    <label class="access-password">密码<input name="password" type="password" required maxlength="4096" autocomplete="new-password"></label></div>
    <p class="small muted">密码仅用于本次初始化。客户端自动生成密钥、安装公钥并验证连接，保留原有授权。首次连接会记录主机指纹。</p>
    <button type="submit" class="primary">添加服务器并配置 SSH</button></form></section>
    <section aria-labelledby="access-github-title"><h3 id="access-github-title">GitHub 账号</h3><div id="access-github-accounts"></div>
    <form id="access-github-form"><label>GitHub 用户名（选填）<input name="username" maxlength="39" placeholder="留空后在 GitHub 页面选择账号" autocomplete="off"></label>
    <button type="submit" class="primary">发起设备授权</button></form>
    <div id="access-github-state" role="status"></div><button id="access-github-cancel" type="button" class="secondary" hidden>取消本次授权</button>
    <p class="small muted">在 GitHub 页面登录并输入设备码，无需在这里填写 GitHub 密码。关闭此窗口后客户端会继续等待授权。</p></section></div>`;
  document.body.append(dialog);
  let client, rpc, revision = 0, timer, ready = false;
  class AccessRPC {
    constructor(id) { this.id = id; this.pending = new Map(); }
    async connect() {
      if (this.ws?.readyState === WebSocket.OPEN) return;
      if (this.connecting) return this.connecting;
      this.connecting = new Promise((resolve, reject) => {
        const ws = new WebSocket(`${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}${basePath}/api/agents/${encodeURIComponent(this.id)}/connect`); this.ws = ws;
        let connected = false;
        const timeout = setTimeout(() => { reject(Error("客户端连接超时")); ws.close(); }, 15000);
        ws.onmessage = e => {
          try {
            const msg = JSON.parse(e.data);
            if (msg.type === "ready") { if (msg.transport !== "relay-v1") throw Error("请升级客户端"); connected = true; clearTimeout(timeout); this.ping = setInterval(() => { if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({type: "ping"})); }, 20000); resolve(); }
            else if (msg.type === "error") throw Error(msg.error || "连接失败");
            else if (msg.type === "event") {
              const event = msg.event, pending = this.pending.get(event?.id); if (!pending) return;
              if (event.type === "chunk") { pending.chunks.push(event.data); if (pending.chunks.length > 128) throw Error("响应过大"); return; }
              if (event.type === "result" && event.data?.chunked) { const chunks = pending.chunks.map(s => Uint8Array.from(atob(s), c => c.charCodeAt(0))); const bytes = new Uint8Array(chunks.reduce((n, s) => n + s.length, 0)); let offset = 0; for (const chunk of chunks) { bytes.set(chunk, offset); offset += chunk.length; } event.data = JSON.parse(new TextDecoder().decode(bytes)); }
              if (["result", "error"].includes(event.type)) { this.pending.delete(event.id); clearTimeout(pending.timer); event.type === "error" ? pending.reject(Error(event.error || "操作失败")) : pending.resolve(event.data); }
            }
          } catch (error) { reject(error); ws.close(); }
        };
        ws.onclose = () => { clearTimeout(timeout); clearInterval(this.ping); if (!connected) reject(Error("客户端暂未上线")); for (const p of this.pending.values()) { clearTimeout(p.timer); p.reject(Error("客户端连接已断开")); } this.pending.clear(); this.connecting = null; };
        ws.onerror = () => ws.close();
      });
      return this.connecting;
    }
    async call(type, body = {}) {
      await this.connect();
      return new Promise((resolve, reject) => {
        const id = crypto.randomUUID(), timer = setTimeout(() => { this.pending.delete(id); reject(Error("操作超时，请刷新检查结果后再重试")); }, 90000);
        this.pending.set(id, {resolve, reject, timer, chunks: []});
        this.ws.send(JSON.stringify({type: "request", request: {id, type, body}}));
      });
    }
    manage(body) { return this.call("access.call", body); }
    close() { clearInterval(this.ping); this.ws?.close(); }
  }
  const form = $("access-server-form"), githubForm = $("access-github-form");
  const current = stamp => dialog.open && stamp === revision && state.user;
  function message(value) { $("access-message").textContent = value; }
  function renderServers(servers) {
    const host = $("access-servers"); host.replaceChildren();
    if (!servers.length) host.append(node("p", "尚未添加服务器。", "muted"));
    for (const server of servers) {
      const card = node("article", undefined, "access-server"), actions = node("div", undefined, "access-actions");
      card.append(node("strong", server.name), node("p", `${server.user}@${server.host}:${server.port}`), node("p", server.fingerprint, "small muted"));
      card.append(node("code", `ssh -F "${server.configPath}" ${server.name}`));
      const test = node("button", "测试连接", "secondary"), remove = node("button", "移除", "secondary"); test.type = remove.type = "button";
      test.onclick = async () => { const stamp = revision; test.disabled = true; try { await rpc.manage({action: "servers.test", name: server.name}); if (current(stamp)) message("密钥登录验证通过"); } catch (e) { if (current(stamp)) message(e.message); } finally { test.disabled = false; } };
      remove.onclick = async () => { if (!confirm(`移除「${server.name}」的本机登记？密钥和配置将归档保留，远端公钥仍有效。`)) return; const stamp = revision; remove.disabled = true; try { await rpc.manage({action: "servers.remove", name: server.name}); if (current(stamp)) await loadServers(stamp); } catch (e) { if (current(stamp)) message(e.message); } finally { remove.disabled = false; } };
      actions.append(test, remove); card.append(actions); host.append(card);
    }
  }
  async function loadServers(stamp) { const servers = await rpc.manage({action: "servers.list"}); if (current(stamp)) renderServers(servers); }
  async function loadAccounts(stamp) {
    const accounts = await rpc.manage({action: "github.accounts"}); if (!current(stamp)) return;
    const host = $("access-github-accounts"); host.replaceChildren(); if (!accounts.length) host.append(node("p", "尚未授权 GitHub 账号。", "muted"));
    for (const username of accounts) { const row = node("p", undefined, "access-actions"), remove = node("button", "移除授权", "secondary"); remove.type = "button"; row.append(node("strong", username), remove); host.append(row); remove.onclick = async () => { if (!confirm(`移除 ${username} 在这个客户端保存的 GitHub 授权？`)) return; remove.disabled = true; try { await rpc.manage({action: "github.logout", username}); if (current(stamp)) await loadAccounts(stamp); } catch (e) { if (current(stamp)) message(e.message); } finally { remove.disabled = false; } }; }
  }
  function renderGitHub(status) {
    const host = $("access-github-state"); host.replaceChildren(node("p", status.message || "可以发起新的设备授权。"));
    if (status.code) { const code = node("code", status.code, "access-device-code"), link = node("a", "打开 GitHub 授权页面", "secondary"); link.href = "https://github.com/login/device"; link.target = "_blank"; link.rel = "noopener noreferrer"; host.append(code, link, node("p", `设备码有效至 ${new Date(status.expiresAt).toLocaleTimeString()}`, "small muted")); }
    const active = ["starting", "waiting"].includes(status.status); $("access-github-cancel").hidden = !active; githubForm.querySelector("button").disabled = active;
    return active;
  }
  async function watch(stamp) {
    clearTimeout(timer); if (!current(stamp) || !ready) return;
    try { const status = await rpc.manage({action: "github.status"}); if (!current(stamp)) return; if (renderGitHub(status)) timer = setTimeout(() => watch(stamp), 2000); else if (status.status === "authorized") await loadAccounts(stamp); }
    catch (e) { if (current(stamp)) { message(e.message); timer = setTimeout(() => watch(stamp), 3000); } }
  }
  async function selectClient() {
    const stamp = ++revision; ready = false; clearTimeout(timer); rpc?.close(); form.elements.password.value = ""; $("access-content").hidden = true;
    client = state.agents.find(d => d.id === $("access-client").value); if (!client) { message("请先添加并连接一个客户端。"); return; } if (!client.online) { message("请等待所选客户端上线。"); return; }
    rpc = new AccessRPC(client.id); message("正在读取客户端的服务器与账号状态……");
    try { const info = await rpc.call("authenticate"); if (!current(stamp)) return; if (!info.credentialManagement) { message("请先升级所选客户端，以启用服务器与 GitHub 管理。"); return; } ready = true; $("access-content").hidden = false; message(""); const results = await Promise.allSettled([loadServers(stamp), loadAccounts(stamp), watch(stamp)]); if (current(stamp)) for (const result of results) if (result.status === "rejected") message(result.reason.message); }
    catch (e) { if (current(stamp)) message(e.message); }
  }
  async function open(next = state.selected) {
    if (!state.user) return; $("device-menu").hidden = true; closeDrawers();
    const select = $("access-client"); select.replaceChildren(...state.agents.map(d => { const option = node("option", `${d.name} · ${d.online ? "在线" : "离线"}`); option.value = d.id; return option; }));
    if (next) select.value = next.id; if (!dialog.open) dialog.showModal(); await selectClient();
  }
  form.onsubmit = async event => {
    event.preventDefault(); if (!ready) return; const stamp = revision, button = form.querySelector("button"); button.disabled = true;
    const request = {action: "servers.add", name: form.elements.name.value, host: form.elements.host.value.trim(), user: form.elements.user.value, port: Number(form.elements.port.value), password: form.elements.password.value}; form.elements.password.value = "";
    message("正在生成密钥、安装公钥并验证连接……");
    try { await rpc.manage(request); request.password = ""; if (current(stamp)) { message("服务器已添加，SSH 密钥登录验证通过。"); await loadServers(stamp); } }
    catch (e) { if (current(stamp)) message(e.message); } finally { request.password = ""; button.disabled = false; }
  };
  githubForm.onsubmit = async event => { event.preventDefault(); if (!ready) return; const stamp = revision, button = githubForm.querySelector("button"); button.disabled = true; try { const status = await rpc.manage({action: "github.start", username: githubForm.elements.username.value.trim()}); if (current(stamp)) { renderGitHub(status); watch(stamp); } } catch (e) { if (current(stamp)) { message(e.message); button.disabled = false; } } };
  $("access-github-cancel").onclick = async () => { const stamp = revision; try { await rpc.manage({action: "github.cancel"}); if (current(stamp)) watch(stamp); } catch (e) { if (current(stamp)) message(e.message); } };
  $("access-client").onchange = selectClient; dialog.querySelector(".icon-button").onclick = () => dialog.close();
  dialog.addEventListener("close", () => { revision++; ready = false; clearTimeout(timer); rpc?.close(); form.elements.password.value = ""; $("access-github-state").replaceChildren(); });
  const originalSignOut = window.platformSignedOut; window.platformSignedOut = () => { originalSignOut?.(); if (dialog.open) dialog.close(); form.reset(); githubForm.reset(); };
  nav.onclick = () => open(); entry.onclick = () => open(menuDevice); window.openAccessSettings = open;
})();
