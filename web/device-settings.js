/* Device actions work for offline devices too; the account owns the identity. */
(() => {
  let device, project, opener;
  const menu = $("device-menu"), dialog = $("device-settings-dialog"), form = $("device-settings-form");
  function closeMenu(restore = false) {
    menu.hidden = true;
    if (restore) opener?.focus();
  }
  window.openDeviceMenu = (event, nextDevice, anchor) => {
    project=null;openMenu(event,nextDevice,anchor);
  };
  window.openProjectMenu=(event,nextDevice,nextProject,anchor)=>{project=nextProject;openMenu(event,nextDevice,anchor)};
  function openMenu(event,nextDevice,anchor) {
    event.preventDefault(); event.stopPropagation(); device = nextDevice; opener = anchor;
    window.prepareMaintenanceMenu?.(device,project);
    $("device-menu-settings").hidden=!!project;
    $("device-menu-channel").hidden=!project;$("device-menu-channel").disabled=!device.online;
    $("device-menu-delete").textContent=project?"删除项目":"删除客户端";
    $("device-menu-delete").disabled=!!project&&!device.online;
    $("device-menu-delete").title=project&&!device.online?"客户端在线后可移除项目登记":"";
    menu.hidden = false;
    const rect = anchor.getBoundingClientRect();
    menu.style.left = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientX : rect.right - menu.offsetWidth, innerWidth - menu.offsetWidth - 8)) + "px";
    menu.style.top = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientY : rect.bottom + 4, innerHeight - menu.offsetHeight - 8)) + "px";
    menu.querySelector("button:not([hidden])").focus();
  }
  $("device-menu-settings").onclick = () => {
    closeMenu(); form.elements.name.value = device.name; $("device-settings-subtitle").textContent = device.name;
    form.elements.os.value=device.os || "";form.elements.os.disabled=!!device.os;
    dialog.showModal(); form.elements.name.focus(); form.elements.name.select();
  };
  window.openDeviceRepair=async(current,nextProject=null)=>{
    const account = state.user?.id;
    closeMenu();
    if(!current?.os){device=current;$("device-menu-settings").click();notice("请先选择这台客户端的操作系统，保存后再打开修复");return}
    try {
      const result = await api(`/api/agents/${current.id}/repair`, "POST", {});
      if (state.user?.id !== account) return;
      showEnroll(); showInstallCommands({...result, repair:true});
      $("enroll-title").textContent = "修复 · " + current.name;
      $("repair-project-info").hidden=!nextProject;
      $("repair-project-info").textContent=nextProject?`项目：${nextProject.name} · ${nextProject.provider} · ID ${nextProject.id}`:"";
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
  };
  $("device-menu-repair").onclick=()=>window.openDeviceRepair(device,project);
  $("device-menu-channel").onclick=()=>{const current=device,nextProject=project;closeMenu();window.openChannelSetup(current,nextProject).catch(e=>notice(e.message))};
  $("device-menu-delete").onclick=async()=>{
    const current=device,target=project,account=state.user?.id;closeMenu();
    const description=target?`删除项目「${target.name}」的登记？关联机器人会解绑，电脑目录、会话和录音文件会保留。`:`删除客户端「${current.name}」？其远程访问和关联 Token 将撤销；本机文件会保留。`;
    if(!confirm(description))return;
    try {
      if(target){await connectDevice(current);await rpc("services.remove",{serviceId:target.id,removeChannel:true});if(state.user?.id!==account)return;if(state.serviceId===target.id)resetWorkspace();refreshServices(await rpc("services.list"))}
      else{await api(`/api/agents/${current.id}`,"DELETE");if(state.user?.id!==account)return;if(state.selected?.id===current.id)disconnect()}
      await refresh();notice(target?"项目登记已删除，本机文件和历史保留":"客户端已删除，本机文件保留");
    }catch(error){notice(error.message)}
  };
  form.onsubmit = async e => {
    e.preventDefault(); const current = device, account = state.user?.id, button = form.querySelector("button[type=submit]");
    button.disabled = true;
    try {
      await api(`/api/agents/${current.id}`, "PATCH", {name:form.elements.name.value,os:current.os || form.elements.os.value});
      if (state.user?.id !== account) return;
      current.os=current.os || form.elements.os.value;
      if (state.selected?.id === current.id) state.selected.name = form.elements.name.value.trim();
      updateSessionContext();
      dialog.close(); await refresh(); notice("设备名称已更新");
    } catch (error) { notice(error.message); } finally { button.disabled = false; }
  };
  document.addEventListener("pointerdown", e => { if (!menu.contains(e.target)) closeMenu(); });
  document.addEventListener("keydown", e => {
    if (menu.hidden) return;
    if (e.key === "Escape") { e.preventDefault(); closeMenu(true); }
    else if (["ArrowDown","ArrowUp","Home","End"].includes(e.key)) {
      e.preventDefault(); const buttons = [...menu.querySelectorAll("button:not([hidden]):not(:disabled)")], i = buttons.indexOf(document.activeElement);
      buttons[e.key === "Home" ? 0 : e.key === "End" ? buttons.length - 1 : (i + (e.key === "ArrowDown" ? 1 : -1) + buttons.length) % buttons.length].focus();
    } else if (e.key === "Tab") closeMenu();
  });
  window.addEventListener("resize", () => closeMenu());
  window.addEventListener("scroll", () => closeMenu(), true);
  window.addEventListener("remote-agent-disconnected", () => closeMenu());
})();
