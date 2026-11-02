/* The Android host exposes only update/settings actions on the trusted site. */
(() => {
  if (!navigator.userAgent.includes("RemoteAgentAndroid/")) return;
  document.documentElement.classList.add("android-app");
  const update = node("button", "↓ 安卓更新与设置");
  update.type = "button"; update.id = "android-update-entry";
  update.onclick = () => { location.href = "remoteagent://settings"; };
  document.querySelector(".platform-nav").append(update);
  $("new-agent").disabled = true; $("empty-add").disabled = true;
})();
