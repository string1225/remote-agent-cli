/* File bytes go to the selected client; localStorage only keeps attachment refs. */
(() => {
  const drafts=new Map();let revision=0;
  const key=()=>JSON.stringify(draftContext());
  const items=()=>drafts.get(key()) || [];
  const ready=rows=>rows.filter(f=>f.id && !f.uploading && !f.error).map(({id,name,mime,size})=>({id,name,mime,size}));
  function persist(context, rows) {
    if(JSON.stringify(context)===key()){saveCurrentDraft();render()}
    else {const draft=WorkspaceCache.read("draft",context)||{};WorkspaceCache.write("draft",context,{text:draft.text||"",attachments:ready(rows)})}
  }
  function render() {
    const host=$("attachment-list");host.replaceChildren(...items().map(file=>{
      const card=node("div",undefined,"attachment-chip"), button=node("button",file.name,"attachment-name");button.type="button";
      button.disabled=!!file.uploading || !!file.error;button.onclick=()=>open(file,state.serviceId,state.session.id);
      const detail=node("span",file.error || (file.uploading ? `上传 ${file.progress||0}%` : `${Math.max(1,Math.ceil(file.size/1024))} KB`),file.error?"attachment-error":"");
      const remove=node("button","×","attachment-remove");remove.type="button";remove.title="移除此附件";remove.disabled=!!file.uploading;
      remove.onclick=()=>{const rows=items();rows.splice(rows.indexOf(file),1);saveCurrentDraft();render()};card.append(button,detail,remove);return card;
    }));
  }
  async function upload(files) {
    if(items().some(f=>f.uploading)){notice("请等待当前附件上传完成");return}
    if(!state.session || state.session.kind==="im" || !state.attachments){notice("请在已更新客户端的 Web 会话中上传附件");return}
    const context=draftContext(), identity=JSON.stringify(context),peer=state.peer,serviceId=state.serviceId,sessionId=state.session.id;
    const rows=items();drafts.set(identity,rows);
    for(const file of files) {
      if(rows.length>=6){notice("每条消息最多 6 个附件");break}
      if(file.size<1 || file.size>10*1024*1024){notice(`${file.name}：单个文件应为 1 B–10 MiB`);continue}
      const item={name:file.name,size:file.size,uploading:true,progress:0};rows.push(item);render();
      try {
        const data=new Uint8Array(await file.arrayBuffer()),chunkSize=32768;
        const send=async(offset,id)=>{
          if(state.peer!==peer)throw Error("设备连接已改变，请重新上传这个附件");
          const chunk=data.subarray(offset,offset+chunkSize);
          return rpc("attachments.write",{serviceId,sessionId,body:{id,name:file.name,size:file.size,offset,data:btoa(String.fromCharCode(...chunk)),complete:offset+chunk.length===data.length}});
        };
        let result=await send(0,"");item.id=result.attachment.id;
        for(let offset=chunkSize;offset<data.length;offset+=chunkSize*4){
          const results=await Promise.all(Array.from({length:Math.min(4,Math.ceil((data.length-offset)/chunkSize))},(_,i)=>send(offset+i*chunkSize,item.id)));
          result=results.at(-1);item.progress=Math.round(result.nextOffset/data.length*100);if(identity===key())render();
        }
        if(!result.complete)throw Error("附件未完整上传，请重试");Object.assign(item,result.attachment,{uploading:false});persist(context,rows);
      } catch(error) {item.uploading=false;item.error=error.message;if(identity===key())render();notice(error.message)}
    }
  }
  async function bytes(ref,serviceId,sessionId) {
    const peer=state.peer, chunks=[];let offset=0;
    while(offset<ref.size){
      if(state.peer!==peer || !peer)throw Error("请先连接附件所在的客户端");
      const result=await rpc("attachments.read",{serviceId,sessionId,path:ref.id,cursor:offset});
      if(result.nextOffset<=offset || result.nextOffset>ref.size)throw Error("附件响应无效");
      chunks.push(Uint8Array.from(atob(result.data),c=>c.charCodeAt(0)));offset=result.nextOffset;
    }
    return new Blob(chunks,{type:ref.mime || "application/octet-stream"});
  }
  async function open(ref,serviceId,sessionId) {
    const stamp=++revision;
    const preview=$("attachment-preview"),content=$("attachment-preview-content");
    $("attachment-preview-title").textContent=ref.name;content.replaceChildren(node("p","正在从客户端读取附件…","muted"));preview.showModal();
    try {
      const blob=await bytes(ref,serviceId,sessionId);if(!preview.open || stamp!==revision)return;
      const url=URL.createObjectURL(blob),download=node("a","下载文件","secondary");download.href=url;download.download=ref.name;
      const clean=()=>{URL.revokeObjectURL(url);preview.removeEventListener("close",clean)};preview.addEventListener("close",clean);
      content.replaceChildren();if(["image/png","image/jpeg","image/webp","image/gif"].includes(ref.mime)){const img=node("img");img.src=url;img.alt=ref.name;content.append(img)}
      content.append(download);
    }catch(error){if(stamp===revision)content.replaceChildren(node("p",error.message,"attachment-error"))}
  }
  $("attachment-preview").addEventListener("close",()=>revision++);
  window.chatAttachments={
    snapshot:()=>ready(items()),busy:()=>items().some(f=>f.uploading),
    restore(refs){const rows=items(),pending=rows.filter(f=>f.uploading);rows.splice(0,rows.length,...refs.filter(f=>f?.id&&typeof f.name==="string"&&f.size>0),...pending);drafts.set(key(),rows);render()},
    renderHistory(article,refs){
      const unique=[...new Map(refs.map(f=>[f.id,f])).values()],signature=JSON.stringify(unique);if(article.attachmentSignature===signature)return;article.attachmentSignature=signature;
      article.querySelector(".message-files")?.remove();if(!unique.length)return;
      const row=node("div",undefined,"message-files"),serviceId=state.serviceId,sessionId=state.session.id;
      for(const ref of unique){const b=node("button",`${ref.mime?.startsWith("image/")?"图片":"文件"} · ${ref.name}`,"attachment-file");b.type="button";b.onclick=()=>open(ref,serviceId,sessionId);row.append(b)}article.append(row);
    },
  };
  $("add-attachment").onclick=()=>$("attachment-input").click();
  $("attachment-input").onchange=e=>{upload([...e.target.files]);e.target.value=""};
  $("prompt").addEventListener("paste",e=>{const images=[...(e.clipboardData?.files||[])].filter(f=>f.type.startsWith("image/"));if(images.length){e.preventDefault();upload(images)}});
  $("run-form").addEventListener("dragover",e=>{if(e.dataTransfer?.types.includes("Files"))e.preventDefault()});
  $("run-form").addEventListener("drop",e=>{if(e.dataTransfer?.files.length){e.preventDefault();upload([...e.dataTransfer.files])}});
})();
