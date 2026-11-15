/* Keep every dialog's heading outside its scrollable content, including dialogs
 * created later by workbench, todos and coordination. Existing nodes/listeners
 * are moved intact; opening a dialog never recreates its form or secret fields. */
(() => {
  function prepare(dialog) {
    if (dialog.classList.contains("fixed-dialog-header")) return;
    let head = [...dialog.children].find(child => child.matches(".dialog-head,.todo-dialog-header,.workbench-dialog-heading"));
    if (!head) {
      const title = dialog.querySelector("h2,h3");
      if (!title) return;
      head = document.createElement("div"); head.className = "dialog-head";
      const close = document.createElement("button"); close.type = "button";
      close.className = "icon-button"; close.textContent = "×"; close.setAttribute("aria-label", "关闭弹窗");
      close.onclick = () => dialog.close();
      head.append(title, close); dialog.prepend(head);
    }
    const body = document.createElement("div"); body.className = "dialog-scroll-body";
    for (const child of [...dialog.childNodes]) if (child !== head) body.append(child);
    head.classList.add("fixed-dialog-heading"); dialog.append(body);
    dialog.classList.add("fixed-dialog-header");
  }
  function scan(root) {
    if (root instanceof HTMLDialogElement) prepare(root);
    root.querySelectorAll?.("dialog").forEach(prepare);
  }
  scan(document);
  new MutationObserver(changes => {
    for (const change of changes) for (const added of change.addedNodes) scan(added);
  }).observe(document.body, {childList:true,subtree:true});
})();
