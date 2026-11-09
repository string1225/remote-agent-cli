/* Project archives belong to the account catalog, independent of client versions. */
(() => {
  const dialog = node("dialog", undefined, "project-archive-dialog");
  dialog.id = "project-archive-dialog";
  dialog.setAttribute("aria-labelledby", "project-archive-title");
  const heading = node("div", undefined, "dialog-head"), title = node("h2", "已归档项目");
  title.id = "project-archive-title";
  const close = node("button", "×", "icon-button"); close.type = "button";
  close.setAttribute("aria-label", "关闭已归档项目"); close.onclick = () => dialog.close();
  heading.append(title, close);
  const description = node("p", "归档只隐藏侧栏项目。会话、录音和正在执行的任务保留，恢复后继续使用。", "muted");
  const status = node("p", undefined, "archive-status"); status.setAttribute("role", "status");
  const list = node("div", undefined, "project-archive-list");
  dialog.append(heading, description, status, list); document.body.append(dialog);
  let generation = 0;
  const pending = new Set();

  async function update(agent, project, archived) {
    const account = state.user?.id;
    if (!account) throw Error("请先登录");
    const key = `${account}/${agent.id}/${project.id}`;
    if (pending.has(key)) return;
    pending.add(key);
    try {
      const result = await api(`/api/agents/${encodeURIComponent(agent.id)}/projects/${encodeURIComponent(project.id)}/archive`, "PATCH", {archived});
      if (state.user?.id !== account) return;
      state.catalogRevision = (state.catalogRevision || 0) + 1;
      const current = state.agents.find(a => a.id === agent.id);
      const target = current?.services?.find(s => s.id === project.id);
      if (target) target.archivedAt = result.archivedAt || 0;
      WorkspaceCache.write("agents", [account], state.agents);
      renderDevices();
      if (dialog.open) render();
      notice(archived ? "项目已归档，可在账号菜单的“已归档项目”中恢复" : "项目已恢复到侧栏");
    } finally { pending.delete(key); }
  }
  window.archiveProject = (agent, project) => update(agent, project, true);
  function render() {
    list.replaceChildren(); status.textContent = "";
    let count = 0;
    for (const agent of state.agents) for (const project of agent.services || []) {
      if (!project.archivedAt || project.scope) continue;
      count++;
      const row = node("div", undefined, "archived-project-row"), info = node("div");
      info.append(node("strong", project.name || "未命名项目"), node("p", `${agent.name} · ${project.provider === "qoder" ? "Qoder" : "Codex"} · ${agent.online ? "在线" : "离线"}`, "muted"));
      const restore = node("button", "恢复", "secondary"); restore.type = "button";
      restore.setAttribute("aria-label", `恢复 ${project.name || "未命名项目"}（${agent.name}）`);
      restore.onclick = async () => {
        const revision = generation, account = state.user?.id;
        restore.disabled = true; status.textContent = "";
        try { await update(agent, project, false); }
        catch (error) { if (dialog.open && generation === revision && state.user?.id === account) status.textContent = error.message; }
        finally { restore.disabled = false; }
      };
      row.append(info, restore); list.append(row);
    }
    if (!count) list.append(node("p", "暂无已归档项目", "archive-empty"));
  }
  window.openArchivedProjects = async () => {
    const account = state.user?.id, revision = ++generation;
    render(); dialog.showModal(); status.textContent = "正在刷新…";
    await refresh();
    if (dialog.open && revision === generation && state.user?.id === account) render();
  };
  dialog.addEventListener("close", () => { generation++; });
  window.addEventListener("remote-agent-disconnected", () => {
    if (!state.user || $("app").hidden) { dialog.close(); list.replaceChildren(); }
  });
})();
