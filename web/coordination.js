/* Logical roles use the existing relay and project/session cache identities. */
(() => {
  let revision = 0;
  window.openAssistant = async ref => {
    const turn = ++revision;
    const navigation = state.navigationRevision;
    const resolved = await api("/api/coordination", "POST", {operation:"resolve",target:ref});
    if (turn !== revision || navigation !== state.navigationRevision || !state.user) return;
    const client = state.agents.find(a=>a.id===resolved.clientId);
    if (!client) throw Error("目标客户端已不可用，请刷新");
    client.services = [...(client.services||[]).filter(s=>s.id!==resolved.service.id),resolved.service];
    if (state.selected?.id===client.id) state.services = [...state.services.filter(s=>s.id!==resolved.service.id),resolved.service];
    const cached = WorkspaceCache.read("client",[state.user.id,client.id]) || {};
    WorkspaceCache.write("client",[state.user.id,client.id],{...cached,services:client.services});
    await openWorkspace(client,resolved.service.id,ref.sessionId);
    renderDevices();
    return resolved;
  };
  window.openCEO = () => window.openAssistant({scope:"account"});
  $("ceo-open").onclick = () => window.openCEO().catch(e=>notice(e.message));
  $("ceo-sayso").onclick = async () => {
    try {await window.openCEO(); await openSayso();} catch(e) {notice(e.message)}
  };
  window.renderHierarchy = () => $("ceo-open").classList.toggle("active",service()?.scope==="account");
  window.appendAssistantButton = (parent,ref,online,compact=false) => {
    const b=node("button",compact?"管家":"◉ 客户端管家",compact?"project-steward":"client-steward");
    b.type="button"; b.title=ref.name; b.setAttribute("aria-label","打开"+ref.name); b.disabled=!online;
    b.onclick=()=>window.openAssistant(ref).catch(e=>notice(e.message)); parent.append(b);
  };
  window.renderWorkReport = (report,existing) => {
    if(existing) return existing;
    const card=node("article",undefined,"message assistant work-report");
    const origin=report.origin;
    const title=node("div",`${origin.name||"下游管家"} · ${report.status==="ready"?"已返回结果":report.status==="progress"?"进度更新":"需检查"}`,"message-label");
    const text=node("div",undefined,"message-text"); renderMarkdown(text,report.summary);
    const meta=node("p",`来源 ${origin.clientId} / ${origin.projectId} · 会话 ${origin.sessionId}`,"work-source");
    const buttons=node("div",undefined,"work-report-actions");
    const reply=node("button","直接回复此结果"),open=node("button","打开执行会话");
    reply.onclick=()=>replyTo(report); open.onclick=async()=>{
      try {await window.openAssistant(origin)}catch(e){notice(e.message)}
    };
    buttons.append(reply,open);
    for(const link of report.related||[]) {
      const direct=node("button","直接回复 "+(link.origin.name||"下游执行管家"));
      direct.onclick=()=>replyTo(link);buttons.append(direct);
    }
    card.append(title,text,meta,buttons); return card;
  };
  function replyTo(report) {
    const context=[state.user.id,report.replyHandle];
    const saved=WorkspaceCache.read("work-reply",context)||{};
    const dialog=node("dialog",undefined,"work-reply-dialog");
    const title=node("h3","回复 "+(report.origin.name||"下游管家"));
    const hint=node("p","指令直达这个结果的原执行会话，完成后继续回报上层。", "muted");
    const form=node("form"),input=node("textarea"),buttons=node("div",undefined,"work-report-actions");
    input.rows=6;input.required=true;input.maxLength=16000;input.value=saved.text||"";input.setAttribute("aria-label","给下游管家的指令");
    const send=node("button","发送到下游"),cancel=node("button","关闭"); send.type="submit";cancel.type="button";
    let requestId=saved.requestId||crypto.randomUUID(),submittedText=saved.submittedText;
    const save=()=>WorkspaceCache.write("work-reply",context,{text:input.value,requestId,submittedText});
    input.oninput=()=>{if(submittedText!==undefined && input.value!==submittedText){requestId=crypto.randomUUID();submittedText=undefined;}save()};
    cancel.onclick=()=>dialog.close(); dialog.onclose=()=>dialog.remove();
    form.onsubmit=async event=>{
      event.preventDefault();if(!input.value.trim())return;send.disabled=true; submittedText=input.value;save();
      try {
        const accepted=await api("/api/coordination","POST",{operation:"reply",replyHandle:report.replyHandle,requestId,prompt:input.value});
        WorkspaceCache.remove("work-reply",context);dialog.close();
        notice("已发送给 "+(accepted.target.name||"下游管家")+"，沿用原执行会话");
      }catch(e){notice(e.message)}finally{send.disabled=false}
    };
    buttons.append(send,cancel);form.append(input,buttons);dialog.append(title,hint,form);document.body.append(dialog);dialog.showModal();input.focus();
  }
  renderHierarchy();
})();
