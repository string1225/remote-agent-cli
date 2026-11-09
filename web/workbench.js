/* The workbench adds bounded inspection to the existing relay, history and drafts. */
(() => {
  const labels = {queued:"排队中", running:"执行中", done:"已完成", failed:"需检查", cancelled:"已停止", interrupted:"已插队中断", paused:"等待继续"};
  const activityLabels = {running:"进行中", completed:"已完成", failed:"失败", stopped:"已结束，状态未确认"};
  let capabilities = {}, freshness, lastScope = "", view;
  const scope = () => JSON.stringify([state.user?.id, state.selected?.id, state.serviceId, state.session?.id]);
  const online = () => !!state.peer && !state.peer.connecting && state.peer.ws.readyState === WebSocket.OPEN;
  const button = (label, action, className = "") => {
    const element = node("button", label, className); element.type = "button"; element.onclick = action; return element;
  };
  const duration = milliseconds => {
    const seconds = Math.max(0, Math.floor(milliseconds / 1000));
    return seconds >= 3600 ? `${Math.floor(seconds/3600)}小时${Math.floor(seconds%3600/60)}分` : seconds >= 60 ? `${Math.floor(seconds/60)}分${seconds%60}秒` : `${seconds}秒`;
  };
  const bar = node("div", undefined, "workbench-bar");
  bar.setAttribute("aria-label", "会话状态和检查工具");
  const status = node("span", "", "workbench-status"), sync = node("span", "", "workbench-sync");
  const actions = node("div", undefined, "workbench-actions");
  const changes = button("工作区变更", () => openChanges());
  const children = button("子任务", () => openChildren());
  const reference = button("引用会话", () => openReferences());
  const diagnostics = button("连接检查", () => openDiagnostics());
  actions.append(changes, children, reference, diagnostics); bar.append(status, sync, actions);
  $("chat-scroll").before(bar);

  function update() {
    const currentScope = scope();
    if (lastScope !== currentScope) { lastScope = currentScope; freshness = null; closeView(); }
    bar.hidden = !state.session || state.mode === "sayso";
    if (bar.hidden) return;
    const session = state.session;
    status.textContent = labels[session.taskStatus] || "会话就绪";
    status.dataset.status = session.taskStatus || "idle";
    if (session.taskStartedAt && (session.taskStatus === "running" || session.taskCompletedAt)) {
      status.textContent += " · " + duration((session.taskCompletedAt || Date.now()) - session.taskStartedAt);
    } else if (session.taskStatus === "queued" && session.taskAcceptedAt) {
      status.textContent += " · 已等待 " + duration(Date.now() - session.taskAcceptedAt);
    }
    const paused = session.queuedMessages?.filter(row => row.status === "paused").length || 0;
    if (paused) status.textContent += ` · ${paused} 条待继续`;
    if (!online() && session.taskStatus) status.textContent += "（上次状态）";
    sync.textContent = !online() ? "离线 · 已保留草稿" : freshness?.error ? "连接正常 · 历史同步失败" : state.historyLocal ? "连接已建立 · 正在校对历史" : freshness?.at ? "历史已同步" : "正在同步历史";
    sync.title = freshness?.error || (freshness?.at ? "最近同步：" + new Date(freshness.at).toLocaleTimeString("zh-CN") : "");
    changes.hidden = !capabilities.workspaceReview;
    children.hidden = !capabilities.runActivities;
    changes.disabled = children.disabled = reference.disabled = !online();
    diagnostics.disabled = !state.selected;
    if (view && view.scope !== currentScope) closeView();
  }

  const baseServices = refreshServices;
  refreshServices = function(data) { capabilities = {workspaceReview:!!data.workspaceReview, runActivities:!!data.runActivities, workbenchDiagnostics:!!data.workbenchDiagnostics}; const result = baseServices(data); update(); return result; };
  const baseRunning = setRunning;
  setRunning = function(...args) { const result = baseRunning(...args); update(); return result; };
  const baseRPC = rpc;
  rpc = async function(type, fields = {}) {
    const currentScope = scope(), peer = state.peer, revision = state.readRevision;
    const watchesHistory = type === "sessions.get" && fields.sessionId === state.session?.id && !state.loadingOlder;
    try {
      const result = await baseRPC(type, fields);
      if (watchesHistory && scope() === currentScope && state.peer === peer && revision === state.readRevision) {
        freshness = result.session?.serviceId === state.serviceId ? {at:Date.now()} : {error:"会话不属于当前项目"}; queueMicrotask(update);
      }
      return result;
    } catch (error) {
      if (watchesHistory && scope() === currentScope && state.peer === peer && revision === state.readRevision) {
        freshness = {error:error.message}; update();
      }
      throw error;
    }
  };
  const baseRender = renderChat;
  renderChat = function(...args) { const result = baseRender(...args); renderActivities(); renderReferences(); update(); return result; };

  function renderActivities() {
    const groups = new Map(), ended = new Set(state.entries.filter(entry => entry.role === "system" && ["done", "error"].includes(entry.kind)).map(entry => entry.turnId));
    if (state.session?.taskId && !["queued", "running"].includes(state.session.taskStatus)) ended.add(state.session.taskId);
    for (const entry of state.entries) {
      if (!entry.activity) continue;
      const group = groups.get(entry.turnId) || new Map();
      const activity = entry.activity.status === "running" && ended.has(entry.turnId) ? {...entry.activity, status:"stopped", completedAt:state.session?.taskId === entry.turnId ? state.session.taskCompletedAt : 0} : entry.activity;
      group.set(activity.id, activity); groups.set(entry.turnId, group);
    }
    for (const article of $("messages").children) {
      const turnId = article.dataset.key?.replace(/\/assistant\/stdout$/, ""), group = groups.get(turnId);
      const details = article.querySelector(".run-details");
      if (!details || !group) continue;
      const fragment = document.createDocumentFragment();
      for (const activity of group.values()) {
        const row = node("div", undefined, "activity-row");
        row.dataset.status = activity.status;
        const header = node("div", undefined, "activity-header");
        header.append(node("span", activityLabels[activity.status] || "状态未知", "activity-state"), node("code", activity.name));
        if (activity.startedAt && (activity.status === "running" || activity.completedAt)) header.append(node("span", duration((activity.completedAt || Date.now()) - activity.startedAt), "activity-duration"));
        row.append(header);
        if (activity.steps?.length) {
          const list = node("ul", undefined, "activity-plan");
          for (const step of activity.steps) list.append(node("li", (step.completed ? "✓ " : "○ ") + step.text));
          row.append(list);
        }
        fragment.append(row);
      }
      // Retain old invocation summaries alongside new metadata, with the same
      // redaction boundary as the existing renderer.
      const legacy = state.entries.filter(entry => entry.turnId === turnId && entry.kind === "tool" && !entry.activity).map(entry => toolInvocation(entry.text)).join("");
      if (legacy) fragment.append(node("pre", legacy.trimEnd(), "activity-legacy"));
      details.replaceChildren(fragment);
      const toggle = article.querySelector(".run-details-toggle");
      const running = [...group.values()].filter(activity => activity.status === "running").length;
      toggle.textContent = `运行详情 · ${group.size}` + (running ? ` · ${running} 项进行中` : "");
      toggle.hidden = false;
    }
  }

  function sessionURL(session = state.session) {
    const url = new URL(location.href); url.hash = new URLSearchParams({client:state.selected.id, project:session.serviceId, session:session.id}).toString(); return url.href;
  }
  function insertContext(text) {
    const prompt = $("prompt"), start = prompt.selectionStart ?? prompt.value.length, end = prompt.selectionEnd ?? start;
    const insertion = (start ? "\n\n" : "") + text + "\n\n";
    const value = prompt.value.slice(0, start) + insertion + prompt.value.slice(end);
    if (new TextEncoder().encode(value).length > 32000) { notice("引用后内容超过消息上限，请先缩短草稿或引用片段。"); return false; }
    prompt.value = value; prompt.setSelectionRange(start + insertion.length, start + insertion.length);
    prompt.dispatchEvent(new Event("input", {bubbles:true})); saveCurrentDraft(); prompt.focus(); return true;
  }
  function quote(text, session, turnId) {
    const bounded = [...text].slice(0, 3000).join("");
    return `引用背景：${session.title || "会话"}\n来源：${sessionURL(session)}${turnId ? `\n轮次：${turnId}` : ""}\n` + bounded.split("\n").map(line => "> " + line).join("\n") + ([...text].length > 3000 ? "\n（引用已截取，完整内容见来源会话）" : "");
  }
  function renderReferences() {
    if (!state.session || state.mode === "sayso") return;
    for (const article of $("messages").querySelectorAll(".message.assistant")) {
      const text = article.querySelector(".message-text"), header = article.querySelector(".message-label");
      if (!text || !text.textContent.trim() || !header || header.querySelector(".quote-response")) continue;
      const targetSession = {...state.session}, turnId = article.dataset.key?.replace(/\/assistant\/stdout$/, "");
      const cite = button("引用", () => {
        if (targetSession.id !== state.session?.id) return;
        const selected = window.getSelection();
        const content = selected && text.contains(selected.anchorNode) && text.contains(selected.focusNode) && selected.toString().trim() ? selected.toString() : text.innerText;
        insertContext(quote(content, targetSession, turnId));
      }, "quote-response");
      cite.title = "引用选中文字或这条回答，添加到草稿"; header.append(cite);
    }
  }

  function closeView() { if (view) { const dialog = view.dialog; view = null; dialog.close(); dialog.remove(); } }
  function openView(title) {
    closeView();
    const dialog = node("dialog", undefined, "workbench-dialog");
    const heading = node("div", undefined, "workbench-dialog-heading"); heading.append(node("h3", title), button("关闭", () => dialog.close()));
    const body = node("div", undefined, "workbench-dialog-body"); dialog.append(heading, body); document.body.append(dialog);
    const current = {dialog, body, scope:scope(), peer:state.peer, project:state.serviceId, session:state.session?.id}; view = current;
    dialog.addEventListener("close", () => { if (view === current) view = null; dialog.remove(); });
    dialog.showModal(); return current;
  }
  const current = target => view === target && target.scope === scope() && target.peer === state.peer;
  async function requestView(target, type, fields) {
    const value = await rpc(type, {serviceId:target.project, ...fields});
    return current(target) ? value : null;
  }
  function failView(target, error) { if (current(target)) target.body.replaceChildren(node("p", error.message, "workbench-error")); }

  async function openChanges() {
    const target = openView("工作区变更");
    target.body.append(node("p", "正在检查…", "muted"));
    try {
      const result = await requestView(target, "workspace.changes"); if (!result) return;
      const refresh = button("刷新变更", () => openChanges());
      target.body.replaceChildren(node("p", "当前项目目录的全部未提交变更，包含暂存区和其他会话的修改。", "muted"), refresh);
      if (!result.git) { target.body.append(node("p", "当前项目尚未启用 Git。")); return; }
      target.body.append(node("p", `${result.branch} · ${result.files.length} 个变更文件`));
      if (result.truncated) target.body.append(node("p", "文件较多，仅显示前 200 项。", "muted"));
      const layout = node("div", undefined, "workspace-review"), list = node("div", undefined, "workspace-change-list"), patch = node("div", undefined, "workspace-patch");
      layout.append(list, patch); target.body.append(layout);
      patch.append(node("p", result.files.length ? "选择文件查看差异" : "工作区没有未提交变更。", "muted"));
      let fileRevision = 0;
      for (const file of result.files) {
        const item = button(`${file.index}${file.work}  ${file.path}`, async () => {
          const revision = ++fileRevision; patch.replaceChildren(node("p", "正在读取差异…", "muted"));
          try {
            const diff = await requestView(target, "workspace.diff", {path:file.path});
            if (!diff || revision !== fileRevision) return;
            list.querySelectorAll("button").forEach(row => row.classList.toggle("active", row === item));
            patch.replaceChildren(node("h4", file.path));
            if (diff.binary) patch.append(node("p", "二进制文件，无法显示文本差异。"));
            for (const [label, content] of [["暂存区", diff.staged], ["工作区", diff.unstaged], ["新增文件", diff.untracked]]) {
              if (!content) continue;
              patch.append(node("h5", label), node("pre", content, "diff-content"));
            }
            if (!diff.binary && !diff.staged && !diff.unstaged && !diff.untracked) patch.append(node("p", "文件没有可显示的文本差异。"));
            if (diff.truncated) patch.append(node("p", "内容较长，预览已截取。请在目标电脑查看完整文件。", "muted"));
          } catch (error) { if (current(target) && revision === fileRevision) patch.replaceChildren(node("p", error.message, "workbench-error")); }
        });
        list.append(item);
      }
    } catch (error) { failView(target, error); }
  }

  async function openChildren() {
    const target = openView("当前会话的子任务"); target.body.append(node("p", "正在读取…", "muted"));
    let count = 0;
    const load = async (cursor = 0, more) => {
      if (more) more.disabled = true;
      try {
        const result = await requestView(target, "sessions.children", {sessionId:target.session, cursor}); if (!result) return;
        if (!cursor) target.body.replaceChildren(); else more?.remove();
        for (const session of result.sessions.filter(session => !session.deletedAt)) {
          count++;
          const row = node("div", undefined, "workbench-child");
          row.append(node("strong", session.title || "子任务"), node("span", labels[session.taskStatus] || "就绪"), button("打开会话", async () => { closeView(); try { await openWorkspace(state.selected, session.serviceId, session.id); } catch(error) { notice(error.message); } })); target.body.append(row);
        }
        if (result.nextCursor) {
          const next = button("加载更多子任务", () => load(result.nextCursor, next)); target.body.append(next);
        } else if (!count) target.body.append(node("p", "当前客户端没有此会话直接创建的子任务。管家派工结果仍在对话中显示。", "muted"));
      } catch(error) { if (more && current(target)) { more.disabled = false; notice(error.message); } else failView(target, error); }
    };
    await load();
  }

  function openReferences() {
    const target = openView("引用同项目会话");
    target.body.append(node("p", "选择会话，将最近一轮的文字和来源加入草稿。", "muted"));
    for (const session of state.sessions.filter(session => session.id !== target.session && !session.deletedAt && !session.archivedAt && session.serviceId === target.project)) {
      const row = button(session.title || "会话", async () => {
        row.disabled = true;
        try {
          const result = await requestView(target, "sessions.get", {sessionId:session.id, turns:1}); if (!result) return;
          const content = result.entries.filter(entry => entry.role === "user" || entry.role === "assistant").map(entry => (entry.role === "user" ? "用户：" : "回答：") + entry.text).join("\n\n");
          if (!content) { notice("这个会话暂时没有可引用的文字。"); return; }
          if (insertContext(quote(content, session))) closeView();
        } catch(error) { if (current(target)) notice(error.message); } finally { row.disabled = false; }
      }, "reference-session"); target.body.append(row);
    }
    if (target.body.children.length === 1) target.body.append(node("p", "当前项目没有其他已加载的会话。", "muted"));
  }

  async function openDiagnostics() {
    const target = openView("连接检查");
    const report = {client:state.selected?.name, project:service()?.name, session:state.session?.id, connection:online()?"已连接":"未连接", history:sync.textContent, task:state.session?.taskStatus || "idle", lastSyncAt:freshness?.at ? new Date(freshness.at).toISOString() : null};
    target.body.append(node("p", "正在检查…", "muted"));
    try {
      if (online() && capabilities.workbenchDiagnostics) {
        const details = await requestView(target, "diagnostics.snapshot"); if (!details) return; report.clientCheck = details;
      }
      if (!current(target)) return;
      const text = JSON.stringify(report, null, 2), copy = button("复制检查结果", async () => { try { await navigator.clipboard.writeText(text); notice("检查结果已复制"); } catch { notice("无法复制，请选择检查结果手动复制。"); } });
      target.body.replaceChildren(node("p", "检查结果包含版本、连接和任务状态，可用于排查连接问题。", "muted"), copy, node("pre", text, "diagnostics-content"));
    } catch(error) { failView(target, error); }
  }
  window.addEventListener("remote-agent-disconnected", () => { capabilities = {}; freshness = null; closeView(); queueMicrotask(update); });
  document.addEventListener("visibilitychange", () => { if (!document.hidden) update(); });
  update();
})();
