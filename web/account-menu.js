/* Move existing actions without replacing their account, permission or update handlers. */
(() => {
  const nav = document.querySelector(".platform-nav"), account = document.querySelector(".account");
  const help = nav.querySelector('[data-platform="setup"]');
  help.id = "help-entry";
  help.className = "icon-button sidebar-help";
  help.textContent = "?";
  help.type = "button";
  help.title = "Codex 接入与帮助";
  help.setAttribute("aria-label", "Codex 接入与帮助");
  document.querySelector(".brand-heading .wordmark").after(help);

  const trigger = node("button", undefined, "account-toggle"), menu = node("div", undefined, "account-popover");
  const name = $("account-name"), avatar = node("span", account.querySelector(".avatar").textContent, "avatar");
  const chevron = node("span", "⌃", "account-chevron"); chevron.setAttribute("aria-hidden", "true");
  trigger.id = "account-toggle"; trigger.type = "button";
  trigger.setAttribute("popovertarget", "account-popover");
  trigger.setAttribute("aria-haspopup", "menu");
  trigger.setAttribute("aria-expanded", "false");
  trigger.setAttribute("aria-controls", "account-popover");
  trigger.append(avatar, name, chevron);
  menu.id = "account-popover";
  menu.setAttribute("popover", "auto");
  menu.setAttribute("role", "menu");
  menu.setAttribute("aria-label", "账号菜单");
  for (const action of [...nav.children]) if (action.id !== "todos-entry") menu.append(action);
  for (const [id, title, action] of [
    ["archived-projects-entry", "已归档项目", () => window.openArchivedProjects()],
    ["appearance-entry", "界面设置", () => window.openAppearanceSettings()],
  ]) {
    const button = node("button", title); button.id = id; button.onclick = action; menu.append(button);
  }
  const logout = $("logout");
  logout.className = "account-logout";
  logout.textContent = "退出登录";
  menu.append(logout);
  for (const action of menu.querySelectorAll("button")) {
    action.type = "button"; action.setAttribute("role", "menuitem"); action.tabIndex = -1;
  }
  account.replaceChildren(trigger, menu);
  const items = () => [...menu.querySelectorAll("button")].filter(action => !action.hidden && !action.disabled);
  const isOpen = () => menu.matches(":popover-open");
  let focusLast = false;
  function position() {
    const bounds = trigger.getBoundingClientRect();
    const width = Math.min(260, innerWidth - 24);
    menu.style.width = `${width}px`;
    menu.style.left = `${Math.max(12, Math.min(bounds.left, innerWidth - width - 12))}px`;
    menu.style.bottom = `${Math.max(12, innerHeight - bounds.top + 8)}px`;
    menu.style.maxHeight = `${Math.max(100, bounds.top - 20)}px`;
  }
  menu.addEventListener("beforetoggle", event => { if (event.newState === "open") position(); });
  menu.addEventListener("toggle", event => {
    const open = event.newState === "open";
    trigger.setAttribute("aria-expanded", String(open));
    if (open) { const rows = items(); (focusLast ? rows.at(-1) : rows[0])?.focus({preventScroll:true}); focusLast = false; }
    else if (menu.contains(document.activeElement)) trigger.focus({preventScroll:true});
  });
  // Close before an existing action opens a dialog or starts signing out.
  menu.addEventListener("click", event => { if (event.target.closest("button") && isOpen()) menu.hidePopover(); }, true);
  trigger.addEventListener("keydown", event => {
    if (!["ArrowDown", "ArrowUp"].includes(event.key)) return;
    event.preventDefault();
    focusLast = event.key === "ArrowUp";
    if (!isOpen()) menu.showPopover();
    else { const rows = items(); (focusLast ? rows.at(-1) : rows[0])?.focus(); focusLast = false; }
  });
  menu.addEventListener("keydown", event => {
    const rows = items(), index = rows.indexOf(document.activeElement);
    const next = ({ArrowDown:(index + 1) % rows.length, ArrowUp:(index + rows.length - 1) % rows.length, Home:0, End:rows.length - 1})[event.key];
    if (next !== undefined) { event.preventDefault(); rows[next]?.focus(); }
    else if (event.key === "Escape" || event.key === "Tab") {
      if (event.key === "Escape") { event.preventDefault(); event.stopPropagation(); }
      menu.hidePopover(); trigger.focus({preventScroll:true});
    }
  });
  window.addEventListener("resize", () => { if (isOpen()) position(); });
  window.addEventListener("remote-agent-disconnected", () => { if (isOpen()) menu.hidePopover(); });
})();
