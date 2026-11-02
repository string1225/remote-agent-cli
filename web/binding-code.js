/* Pending device codes are fetched only on demand by their owning account. */
(() => {
  const dialog = $("enroll-dialog"), field = $("device-binding-code"), box = $("device-binding-code-box");
  let target;
  function clear() { target = null; field.value = ""; box.hidden = true; }
  window.renderDeviceBindingCode = result => {
    clear();
    if (!result.bindingCode || result.agent?.enrolled || result.agent?.os !== "windows") return;
    target = {id:result.agent.id, account:state.user?.id};
    field.value = result.bindingCode; box.hidden = false;
    $("device-binding-code-expiry").textContent = result.expiresAt ? "有效期至 " + new Date(result.expiresAt * 1000).toLocaleTimeString() : "一次性绑定码 · 15 分钟有效";
    $("install-command-label").firstChild.textContent = "安装并预填绑定码（可选）";
    $("install-command").value = result.windowsInstall || result.windows;
    $("install-guide").textContent = "下载并安装 MSI，在安装器最后一页粘贴下方绑定码完成激活；也可复制安装指令，自动下载对应架构的 MSI 并预填绑定码。";
    $("binding-step-title").textContent = "2. 在 MSI 中填写设备绑定码";
  };
  $("copy-device-binding-code").onclick = async () => {
    try { await navigator.clipboard.writeText(field.value); notice("绑定码已复制"); }
    catch { field.focus(); field.select(); notice("请按 Ctrl / ⌘ + C 复制绑定码"); }
  };
  $("renew-device-binding-code").onclick = async event => {
    const previous = target; if (!previous) return;
    event.target.disabled = true;
    try {
      const result = await api(`/api/agents/${previous.id}/binding-code`, "POST", {renew:true});
      if (target !== previous || state.user?.id !== previous.account || !dialog.open) return;
      showInstallCommands(result); notice("新绑定码已生成，原码已失效");
    } catch (error) { notice(error.message); } finally { event.target.disabled = false; }
  };
  const original = window.openDeviceMenu;
  const item = node("button", "查看绑定码", ""); item.type = "button"; item.setAttribute("role", "menuitem");
  window.openDeviceMenu = (event, device, opener) => {
    item.hidden = true;
    original(event, device, opener);
    if (device.enrolled || device.os !== "windows") return;
    $("device-menu").prepend(item); item.hidden = false;
    item.onclick = async () => {
      const account = state.user?.id;
      $("device-menu").hidden = true;
      try {
        const result = await api(`/api/agents/${device.id}/binding-code`, "POST", {});
        if (state.user?.id !== account) return;
        showEnroll(); showInstallCommands(result);
      } catch (error) { notice(error.message); }
    };
  };
  const projectMenu = window.openProjectMenu;
  window.openProjectMenu = (...args) => { item.hidden = true; projectMenu(...args); };
  dialog.addEventListener("close", clear);
})();
