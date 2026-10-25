/* Client history is authoritative; the browser keeps account-scoped drafts and a bounded cache. */
const $ = (id) => document.getElementById(id);
const basePath = new URL(".", location.href).pathname.replace(/\/$/, "");
const state = {
  register: false,
  mode: "web",
  agents: [],
  selected: null,
  peer: null,
  services: [],
  providers: {},
  roots: [],
  allowWrite: false,
  permissionMode: "",
  attachments: false,
  serviceId: null,
  sessions: [],
  session: null,
  entries: [],
  running: null,
  runSession: null,
  pending: new Map(),
  expanded: new Set(),
  listRevision: 0,
  readRevision: 0,
  listCursor: 0,
  listExtended: false,
};
let noticeTimer, searchTimer, drawerOpener, closeSayso;
function draftContext(sessionID=state.session?.id) { return [state.user?.id,state.selected?.id,state.serviceId,sessionID]; }
function saveCurrentDraft() {
  if(!state.session || !state.user || !state.selected || !state.serviceId)return;
  WorkspaceCache.write("draft",draftContext(),{text:$("prompt").value,attachments:window.chatAttachments?.snapshot() || []});
}
function restoreCurrentDraft() {
  const context=draftContext(), saved=WorkspaceCache.read("draft",context), pending=WorkspaceCache.read("pending",context);
  const draft=saved?.text || saved?.attachments?.length ? saved : pending || saved;
  $("prompt").value=draft?.text || "";
  window.chatAttachments?.restore(draft?.attachments || []);
}
function cacheCurrentSession() {if(state.session && state.user && state.selected)WorkspaceCache.history(draftContext(),state.session,state.entries)}
function reconcilePendingDraft() {
  const pending=WorkspaceCache.read("pending",draftContext());
  if(pending && state.entries.some(e=>e.role==="user" && e.turnId===pending.turnId)){
    WorkspaceCache.remove("pending",draftContext());
    const draft=WorkspaceCache.read("draft",draftContext());
    if(!draft?.text && !draft?.attachments?.length && $("prompt").value===pending.text){$("prompt").value="";window.chatAttachments?.restore([])}
  }
}
function updateSessionContext() {
  const project=service();if(!project)return;
  const target=$("selected-name");target.replaceChildren(document.createTextNode(`${state.selected?.name || "客户端"} · ${project.provider} · ${folderName(project.workspace)}`));
  if(state.session){const sessionID=state.session.id,id=node("code","Session "+sessionID,"session-id");id.title="会话 ID（点击复制）："+sessionID;id.tabIndex=0;id.setAttribute("role","button");id.onclick=async()=>{try{await navigator.clipboard.writeText(sessionID);notice("Session ID 已复制")}catch{notice(sessionID)}};id.onkeydown=e=>{if(e.key==="Enter"||e.key===" "){e.preventDefault();id.click()}};target.append(id)}
  if(state.historyLocal)target.append(node("span","本地缓存","cache-label"));
}
function notice(message) {
  $("notice").textContent = message;
  $("notice").hidden = false;
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => ($("notice").hidden = true), 8000);
}
window.onWorkspaceCacheError=()=>notice("浏览器存储暂不可用，草稿当前保留在本页；刷新前请复制保存。");
async function api(path, method = "GET", body) {
  const res = await fetch(basePath + path, {
    method,
    credentials: "same-origin",
    headers: body ? { "Content-Type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json();
  if (!res.ok) {
    if (res.status === 401 && !["/api/login", "/api/register"].includes(path))
      signedOut();
    throw Error(data.error || "请求失败");
  }
  return data;
}
function node(tag, text, className) {
  const el = document.createElement(tag);
  if (text !== undefined) el.textContent = text;
  if (className) el.className = className;
  return el;
}
function signedOut() {
  saveCurrentDraft();
  window.platformSignedOut?.();
  state.user = null;
  disconnect();
  if (closeSayso) {
    closeSayso();
    closeSayso = null;
  }
  $("sayso-root").hidden = true;
  $("app").hidden = true;
  $("auth").hidden = false;
  closeDrawers();
  for (const d of document.querySelectorAll("dialog[open]")) d.close();
}
async function signedIn() {
  state.user = await api("/api/me");
  $("account-name").textContent = state.user.username;
  window.platformSignedIn?.(state.user);
  $("auth").hidden = true;
  $("app").hidden = false;
  await refresh();
  await openLinkedSession();
  if(!location.hash.startsWith("#client=")){
    const saved=WorkspaceCache.read("selection",[state.user.id]),client=state.agents.find(a=>a.id===saved?.client);
    if(client){try{await connectDevice(client);await selectProject(saved.project);if(saved.session)await loadSession({id:saved.session})}catch(error){notice(error.message)}}
  }
}
async function refresh() {
  if ($("app").hidden) return;
  try {
    const agents = await api("/api/agents");
    const signature = rows => JSON.stringify(rows.map(a => [a.id,a.name,a.os,a.online,a.enrolled,a.services]));
    const changed = signature(agents) !== signature(state.agents);
    state.agents = agents;
    if (changed) renderDevices();
  } catch (e) {
    notice(e.message);
  }
}
function folderName(path) {
  return (
    path
      ?.replace(/[\\/]+$/, "")
      .split(/[\\/]/)
      .pop() ||
    path ||
    "项目"
  );
}
function service() {
  return state.services.find((s) => s.id === state.serviceId);
}
function renderDevices() {
  $("device-count").textContent = state.agents.length;
  $("device-empty").hidden = state.agents.length > 0;
  const list = $("device-list"),
    fragment = document.createDocumentFragment();
  for (const a of state.agents) {
    const selected = state.selected?.id === a.id;
    const card = node(
      "details",
      undefined,
      "device-node" + (selected ? " selected" : ""),
    );
    card.open = state.expanded.has(a.id) || selected;
    card.ontoggle = () => {
      if (card.open) state.expanded.add(a.id);
      else state.expanded.delete(a.id);
    };
    const row = node("summary", undefined, "device-row");
    const os = ({windows:["windows","Windows"],darwin:["macos","macOS"],linux:["linux","Linux"]})[a.os] || ["windows","客户端"];
    const icon = node("img", undefined, "device-icon"); icon.src = `${basePath}/icons/${os[0]}.svg`; icon.alt = os[1]; icon.title = os[1];
    row.append(icon, node("span", a.name, "device-name"));
    const add = node("button", "＋", "device-add");
    add.title = "新增项目"; add.setAttribute("aria-label", `在 ${a.name} 新增项目`); add.disabled = !a.online;
    add.onclick = async e => { e.preventDefault(); e.stopPropagation(); try { await connectDevice(a); showService(); } catch (e) { notice(e.message); } };
    const status = node("span", undefined, "device-status" + (a.online ? "" : " offline"));
    status.title = a.online ? "在线" : "离线";
    status.append(node("span", undefined, "live-dot" + (a.online ? "" : " offline")), document.createTextNode(status.title));
    const menu = node("button", "⋯", "device-menu-button"); menu.title = "设备设置与修复";
    menu.setAttribute("aria-label", `${a.name} 的设备操作`); menu.setAttribute("aria-haspopup", "menu");
    menu.onclick = e => window.openDeviceMenu(e, a, menu);
    row.oncontextmenu = e => window.openDeviceMenu(e, a, menu);
    row.onkeydown = e => { if ((e.shiftKey && e.key === "F10") || e.key === "ContextMenu") window.openDeviceMenu(e, a, menu); };
    row.append(status, add, menu); card.append(row);
    const children = node("div", undefined, "device-children");
    const services = selected ? state.services : a.services || [];
    for (const project of services) {
      const button = node("button", undefined, "project-button" + (selected && state.serviceId === project.id ? " active" : ""));
      button.title = project.workspace;
      const provider = node("span", project.provider === "qoder" ? "◇" : "◈", "project-provider");
      provider.title = project.provider === "qoder" ? "Qoder" : "Codex"; provider.setAttribute("aria-label", provider.title);
      button.append(provider, node("span", project.name || folderName(project.workspace)));
      button.onclick = async () => { try { await connectDevice(a); await selectProject(project.id); closeDrawers(); } catch (e) { notice(e.message); } };
      const projectRow = node("div", undefined, "project-row");
      const settings = node("button", "⚙", "project-settings");
      settings.title = "项目操作"; settings.setAttribute("aria-label", `${project.name} 的项目操作`);
      settings.onclick=e=>window.openProjectMenu(e,a,project,settings);
      projectRow.oncontextmenu=e=>window.openProjectMenu(e,a,project,settings);
      projectRow.append(button, settings);
      children.append(projectRow);
    }
    card.append(children);
    fragment.append(card);
  }
  list.replaceChildren(fragment);
}
function connectionStatus(text, good = false) {
  $("connection-status").textContent = text;
  $("connection-status").className = "status" + (good ? "" : " neutral");
}
function setRunning(id) {
  state.running = id;
  const background = ["queued","running"].includes(state.session?.taskStatus);
  $("cancel-run").hidden = !id && !(background && state.session?.taskId);
  $("send-run").disabled = !!id || background || !state.session || !service() || !state.peer || state.permissionMode!=="full";
  $("new-session").disabled = !!id || !service() || !state.peer || state.permissionMode!=="full";
}
function resetWorkspace() {
  saveCurrentDraft();
  state.historyLocal=false;
  state.mode = "web";
  state.serviceId = null;
  state.sessions = [];
  state.session = null;
  state.entries = [];
  $("prompt").value = "";
  state.listRevision++;
  state.readRevision++;
  state.listCursor = 0;
  renderSessions();
  renderChat();
  $("project-label").textContent = "先从左侧选择一个项目";
  $("selected-name").textContent =
    state.selected?.name || "连接设备，打开一个项目";
}
function disconnect(reason) {
  saveCurrentDraft();
  closeDirectoryPicker();
  window.dispatchEvent(new Event("remote-agent-disconnected"));
  const peer = state.peer;
  state.peer = null;
  if (peer) {
    clearInterval(peer.ping);
    clearTimeout(peer.timeout);
    peer.reject?.(Error(reason || "连接已断开"));
    peer.ws.close();
  }
  for (const entry of state.pending.values()) {
    clearTimeout(entry.timer);
    entry.reject(Error(reason || "连接已断开"));
  }
  state.pending.clear();
  state.selected = null;
  state.services = [];
  state.providers = {};
  state.permissionMode="";state.clientVersion="";state.attachments=false;
  state.roots = [];
  state.runSession = null;
  resetWorkspace();
  setRunning(null);
  connectionStatus(reason || "未连接");
  renderDevices();
}
function rpc(type, fields = {}) {
  return new Promise((resolve, reject) => {
    const peer = state.peer;
    if (!peer || peer.ws.readyState !== WebSocket.OPEN) {
      reject(Error("设备尚未连接"));
      return;
    }
    const id = crypto.randomUUID();
    const timer = setTimeout(() => {
      state.pending.delete(id);
      reject(Error("设备响应超时"));
    }, type === "channels.configure" ? 60000 : 20000);
    state.pending.set(id, { resolve, reject, timer, chunks: [], chunkSize: 0 });
    try {
      sendRequest(peer, { id, type, ...fields });
    } catch (e) {
      clearTimeout(timer);
      state.pending.delete(id);
      reject(e);
    }
  });
}
function refreshServices(data) {
  state.directorySearch = !!data.directorySearch;
  state.sayso = !!data.sayso;
  state.services = data.services || [];
  state.providers = data.providers || {};
  state.roots = data.allowedRoots || [];
  state.allowWrite = !!data.allowWrite;
  state.permissionMode=data.permissionMode || "";
  state.clientVersion=data.clientVersion || "";
  state.attachments=!!data.attachments;
  if(state.user && state.selected && state.peer)WorkspaceCache.write("client",[state.user.id,state.selected.id],data);
  renderDevices();
}
function sendRequest(peer, request) {
  const message = JSON.stringify({ type: "request", request });
  if (new TextEncoder().encode(JSON.stringify(request)).length > 64 * 1024)
    throw Error("请求超过 64 KiB 上限");
  if (peer.ws.bufferedAmount > 1024 * 1024) throw Error("连接繁忙，请稍后重试");
  peer.ws.send(message);
}
function receiveEvent(msg) {
          const pending = state.pending.get(msg.id);
          if (pending && msg.type === "chunk") {
            const bytes = Uint8Array.from(atob(msg.data), (c) =>
              c.charCodeAt(0),
            );
            pending.chunkSize += bytes.length;
            if (pending.chunkSize > 32 * 1024 * 1024) {
              clearTimeout(pending.timer);
              state.pending.delete(msg.id);
              pending.reject(Error("设备响应过大"));
              return;
            }
            pending.chunks.push(bytes);
            clearTimeout(pending.timer);
            pending.timer = setTimeout(() => {
              state.pending.delete(msg.id);
              pending.reject(Error("设备响应超时"));
            }, 20000);
            return;
          }
          if (pending && (msg.type === "result" || msg.type === "error")) {
            clearTimeout(pending.timer);
            state.pending.delete(msg.id);
            if (msg.error) {
              pending.reject(
                Error(
                  msg.error === "unknown request type"
                    ? "请更新目标电脑上的 Agent，以支持这项功能"
                    : msg.error,
                ),
              );
            } else {
              try {
                pending.resolve(
                  msg.data?.chunked ? decodeChunks(pending) : msg.data,
                );
              } catch {
                pending.reject(Error("设备返回了无效的分片数据"));
              }
            }
            return;
          }
          if (msg.id !== state.running) return;
          if(msg.type==="started" && state.runContext)WorkspaceCache.remove("pending",state.runContext);
          if (msg.type === "output") {
            if (state.session?.id === state.runSession) {
              if (msg.data?.entries) mergeEntries(msg.data.entries, true);
              else if (msg.data?.text)
                mergeEntries(
                  [
                    {
                      seq: Date.now(),
                      turnId: msg.id,
                      role: "assistant",
                      text: msg.data.text,
                      kind: "stdout",
                    },
                  ],
                  true,
                );
            }
          } else if (msg.type === "done" || msg.type === "error") {
            const sid = state.runSession;
            setRunning(null);
            state.runSession = null;
            if (msg.error) notice(msg.error);
            if (state.session?.id === sid)
              loadSession(state.session).catch((e) => notice(e.message));
            refreshSessions().catch((e) => notice(e.message));
          }
}
async function connectDevice(agent) {
  if (state.selected?.id === agent.id && state.peer) return state.peer.ready;
  if (!agent.online) {
    const cached=WorkspaceCache.read("client",[state.user?.id,agent.id]);
    if(!cached)throw Error("客户端离线，尚无本地会话缓存");
    if(state.running)throw Error("请先停止当前任务再打开离线缓存");
    if(closeSayso && !(await closeSayso()))throw Error("请先保存录音");
    disconnect();state.selected=agent;state.expanded.add(agent.id);refreshServices(cached);connectionStatus("离线 · 本地缓存");return cached;
  }
  if (state.running && !confirm("切换设备会停止当前任务，继续吗？"))
    return Promise.reject(Error("已取消切换"));
  if (closeSayso && !(await closeSayso())) throw Error("请先保存 SaySo 录音");
  disconnect();
  state.selected = agent;
  state.expanded.add(agent.id);
  renderDevices();
  connectionStatus("正在连接设备…");
  const ws = new WebSocket(`${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}${basePath}/api/agents/${agent.id}/connect`);
  const peer = { ws, lastPong: Date.now() };
  peer.ready = new Promise((resolve, reject) => { peer.resolve = resolve; peer.reject = reject; });
  state.peer = peer;
  const fail = (msg) => { if (state.peer === peer) { notice(msg); disconnect(msg); } };
  peer.timeout = setTimeout(() => fail("连接超时，请检查客户端是否在线并已更新"), 20000);
  ws.onclose = () => fail("设备连接已断开，请重新连接");
  ws.onerror = () => fail("连接暂时不可用，请刷新设备状态");
  ws.onopen = () => {
    peer.ping = setInterval(() => {
      if (Date.now() - peer.lastPong > 65000) { fail("设备连接超时，请重新连接"); return; }
      if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: "ping" }));
    }, 20000);
  };
  ws.onmessage = async (event) => {
    if (state.peer !== peer) return;
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === "ready") {
        if (msg.transport !== "relay-v1") throw Error("请更新服务器和客户端");
        const data = await rpc("services.list");
        if (state.peer !== peer) return;
        clearTimeout(peer.timeout);
        refreshServices(data);
        $("selected-name").textContent = agent.name;
        connectionStatus("已连接", true);
        peer.resolve(data);
        peer.reject = null;
      } else if (msg.type === "event" && msg.event) receiveEvent(msg.event);
      else if (msg.type === "pong") peer.lastPong = Date.now();
      else if (msg.type === "error") fail(msg.error || "连接失败");
    } catch (e) { fail(e.message); }
  };
  return peer.ready;
}
async function selectProject(id) {
  if (state.running && id !== state.serviceId)
    throw Error("请先停止当前任务，再切换项目");
  if (!state.services.some((s) => s.id === id))
    throw Error("项目已被移除，请刷新设备");
  if (id === state.serviceId) return;
  saveCurrentDraft();
  if (closeSayso && !(await closeSayso())) throw Error("请先保存 SaySo 录音");
  state.mode = "web";
  state.serviceId = id;
  state.session = null;
  state.entries = [];
  state.readRevision++;
  $("session-search").value = "";
  const s = service();
  $("project-label").textContent = `${s.provider.toUpperCase()} / ${s.name}`;
  $("project-label").title = s.workspace;
  $("selected-name").textContent =
    `${state.selected.name} · ${s.provider} · ${folderName(s.workspace)}`;
  setRunning(state.running);
  renderDevices();
  renderChat();
  await refreshSessions();
}
async function refreshSessions(more = false, quiet = false) {
  if (!state.serviceId) return;
  const cacheScope=[state.user?.id,state.selected?.id,state.serviceId];
  if (!more && !quiet && !state.sessions.length) {const cached=WorkspaceCache.read("sessions",cacheScope);if(cached){state.sessions=cached;renderSessions();}}
  if(!state.peer)return;
  const revision = ++state.listRevision;
  const sid = state.serviceId;
  const page = await rpc("sessions.list", {
    serviceId: sid,
    search: $("session-search").value.trim(),
    cursor: more ? state.listCursor : 0,
  });
  if (revision !== state.listRevision || sid !== state.serviceId) return;
  const oldCursor = state.listCursor;
  const rows = more ? [...state.sessions,...page.sessions] : quiet ? [...page.sessions,...state.sessions.filter(s => !page.sessions.some(n => n.id === s.id))] : page.sessions;
  state.sessions = [...new Map(rows.map(s => [s.id,s])).values()];
  WorkspaceCache.write("sessions",cacheScope,state.sessions);
  if (more) state.listExtended = true;
  else if (!quiet) state.listExtended = false;
  state.listCursor = quiet && state.listExtended ? oldCursor : page.nextCursor || 0;
  renderSessions();
}
let sessionSignature;
function renderSessions() {
  $("conversation-count").textContent = state.sessions.length + (service() ? 1 : 0);
  $("more-sessions").hidden = !state.listCursor;
  $("conversation-empty").hidden = state.sessions.length > 0;
  $("conversation-empty").textContent = state.serviceId
    ? $("session-search").value
      ? "没有找到匹配的会话。"
      : "还没有会话。\n点击右上角 ＋ 开始。"
    : "连接设备，选择项目，\n从一个新会话开始。";
  const signature = JSON.stringify([state.serviceId,state.mode,state.session?.id,state.runSession,state.sessions,service()?.channel]);
  if (signature === sessionSignature) return;
  sessionSignature = signature;
  const fragment = document.createDocumentFragment();
  if (service()) {
    const sayso = node("button", undefined, "conversation-card pinned" + (state.mode === "sayso" ? " active" : ""));
    sayso.append(node("strong", "♫ SaySo"), node("p", "置顶 · 讨论、录音与 Action"));
    sayso.onclick = () => openSayso().catch(e => notice(e.message)); fragment.append(sayso);
  }
  const channel = service()?.channel;
  if (channel && !state.sessions.some(s => s.kind === "im")) {
    const im = node("button", undefined, "conversation-card pinned" + (state.session?.id === channel.sessionId && state.mode !== "sayso" ? " active" : ""));
    im.append(node("strong", channel.provider === "feishu" ? "飞书 · 渠道" : "钉钉 · 渠道"), node("p", "等待第一条消息"));
    im.onclick = () => loadSession({id:channel.sessionId,title:"渠道会话",kind:"im"}).then(closeDrawers).catch(e => notice(e.message)); fragment.append(im);
  }
  for (const s of [...state.sessions].sort((a,b)=>(b.kind==="im")-(a.kind==="im"))) {
    const b = node(
      "button",
      undefined,
      "conversation-card" + (s.kind === "im" ? " im-session" : "") + (state.session?.id === s.id ? " active" : ""),
    );
    const title = sessionTitle(s);
    b.title = title;
    b.append(node("strong", title));
    const meta = node("p");
    meta.append(
      node(
        "span",
        new Date(s.updatedAt).toLocaleString("zh-CN", {
          month: "numeric",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit",
        }),
      ),
      node("span", s.id === state.runSession ? "生成中…" : ({queued:"排队中",running:"执行中",failed:"执行失败",cancelled:"已取消"})[s.taskStatus] || (s.kind === "im" ? s.channelChatType === "group" ? "群聊" : s.channelChatId ? "私聊" : "渠道历史" : "Web 会话")),
    );
    b.append(meta);
    b.onclick = () =>
      loadSession(s)
        .then(closeDrawers)
        .catch((e) => notice(e.message));
    fragment.append(b);
  }
  $("conversation-list").replaceChildren(fragment);
}
async function loadSession(session) {
  if (closeSayso && !(await closeSayso())) return;
  state.mode = "web";
  const revision = ++state.readRevision;
  saveCurrentDraft();
  const context=draftContext(session.id), cached=WorkspaceCache.read("history",context);
  state.historyLocal=!!cached;
  state.session = cached?.session?.serviceId===state.serviceId ? cached.session : session;
  state.entries = cached?.entries || [];
  WorkspaceCache.write("selection",[state.user?.id],{client:state.selected?.id,project:state.serviceId,session:session.id});
  restoreCurrentDraft();
  renderSessions();
  renderChat();
  $("chat-scroll").scrollTop = 0;
  setRunning(state.running);
  if(!state.peer)return;
  let cursor = 0,
    entries = [];
  do {
    const page = await rpc("sessions.get", { sessionId: session.id, cursor });
    if (revision !== state.readRevision) return;
    if (page.session.serviceId !== state.serviceId) throw Error("会话不属于当前项目");
    state.session = page.session;
    entries.push(...page.entries);
    cursor = page.nextCursor || 0;
  } while (cursor);
  if (revision !== state.readRevision) return;
  const seq = new Set(entries.map((e) => e.seq));
  entries.push(...state.entries.filter((e) => e.turnId===state.running && e.seq > 0 && !seq.has(e.seq)));
  state.entries = entries.sort((a, b) => a.seq - b.seq);
  state.historyLocal=false;
  reconcilePendingDraft(); cacheCurrentSession();
  renderChat();
  $("chat-scroll").scrollTop = $("chat-scroll").scrollHeight;
}
function sessionTitle(session) {
  if (session?.kind !== "im") return session?.title || "新会话";
  const channel = ({feishu:"飞书",dingtalk:"钉钉"})[session.channelName] || session.channelName || (session.channelProvider === "dingtalk" ? "钉钉" : "飞书");
  return session.channelPeerName ? `${channel} · ${session.channelPeerName}` : session.title || `${channel} · 渠道历史`;
}
function mergeEntries(entries, live = false) {
  const seq = new Set(state.entries.map((e) => e.seq));
  state.entries.push(...entries.filter((e) => !seq.has(e.seq)));
  state.entries.sort((a, b) => a.seq - b.seq);
  if(live)reconcilePendingDraft(); cacheCurrentSession();
  renderChat(live);
}
function renderChat(live = false) {
  updateSessionContext();
  $("send-run").textContent = state.session?.kind === "im" ? "发送到渠道 ↗" : "发送 ↗";
  $("prompt").placeholder = state.session?.kind === "im" ? "向这个项目的 IM 渠道发送消息…" : "输入你的想法，交给项目的执行器…";
  $("sayso-root").hidden = state.mode !== "sayso";
  if (state.mode === "sayso") {
    $("session-empty").hidden = true; $("session-body").hidden = true; $("chat-title").textContent = "SaySo · " + (service()?.name || "讨论"); return;
  }
  $("session-empty").hidden = !!state.session;
  $("session-body").hidden = !state.session;
  $("run-form").hidden = !state.session;
  $("chat-title").textContent =
    state.session ? sessionTitle(state.session) : (service() ? service().name : "开始你的下一件事");
  $("policy").textContent = !state.peer ? "离线草稿 · 连接后可发送" : state.permissionMode!=="full" ? "客户端待更新，尚未确认完整授权" : state.session?.kind === "im" ? "发送到此渠道会话" : "完整授权 · 在你的电脑执行";
  $("policy").title = "客户端版本："+(state.clientVersion||"旧客户端未上报");
  $("fix-client").hidden = state.permissionMode==="full";
  $("add-attachment").disabled = !state.peer || !state.attachments || state.session?.kind==="im";
  $("chat-hint").hidden = state.entries.length > 0;
  const container = $("messages"), scroll = $("chat-scroll"),
    atBottom =
      scroll.scrollHeight - scroll.scrollTop - scroll.clientHeight <
      90;
  if (!state.session) {
    container.replaceChildren();
    return;
  }
  const groups = [];
  for (const e of state.entries) {
    if (e.kind === "done") continue;
    const key = `${e.turnId}/${e.role}/${e.kind}`;
    let g = groups.find((g) => g.key === key);
    if (!g) {
      g = { key, role: e.role, text: "", turnId: e.turnId, kind: e.kind };
      groups.push(g);
    }
    g.text += e.text;
  }
  const existing = new Map(
    [...container.children].map((el) => [el.dataset.key, el]),
  );
  const keep = new Set();
  for (const g of groups) {
    keep.add(g.key);
    let el = existing.get(g.key);
    if (!el) {
      el = node(
        g.role === "system" ? "details" : "article",
        undefined,
        "message " + g.role,
      );
      el.dataset.key = g.key;
      const label =
        g.role === "user"
          ? "你"
          : g.role === "assistant"
            ? service()?.provider.toUpperCase() || "AGENT"
            : g.kind === "error"
              ? "运行错误"
              : "运行详情";
      el.append(
        node(g.role === "system" ? "summary" : "div", label, "message-label"),
        node("div", "", "message-text"),
      );
      container.append(el);
    }
    const text = el.querySelector(".message-text");
    if (g.role === "assistant") renderMarkdown(text, g.text);
    else text.textContent = g.text;
    window.chatAttachments?.renderHistory(el, state.entries.filter(e=>e.turnId===g.turnId && e.role===g.role).flatMap(e=>e.attachments || []));
    el.classList.toggle(
      "typing",
      g.role === "assistant" && g.turnId === state.running,
    );
  }
  for (const [key, el] of existing) if (!keep.has(key)) el.remove();
  if (atBottom) scroll.scrollTop = scroll.scrollHeight;
}

function initPaneResizers() {
  const app = $("app"),
    panes = [$("devices-pane"), $("conversations-pane"), $("chat-pane")],
    resizers = [$("devices-resizer"), $("conversations-resizer")],
    properties = ["--devices-width", "--conversations-width", "--chat-width"],
    defaults = [0.21, 0.23, 0.56],
    storageKey = `remote-agent:column-widths:${basePath || "/"}`;
  let ratios = [...defaults],
    drag = null;
  try {
    const saved = JSON.parse(localStorage.getItem(storageKey));
    if (
      Array.isArray(saved) &&
      saved.length === 3 &&
      saved.every(
        (value) => Number.isFinite(value) && value > 0 && value < 1,
      ) &&
      Math.abs(saved.reduce((sum, value) => sum + value, 0) - 1) < 0.001
    )
      ratios = saved.map(
        (value) => value / saved.reduce((sum, item) => sum + item, 0),
      );
  } catch {}

  function save() {
    try {
      localStorage.setItem(storageKey, JSON.stringify(ratios));
    } catch {}
  }
  function availableWidth() {
    return (
      app.clientWidth - resizers.reduce((sum, el) => sum + el.offsetWidth, 0)
    );
  }
  function minimums(width) {
    const scale = Math.min(1, width / 660);
    return [180, 180, 300].map((value) => value * scale);
  }
  function render() {
    if (app.hidden || innerWidth <= 760) return;
    const width = availableWidth();
    if (width <= 0) return;
    const mins = minimums(width),
      desired = ratios.map((ratio) => ratio * width),
      deficit = desired.reduce(
        (sum, value, i) => sum + Math.max(0, mins[i] - value),
        0,
      ),
      excess = desired.reduce(
        (sum, value, i) => sum + Math.max(0, value - mins[i]),
        0,
      ),
      widths = desired.map((value, i) =>
        value < mins[i]
          ? mins[i]
          : value - (excess ? (deficit * (value - mins[i])) / excess : 0),
      );
    widths.forEach((value, i) =>
      app.style.setProperty(properties[i], `${value}px`),
    );
    resizers.forEach((el, i) => {
      el.setAttribute("aria-valuemin", Math.round((100 * mins[i]) / width));
      el.setAttribute(
        "aria-valuemax",
        Math.round((100 * (widths[i] + widths[i + 1] - mins[i + 1])) / width),
      );
      el.setAttribute("aria-valuenow", Math.round((100 * widths[i]) / width));
      el.setAttribute("aria-valuetext", `${Math.round(widths[i])} 像素`);
    });
  }
  function resize(index, start, delta) {
    const width = start.reduce((sum, value) => sum + value, 0),
      mins = minimums(width),
      pair = start[index] + start[index + 1],
      next = [...start];
    next[index] = Math.max(
      mins[index],
      Math.min(pair - mins[index + 1], start[index] + delta),
    );
    next[index + 1] = pair - next[index];
    ratios = next.map((value) => value / width);
    render();
  }
  function finishDrag() {
    if (!drag) return;
    const { el, pointerId } = drag;
    drag = null;
    el.classList.remove("resizing");
    document.body.classList.remove("resizing-panes");
    if (el.hasPointerCapture(pointerId)) el.releasePointerCapture(pointerId);
    save();
  }
  resizers.forEach((el, index) => {
    el.addEventListener("pointerdown", (event) => {
      if (event.button !== 0 || !event.isPrimary || drag || innerWidth <= 760)
        return;
      event.preventDefault();
      el.focus({ preventScroll: true });
      drag = {
        el,
        pointerId: event.pointerId,
        x: event.clientX,
        widths: panes.map((pane) => pane.getBoundingClientRect().width),
      };
      el.setPointerCapture(event.pointerId);
      el.classList.add("resizing");
      document.body.classList.add("resizing-panes");
    });
    el.addEventListener("pointermove", (event) => {
      if (drag?.el === el && drag.pointerId === event.pointerId)
        resize(index, drag.widths, event.clientX - drag.x);
    });
    for (const type of ["pointerup", "pointercancel", "lostpointercapture"])
      el.addEventListener(type, (event) => {
        if (drag?.el === el && drag.pointerId === event.pointerId) finishDrag();
      });
    el.addEventListener("keydown", (event) => {
      if (
        !["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key) ||
        drag
      )
        return;
      event.preventDefault();
      const delta =
        event.key === "Home"
          ? -Infinity
          : event.key === "End"
            ? Infinity
            : (event.key === "ArrowLeft" ? -1 : 1) * (event.shiftKey ? 48 : 16);
      resize(
        index,
        panes.map((pane) => pane.getBoundingClientRect().width),
        delta,
      );
      save();
    });
    el.addEventListener("dblclick", () => {
      ratios = [...defaults];
      render();
      save();
    });
  });
  new ResizeObserver(() => {
    finishDrag();
    render();
  }).observe(app);
  window.addEventListener("blur", finishDrag);
  render();
}
initPaneResizers();

function closeDrawers() {
  for (const p of [$("devices-pane"), $("conversations-pane")]) {
    p.classList.remove("drawer-open");
    p.removeAttribute("aria-modal");
    p.removeAttribute("role");
  }
  $("drawer-backdrop").hidden = true;
  $("open-devices").setAttribute("aria-expanded", "false");
  $("open-conversations").setAttribute("aria-expanded", "false");
  if (drawerOpener) {
    drawerOpener.focus();
    drawerOpener = null;
  }
}
function openDrawer(id, opener) {
  closeDrawers();
  drawerOpener = opener;
  const pane = $(id);
  pane.classList.add("drawer-open");
  pane.setAttribute("role", "dialog");
  pane.setAttribute("aria-modal", "true");
  $("drawer-backdrop").hidden = false;
  opener.setAttribute("aria-expanded", "true");
  pane.querySelector(".close-drawer").focus();
}
$("open-devices").onclick = () => openDrawer("devices-pane", $("open-devices"));
$("open-conversations").onclick = () =>
  openDrawer("conversations-pane", $("open-conversations"));
$("drawer-backdrop").onclick = closeDrawers;
for (const b of document.querySelectorAll(".close-drawer"))
  b.onclick = closeDrawers;
document.addEventListener("keydown", (e) => {
  const pane = document.querySelector(".drawer-open");
  if (!pane) return;
  if (e.key === "Escape") {
    closeDrawers();
    return;
  }
  if (e.key === "Tab") {
    const targets = [...pane.querySelectorAll("a,button,input,summary")].filter(
      (el) => !el.disabled && el.getClientRects().length,
    );
    const first = targets[0],
      last = targets.at(-1);
    if (e.shiftKey && document.activeElement === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && document.activeElement === last) {
      e.preventDefault();
      first.focus();
    }
  }
});
window.addEventListener("resize", () => {
  if (innerWidth > 760) closeDrawers();
});
$("session-search").oninput = () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(
    () => refreshSessions().catch((e) => notice(e.message)),
    200,
  );
};
$("more-sessions").onclick = () =>
  refreshSessions(true).catch((e) => notice(e.message));
$("new-session").onclick = async () => {
  try {
    const s = await rpc("sessions.create", { serviceId: state.serviceId });
    await refreshSessions();
    await loadSession(s);
    closeDrawers();
    $("prompt").focus();
  } catch (e) {
    notice(e.message);
  }
};
function showService() {
  const select = $("provider-select");
  select.replaceChildren();
  for (const name of ["codex", "qoder"]) {
    const p = state.providers[name],
      o = new Option(
        name + (p?.available ? " · 已检测到" : " · 未安装（可选）"),
        name,
      );
    o.disabled = !p?.available;
    select.append(o);
  }
  select.selectedIndex = [...select.options].findIndex((o) => !o.disabled);
  $("provider-info").textContent =
    select.selectedIndex < 0
      ? "设备连接正常。需要执行 AI 任务时，安装任意一种 CLI 并重启 Agent 即可。"
      : "选择本次需要的 Agent，另一种无需安装。";
  $("roots-info").textContent = "允许的根目录：" + state.roots.join("、");
  $("service-submit").disabled = select.selectedIndex < 0;
  $("service-dialog").showModal();
  closeDirectoryPicker();
  $("directory-status").textContent = state.directorySearch
    ? "输入目录名搜索，或输入路径补全；也可以直接填写绝对路径。"
    : "这台设备的 Agent 尚不支持目录搜索，可先填写绝对路径，或更新 Agent。";
}

$("auth-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = $("auth-submit");
  button.disabled = true;
  try {
    const data = Object.fromEntries(new FormData(e.target));
    await api("/api/login", "POST", data);
    $("account-name").textContent = data.username;
    e.target.elements.password.value = "";
    await signedIn();
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
$("logout").onclick = async () => {
  disconnect();
  try {
    await api("/api/logout", "POST", {});
    signedOut();
  } catch (e) {
    notice(e.message);
  }
};
$("refresh").onclick = refresh;
function showEnroll() {
  $("enroll-title").textContent = "添加一台设备";
  $("enroll-form").reset();
  $("enroll-form").hidden = false;
  $("install-result").hidden = true;
  $("repair-steps").hidden = true;
  $("install-command").value = "";
  $("repair-project-info").hidden=true;
  $("enroll-dialog").showModal();
}
$("new-agent").onclick = showEnroll;
$("empty-add").onclick = showEnroll;
function showInstallCommands(result) {
  if (result.repair) $("enroll-title").textContent = "修复设备绑定";
  $("repair-steps").hidden = !result.repair;
  const os=result.agent?.os, name=({windows:"Windows",darwin:"macOS",linux:"Linux"})[os];
  if(!name)throw Error("这台客户端尚未选择操作系统，请先在设置中选择");
  $("enroll-os-badge").textContent=name+" · 系统类型已固定";
  $("install-command").value=({windows:result.windows,darwin:result.macos,linux:result.linux})[os];
  $("install-command-label").firstChild.textContent=({windows:"在新 PowerShell 中执行",darwin:"在这台 Mac 的终端执行",linux:"在这台 Linux 的终端执行"})[os];
  $("install-guide").textContent=({windows:"下载适合这台电脑架构的 MSI 并安装。更新会保留原设备绑定和本机数据。",darwin:"下方命令会安装或更新客户端，并配置登录后自动启动。",linux:"下方命令会安装或更新客户端，并配置 systemd 用户服务。"})[os];
  $("enroll-msi-downloads").hidden=os!=="windows";
  $("binding-step-title").textContent=result.repair?"2. 修复本机绑定与服务":"2. 绑定这台设备";
  $("repair-log-path").textContent=os==="windows"?"运行日志：%USERPROFILE%\\.remote-agent\\agent.log；管理员配置日志：同目录 service-setup.log。":"运行日志：~/.remote-agent/agent.log。";
  $("repair-permission-note").hidden=os!=="windows";
  $("repair-recovery").hidden=!result.existingBinding;
  const downloadStamp = Date.now();
  $("enroll-msi-x64").href = result.windowsMsi + "?v=" + downloadStamp;
  $("enroll-msi-arm64").href = result.windowsArm64Msi + "?v=" + downloadStamp;
  $("enroll-command-note").textContent = result.rebind || result.recovery
    ? "恢复凭证 · 15 分钟有效 · 在这台设备执行，保留数据与设备身份；成功后旧凭证失效。"
    : result.existingBinding
    ? "在原设备执行，复用本地绑定并修复后台服务；保留项目、会话和录音。若本地配置丢失，请先恢复备份，勿创建重复设备。"
    : "一次性设备绑定凭证 · 15 分钟有效 · 请勿分享";
  $("enroll-form").hidden = true;
  $("install-result").hidden = false;
}
$("enroll-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = e.target.querySelector("button");
  button.disabled = true;
  try {
    const result = await api(
      "/api/agents",
      "POST",
      Object.fromEntries(new FormData(e.target)),
    );
    showInstallCommands(result);
    await refresh();
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
for (const button of document.querySelectorAll(".close-dialog"))
  button.onclick = () => button.closest("dialog").close();
$("enroll-dialog").addEventListener("close", () => {
  $("install-command").value = "";
});
for (const button of document.querySelectorAll(".copy"))
  button.onclick = async () => {
    try {
      await navigator.clipboard.writeText($(button.dataset.copy).value);
      notice("已复制命令");
    } catch {
      $(button.dataset.copy).select();
      notice("请按 Ctrl / ⌘ + C 复制");
    }
  };

$("service-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = $("service-submit");
  button.disabled = true;
  try {
    const s = await rpc("services.add", {
      service: Object.fromEntries(new FormData(e.target)),
    });
    refreshServices(await rpc("services.list"));
    $("service-dialog").close();
    e.target.reset();
    await selectProject(s.id);
    if (s.session) await loadSession(s.session);
    else if (state.sessions.length) await loadSession(state.sessions[0]);
    notice("项目已就绪，开始新会话吧");
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
$("run-form").onsubmit = async (e) => {
  e.preventDefault();
  if (state.running || $("send-run").disabled) return;
  const files=window.chatAttachments?.snapshot() || [];
  const prompt = $("prompt").value.trim() || (files.length ? "请查看附件。" : "");
  if (!prompt || !state.session || !service()) return;
  if(window.chatAttachments?.busy()){notice("附件正在上传，请稍候");return}
  if (!state.peer || state.peer.ws.readyState !== WebSocket.OPEN) {
    notice("请先连接设备");
    return;
  }
  if (state.session.kind === "im") {
    const context=draftContext(), session=state.session, text=$("prompt").value, peer=state.peer;
    const button = $("send-run"); button.disabled = true;
    try { await rpc("channels.notify", {body:{serviceId:state.serviceId,sessionId:session.id,message:prompt}}); WorkspaceCache.remove("draft",context); if(state.peer===peer && state.session?.id===session.id && $("prompt").value===text){$("prompt").value="";await loadSession(session)} }
    catch (e) { notice(e.message); } finally { button.disabled = false; }
    return;
  }
  const id = crypto.randomUUID();
  const context=draftContext();
  state.runContext=context;
  WorkspaceCache.write("pending",context,{text:$("prompt").value,attachments:files,turnId:id});
  state.runSession = state.session.id;
  setRunning(id);
  mergeEntries(
    [
      {
        seq: (state.entries.at(-1)?.seq || 0) + 1,
        turnId: id,
        role: "user",
        text: prompt,
        kind: "message",
        attachments:files,
      },
    ],
    false,
  );
  try {
    sendRequest(state.peer, {
        id,
        type: "run",
        serviceId: state.serviceId,
        sessionId: state.session.id,
        prompt,
        attachments:files.map(f=>f.id),
    });
    $("prompt").value = "";
    window.chatAttachments?.restore([]);
    WorkspaceCache.remove("draft",context);
    renderSessions();
  } catch (err) {
    setRunning(null);
    state.runSession = null;
    notice(err.message);
  }
};
$("prompt").onkeydown = (e) => {
  if (e.key === "Enter" && !e.shiftKey && !e.isComposing && e.keyCode !== 229) {
    e.preventDefault();
    $("run-form").requestSubmit();
  }
};
$("prompt").addEventListener("input",saveCurrentDraft);
document.addEventListener("visibilitychange",()=>{if(document.hidden)saveCurrentDraft()});
$("fix-client").onclick=()=>window.openDeviceRepair?.(state.selected).catch(e=>notice(e.message));
$("cancel-run").onclick = async () => {
  try {
    if (state.running) await rpc("cancel", {runId:state.running});
    else if (state.session?.taskId) await rpc("tasks.cancel", {runId:state.session.taskId});
  } catch (e) {
    notice(e.message);
  }
};
window.addEventListener("beforeunload", () => {
  saveCurrentDraft();
  state.peer?.ws.close();
});
setInterval(refresh, 10000);
let historyPolling = false;
setInterval(async () => {
  if (historyPolling || !state.peer || !service() || $("app").hidden) return;
  historyPolling = true;
  try { await refreshSessions(false, true); if (state.mode !== "sayso" && state.session && !state.running) await pollSession(); }
  catch {} finally { historyPolling = false; }
}, 4000);
(async () => {
  try {
    await signedIn();
  } catch (e) {
    if (e.message !== "please sign in") notice(e.message);
  }
})();

function decodeChunks(pending) {
  const data = new Uint8Array(pending.chunkSize);
  let offset = 0;
  for (const chunk of pending.chunks) {
    data.set(chunk, offset);
    offset += chunk.length;
  }
  return JSON.parse(new TextDecoder().decode(data));
}
async function openSayso() {
  saveCurrentDraft();
  if (state.mode === "sayso") { closeDrawers(); return; }
  if (!state.sayso || !state.peer || !service()) throw Error("请先选择项目并连接已更新的客户端");
  if (state.running) throw Error("请先等待当前任务结束，再打开 SaySo");
  const peer = state.peer, project = service();
  const binding = await rpc("sayso.open", {serviceId:project.id});
  const module = await import(`${basePath}/sayso/sayso.js`);
  if (state.peer !== peer || state.serviceId !== project.id) throw Error("项目已改变，请重新打开");
  closeDrawers(); state.mode = "sayso";
  const host = $("sayso-root");
  const dispose = module.mount(host, async (method,path,body) => {
    if (state.peer !== peer) throw Error("客户端连接已断开，请保存录音备份后重新连接");
    return rpc("sayso.call", {method,path,body,serviceId:project.id});
  }, () => {
    dispose(); if (state.peer !== peer) return; host.hidden = true; state.mode = "web"; closeSayso = null; renderSessions(); renderChat();
  }, `${project.name} · 完整授权`, {
    projectId:binding.id, sessionId:binding.pinnedSessionId,
    onTaskCreated:async id => {
      if (state.peer !== peer) return;
      await refreshSessions();
      notice("已创建项目下的新 Web 会话，可在会话列表查看执行进度");
    },
  });
  closeSayso = () => dispose.close();
  renderSessions(); renderChat();
}

// Directory suggestions are transient and fetched only from the connected Agent.
let directoryRevision = 0,
  directorySkipFocus = false,
  directoryTimer,
  directoryItems = [],
  directoryActive = -1,
  directoryPage;
function closeDirectoryPicker() {
  directoryRevision++;
  clearTimeout(directoryTimer);
  directoryItems = [];
  directoryActive = -1;
  $("directory-picker").hidden = true;
  $("directory-options").replaceChildren();
  $("workspace-input").setAttribute("aria-expanded", "false");
  $("workspace-input").removeAttribute("aria-activedescendant");
}
function activateDirectory(index) {
  directoryActive = index;
  const input = $("workspace-input");
  for (const [i, row] of [...$("directory-options").children].entries()) {
    row.setAttribute("aria-selected", String(i === index));
    row.classList.toggle("active", i === index);
  }
  const row = $("directory-options").children[index];
  if (row) {
    input.setAttribute("aria-activedescendant", row.id);
    row.scrollIntoView({ block: "nearest" });
  } else input.removeAttribute("aria-activedescendant");
}
function chooseDirectory(item) {
  $("workspace-input").value = item.path;
  closeDirectoryPicker();
  directorySkipFocus = true;
  $("workspace-input").focus({ preventScroll: true });
  directorySkipFocus = false;
  $("directory-status").textContent = "已选择目录。再次输入可继续搜索或修改。";
}
async function findDirectories(path = "", query = $("workspace-input").value) {
  if (!state.directorySearch || !$("service-dialog").open) return;
  const revision = ++directoryRevision,
    peer = state.peer;
  clearTimeout(directoryTimer);
  directoryItems = [];
  directoryActive = -1;
  directoryPage = null;
  $("directory-parent").hidden = true;
  $("directory-base").textContent = "正在查找…";
  $("directory-options").replaceChildren();
  $("workspace-input").removeAttribute("aria-activedescendant");
  $("directory-picker").hidden = false;
  $("workspace-input").setAttribute("aria-expanded", "true");
  $("directory-options").setAttribute("aria-busy", "true");
  $("directory-status").textContent = "正在目标电脑查找目录…";
  try {
    const result = await rpc("directories.search", { path, search: query });
    if (
      revision !== directoryRevision ||
      state.peer !== peer ||
      !$("service-dialog").open
    )
      return;
    directoryPage = result;
    directoryItems = result.directories;
    $("directory-parent").hidden = !result.parent;
    $("directory-base").textContent = result.base || "根目录与已注册项目";
    $("directory-base").title = result.base || "";
    if (result.recursive) $("directory-base").textContent = "按目录名搜索";
    const fragment = document.createDocumentFragment();
    for (const [index, item] of directoryItems.entries()) {
      const row = node("div", undefined, "directory-option");
      row.id = `directory-option-${index}`;
      row.setAttribute("role", "option");
      row.setAttribute("aria-selected", "false");
      row.title = item.path;
      const select = node("button", undefined, "directory-select");
      select.type = "button";
      select.tabIndex = -1;
      select.append(node("strong", "▱ " + item.name), node("span", item.path));
      select.onclick = () => chooseDirectory(item);
      const browse = node("button", "›", "directory-browse");
      browse.type = "button";
      browse.tabIndex = -1;
      browse.setAttribute("aria-label", "浏览子目录 " + item.name);
      browse.title = "浏览子目录";
      browse.onclick = () => browseDirectory(item.path);
      row.onpointerenter = () => activateDirectory(index);
      row.append(select, browse);
      fragment.append(row);
    }
    $("directory-options").replaceChildren(fragment);
    const hint = result.recursive
      ? "已搜索最多 5 层目录；输入完整路径可继续缩小范围。"
      : "点击名称选定，点击 › 浏览子目录。";
    $("directory-status").textContent =
      (directoryItems.length
        ? `${directoryItems.length} 个候选。`
        : "没有匹配目录。尝试更完整的路径，或从授权根目录浏览。") +
      (result.truncated ? "结果不完整，请细化目录名或路径。" : hint);
  } catch (error) {
    if (revision === directoryRevision && state.peer === peer)
      $("directory-status").textContent = error.message;
  } finally {
    if (revision === directoryRevision)
      $("directory-options").removeAttribute("aria-busy");
  }
}
function browseDirectory(path) {
  $("workspace-input").value = path;
  findDirectories(path, "");
  $("workspace-input").focus({ preventScroll: true });
}
$("workspace-input").addEventListener("focus", () => {
  if (!directorySkipFocus && $("directory-picker").hidden) findDirectories();
});
$("workspace-input").addEventListener("input", () => {
  closeDirectoryPicker();
  directoryTimer = setTimeout(() => findDirectories(), 250);
});
$("workspace-input").addEventListener("keydown", (event) => {
  if (event.isComposing) return;
  if (event.key === "Escape" && !$("directory-picker").hidden) {
    event.preventDefault();
    event.stopPropagation();
    closeDirectoryPicker();
    return;
  }
  if (["ArrowDown", "ArrowUp"].includes(event.key)) {
    event.preventDefault();
    if ($("directory-picker").hidden) {
      findDirectories();
      return;
    }
    if (!directoryItems.length) return;
    activateDirectory(
      directoryActive < 0
        ? event.key === "ArrowDown"
          ? 0
          : directoryItems.length - 1
        : (directoryActive +
            (event.key === "ArrowDown" ? 1 : -1) +
            directoryItems.length) %
            directoryItems.length,
    );
  } else if (event.key === "Enter" && directoryActive >= 0) {
    event.preventDefault();
    chooseDirectory(directoryItems[directoryActive]);
  } else if (event.key === "ArrowRight" && directoryActive >= 0) {
    event.preventDefault();
    browseDirectory(directoryItems[directoryActive].path);
  }
});
$("service-form").addEventListener("focusout", (event) => {
  if (
    event.relatedTarget &&
    event.relatedTarget !== $("workspace-input") &&
    !$("directory-picker").contains(event.relatedTarget)
  )
    closeDirectoryPicker();
});
$("directory-roots").onclick = () => {
  $("workspace-input").value = "";
  findDirectories("", "");
  $("workspace-input").focus({ preventScroll: true });
};
$("directory-parent").onclick = () => {
  if (directoryPage?.parent) browseDirectory(directoryPage.parent);
};
$("service-dialog").addEventListener("close", closeDirectoryPicker);
document.addEventListener("pointerdown", (event) => {
  if (
    !$("directory-picker").contains(event.target) &&
    event.target !== $("workspace-input")
  )
    closeDirectoryPicker();
});

async function openLinkedSession() {
  if (!location.hash.startsWith("#client=") || $("app").hidden) return;
  const route = new URLSearchParams(location.hash.slice(1));
  const client = state.agents.find(a => a.id === route.get("client"));
  if (!client || !client.online) { notice("链接中的客户端不在线或不属于当前账号"); return; }
  try { await connectDevice(client); await selectProject(route.get("project")); if (route.get("session")) await loadSession({id:route.get("session")}); }
  catch (e) { notice(e.message); }
}

async function pollSession() {
  const session = state.session, peer = state.peer, revision = state.readRevision;
  let cursor = state.entries.at(-1)?.seq || 0;
  do {
    const page = await rpc("sessions.get", {sessionId:session.id,cursor});
    if (state.peer !== peer || state.session?.id !== session.id || state.readRevision !== revision || state.mode === "sayso") return;
    state.session = page.session;
    setRunning(state.running);
    if (page.entries.length) mergeEntries(page.entries, true);
    cursor = page.nextCursor || 0;
  } while (cursor);
}
