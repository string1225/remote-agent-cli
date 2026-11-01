/* Project and session actions use the connected client's durable state. */
(() => {
  const projectDialog = $("project-settings-dialog"), projectForm = $("project-settings-form");
  let editingProject;
  window.openProjectSettings = async (client, project) => {
    await connectDevice(client);
    if (!state.projectSettings) throw Error("请更新目标客户端后使用项目设置");
    const current = state.services.find(row => row.id === project.id);
    if (!current) throw Error("项目已被移除，请刷新客户端");
    editingProject = {id:current.id, peer:state.peer, account:state.user?.id};
    projectForm.elements.name.value = current.name;
    projectForm.elements.workspace.value = current.workspace;
    $("project-settings-roots").textContent = "允许的根目录：" + state.roots.join("、");
    projectDialog.showModal();
    projectForm.elements.name.focus();
  };
  projectForm.onsubmit = async event => {
    event.preventDefault();
    const target = editingProject, button = projectForm.querySelector("button[type=submit]");
    if (!target || target.peer !== state.peer || target.account !== state.user?.id) { notice("客户端连接已改变，请重新打开项目设置"); return; }
    button.disabled = true;
    try {
      await rpc("services.update", {service:{id:target.id, name:projectForm.elements.name.value, workspace:projectForm.elements.workspace.value}});
      if (target.peer !== state.peer || target.account !== state.user?.id) return;
      const data = await rpc("services.list");
      if (target.peer !== state.peer) return;
      refreshServices(data);
      if (state.serviceId === target.id) {
        $("project-label").textContent = `${service().provider.toUpperCase()} / ${service().name}`;
        $("project-label").title = service().workspace;
        renderChat();
      }
      projectDialog.close(); await refresh(); notice("项目设置已保存");
    } catch (error) { notice(error.message); }
    finally { button.disabled = false; }
  };

  const menu = $("session-menu"), renameDialog = $("session-rename-dialog"), renameForm = $("session-rename-form");
  let menuTarget, renameTarget, archiveView;
  const scope = () => JSON.stringify([state.user?.id, state.selected?.id, state.serviceId]);
  const isCurrent = target => target?.peer === state.peer && target.scope === scope();
  function closeMenu(restore = false) { menu.hidden = true; if (restore) menuTarget?.opener?.focus(); }
  window.openSessionMenu = (event, session, opener) => {
    event.preventDefault(); event.stopPropagation();
    menuTarget = {session, opener, peer:state.peer, scope:scope()};
    $("device-menu").hidden = true;
    $("session-menu-archive").textContent = session.archivedAt ? "取消归档" : "归档";
    for (const button of menu.querySelectorAll("button")) button.disabled = !state.peer || !state.sessionManagement;
    menu.hidden = false;
    const rect = opener.getBoundingClientRect();
    menu.style.left = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientX : rect.left, innerWidth - menu.offsetWidth - 8)) + "px";
    menu.style.top = Math.max(8, Math.min(event.type === "contextmenu" ? event.clientY : rect.bottom, innerHeight - menu.offsetHeight - 8)) + "px";
    menu.querySelector("button").focus({preventScroll:true});
    if (!state.sessionManagement) notice("客户端在线并升级后，可管理会话");
  };
  async function applySessionAction(target, type, fields = {}) {
    if (!isCurrent(target)) throw Error("客户端或项目已切换，请重新打开会话菜单");
    const session = target.session;
    const result = await rpc(type, {serviceId:session.serviceId || state.serviceId, sessionId:session.id, ...fields});
    if (!isCurrent(target)) return;
    state.sessions = state.sessions.filter(row => row.id !== session.id);
    if (type === "sessions.rename" && !result.archivedAt) state.sessions.push(result);
    if (state.session?.id === session.id) {
      if (type === "sessions.delete") {
        for (const kind of ["history", "draft", "pending"]) WorkspaceCache.remove(kind, draftContext());
        state.session = null; state.entries = []; state.readRevision++;
        $("prompt").value = ""; window.chatAttachments?.restore([]);
        WorkspaceCache.remove("selection", [state.user?.id]);
      } else state.session = {...state.session, ...result};
      renderChat();
    }
    if (archiveView?.scope === scope()) { archiveView.rows = archiveView.rows.filter(row => row.id !== session.id); archiveView.loadedAt = 0; }
    await refreshSessions(false, true);
    renderSessions();
    if (archiveView?.box.open) await loadArchives(archiveView);
  }
  $("session-menu-rename").onclick = () => {
    renameTarget = menuTarget; closeMenu();
    renameForm.elements.title.value = sessionTitle(renameTarget.session);
    renameDialog.showModal(); renameForm.elements.title.focus(); renameForm.elements.title.select();
  };
  renameForm.onsubmit = async event => {
    event.preventDefault(); const button = renameForm.querySelector("button[type=submit]"); button.disabled = true;
    try { await applySessionAction(renameTarget, "sessions.rename", {title:renameForm.elements.title.value}); renameDialog.close(); }
    catch (error) { notice(error.message); } finally { button.disabled = false; }
  };
  $("session-menu-archive").onclick = async () => {
    const target = menuTarget; closeMenu();
    try { await applySessionAction(target, "sessions.archive", {archived:!target.session.archivedAt}); }
    catch (error) { notice(error.message); }
  };
  $("session-menu-delete").onclick = async () => {
    const target = menuTarget; closeMenu();
    if (!confirm(`删除会话「${sessionTitle(target.session)}」？会话将从列表移除，本机历史与附件保留用于恢复。`)) return;
    try { await applySessionAction(target, "sessions.delete"); }
    catch (error) { notice(error.message); }
  };
  document.addEventListener("pointerdown", event => { if (!menu.contains(event.target)) closeMenu(); });
  document.addEventListener("keydown", event => {
    if (menu.hidden) return;
    if (event.key === "Escape") { event.preventDefault(); closeMenu(true); }
    if (["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) {
      event.preventDefault(); const buttons = [...menu.querySelectorAll("button:not(:disabled)")];
      if (!buttons.length) return;
      const index = buttons.indexOf(document.activeElement);
      buttons[event.key === "Home" ? 0 : event.key === "End" ? buttons.length - 1 : (index + (event.key === "ArrowDown" ? 1 : -1) + buttons.length) % buttons.length].focus();
    }
    if (event.key === "Tab") closeMenu();
  });
  window.addEventListener("resize", () => closeMenu());
  window.addEventListener("wheel", () => closeMenu(), {passive:true, capture:true});
  window.addEventListener("touchmove", () => closeMenu(), {passive:true, capture:true});
  window.addEventListener("remote-agent-disconnected", () => { closeMenu(); archiveView = null; editingProject = null; });

  function drawArchives(view) {
    view.list.replaceChildren();
    for (const session of view.rows) {
      const card = node("button", undefined, "conversation-card" + (state.session?.id === session.id ? " active" : ""));
      card.type = "button"; card.append(node("strong", sessionTitle(session)), node("p", "已归档"));
      card.onclick = () => loadSession(session).then(closeDrawers).catch(error => notice(error.message));
      card.oncontextmenu = event => window.openSessionMenu(event, session, card);
      card.onkeydown = event => { if ((event.shiftKey && event.key === "F10") || event.key === "ContextMenu") window.openSessionMenu(event, session, card); };
      view.list.append(card);
    }
    view.summary.textContent = "归档" + (view.rows.length ? ` · ${view.rows.length}` : "");
    if (!view.rows.length) view.list.append(node("p", view.loading ? "正在加载…" : "没有已归档的会话", "archive-empty"));
  }
  async function loadArchives(view) {
    if (archiveView !== view || view.scope !== scope() || !state.peer || view.loading) return;
    const peer = state.peer, project = state.serviceId, search = $("session-search").value.trim();
    view.loading = true; drawArchives(view);
    try {
      let cursor = 0, rows = [];
      do {
        const page = await rpc("sessions.list", {serviceId:project, search, cursor, archived:true});
        if (archiveView !== view || peer !== state.peer || view.scope !== scope() || search !== $("session-search").value.trim()) return;
        rows.push(...page.sessions); cursor = page.nextCursor || 0;
      } while (cursor);
      view.rows = rows; view.loadedAt = Date.now(); view.search = search;
    } catch (error) { notice(error.message); }
    finally { view.loading = false; if (archiveView === view && view.scope === scope()) drawArchives(view); }
  }
  window.renderArchivedSessions = container => {
    if (!state.serviceId || !state.sessionManagement) return;
    if (archiveView?.scope !== scope()) {
      const box = node("details", undefined, "archived-sessions"), summary = node("summary", "归档"), list = node("div");
      box.append(summary, list); archiveView = {scope:scope(), box, summary, list, rows:[], loadedAt:0};
      const view = archiveView;
      box.ontoggle = () => { if (box.open) loadArchives(view); };
    }
    container.append(archiveView.box); drawArchives(archiveView);
    if (archiveView.box.open && (Date.now() - archiveView.loadedAt > 4000 || archiveView.search !== $("session-search").value.trim())) loadArchives(archiveView);
  };
  setInterval(() => { if (!document.hidden && archiveView?.box.open && state.peer && state.serviceId && !$("app").hidden) loadArchives(archiveView); }, 4000);

  const queue = $("queued-messages");
  const queueSignature = new WeakMap();
  const queueBusySessions = new Set();
  let queuePeer;
  window.updateQueuedComposer = () => {
    const web = state.session && state.session.kind !== "im" && state.mode !== "sayso";
    if (state.taskQueue && web) {
      $("send-run").disabled = state.submitting.has(state.session.id) || queueBusySessions.has(state.session.id) || !service() || !state.peer || state.permissionMode !== "full";
    }
    const rows = state.taskQueue && web ? state.session.queuedMessages || [] : [];
    queue.hidden = !rows.length;
    const busy = state.submitting.has(state.session?.id) || queueBusySessions.has(state.session?.id);
    const signature = JSON.stringify([draftContext(), rows, !!state.peer, state.permissionMode, busy]);
    if (queueSignature.get(queue) === signature && queuePeer === state.peer) return;
    queuePeer = state.peer;
    queueSignature.set(queue, signature); queue.replaceChildren();
    if (!rows.length) return;
    queue.append(node("h3", `待处理消息 · ${rows.length}`));
    for (const [index, message] of rows.entries()) {
      const card = node("article", undefined, "queued-message"), header = node("div", undefined, "queued-message-header");
      const label = message.status === "paused" ? "已暂停，可插队继续" : message.status === "editing" ? "编辑内容待取回" : message.priority ? "优先执行" : `排队中 · ${index + 1}`;
      header.append(node("span", label));
      const promote = node("button", "插队", "queue-action"), edit = node("button", "编辑", "queue-action");
      promote.type = edit.type = "button"; promote.hidden = message.status === "editing";
      promote.disabled = edit.disabled = !state.peer || busy;
      const target = {peer:state.peer, scope:scope(), session:state.session, context:draftContext()};
      const act = async operation => {
        if (!isCurrent(target) || state.session?.id !== target.session.id) { notice("会话已切换，请重新操作"); return; }
        if (state.submitting.has(target.session.id) || queueBusySessions.has(target.session.id)) return;
        queueBusySessions.add(target.session.id); window.updateQueuedComposer();
        promote.disabled = edit.disabled = true;
        try {
          const result = await rpc(operation, {serviceId:target.session.serviceId || state.serviceId, sessionId:target.session.id, runId:message.id});
          if (operation === "tasks.edit") {
            const previous = WorkspaceCache.read("draft", target.context), savedEdits = WorkspaceCache.read("queue-edits", target.context) || [];
            const visible = isCurrent(target) && state.session?.id === target.session.id;
            const draft = visible ? {text:$("prompt").value, attachments:window.chatAttachments?.snapshot() || []} : previous || {text:"", attachments:[]};
            if (!savedEdits.includes(message.id)) {
              draft.text = [draft.text, result.prompt].filter(Boolean).join("\n\n");
              draft.attachments = [...new Map([...(draft.attachments || []), ...(result.attachments || [])].map(file => [file.id, file])).values()];
              WorkspaceCache.write("draft", target.context, draft);
              WorkspaceCache.write("queue-edits", target.context, [...savedEdits, message.id].slice(-64));
            }
            if (visible) { $("prompt").value = draft.text; window.chatAttachments?.restore(draft.attachments); $("prompt").focus(); }
            if (isCurrent(target)) await rpc("tasks.edit", {serviceId:target.session.serviceId || state.serviceId, sessionId:target.session.id, runId:message.id, body:{acknowledged:true}});
          }
          if (isCurrent(target) && state.session?.id === target.session.id) await loadSession(state.session);
        } catch (error) { notice(error.message); }
        finally { queueBusySessions.delete(target.session.id); setRunning(state.running); }
      };
      promote.onclick = () => act("tasks.promote"); edit.onclick = () => act("tasks.edit");
      header.append(promote, edit); card.append(header, node("p", message.prompt, "queued-message-text"));
      if (message.attachments?.length) card.append(node("p", message.attachments.map(file => file.name).join(" · "), "queued-message-files"));
      queue.append(card);
    }
  };
  setRunning(state.running);
})();
