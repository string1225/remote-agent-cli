/* Account-wide todos use the site's authenticated relay, independent of the
 * currently selected project. Only the default host stores their contents. */
(() => {
  let host = null, account = null, revision = 0, cursors = {}, rows = [], busy = false;
  const statuses = new Set(["open"]), statusLabels = {open:"待处理",completed:"已处理",cancelled:"已删除"};
  const button = node("button", "☑ 待办任务"), dialog = node("dialog", undefined, "todo-dialog"),
    heading = node("div", undefined, "todo-dialog-header"), title = node("h2", "待办任务"), count = node("span", "", "todo-count"),
    headerTitle = node("div", undefined, "todo-title"), headerActions = node("div", undefined, "todo-header-actions"),
    close = node("button", "×", "icon-button"), hostLine = node("button", "请选择默认主机", "todo-host-name"),
    hostBar = node("div", undefined, "todo-host"), hostInfo = node("div", undefined, "todo-host-info"),
    toolbar = node("div", undefined, "todo-toolbar"),
    searchBox = node("div", undefined, "todo-search"),
    filter = node("div", undefined, "todo-status-filters"), search = node("input"), refreshButton = node("button", undefined, "secondary todo-refresh"),
    createButton = node("button", "＋ 新建待办", "primary"), errorLine = node("p", "", "todo-error"),
    list = node("div", undefined, "todo-list"), more = node("button", "加载更多", "text-button");
  button.id = "todos-entry"; button.type = "button";
  document.querySelector(".platform-nav").prepend(button);
  dialog.id = "todos-dialog"; close.type = "button"; close.setAttribute("aria-label", "关闭待办");
  headerTitle.append(title, count); headerActions.append(hostBar, createButton, close); heading.append(headerTitle, headerActions); close.onclick = () => dialog.close();
  hostInfo.append(node("span", "默认主机", "todo-host-label"), hostLine); hostBar.append(hostInfo);
  hostLine.type = createButton.type = refreshButton.type = "button";
  refreshButton.append(icon("refresh")); refreshButton.title = "刷新待办"; refreshButton.setAttribute("aria-label", "刷新待办");
  dialog.setAttribute("aria-label", "账号待办清单");
  for (const [value, label] of Object.entries(statusLabels)) {
    const option = node("button", label); option.type = "button"; option.dataset.status = value;
    option.setAttribute("aria-pressed", String(statuses.has(value)));
    option.onclick = () => { if (statuses.has(value)) statuses.delete(value); else statuses.add(value); option.setAttribute("aria-pressed", String(statuses.has(value))); load(); };
    filter.append(option);
  }
  filter.setAttribute("role", "group"); filter.setAttribute("aria-label", "待办状态"); search.type = "search"; search.placeholder = "搜索待办…"; search.setAttribute("aria-label", "搜索待办");
  const searchIcon = node("span", "⌕", "todo-search-icon"); searchIcon.setAttribute("aria-hidden", "true"); searchBox.append(searchIcon, search);
  toolbar.append(searchBox, filter, refreshButton);
  errorLine.setAttribute("role", "alert"); list.setAttribute("aria-label", "待办列表");
  dialog.append(heading, toolbar, errorLine, list, more); document.body.append(dialog);
  more.hidden = true;
  async function readHost() {
    const user = state.user?.id; if (!user) return;
    const next = await api("/api/default-host");
    if (state.user?.id !== user) return;
    host = next; account = user;
    hostLine.textContent = next.agentId ? `${next.name} · ${next.online ? "在线" : "离线"}` : "未选择 · 选择一台电脑保存待办";
    hostLine.dataset.state = next.agentId ? next.online ? "online" : "offline" : "unselected";
    hostLine.title = next.agentId ? "更换默认主机" : "选择默认主机";
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
    const actions = node("div", undefined, "todo-form-actions"); actions.append(cancel, save);
    form.append(node("h2", "选择默认主机"), description, label, error, actions); picker.append(form); document.body.append(picker);
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
  hostLine.onclick = () => chooseHost().catch(e => notice(e.message));

  function icon(name) {
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("viewBox", "0 0 24 24"); svg.setAttribute("aria-hidden", "true");
    svg.setAttribute("fill", "none"); svg.setAttribute("stroke", "currentColor"); svg.setAttribute("stroke-width", "1.7");
    const paths = {refresh:"M20 7v5h-5 M20 12a8 8 0 1 0-2 5 M20 7l-3-3",edit:"m15 5 4 4 M4 20l4-1L20 7a2.8 2.8 0 0 0-4-4L4 15Z",save:"m5 12 4 4L19 6",open:"M14 4h6v6 M20 4 10 14 M10 4H5a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-5",close:"m6 6 12 12 M18 6 6 18"};
    const path = document.createElementNS(svg.namespaceURI, "path"); path.setAttribute("d", paths[name]);
    path.setAttribute("stroke-linecap", "round"); path.setAttribute("stroke-linejoin", "round"); svg.append(path); return svg;
  }
  function iconAction(label, name, callback) {
    const b = action(label, callback); b.className = "icon-button todo-icon"; b.replaceChildren(icon(name));
    b.title = label; b.setAttribute("aria-label", label); return b;
  }
  function openSourceSession(row) {
    const url = new URL(location.pathname, location.origin);
    url.search = new URLSearchParams({client:row.clientId,project:row.projectId,session:row.sessionId});
    if (!document.documentElement.classList.contains("android-app")) {
      window.open(url.href, "_blank", "popup=yes,width=1200,height=860,noopener,noreferrer"); return;
    }
    url.searchParams.set("view", "todo-session");
    // Android's single WebView cannot create a browser window. Keep the list
    // alive underneath an independent, same-site workspace instead.
    const viewer = node("dialog", undefined, "todo-session-dialog"), head = node("div", undefined, "dialog-head"), frame = node("iframe");
    viewer.setAttribute("aria-label","来源会话");
    const close = iconAction("关闭来源会话", "close", () => viewer.close());
    head.append(node("h3", row.title),close); frame.title = "待办来源会话"; frame.src = url.href;
    viewer.append(head,frame); document.body.append(viewer);
    viewer.addEventListener("close",()=>viewer.remove(),{once:true}); viewer.showModal();
  }
  async function load(append = false) {
    if (!state.user || !dialog.open) return;
    const stamp = ++revision, user = state.user.id, selected = [...statuses];
    errorLine.textContent = ""; list.setAttribute("aria-busy", "true"); list.inert = true; count.textContent = "加载中…";
    refreshButton.classList.add("loading");
    try {
      const pages = await Promise.all(selected.filter(status => !append || cursors[status]).map(async status => {
        const cursor = append ? cursors[status] : "";
        const page = await api(`/api/todos?status=${encodeURIComponent(status)}&search=${encodeURIComponent(search.value)}&cursor=${encodeURIComponent(cursor || "")}`);
        if (page.nextCursor && page.nextCursor === cursor) throw Error("待办分页没有推进，请刷新后重试");
        return {status,page};
      }));
      if (stamp !== revision || user !== state.user?.id || !dialog.open) return;
      if (!append) { rows = []; cursors = {}; }
      const merged = new Map(rows.map(row => [row.id,row]));
      for (const {status,page} of pages) { cursors[status] = page.nextCursor || ""; for (const row of page.todos) merged.set(row.id,row); }
      rows = [...merged.values()].filter(row => statuses.has(row.status)).sort((a,b) => (b.createdAt || 0) - (a.createdAt || 0) || a.id.localeCompare(b.id));
      list.replaceChildren(...rows.map(renderTodo));
      more.hidden = !selected.some(status => cursors[status]);
      count.textContent = `${rows.length}${more.hidden ? "" : "+"} 项`;
      if (!rows.length) {
        const empty = node("div", undefined, "todo-empty"), check = node("span", "✓", "todo-empty-icon"); check.setAttribute("aria-hidden", "true");
        const label = !selected.length ? "请选择待办状态" : search.value.trim() ? "没有匹配的待办" : selected.length === 1 ? `暂无${statusLabels[selected[0]]}任务` : "暂无这些状态的待办";
        empty.append(check, node("h3", label), node("p", search.value.trim() ? "换个关键词，或调整状态筛选。" : "需要你判断的事项和通知会出现在这里。")); list.append(empty);
      }
    } catch (e) { if (stamp === revision && user === state.user?.id) { errorLine.textContent = e.message; count.textContent = "暂不可用"; } }
    finally { if (stamp === revision) { list.setAttribute("aria-busy", "false"); list.inert = false; refreshButton.classList.remove("loading"); } }
  }
  async function mutate(command, reloadOnError = false) {
    if (busy) throw Error("正在保存其他待办，请稍候");
    busy = true; const user = state.user?.id;
    try {
      const result = await api("/api/todos", "POST", command);
      if (state.user?.id !== user) return;
      await load(); return result;
    } catch (e) {
      if (state.user?.id === user) { if (reloadOnError) await load(); errorLine.textContent = e.message; }
      throw e;
    } finally { busy = false; }
  }
  function action(label, callback, primary = false) {
    const b = node("button", label, "secondary" + (primary ? " todo-action-primary" : "")); b.type = "button";
    b.onclick = async () => { b.disabled = true; try { await callback(); } catch (e) { notice(e.message); } finally { b.disabled = false; } }; return b;
  }
  function renderTodo(row) {
    const card = node("article", undefined, "todo-card"), top = node("div", undefined, "todo-heading"),
      title = node("h3", row.title), tools = node("div", undefined, "todo-card-tools"),
      actions = node("div", undefined, "todo-actions"), footer = node("div", undefined, "todo-footer"), body = node("div", "", "todo-body");
    card.dataset.todoId = row.id; card.dataset.status = row.status;
    const command = operation => ({operation,id:row.id,version:row.version});
    const edit = iconAction("编辑待办", "edit", () => toggleEdit());
    edit.disabled = row.reply?.state === "pending";
    tools.append(node("span", statusLabels[row.status] || row.status, "todo-status"),edit);
    if (row.sessionId) tools.append(iconAction("在新窗口打开会话", "open", () => openSourceSession(row)));
    top.append(title, tools); renderMarkdown(body, row.body || ""); card.append(top, body);
    const device = state.agents.find(a => a.id === row.clientId), project = device?.services?.find(p => p.id === row.projectId);
    if (row.sessionId) {
      const route = node("div", undefined, "todo-route"), session = node("span", `会话 ${row.sessionId.slice(0, 8)}${row.sessionId.length > 8 ? "…" : ""}`, "todo-session-ref");
      session.title = `Session ${row.sessionId}`; route.append(node("span", `${device?.name || row.clientId} / ${project?.name || row.projectId}`), session); card.append(route);
    }
    if (row.reply) card.append(node("p", `上次回复 · ${({pending:"等待送达确认",delivered:"已送入原会话",cancelled:"已停止重试"})[row.reply.state] || row.reply.state}：${row.reply.text}`, "todo-reply"));
    if (row.reply?.state === "pending") {
      actions.append(action("重试回复", () => mutate({...command("reply"),requestId:row.reply.id,message:row.reply.text}, true)));
    } else if (row.status === "open" && row.sessionId) footer.append(responseForm(row));
    if (row.status === "open") {
      const complete = action("标记完成", () => mutate(command("complete")));
      complete.disabled = row.reply?.state === "pending";
      actions.append(complete);
    } else actions.append(action(row.status === "cancelled" ? "恢复" : "重新打开", () => mutate(command("reopen"))));
    if (row.status !== "cancelled") actions.append(action("删除", () => mutate(command("delete"))));
    footer.append(actions); card.append(footer);

    let editing = false, titleInput, bodyInput, cancelEdit;
    const context = [state.user?.id,row.id];
    function closeEdit(discard = false) {
      if (discard) WorkspaceCache.remove("todo-edit",context);
      titleInput?.replaceWith(title); bodyInput?.replaceWith(body); cancelEdit?.remove();
      editing = false; card.classList.remove("editing");
      edit.replaceChildren(icon("edit")); edit.title = "编辑待办"; edit.setAttribute("aria-label","编辑待办");
    }
    async function toggleEdit() {
      if (editing) {
        if (!titleInput.value.trim()) { titleInput.setCustomValidity("请填写待办标题"); titleInput.reportValidity(); return; }
        const draft = WorkspaceCache.read("todo-edit",context);
        titleInput.disabled = bodyInput.disabled = true;
        try {
          const result = await mutate({...command("update"),version:draft?.version || row.version,title:titleInput.value,body:bodyInput.value});
          if (result) WorkspaceCache.remove("todo-edit",context);
        } finally { titleInput.disabled = bodyInput.disabled = false; }
        return;
      }
      const draft = WorkspaceCache.read("todo-edit",context);
      titleInput = node("input"); titleInput.required = true; titleInput.maxLength = 160;
      titleInput.setAttribute("aria-label","待办标题"); titleInput.value = draft?.title ?? row.title;
      bodyInput = node("textarea"); bodyInput.rows = 5; bodyInput.maxLength = 8000;
      bodyInput.setAttribute("aria-label","待办详情"); bodyInput.value = draft?.body ?? row.body;
      titleInput.oninput = bodyInput.oninput = () => { titleInput.setCustomValidity(""); WorkspaceCache.write("todo-edit",context,{version:draft?.version || row.version,title:titleInput.value,body:bodyInput.value}); };
      cancelEdit = iconAction("取消编辑", "close", () => closeEdit(true)); tools.append(cancelEdit);
      title.replaceWith(titleInput); body.replaceWith(bodyInput); editing = true; card.classList.add("editing");
      edit.replaceChildren(icon("save")); edit.title = "保存待办"; edit.setAttribute("aria-label","保存待办"); titleInput.focus();
    }
    return card;
  }
  function responseForm(row) {
    const form = node("form", undefined, "todo-response"), options = node("div", undefined, "todo-options"),
      text = node("textarea"), send = node("button", "发送回复", "primary"), context = [state.user?.id,row.id],
      draft = WorkspaceCache.read("todo-reply",context), requestId = draft?.id || crypto.randomUUID(), user = state.user?.id;
    let selected = Number.isInteger(draft?.option) ? draft.option : -1;
    const choices = row.options || [];
    if (!choices[selected]) selected = -1;
    const storeDraft = () => { text.setCustomValidity(""); WorkspaceCache.write("todo-reply",context,{id:requestId,text:text.value,option:selected}); };
    for (const [index,option] of choices.entries()) {
      const label = node("label", undefined, "todo-option"), radio = node("input"), copy = node("span");
      radio.type = "radio"; radio.name = `todo-option-${row.id}`; radio.value = index; radio.checked = index === selected;
      copy.append(node("strong",option.label)); if (option.description) copy.append(node("span",option.description));
      label.append(radio,copy); options.append(label);
      radio.onchange = () => { selected = index; storeDraft(); };
    }
    text.rows = 2; text.maxLength = 8000; text.value = draft?.text || "";
    text.placeholder = choices.length ? "补充你的判断，或直接输入其他方案…" : "有后续指令的话，写在这里…";
    text.setAttribute("aria-label","补充指令"); text.oninput = storeDraft; send.type = "submit";
    form.append(options,text,send);
    form.onsubmit = async event => {
      event.preventDefault();
      const option = choices[selected], extra = text.value.trim();
      if (!option && !extra) { text.setCustomValidity("请选择一个选项，或填写补充指令"); text.reportValidity(); return; }
      const message = option ? `${option.label}${option.description ? "："+option.description : ""}${extra ? "\n\n补充："+extra : ""}` : extra;
      send.disabled = true; storeDraft();
      try { const result = await mutate({operation:"reply",id:row.id,version:row.version,requestId,message}, true); if (result && state.user?.id === user) WorkspaceCache.remove("todo-reply",context); }
      catch {} finally { send.disabled = false; }
    };
    return form;
  }

  function editForm(row) {
    const editor = node("dialog", undefined, "todo-editor"), form = node("form"), title = node("input"), body = node("textarea"),
      save = node("button", "保存", "primary"), cancel = node("button", "取消", "secondary"), linked = node("input"), linkLabel = node("label", "关联当前会话（回复将让 Agent 继续执行）"), error = node("p", "", "todo-error");
    title.required = true; title.maxLength = 160; title.placeholder = "待办标题"; title.setAttribute("aria-label", "待办标题"); title.value = row?.title || "";
    body.maxLength = 8000; body.rows = 6; body.placeholder = "写清重点、下一步和验收标准，支持 Markdown…"; body.setAttribute("aria-label", "待办正文"); body.value = row?.body || "";
    linked.type = "checkbox"; linked.checked = !!state.session; linked.disabled = !state.session; linkLabel.prepend(linked); linkLabel.hidden = !!row;
    const route = {clientId:state.selected?.id,projectId:state.serviceId,sessionId:state.session?.id}, requestId = crypto.randomUUID(), user = state.user?.id;
    save.type = "submit"; cancel.type = "button"; cancel.onclick = () => editor.close();
    const titleLabel = node("label", "标题"), bodyLabel = node("label", "任务说明"), actions = node("div", undefined, "todo-form-actions");
    titleLabel.append(title); bodyLabel.append(body); linkLabel.className = "todo-link-session"; actions.append(cancel, save);
    form.append(node("h2", row ? "编辑待办" : "新建待办"), titleLabel, bodyLabel, linkLabel, error, actions); editor.append(form); document.body.append(editor);
    editor.addEventListener("close", () => editor.remove(), {once:true});
    form.onsubmit = async e => { e.preventDefault(); save.disabled = true; try {
      await mutate({operation:row ? "update" : "create",id:row?.id,version:row?.version,requestId,title:title.value,body:body.value,...(!row && linked.checked ? route : {})});
      if (user === state.user?.id) editor.close();
    } catch (e) { error.textContent = e.message; } finally { save.disabled = false; } };
    editor.showModal(); title.focus();
  }
  createButton.onclick = () => editForm(); refreshButton.onclick = () => { readHost().catch(e => notice(e.message)); load(); };
  search.onsearch = () => load();
  let searchTimer; search.oninput = () => { clearTimeout(searchTimer); searchTimer = setTimeout(() => load(), 250); };
  more.onclick = () => load(true);
  button.onclick = async () => { dialog.showModal(); closeDrawers(); await readHost().catch(e => { errorLine.textContent = e.message; }); await load(); };
  dialog.addEventListener("close", () => { revision++; });
  const signedIn = window.platformSignedIn, signedOut = window.platformSignedOut;
  window.platformSignedIn = user => { signedIn?.(user); readHost().catch(() => {}); };
  window.platformSignedOut = () => { signedOut?.(); host = null; account = null; revision++; list.replaceChildren(); };
  if (state.user) readHost().catch(() => {});
  setInterval(() => { if (dialog.open && !document.hidden && !busy && !dialog.querySelector(".editing") && !dialog.querySelector(".todo-response textarea:focus") && !dialog.querySelector(".todo-response input:focus") && !dialog.contains(document.activeElement)) load(); }, 6000);
})();
