/* Account-wide todos use the site's authenticated relay, independent of the
 * currently selected project. Only the default host stores their contents. */
(() => {
  let host = null, account = null, revision = 0, cursor = "", busy = false;
  const button = node("button", "☑ 待办任务"), dialog = node("dialog", undefined, "todo-dialog"),
    heading = node("div", undefined, "todo-heading"), title = node("h2", "待办任务"),
    close = node("button", "×", "icon-button"), hostLine = node("p", "请选择默认主机", "muted"),
    hostButton = node("button", "选择默认主机", "secondary"), toolbar = node("div", undefined, "todo-toolbar"),
    filter = node("select"), search = node("input"), refreshButton = node("button", "刷新", "secondary"),
    createButton = node("button", "＋ 新建待办", "primary"), errorLine = node("p", "", "todo-error"),
    list = node("div", undefined, "todo-list"), more = node("button", "加载更多", "text-button");
  button.id = "todos-entry"; button.type = "button";
  document.querySelector(".platform-nav").prepend(button);
  dialog.id = "todos-dialog"; close.type = "button"; close.setAttribute("aria-label", "关闭待办");
  heading.append(title, close); close.onclick = () => dialog.close();
  dialog.setAttribute("aria-label", "账号待办清单");
  for (const [value, label] of [["open", "待处理"], ["completed", "已完成"], ["cancelled", "已取消"], ["", "全部"]]) {
    const option = node("option", label); option.value = value; filter.append(option);
  }
  filter.setAttribute("aria-label", "待办状态"); search.type = "search"; search.placeholder = "搜索待办…"; search.setAttribute("aria-label", "搜索待办");
  toolbar.append(filter, search, refreshButton, createButton);
  dialog.append(heading, hostLine, hostButton, toolbar, errorLine, list, more); document.body.append(dialog);
  more.hidden = true;
  async function readHost() {
    const user = state.user?.id; if (!user) return;
    const next = await api("/api/default-host");
    if (state.user?.id !== user) return;
    host = next; account = user;
    hostLine.textContent = next.agentId ? `保存在默认主机：${next.name} · ${next.online ? "在线" : "离线"}` : "请先选择一台已绑定的电脑保存本账号的唯一待办清单。";
    hostButton.textContent = next.agentId ? "更换默认主机" : "选择默认主机";
    renderDevices();
  }
  const render = renderDevices;
  renderDevices = function () {
    render();
    if (account !== state.user?.id || !host?.agentId) return;
    for (const card of $("device-list").children) {
      const name = card.querySelector(".device-name"), device = state.agents[[...$("device-list").children].indexOf(card)];
      if (name && device?.id === host.agentId) { const badge = node("span", "默认", "default-host-badge"); badge.title = "保存本账号唯一待办清单的默认主机"; name.append(badge); }
    }
  };
  const menuButton = node("button", "设为默认主机"); menuButton.type = "button"; menuButton.setAttribute("role", "menuitem");
  $("device-menu").append(menuButton);
  const deviceMenu = window.openDeviceMenu, projectMenu = window.openProjectMenu;
  window.openDeviceMenu = (event, device, anchor) => {
    menuButton.hidden = false; menuButton.disabled = !device.enrolled || !device.online || device.id === host?.agentId;
    menuButton.textContent = device.id === host?.agentId ? "当前默认主机" : "设为默认主机";
    menuButton.onclick = () => { $("device-menu").hidden = true; chooseHost(device.id).catch(e => notice(e.message)); };
    deviceMenu(event, device, anchor);
  };
  window.openProjectMenu = (...args) => { menuButton.hidden = true; projectMenu(...args); };

  async function chooseHost(preselected) {
    const user = state.user?.id;
    await refresh(); if (user !== state.user?.id) return;
    const picker = node("dialog", undefined, "todo-host-dialog"), form = node("form"), label = node("label", "默认主机"), select = node("select"),
      description = node("p", "待办清单保存在这台电脑。更换时会迁移全部待办及回复记录，保留旧数据；两台电脑都需要在线。", "muted"),
      save = node("button", "保存并迁移", "primary"), cancel = node("button", "取消", "secondary"), error = node("p", "", "todo-error");
    picker.setAttribute("aria-label", "选择默认主机");
    for (const device of state.agents.filter(a => a.enrolled)) {
      const option = node("option", `${device.name} · ${device.online ? "在线" : "离线"}`); option.value = device.id; option.disabled = !device.online && device.id !== host?.agentId; select.append(option);
    }
    if (preselected || host?.agentId) select.value = preselected || host.agentId;
    label.append(select); save.type = "submit"; cancel.type = "button"; cancel.onclick = () => picker.close();
    form.append(node("h2", "选择默认主机"), description, label, error, save, cancel); picker.append(form); document.body.append(picker);
    picker.addEventListener("close", () => picker.remove(), {once:true});
    form.onsubmit = async e => {
      e.preventDefault(); save.disabled = true; cancel.disabled = true;
      try {
        if (!select.value) throw Error("请先绑定并连接一台电脑");
        await api("/api/default-host", "PUT", {agentId:select.value});
        if (state.user?.id !== user) return;
        picker.close(); await readHost(); if (dialog.open) await load(); notice("默认主机已保存，待办清单已就绪");
      } catch (e) { error.textContent = e.message; } finally { save.disabled = false; cancel.disabled = false; }
    };
    picker.showModal();
  }
  hostButton.onclick = () => chooseHost().catch(e => notice(e.message));

  async function load(append = false) {
    if (!state.user || !dialog.open) return;
    const stamp = ++revision, user = state.user.id;
    errorLine.textContent = "";
    try {
      const page = await api(`/api/todos?status=${encodeURIComponent(filter.value)}&search=${encodeURIComponent(search.value)}&cursor=${encodeURIComponent(append ? cursor : "")}`);
      if (stamp !== revision || user !== state.user?.id || !dialog.open) return;
      if (!append) list.replaceChildren();
      list.append(...page.todos.map(renderTodo)); cursor = page.nextCursor || ""; more.hidden = !cursor;
      if (!list.children.length) list.append(node("p", "这里还没有待办。Agent 需要你判断时会在这里提问，也可以手动新建。", "muted"));
    } catch (e) { if (stamp === revision && user === state.user?.id) { errorLine.textContent = e.message; if (!append) list.replaceChildren(); more.hidden = true; } }
  }
  async function mutate(command) {
    if (busy) return;
    busy = true; const user = state.user?.id;
    try { const result = await api("/api/todos", "POST", command); if (state.user?.id !== user) return; await load(); return result; }
    catch (e) { if (state.user?.id === user) { await load(); errorLine.textContent = e.message; } throw e; }
    finally { busy = false; }
  }
  function action(label, callback) { const b = node("button", label, "secondary"); b.type = "button"; b.onclick = async () => { b.disabled = true; try { await callback(); } catch (e) { notice(e.message); } finally { b.disabled = false; } }; return b; }
  function renderTodo(row) {
    const card = node("article", undefined, "todo-card"), top = node("div", undefined, "todo-heading"), actions = node("div", undefined, "todo-actions");
    card.dataset.todoId = row.id;
    top.append(node("h3", row.title), node("span", ({open:"待处理",completed:"已完成",cancelled:"已取消"})[row.status] || row.status, "status neutral"));
    card.append(top, node("p", row.body, "todo-body"));
    const device = state.agents.find(a => a.id === row.clientId), project = device?.services?.find(p => p.id === row.projectId);
    if (row.sessionId) card.append(node("p", `${device?.name || row.clientId} / ${project?.name || row.projectId} · Session ${row.sessionId}`, "small muted todo-route"));
    const command = operation => ({operation, id:row.id, version:row.version});
    if (row.reply) card.append(node("p", `上次回复 · ${({pending:"等待送达确认",delivered:"已送入原会话",cancelled:"已取消重试，请检查来源会话"})[row.reply.state] || row.reply.state}：${row.reply.text}`, "todo-reply"));
    if (row.reply?.state === "pending") actions.append(action("重试回复", () => mutate({...command("reply"), requestId:row.reply.id, message:row.reply.text})), action("取消", () => mutate(command("cancel"))));
    else {
      if (row.status === "open") {
        actions.append(action("完成", () => mutate(command("complete"))), action("取消", () => mutate(command("cancel"))));
        if (row.sessionId) actions.append(action("回复", () => replyForm(card, row)));
      } else actions.append(action("重新打开", () => mutate(command("reopen"))));
      actions.append(action("编辑", () => editForm(row)));
    }
    if (row.sessionId) actions.append(action("打开会话", async () => { dialog.close(); location.hash = `client=${encodeURIComponent(row.clientId)}&project=${encodeURIComponent(row.projectId)}&session=${encodeURIComponent(row.sessionId)}`; await openLinkedSession(); closeDrawers(); }));
    card.append(actions); return card;
  }
  function replyForm(card, row) {
    if (card.querySelector("form")) return;
    const form = node("form"), text = node("textarea"), send = node("button", "发送到原会话", "primary");
    text.required = true; text.maxLength = 8000; text.placeholder = "给 Agent 的判断、建议或补充信息…"; text.setAttribute("aria-label", "回复内容"); send.type = "submit";
    const context = [state.user?.id,row.id], draft = WorkspaceCache.read("todo-reply",context), id = draft?.id || crypto.randomUUID();
    text.value = draft?.text || "";
    text.oninput = () => WorkspaceCache.write("todo-reply",context,{id,text:text.value});
    form.onsubmit = async e => { e.preventDefault(); send.disabled = true; try { await mutate({operation:"reply",id:row.id,version:row.version,requestId:id,message:text.value}); WorkspaceCache.remove("todo-reply",context); } catch {} finally { send.disabled = false; } };
    form.append(text, send, action("收起", () => form.remove())); card.append(form); text.focus();
  }
  function editForm(row) {
    const editor = node("dialog", undefined, "todo-editor"), form = node("form"), title = node("input"), body = node("textarea"),
      save = node("button", "保存", "primary"), cancel = node("button", "取消", "secondary"), linked = node("input"), linkLabel = node("label", "关联当前会话（回复将让 Agent 继续执行）"), error = node("p", "", "todo-error");
    title.required = true; title.maxLength = 160; title.placeholder = "待办标题"; title.setAttribute("aria-label", "待办标题"); title.value = row?.title || "";
    body.maxLength = 8000; body.rows = 7; body.placeholder = "背景、需要判断的问题和建议…"; body.setAttribute("aria-label", "待办正文"); body.value = row?.body || "";
    linked.type = "checkbox"; linked.checked = !!state.session; linked.disabled = !state.session; linkLabel.prepend(linked); linkLabel.hidden = !!row;
    const route = {clientId:state.selected?.id,projectId:state.serviceId,sessionId:state.session?.id}, requestId = crypto.randomUUID(), user = state.user?.id;
    save.type = "submit"; cancel.type = "button"; cancel.onclick = () => editor.close();
    form.append(node("h2", row ? "编辑待办" : "新建待办"), title, body, linkLabel, error, save, cancel); editor.append(form); document.body.append(editor);
    editor.addEventListener("close", () => editor.remove(), {once:true});
    form.onsubmit = async e => { e.preventDefault(); save.disabled = true; try {
      await mutate({operation:row ? "update" : "create",id:row?.id,version:row?.version,requestId,title:title.value,body:body.value,...(!row && linked.checked ? route : {})});
      if (user === state.user?.id) editor.close();
    } catch (e) { error.textContent = e.message; } finally { save.disabled = false; } };
    editor.showModal(); title.focus();
  }
  createButton.onclick = () => editForm(); refreshButton.onclick = () => { readHost().catch(e => notice(e.message)); load(); };
  filter.onchange = () => load(); search.onsearch = () => load();
  let searchTimer; search.oninput = () => { clearTimeout(searchTimer); searchTimer = setTimeout(() => load(), 250); };
  more.onclick = () => load(true);
  button.onclick = async () => { dialog.showModal(); closeDrawers(); await readHost().catch(e => { errorLine.textContent = e.message; }); await load(); };
  dialog.addEventListener("close", () => { revision++; });
  const signedIn = window.platformSignedIn, signedOut = window.platformSignedOut;
  window.platformSignedIn = user => { signedIn?.(user); readHost().catch(() => {}); };
  window.platformSignedOut = () => { signedOut?.(); host = null; account = null; revision++; list.replaceChildren(); };
  if (state.user) readHost().catch(() => {});
  setInterval(() => { if (dialog.open && !document.hidden && !busy && !dialog.querySelector("form") && !dialog.contains(document.activeElement)) load(); }, 6000);
})();
