/* History is stored on the Agent, never in browser storage or the control server. */
const $ = (id) => document.getElementById(id);
const basePath = new URL(".", location.href).pathname.replace(/\/$/, "");
const state = {
  register: false,
  agents: [],
  selected: null,
  peer: null,
  services: [],
  providers: {},
  roots: [],
  allowWrite: false,
  serviceId: null,
  sessions: [],
  session: null,
  entries: [],
  drafts: new Map(),
  running: null,
  runSession: null,
  pending: new Map(),
  expanded: new Set(),
  listRevision: 0,
  readRevision: 0,
  listCursor: 0,
};
let noticeTimer, searchTimer, drawerOpener;
function notice(message) {
  $("notice").textContent = message;
  $("notice").hidden = false;
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => ($("notice").hidden = true), 8000);
}
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
  disconnect();
  $("app").hidden = true;
  $("auth").hidden = false;
  closeDrawers();
  for (const d of document.querySelectorAll("dialog[open]")) d.close();
}
async function signedIn() {
  $("auth").hidden = true;
  $("app").hidden = false;
  await refresh();
}
async function refresh() {
  if ($("app").hidden) return;
  try {
    state.agents = await api("/api/agents");
    renderDevices();
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
    row.append(
      node("span", a.os === "darwin" ? "⌘" : "▣", "device-icon"),
      node("span", a.name, "device-name"),
    );
    const dot = node(
      "span",
      undefined,
      "live-dot" + (a.online ? "" : " offline"),
    );
    dot.title = a.online ? "在线" : "离线";
    row.append(dot);
    card.append(row);
    const children = node("div", undefined, "device-children");
    children.append(
      node(
        "p",
        a.enrolled ? `${a.os} · ${a.online ? "在线" : "离线"}` : "等待安装绑定",
        "device-detail",
      ),
    );
    const actions = node("div", undefined, "device-actions"),
      connect = node(
        "button",
        selected ? "已连接" : "连接设备 ↗",
        "text-button",
      );
    connect.disabled = !a.online || selected;
    connect.onclick = () => connectDevice(a).catch((e) => notice(e.message));
    actions.append(connect);
    const revoke = node("button", "撤销", "text-button");
    revoke.onclick = async () => {
      if (!confirm(`撤销「${a.name}」？设备凭证将失效。`)) return;
      try {
        await api(`/api/agents/${a.id}`, "DELETE");
        if (selected) disconnect();
        await refresh();
      } catch (e) {
        notice(e.message);
      }
    };
    actions.append(revoke);
    children.append(actions);
    const services = selected ? state.services : a.services || [];
    for (const name of ["codex", "qoder"]) {
      const projects = services.filter((s) => s.provider === name);
      if (!projects.length && !selected) continue;
      if (!projects.length && !state.providers[name]?.available) continue;
      const group = node("div", undefined, "provider-group");
      group.append(
        node("p", name === "codex" ? "◈ CODEX" : "◇ QODER", "provider-label"),
      );
      if (!projects.length)
        group.append(node("p", "尚未添加项目", "project-empty"));
      for (const s of projects) {
        const button = node(
          "button",
          undefined,
          "project-button" +
            (selected && state.serviceId === s.id ? " active" : ""),
        );
        button.title = s.workspace;
        button.append(node("span", "▱"), node("span", folderName(s.workspace)));
        button.onclick = async () => {
          try {
            await connectDevice(a);
            await selectProject(s.id);
            closeDrawers();
          } catch (e) {
            notice(e.message);
          }
        };
        const projectRow = node("div", undefined, "project-row");
        const remove = node("button", "×", "project-remove");
        remove.title = "移除项目";
        remove.setAttribute(
          "aria-label",
          `移除项目 ${folderName(s.workspace)}`,
        );
        remove.onclick = async () => {
          if (!confirm("移除这个项目？本机文件和会话历史会保留。")) return;
          try {
            await connectDevice(a);
            await rpc("services.remove", { serviceId: s.id });
            refreshServices(await rpc("services.list"));
            if (state.serviceId === s.id) resetWorkspace();
            setRunning(state.running);
          } catch (e) {
            notice(e.message);
          }
        };
        projectRow.append(button, remove);
        group.append(projectRow);
      }
      children.append(group);
    }
    if (selected) {
      const add = node("button", "＋ 添加项目 / 服务", "text-button");
      add.onclick = showService;
      children.append(add);
      if (!services.length)
        children.append(
          node(
            "p",
            "Codex、Qoder 均为可选。\n设备可以独立保持在线。",
            "project-empty",
          ),
        );
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
  $("cancel-run").hidden = !id;
  $("send-run").disabled = !!id || !state.session || !service();
  $("new-session").disabled = !!id || !service();
}
function resetWorkspace() {
  state.serviceId = null;
  state.sessions = [];
  state.session = null;
  state.entries = [];
  state.drafts.clear();
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
  const peer = state.peer;
  state.peer = null;
  if (peer) {
    clearInterval(peer.ping);
    clearTimeout(peer.timeout);
    peer.reject?.(Error(reason || "连接已断开"));
    peer.dc?.close();
    peer.pc?.close();
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
  state.roots = [];
  state.runSession = null;
  resetWorkspace();
  setRunning(null);
  $("disconnect").hidden = true;
  connectionStatus(reason || "未连接");
  renderDevices();
}
function rpc(type, fields = {}) {
  return new Promise((resolve, reject) => {
    const peer = state.peer;
    if (!peer || peer.dc?.readyState !== "open") {
      reject(Error("设备尚未连接"));
      return;
    }
    const id = crypto.randomUUID();
    const timer = setTimeout(() => {
      state.pending.delete(id);
      reject(Error("设备响应超时"));
    }, 20000);
    state.pending.set(id, { resolve, reject, timer });
    try {
      peer.dc.send(JSON.stringify({ id, type, ...fields }));
    } catch (e) {
      clearTimeout(timer);
      state.pending.delete(id);
      reject(e);
    }
  });
}
function refreshServices(data) {
  state.services = data.services || [];
  state.providers = data.providers || {};
  state.roots = data.allowedRoots || [];
  state.allowWrite = !!data.allowWrite;
  renderDevices();
}
async function updateTransport(peer) {
  if (state.peer !== peer) return;
  try {
    const stats = await peer.pc.getStats();
    for (const report of stats.values()) {
      if (report.type === "transport" && report.selectedCandidatePairId) {
        const pair = stats.get(report.selectedCandidatePairId);
        const local = stats.get(pair?.localCandidateId),
          remote = stats.get(pair?.remoteCandidateId);
        connectionStatus(
          local?.candidateType === "relay" || remote?.candidateType === "relay"
            ? "TURN 加密中继"
            : "P2P 加密直连",
          true,
        );
        return;
      }
    }
    connectionStatus("加密通道已连接", true);
  } catch {
    connectionStatus("加密通道已连接", true);
  }
}
function connectDevice(agent) {
  if (state.selected?.id === agent.id && state.peer) return state.peer.ready;
  if (state.running && !confirm("切换设备会停止当前任务，继续吗？"))
    return Promise.reject(Error("已取消切换"));
  disconnect();
  state.selected = agent;
  state.expanded.add(agent.id);
  renderDevices();
  connectionStatus("正在连接…");
  const ws = new WebSocket(
    `${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}${basePath}/api/agents/${agent.id}/signal`,
  );
  const peer = {
    ws,
    remoteQueue: [],
    localQueue: [],
    offerSent: false,
    pc: null,
    dc: null,
  };
  peer.ready = new Promise((resolve, reject) => {
    peer.resolve = resolve;
    peer.reject = reject;
  });
  state.peer = peer;
  const fail = (msg) => {
    if (state.peer === peer) {
      notice(msg);
      disconnect(msg);
    }
  };
  peer.timeout = setTimeout(
    () => fail("连接超时，请检查网络或配置 TURN"),
    35000,
  );
  ws.onclose = () => fail("设备控制连接已断开");
  ws.onerror = () => fail("无法连接设备，请刷新设备状态");
  ws.onmessage = async (event) => {
    if (state.peer !== peer) return;
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === "ready") {
        peer.secret = msg.secret;
        peer.pc = new RTCPeerConnection({ iceServers: msg.iceServers });
        peer.dc = peer.pc.createDataChannel("remote-agent", { ordered: true });
        peer.pc.onicecandidate = (event) => {
          if (!event.candidate) return;
          const m = JSON.stringify({
            type: "candidate",
            candidate: event.candidate.toJSON(),
          });
          if (peer.offerSent && ws.readyState === WebSocket.OPEN) ws.send(m);
          else peer.localQueue.push(m);
        };
        peer.pc.onconnectionstatechange = () => {
          if (peer.pc.connectionState === "failed")
            fail("直连失败，请检查 STUN / TURN 配置");
        };
        peer.dc.onclose = () => fail("数据通道已断开");
        peer.dc.onmessage = (event) => {
          let msg;
          try {
            msg = JSON.parse(event.data);
          } catch {
            fail("设备返回了无效数据");
            return;
          }
          const pending = state.pending.get(msg.id);
          if (pending && (msg.type === "result" || msg.type === "error")) {
            clearTimeout(pending.timer);
            state.pending.delete(msg.id);
            msg.error
              ? pending.reject(
                  Error(
                    msg.error === "unknown request type"
                      ? "请更新目标电脑上的 Agent，以支持会话历史"
                      : msg.error,
                  ),
                )
              : pending.resolve(msg.data);
            return;
          }
          if (msg.id !== state.running) return;
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
        };
        peer.dc.onopen = async () => {
          if (state.peer !== peer) return;
          try {
            const data = await rpc("authenticate", { secret: peer.secret });
            if (state.peer !== peer) return;
            peer.secret = null;
            clearTimeout(peer.timeout);
            refreshServices(data);
            $("disconnect").hidden = false;
            $("selected-name").textContent = agent.name;
            await updateTransport(peer);
            peer.resolve(data);
            peer.reject = null;
          } catch (e) {
            fail(e.message);
          }
        };
        peer.ping = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN)
            ws.send(JSON.stringify({ type: "ping" }));
          updateTransport(peer);
        }, 20000);
        const offer = await peer.pc.createOffer();
        await peer.pc.setLocalDescription(offer);
        ws.send(JSON.stringify({ type: "offer", sdp: offer }));
        peer.offerSent = true;
        for (const c of peer.localQueue) ws.send(c);
        peer.localQueue = [];
      } else if (msg.type === "answer") {
        await peer.pc.setRemoteDescription(msg.sdp);
        for (const c of peer.remoteQueue) await peer.pc.addIceCandidate(c);
        peer.remoteQueue = [];
      } else if (msg.type === "candidate") {
        if (peer.pc?.remoteDescription)
          await peer.pc.addIceCandidate(msg.candidate);
        else peer.remoteQueue.push(msg.candidate);
      } else if (msg.type === "error") fail(msg.error || "连接失败");
    } catch (e) {
      fail(e.message);
    }
  };
  return peer.ready;
}
async function selectProject(id) {
  if (state.running && id !== state.serviceId)
    throw Error("请先停止当前任务，再切换项目");
  if (!state.services.some((s) => s.id === id))
    throw Error("项目已被移除，请刷新设备");
  if (id === state.serviceId) return;
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
async function refreshSessions(more = false) {
  if (!state.serviceId) return;
  const revision = ++state.listRevision;
  const sid = state.serviceId;
  const page = await rpc("sessions.list", {
    serviceId: sid,
    search: $("session-search").value.trim(),
    cursor: more ? state.listCursor : 0,
  });
  if (revision !== state.listRevision || sid !== state.serviceId) return;
  state.sessions = more ? [...state.sessions, ...page.sessions] : page.sessions;
  state.listCursor = page.nextCursor || 0;
  renderSessions();
}
function renderSessions() {
  $("conversation-count").textContent = state.sessions.length;
  $("more-sessions").hidden = !state.listCursor;
  $("conversation-empty").hidden = state.sessions.length > 0;
  $("conversation-empty").textContent = state.serviceId
    ? $("session-search").value
      ? "没有找到匹配的会话。"
      : "还没有会话。\n点击右上角 ＋ 开始。"
    : "连接设备，选择项目，\n从一个新会话开始。";
  const fragment = document.createDocumentFragment();
  for (const s of state.sessions) {
    const b = node(
      "button",
      undefined,
      "conversation-card" + (state.session?.id === s.id ? " active" : ""),
    );
    b.title = s.title;
    b.append(node("strong", s.title));
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
      node("span", s.id === state.runSession ? "生成中…" : "↗"),
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
  const revision = ++state.readRevision;
  if (state.session) state.drafts.set(state.session.id, $("prompt").value);
  state.session = session;
  $("prompt").value = state.drafts.get(session.id) || "";
  state.entries = [];
  renderSessions();
  renderChat();
  setRunning(state.running);
  let cursor = 0,
    entries = [];
  do {
    const page = await rpc("sessions.get", { sessionId: session.id, cursor });
    if (revision !== state.readRevision) return;
    state.session = page.session;
    entries.push(...page.entries);
    cursor = page.nextCursor || 0;
  } while (cursor);
  if (revision !== state.readRevision) return;
  const seq = new Set(entries.map((e) => e.seq));
  entries.push(...state.entries.filter((e) => e.seq > 0 && !seq.has(e.seq)));
  state.entries = entries.sort((a, b) => a.seq - b.seq);
  renderChat();
}
function mergeEntries(entries, live = false) {
  const seq = new Set(state.entries.map((e) => e.seq));
  state.entries.push(...entries.filter((e) => !seq.has(e.seq)));
  state.entries.sort((a, b) => a.seq - b.seq);
  renderChat(live);
}
function renderChat(live = false) {
  $("session-empty").hidden = !!state.session;
  $("session-body").hidden = !state.session;
  $("run-form").hidden = !state.session;
  $("chat-title").textContent =
    state.session?.title || (service() ? service().name : "开始你的下一件事");
  $("policy").textContent = state.allowWrite
    ? "已授权工作区写入"
    : "只读 / 规划模式";
  $("chat-hint").hidden = state.entries.length > 0;
  const container = $("messages"),
    atBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight <
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
    text.targetText = g.text;
    if (
      !live ||
      g.role !== "assistant" ||
      matchMedia("(prefers-reduced-motion: reduce)").matches
    )
      text.textContent = g.text;
    el.classList.toggle(
      "typing",
      g.role === "assistant" && g.turnId === state.running,
    );
  }
  for (const [key, el] of existing) if (!keep.has(key)) el.remove();
  if (atBottom) container.scrollTop = container.scrollHeight;
}
setInterval(() => {
  const container = $("messages"),
    atBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight <
      90;
  for (const el of container.querySelectorAll(
    ".message.assistant .message-text",
  )) {
    const target = el.targetText || "",
      current = el.textContent;
    if (current === target) continue;
    if (!target.startsWith(current)) {
      el.textContent = target;
      continue;
    }
    const chars = Array.from(target.slice(current.length));
    el.textContent += chars
      .slice(0, Math.max(2, Math.ceil(chars.length / 15)))
      .join("");
  }
  if (atBottom) container.scrollTop = container.scrollHeight;
}, 25);
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
  $("service-form").elements.workspace.placeholder =
    state.roots[0] || "绝对路径";
  $("service-form").querySelector("button").disabled = select.selectedIndex < 0;
  $("service-dialog").showModal();
}

$("auth-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = $("auth-submit");
  button.disabled = true;
  try {
    const data = Object.fromEntries(new FormData(e.target));
    await api(state.register ? "/api/register" : "/api/login", "POST", data);
    $("account-name").textContent = data.username;
    e.target.elements.password.value = "";
    await signedIn();
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
$("auth-toggle").onclick = () => {
  state.register = !state.register;
  $("auth-title").textContent = state.register ? "创建你的工作台" : "欢迎回来";
  $("auth-subtitle").textContent = state.register
    ? "一个账号，连接你所有的工作设备。"
    : "登录账号，连接你的工作设备。";
  $("auth-submit").textContent = state.register ? "创建账号 ↗" : "登录工作台 ↗";
  $("auth-toggle").textContent = state.register
    ? "已有账号？去登录"
    : "还没有账号？创建账号";
  $("auth-form").elements.password.autocomplete = state.register
    ? "new-password"
    : "current-password";
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
  $("enroll-form").reset();
  $("enroll-form").hidden = false;
  $("install-result").hidden = true;
  $("install-windows").value = "";
  $("install-macos").value = "";
  $("enroll-dialog").showModal();
}
$("new-agent").onclick = showEnroll;
$("empty-add").onclick = showEnroll;
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
    $("install-windows").value = result.windows;
    $("install-macos").value = result.macos;
    $("enroll-form").hidden = true;
    $("install-result").hidden = false;
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
  $("install-windows").value = "";
  $("install-macos").value = "";
});
for (const button of document.querySelectorAll(".copy"))
  button.onclick = async () => {
    try {
      await navigator.clipboard.writeText($(button.dataset.copy).value);
      notice("已复制安装命令");
    } catch {
      $(button.dataset.copy).select();
      notice("请按 Ctrl / ⌘ + C 复制");
    }
  };

$("disconnect").onclick = () => {
  if (!state.running || confirm("断开连接会停止当前任务，继续吗？"))
    disconnect();
};
$("service-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = e.target.querySelector("button");
  button.disabled = true;
  try {
    const s = await rpc("services.add", {
      service: Object.fromEntries(new FormData(e.target)),
    });
    refreshServices(await rpc("services.list"));
    $("service-dialog").close();
    e.target.reset();
    await selectProject(s.id);
    notice("项目已添加");
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
$("run-form").onsubmit = (e) => {
  e.preventDefault();
  if (state.running) return;
  const prompt = $("prompt").value.trim();
  if (!prompt || !state.session || !service()) return;
  if (!state.peer || state.peer.dc.readyState !== "open") {
    notice("请先连接设备");
    return;
  }
  const id = crypto.randomUUID();
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
      },
    ],
    false,
  );
  try {
    state.peer.dc.send(
      JSON.stringify({
        id,
        type: "run",
        serviceId: state.serviceId,
        sessionId: state.session.id,
        prompt,
      }),
    );
    $("prompt").value = "";
    state.drafts.delete(state.session.id);
    renderSessions();
  } catch (err) {
    setRunning(null);
    state.runSession = null;
    notice(err.message);
  }
};
$("prompt").onkeydown = (e) => {
  if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
    e.preventDefault();
    $("run-form").requestSubmit();
  }
};
$("cancel-run").onclick = async () => {
  try {
    await rpc("cancel", { runId: state.running });
  } catch (e) {
    notice(e.message);
  }
};
window.addEventListener("beforeunload", () => {
  state.peer?.dc?.close();
  state.peer?.ws.close();
});
setInterval(refresh, 10000);
(async () => {
  try {
    const config = await api("/api/config");
    $("auth-toggle").hidden = !config.allowSignup;
    await api("/api/me");
    await signedIn();
  } catch (e) {
    if (e.message !== "please sign in") notice(e.message);
  }
})();
