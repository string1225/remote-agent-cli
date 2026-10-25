/* Device actions work for offline devices too; the account owns the identity. */
(() => {
  let device, opener;
  const menu = $("device-menu"), dialog = $("device-settings-dialog"), form = $("device-settings-form");
  function closeMenu(restore = false) {
    menu.hidden = true;
    if (restore) opener?.focus();
  }
  window.openDeviceMenu = (event, nextDevice, anchor) => {
    event.preventDefault(); event.stopPropagation(); device = nextDevice; opener = anchor;
    menu.hidden = false;
    const rect = anchor.getBoundingClientRect();
    menu.style.left = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientX : rect.right - menu.offsetWidth, innerWidth - menu.offsetWidth - 8)) + "px";
    menu.style.top = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientY : rect.bottom + 4, innerHeight - menu.offsetHeight - 8)) + "px";
    menu.querySelector("button").focus();
  };
  $("device-menu-settings").onclick = () => {
    closeMenu(); form.elements.name.value = device.name; $("device-settings-subtitle").textContent = device.name;
    dialog.showModal(); form.elements.name.focus(); form.elements.name.select();
  };
  async function repair() {
    const current = device, account = state.user?.id;
    closeMenu();
    try {
      const result = await api(`/api/agents/${current.id}/repair`, "POST", {});
      if (state.user?.id !== account) return;
      showEnroll(); showInstallCommands({...result, repair:true});
      $("enroll-title").textContent = "修复 · " + current.name;
      $("repair-rebind").onclick = async e => {
        e.target.disabled = true;
        try {
          const next = await api(`/api/agents/${current.id}/rebind`, "POST", {});
          if (state.user?.id !== account || !$("enroll-dialog").open) return;
          showInstallCommands({...next, repair:true, recovery:true});
          notice("恢复命令已生成，请在这台设备执行。现有连接将在恢复成功后失效。");
        } catch (error) { notice(error.message); } finally { e.target.disabled = false; }
      };
    } catch (error) { notice(error.message); }
  }
  $("device-menu-repair").onclick = repair;
  form.onsubmit = async e => {
    e.preventDefault(); const current = device, account = state.user?.id, button = form.querySelector("button[type=submit]");
    button.disabled = true;
    try {
      await api(`/api/agents/${current.id}`, "PATCH", {name:form.elements.name.value});
      if (state.user?.id !== account) return;
      if (state.selected?.id === current.id) state.selected.name = form.elements.name.value.trim();
      dialog.close(); await refresh(); notice("设备名称已更新");
    } catch (error) { notice(error.message); } finally { button.disabled = false; }
  };
  document.addEventListener("pointerdown", e => { if (!menu.contains(e.target)) closeMenu(); });
  document.addEventListener("keydown", e => {
    if (menu.hidden) return;
    if (e.key === "Escape") { e.preventDefault(); closeMenu(true); }
    else if (["ArrowDown","ArrowUp","Home","End"].includes(e.key)) {
      e.preventDefault(); const buttons = [...menu.querySelectorAll("button")], i = buttons.indexOf(document.activeElement);
      buttons[e.key === "Home" ? 0 : e.key === "End" ? buttons.length - 1 : (i + (e.key === "ArrowDown" ? 1 : -1) + buttons.length) % buttons.length].focus();
    } else if (e.key === "Tab") closeMenu();
  });
  window.addEventListener("resize", () => closeMenu());
  window.addEventListener("scroll", () => closeMenu(), true);
  window.addEventListener("remote-agent-disconnected", () => closeMenu());
})();
