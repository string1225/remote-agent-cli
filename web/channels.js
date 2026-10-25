/* Channel secrets are never cached in browser storage or returned by the client. */
(() => {
  const dialog = $("channel-dialog"), form = $("channel-form"), progress = $("channel-progress-dialog");
  let project, peer, revision = 0, progressRevision = 0, pollTimer;
  const update = () => {
    const feishu = form.elements.provider.value === "feishu";
    $("channel-robot-code").hidden = feishu;
    $("channel-reaction-help").textContent = feishu ? "收到消息后显示思考中，回复完成后撤下。" : "钉钉以 Markdown 回复；该平台的机器人暂不支持消息思考表情。";
  };
  form.elements.provider.onchange = update;
  window.openChannelSetup = async (client, nextProject) => {
    const opening = ++revision;
    await connectDevice(client);
    await selectProject(nextProject.id);
    if (opening !== revision) return;
    project = nextProject; peer = state.peer;
    const openingPeer = peer;
    const config = await rpc("channels.get", {serviceId:project.id});
    if (opening !== revision) return;
    if (state.peer !== openingPeer) throw Error("客户端连接已改变");
    form.reset();
    for (const key of ["provider","name","appId","robotCode"]) if (config[key]) form.elements[key].value = config[key];
    form.elements.appSecret.value = "";
    form.elements.appSecret.placeholder = config.secretConfigured ? "已保存；留空保留原 Secret" : "填写机器人应用 Secret";
    form.elements.appSecret.required = !config.secretConfigured;
    $("channel-title").textContent = project.name + " · IM 渠道";
    $("channel-state").textContent = config.configured
      ? (({connected:"已连接，可以发消息了",connecting:"正在连接",retrying:"正在重试连接",stopped:"已停止",error:"连接或消息处理遇到问题"})[config.state] || "等待连接") + (config.error ? "。" + config.error : config.waitingForMessage ? "。收到消息后，会自动显示对应的私聊或群聊会话。" : "") + (config.reactionError ? "。" + config.reactionError : "")
      : "保存后会连接机器人，私聊与群聊自动识别。";
    $("channel-unbind").hidden = !config.configured;
    update(); closeDrawers(); dialog.showModal();
  };
  function setProgress(step, message, failed = false) {
    const steps = ["config","start","ready"];
    steps.forEach((id, i) => $("channel-progress-" + id).className = i < step ? "done" : i === step ? "current" : "");
    $("channel-progress-message").textContent = message;
    $("channel-progress-message").className = "progress-message" + (failed ? " error" : "");
    $("channel-progress-title").textContent = step === 3 ? "渠道已连接" : failed ? "渠道需要检查" : "正在连接渠道";
  }
  async function watchConnection(serviceId, connection, stamp) {
    if (!progress.open || progressRevision !== stamp) return;
    if (state.peer !== connection) { setProgress(1, "设备连接已改变。重新连接设备后，可在渠道设置中查看状态。", true); return; }
    try {
      const config = await rpc("channels.get", {serviceId});
      if (!progress.open || progressRevision !== stamp) return;
      if (config.state === "connected") {
        setProgress(3, "长连接启动成功。现在可以私聊机器人，或在群聊中 @ 它。此窗口由你关闭。" + (config.reactionError ? " " + config.reactionError : ""));
        await refreshSessions(); return;
      }
      setProgress(1, config.error || (config.state === "retrying" ? "连接暂时中断，客户端正在重试。可检查应用是否发布、事件订阅权限和网络。" : "正在配置长连接，等待机器人连接成功。"), !!config.error);
      pollTimer = setTimeout(() => watchConnection(serviceId, connection, stamp), 1500);
    } catch (error) {
      if (progress.open && progressRevision === stamp) setProgress(1, error.message, true);
    }
  }
  async function save(body) {
    if (state.peer !== peer) throw Error("客户端连接已改变，请重新打开配置");
    const savingPeer = peer, savingRevision = revision, serviceId = project.id, stamp = ++progressRevision;
    clearTimeout(pollTimer);
    if (body.provider) { setProgress(0, "正在验证应用凭据并保存配置，请稍候。"); progress.showModal(); }
    try {
      await rpc("channels.configure", {body:{...body,serviceId}});
      if (state.peer !== savingPeer) throw Error("设备连接已改变，请重新打开渠道设置查看结果");
      if (revision === savingRevision) form.elements.appSecret.value = "";
      const services = await rpc("services.list");
      if (state.peer !== savingPeer) return;
      refreshServices(services); await refreshSessions();
      if (revision === savingRevision) dialog.close();
      if (body.provider) {
        if (progress.open && stamp === progressRevision) { setProgress(1, "配置已保存，正在配置长连接。"); await watchConnection(serviceId, savingPeer, stamp); }
      } else notice("渠道已解绑，会话历史已保留");
    } catch (error) {
      if (body.provider && progress.open && stamp === progressRevision) { setProgress(0, error.message, true); $("channel-progress-config").className = ""; }
      else throw error;
    }
  }
  form.onsubmit = async e => {
    e.preventDefault(); const button = form.querySelector("button[type=submit]"); button.disabled = true;
    try { await save(Object.fromEntries(new FormData(form))); }
    catch (e) { notice(e.message); } finally { button.disabled = false; }
  };
  $("channel-unbind").onclick = async e => {
    if (!confirm("解绑这个项目的 IM 渠道？本机渠道会话与历史会保留。")) return;
    e.target.disabled = true;
    try { await save({provider:""}); } catch (e) { notice(e.message); } finally { e.target.disabled = false; }
  };
  $("channel-progress-close").onclick = () => progress.close();
  progress.addEventListener("close", () => { progressRevision++; clearTimeout(pollTimer); });
  dialog.addEventListener("close", () => { revision++; form.reset(); form.elements.appSecret.value = ""; project = null; peer = null; });
})();
