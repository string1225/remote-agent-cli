/* Channel secrets are never cached in browser storage or returned by the client. */
(() => {
  const dialog = $("channel-dialog"), form = $("channel-form");
  let project, peer, revision = 0;
  const update = () => { $("channel-robot-code").hidden = form.elements.provider.value !== "dingtalk"; };
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
    for (const key of ["provider","appId","chatId","chatType","robotCode"]) if (config[key]) form.elements[key].value = config[key];
    form.elements.appSecret.value = "";
    form.elements.appSecret.placeholder = config.secretConfigured ? "已保存；留空保留原 Secret" : "填写机器人应用 Secret";
    form.elements.appSecret.required = !config.secretConfigured;
    $("channel-title").textContent = project.name + " · IM 渠道";
    $("channel-state").textContent = config.configured ? "当前状态：" + ({connected:"长连接已建立",connecting:"正在连接",retrying:"正在重连，请检查机器人权限与网络",stopped:"已停止"})[config.state] : "尚未绑定渠道。保存成功后，项目下会置顶一个固定渠道会话。";
    $("channel-unbind").hidden = !config.configured;
    update(); closeDrawers(); dialog.showModal();
  };
  async function save(body) {
    if (state.peer !== peer) throw Error("客户端连接已改变，请重新打开配置");
    const saving = revision, savingPeer = peer, serviceId = project.id;
    await rpc("channels.configure", {body:{...body,serviceId}});
    if (saving !== revision || state.peer !== savingPeer) return;
    form.elements.appSecret.value = "";
    const services = await rpc("services.list");
    if (saving !== revision || state.peer !== savingPeer) return;
    refreshServices(services);
    await refreshSessions();
    if (saving !== revision || state.peer !== savingPeer) return;
    dialog.close(); notice(body.provider ? "渠道已保存，客户端正在建立长连接" : "渠道已解绑，会话历史保留在客户端");
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
  dialog.addEventListener("close", () => { revision++; form.reset(); form.elements.appSecret.value = ""; project = null; peer = null; });
})();
