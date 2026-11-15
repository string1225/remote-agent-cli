/* Phone preferences live in the workbench; only device actions cross to Android. */
(() => {
  if (!navigator.userAgent.includes("RemoteAgentAndroid/")) return;
  document.documentElement.classList.add("android-app");
  $("new-agent").disabled = true; $("empty-add").disabled = true;
  // Embedded todo sessions share the host, but cannot issue native commands.
  if (window.top !== window) return;
  const integrated = navigator.userAgent.includes("RemoteAgentNative/1");
  const stylesheet = document.createElement("link");
  stylesheet.rel = "stylesheet"; stylesheet.href = new URL("./android-app.css", document.baseURI); document.head.append(stylesheet);
  const icon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m9 3-.6 2.4-2 .9L4.2 5.6 2 9.4l1.8 1.7v1.8L2 14.6l2.2 3.8 2.2-.7 2 .9L9 21h6l.6-2.4 2-.9 2.2.7 2.2-3.8-1.8-1.7v-1.8L22 9.4l-2.2-3.8-2.2.7-2-.9L15 3Z"/><circle cx="12" cy="12" r="3"/></svg>';
  let host, dialog;
  function command(action, params) {
    const query = new URLSearchParams(params).toString();
    location.href = `remoteagent://${action}${query ? `?${query}` : ""}`;
  }
  function render() {
    if (!dialog) return;
    $("phone-version").textContent = host ? `版本 ${host.version}` : "正在读取手机设置…";
    $("phone-update-status").textContent = host?.updateStatus || "";
    $("phone-check-update").disabled = !host || host.updateBusy;
    $("phone-check-update").textContent = host?.updateBusy ? "正在检查 / 下载…" : "检查更新";
    $("phone-auto-update").disabled = !host;
    $("phone-auto-update").checked = !!host?.autoUpdate;
    const site = $("phone-site");
    if (host && !site.dataset.edited) site.value = host.site;
    $("phone-save-site").disabled = !host;
  }
  function openSettings() {
    if (!integrated) { location.href = "remoteagent://settings"; return; }
    if (!dialog) {
      dialog = document.createElement("dialog"); dialog.id = "phone-settings-dialog"; dialog.className = "phone-settings-dialog";
      dialog.setAttribute("aria-labelledby", "phone-settings-title");
      dialog.innerHTML = `<div class="dialog-head"><div><p class="phone-eyebrow">REMOTE AGENT · ANDROID</p><h2 id="phone-settings-title">手机设置</h2></div><button type="button" class="icon-button" aria-label="关闭手机设置">×</button></div>
        <section class="phone-settings-section" aria-labelledby="phone-app-title"><div class="phone-app-heading"><span class="phone-app-icon">${icon}</span><div><h3 id="phone-app-title">Remote Agent</h3><p id="phone-version" class="small muted"></p></div></div>
          <button type="button" id="phone-check-update" class="secondary full">检查更新</button><p id="phone-update-status" class="phone-status small muted" role="status"></p>
          <label class="phone-toggle"><span><strong>自动检查更新</strong><small>启动时检查并下载，安装由你确认</small></span><input id="phone-auto-update" type="checkbox" role="switch"></label></section>
        <section class="phone-settings-section" aria-labelledby="phone-site-title"><h3 id="phone-site-title">工作台站点</h3><form id="phone-site-form"><label for="phone-site">HTTPS 地址</label><input id="phone-site" type="url" required autocomplete="off" autocapitalize="off" spellcheck="false" inputmode="url"><p class="small muted">切换站点后需要重新登录。</p><p id="phone-site-error" class="phone-error small" role="alert" hidden></p><button id="phone-save-site" type="submit" class="secondary full">保存站点</button></form></section>
        <section class="phone-settings-section phone-actions"><button type="button" id="phone-appearance" class="secondary">界面主题<span aria-hidden="true">›</span></button><button type="button" id="phone-reload" class="secondary">重新加载工作台<span aria-hidden="true">↻</span></button></section>
        <p class="small muted phone-footnote">手机与电脑共用项目和会话。退出手机，电脑任务继续运行。</p>`;
      dialog.querySelector(".dialog-head button").onclick = () => dialog.close();
      dialog.querySelector("#phone-check-update").onclick = () => { $("phone-check-update").disabled = true; command("update"); };
      dialog.querySelector("#phone-auto-update").onchange = event => command("auto-update", {enabled:event.target.checked});
      dialog.querySelector("#phone-site").oninput = event => { event.target.dataset.edited = "true"; $("phone-site-error").hidden = true; };
      dialog.querySelector("#phone-site-form").onsubmit = event => {
        event.preventDefault();
        try {
          const site = $("phone-site").value.trim(), url = new URL(site);
          if (url.protocol !== "https:" || !url.hostname || url.username || url.password || site.includes("?") || site.includes("#")) throw Error();
          $("phone-site-error").hidden = true; command("site", {url:site});
        } catch { $("phone-site-error").textContent = "请输入完整 HTTPS 地址，不包含账号、查询参数或片段。"; $("phone-site-error").hidden = false; $("phone-site").focus(); }
      };
      dialog.querySelector("#phone-appearance").onclick = () => window.openAppearanceSettings();
      dialog.querySelector("#phone-reload").onclick = () => command("reload");
      document.body.append(dialog);
    }
    closeDrawers(); render();
    if (!dialog.open) dialog.showModal();
    command("state");
  }
  const entry = node("button", integrated ? "手机设置" : "↓ 安卓更新与设置");
  entry.type = "button"; entry.id = "android-update-entry"; entry.onclick = openSettings;
  document.querySelector(".platform-nav").append(entry);
  if (!integrated) return;
  const header = node("button", undefined, "icon-button phone-settings-entry");
  header.id = "phone-settings-entry"; header.type = "button"; header.innerHTML = icon;
  header.title = "手机设置"; header.setAttribute("aria-label", "手机设置"); header.onclick = openSettings;
  document.querySelector(".chat-header").append(header);
  const auth = node("button", "手机设置", "text-button phone-auth-settings"); auth.type = "button"; auth.onclick = openSettings;
  document.querySelector(".auth-card").append(auth);
  window.addEventListener("remote-agent-android-state", event => { host = event.detail; render(); });
  window.addEventListener("remote-agent-theme", () => command("theme", {value:document.documentElement.dataset.theme}));
  window.remoteAgentAndroidBack = () => {
    const popover = document.querySelector(":popover-open");
    if (popover) { popover.hidePopover(); return true; }
    const modals = [...document.querySelectorAll("dialog:modal")];
    // The focused modal is on top, including theme settings opened from this sheet.
    const modal = document.activeElement?.closest("dialog:modal") || modals.at(-1);
    if (modal) {
      if (modal.dispatchEvent(new Event("cancel", {cancelable:true}))) modal.close();
      return true;
    }
    const menu = document.querySelector('.context-menu:not([hidden])');
    if (menu) { document.dispatchEvent(new KeyboardEvent("keydown", {key:"Escape",bubbles:true})); return true; }
    if (document.querySelector(".drawer-open")) { closeDrawers(); return true; }
    return false;
  };
})();
