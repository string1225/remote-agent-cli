import {useEffect,useState} from "react";
import {api} from "./api";
import type {ActionItem,AssistantRef} from "./types";

export default function DispatchTargetModal({action,onClose,onSend}:{action:ActionItem;onClose:()=>void;onSend:(target:AssistantRef)=>Promise<void>}) {
 const [targets,setTargets]=useState<AssistantRef[]>([]),[selected,setSelected]=useState(""),[busy,setBusy]=useState(false),[error,setError]=useState("");
 useEffect(()=>{let active=true;api.assistants().then(data=>{if(active){setTargets(data.assistants.filter(t=>t.scope!=="account"));}}).catch(e=>{if(active)setError(e.message)});return()=>{active=false}},[]);
 return <div className="modal-backdrop"><section className="modal" role="dialog" aria-modal="true" aria-label="选择执行管家">
  <header className="modal-header"><h2>选择执行管家</h2><button onClick={onClose} disabled={busy} aria-label="关闭">×</button></header>
  <p>{action.title}</p><p>Action 将交给选定管家，完成后回报 CEO；讨论和录音保存在 CEO 主机。</p>
  <label>执行目标<select aria-label="执行目标" value={selected} onChange={e=>setSelected(e.target.value)} disabled={busy}><option value="">请选择客户端管家或项目管家</option>{targets.map((t,i)=><option key={i} value={String(i)}>{t.name||t.projectId||t.clientId}{t.online===false?"（离线，等待重连）":""}</option>)}</select></label>
  {error&&<p role="alert">{error}</p>}
  <footer className="modal-footer"><button onClick={onClose} disabled={busy}>取消</button><button className="primary-button" disabled={busy||selected===""} onClick={async()=>{setBusy(true);setError("");try{await onSend(targets[Number(selected)])}catch(e){setError(e instanceof Error?e.message:"派工失败")}finally{setBusy(false)}}}>{busy?"正在加入派工队列…":"确认派工"}</button></footer>
 </section></div>;
}
