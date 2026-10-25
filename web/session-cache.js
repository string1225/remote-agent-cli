/* Account-scoped convenience cache; the client remains authoritative. */
(() => {
  const prefix = "remote-agent:workspace:v2:" + new URL(".", location.href).pathname + ":";
  const memory = new Map();let warned=false;
  const key = (kind, context) => prefix + JSON.stringify([kind,...context]);
  function read(kind, context) {
    if (!context[0]) return null;
    const name = key(kind,context);
    try { const value = JSON.parse(localStorage.getItem(name)); if (value?.version === 2 && (!memory.has(name) || value.at>memory.get(name).at)) return value.data; } catch {}
    return memory.get(name)?.data || null;
  }
  function trim() {
    try {
      const rows = Object.keys(localStorage).filter(k=>k.startsWith(prefix) && !k.includes('"draft"') && !k.includes('"pending"')).map(k=>({key:k,at:JSON.parse(localStorage.getItem(k))?.at || 0})).sort((a,b)=>a.at-b.at);
      while(rows.length>16) localStorage.removeItem(rows.shift().key);
    } catch {}
  }
  function write(kind, context, data) {
    if (!context[0]) return;
    const name=key(kind,context), value={version:2,at:Date.now(),data};memory.set(name,value);
    try { trim(); localStorage.setItem(name,JSON.stringify(value)); }
    catch { if(!warned){warned=true;window.onWorkspaceCacheError?.()} }
  }
  function remove(kind, context) {const name=key(kind,context);memory.delete(name);try{localStorage.removeItem(name)}catch{}}
  window.WorkspaceCache = {
    read,write,remove,
    history(context, session, entries) {
      let rows=entries.slice(-250);while(rows.length && JSON.stringify(rows).length>180000) rows=rows.slice(Math.max(1,Math.floor(rows.length/5)));
      write("history",context,{session,entries:rows});
    },
  };
})();
