/* No prompt, output, CLI credential, or peer secret is persisted in browser storage. */
const $ = (id) => document.getElementById(id);
const state = {
  register: false,
  agents: [],
  selected: null,
  peer: null,
  services: [],
  providers: {},
  roots: [],
  running: null,
  pending: new Map(),
};
let noticeTimer;
function notice(message) {
  $("notice").textContent = message;
  $("notice").hidden = false;
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => ($("notice").hidden = true), 6500);
}
async function api(path, method = "GET", body) {
  const res = await fetch(path, {
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
function signedOut() {
  disconnect();
  $("app").hidden = true;
  $("auth").hidden = false;
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
function node(tag, text, className) {
  const el = document.createElement(tag);
  if (text !== undefined) el.textContent = text;
  if (className) el.className = className;
  return el;
}
function renderDevices() {
  $("stat-total").textContent = String(
    state.agents.filter((a) => a.enrolled).length,
  ).padStart(2, "0");
  $("stat-online").textContent = String(
    state.agents.filter((a) => a.online).length,
  ).padStart(2, "0");
  $("device-count").textContent = state.agents.length;
  $("device-empty").hidden = state.agents.length > 0;
  const list = $("device-list");
  list.replaceChildren();
  for (const a of state.agents) {
    const card = node(
      "article",
      undefined,
      "device-card" + (state.selected?.id === a.id ? " selected" : ""),
    );
    const top = node("div", undefined, "device-top");
    top.append(node("div", a.os === "darwin" ? "⌘" : "▣", "device-icon"));
    const title = node("div");
    title.append(
      node("h3", a.name),
      node("p", a.enrolled ? `${a.os} / ${a.arch}` : "等待安装"),
    );
    top.append(
      title,
      node(
        "span",
        a.online ? "在线" : a.enrolled ? "离线" : "待绑定",
        "status" + (a.online ? "" : " neutral"),
      ),
    );
    card.append(
      top,
      node(
        "p",
        `${a.services?.length || 0} 个服务 · ${a.lastSeen ? new Date(a.lastSeen * 1000).toLocaleString("zh-CN", { month: "numeric", day: "numeric", hour: "2-digit", minute: "2-digit" }) : "尚未连接"}`,
        "device-meta",
      ),
    );
    const actions = node("div", undefined, "device-actions");
    const revoke = node("button", "撤销设备", "text-button");
    revoke.onclick = async () => {
      if (!confirm(`撤销「${a.name}」？该设备的连接和安装凭证将立即失效。`))
        return;
      try {
        await api("/api/agents/" + a.id, "DELETE");
        if (state.selected?.id === a.id) disconnect();
        await refresh();
      } catch (e) {
        notice(e.message);
      }
    };
    const connect = node(
      "button",
      state.selected?.id === a.id ? "已选择" : "连接设备 ↗",
      "secondary",
    );
    connect.disabled = !a.online;
    connect.onclick = () => connectDevice(a);
    actions.append(revoke, connect);
    card.append(actions);
    list.append(card);
  }
}
function connectionStatus(text, good = false) {
  $("connection-status").textContent = text;
  $("connection-status").className = "status" + (good ? "" : " neutral");
}
function setRunning(id) {
  state.running = id;
  $("send-run").disabled = !!id || !state.services.length;
  $("cancel-run").hidden = !id;
  $("service-select").disabled = !!id;
  $("remove-service").disabled = !!id;
}
function appendOutput(text) {
  const el = $("output");
  el.textContent = (el.textContent + text).slice(-400000);
  el.scrollTop = el.scrollHeight;
}
function disconnect(reason) {
  const peer = state.peer;
  state.peer = null;
  if (peer) {
    clearInterval(peer.ping);
    clearTimeout(peer.timeout);
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
  setRunning(null);
  $("session-body").hidden = true;
  $("session-empty").hidden = false;
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
  const select = $("service-select"),
    value = select.value;
  select.replaceChildren();
  if (!state.services.length)
    select.append(new Option("先注册一个 Codex / Qoder 服务", ""));
  for (const s of state.services)
    select.append(new Option(`${s.name} · ${s.provider}`, s.id));
  if (state.services.some((s) => s.id === value)) select.value = value;
  $("policy").textContent = data.allowWrite
    ? "已由本机授权：Codex 工作区写入 / Qoder 文件编辑"
    : "本机权限：Codex 只读 / Qoder 规划模式";
  setRunning(state.running);
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
async function connectDevice(agent) {
  if (state.selected?.id === agent.id && state.peer) return;
  if (state.running && !confirm("切换设备会停止当前任务，继续吗？")) return;
  disconnect();
  state.selected = agent;
  renderDevices();
  connectionStatus("正在协商连接…");
  $("selected-name").textContent = agent.name;
  const ws = new WebSocket(
    `${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}/api/agents/${agent.id}/signal`,
  );
  const peer = {
    ws,
    remoteQueue: [],
    localQueue: [],
    offerSent: false,
    pc: null,
    dc: null,
    lines: { stdout: "", stderr: "" },
  };
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
          const message = JSON.stringify({
            type: "candidate",
            candidate: event.candidate.toJSON(),
          });
          if (peer.offerSent && ws.readyState === WebSocket.OPEN)
            ws.send(message);
          else peer.localQueue.push(message);
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
            return;
          }
          const pending = state.pending.get(msg.id);
          if (pending && (msg.type === "result" || msg.type === "error")) {
            clearTimeout(pending.timer);
            state.pending.delete(msg.id);
            msg.error
              ? pending.reject(Error(msg.error))
              : pending.resolve(msg.data);
            return;
          }
          if (msg.id !== state.running) return;
          if (msg.type === "output") {
            displayOutput(peer, msg.data);
          } else if (msg.type === "done" || msg.type === "error") {
            for (const stream of ["stdout", "stderr"]) {
              if (peer.lines[stream]) appendOutput(peer.lines[stream] + "\n");
              peer.lines[stream] = "";
            }
            appendOutput(
              msg.error ? `\n任务结束：${msg.error}\n` : "\n✓ 任务已完成\n",
            );
            setRunning(null);
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
            $("session-empty").hidden = true;
            $("session-body").hidden = false;
            $("output").textContent =
              "已建立加密数据通道。注册或选择服务，开始一个新任务。\n";
            await updateTransport(peer);
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
        for (const candidate of peer.localQueue) ws.send(candidate);
        peer.localQueue = [];
      } else if (msg.type === "answer") {
        await peer.pc.setRemoteDescription(msg.sdp);
        for (const c of peer.remoteQueue) await peer.pc.addIceCandidate(c);
        peer.remoteQueue = [];
      } else if (msg.type === "candidate") {
        if (peer.pc?.remoteDescription)
          await peer.pc.addIceCandidate(msg.candidate);
        else peer.remoteQueue.push(msg.candidate);
      } else if (msg.type === "error") {
        fail(msg.error || "连接失败");
      }
    } catch (e) {
      fail(e.message);
    }
  };
}
function displayOutput(peer, data) {
  const stream = data.stream === "stderr" ? "stderr" : "stdout";
  peer.lines[stream] += data.text;
  let at;
  while ((at = peer.lines[stream].indexOf("\n")) >= 0) {
    const line = peer.lines[stream].slice(0, at);
    peer.lines[stream] = peer.lines[stream].slice(at + 1);
    if (!line.trim()) continue;
    try {
      const event = JSON.parse(line);
      if (event.item?.text) appendOutput(event.item.text + "\n");
      else if (event.item?.command)
        appendOutput(
          "$ " +
            event.item.command +
            "\n" +
            (event.item.aggregated_output || ""),
        );
      else if (event.message?.content)
        appendOutput(
          event.message.content
            .filter((x) => x.text)
            .map((x) => x.text)
            .join("\n") + "\n",
        );
      else if (event.result) appendOutput(String(event.result) + "\n");
      else appendOutput(line + "\n");
    } catch {
      appendOutput(line + "\n");
    }
  }
  if (peer.lines[stream].length > 64000) {
    appendOutput(peer.lines[stream]);
    peer.lines[stream] = "";
  }
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
  $("auth-submit").textContent = state.register
    ? "创建账号 ↗"
    : "登录工作台 ↗";
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
$("add-service").onclick = () => {
  const select = $("provider-select");
  select.replaceChildren();
  for (const name of ["codex", "qoder"]) {
    const p = state.providers[name];
    const option = new Option(
      name + (p?.available ? " · 已检测到" : " · 未安装"),
      name,
    );
    option.disabled = !p?.available;
    select.append(option);
  }
  select.selectedIndex = [...select.options].findIndex((o) => !o.disabled);
  $("provider-info").textContent = Object.values(state.providers)
    .filter((p) => !p.available)
    .map((p) => `${p.name}: ${p.error}`)
    .join(" / ");
  $("roots-info").textContent = "允许的根目录：" + state.roots.join("、");
  $("service-form").elements.workspace.placeholder =
    state.roots[0] || "绝对路径";
  $("service-form").querySelector("button").disabled = select.selectedIndex < 0;
  $("service-dialog").showModal();
};
$("service-form").onsubmit = async (e) => {
  e.preventDefault();
  const button = e.target.querySelector("button");
  button.disabled = true;
  try {
    await rpc("services.add", {
      service: Object.fromEntries(new FormData(e.target)),
    });
    refreshServices(await rpc("services.list"));
    $("service-dialog").close();
    e.target.reset();
    notice("服务已注册");
  } catch (err) {
    notice(err.message);
  } finally {
    button.disabled = false;
  }
};
$("remove-service").onclick = async () => {
  if (
    !$("service-select").value ||
    !confirm("移除这个服务？本机 CLI 和项目文件会保留。")
  )
    return;
  try {
    await rpc("services.remove", { serviceId: $("service-select").value });
    refreshServices(await rpc("services.list"));
  } catch (e) {
    notice(e.message);
  }
};
$("run-form").onsubmit = (e) => {
  e.preventDefault();
  if (state.running) return;
  const prompt = $("prompt").value.trim();
  if (!prompt || !$("service-select").value) return;
  if (!state.peer || state.peer.dc.readyState !== "open") {
    notice("请先连接设备");
    return;
  }
  const id = crypto.randomUUID();
  setRunning(id);
  state.peer.lines = { stdout: "", stderr: "" };
  appendOutput("\n› " + prompt + "\n\n");
  try {
    state.peer.dc.send(
      JSON.stringify({
        id,
        type: "run",
        serviceId: $("service-select").value,
        prompt,
      }),
    );
    $("prompt").value = "";
  } catch (err) {
    setRunning(null);
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
$("clear-output").onclick = () => ($("output").textContent = "");
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
