package web

import "embed"

//go:embed index.html app.js platform.js channels.js markdown.js device-settings.js device-maintenance.js session-cache.js attachments.js workspace-actions.js todos.js android-app.js style.css workspace.css maintenance.css todos.css vendor icons sayso
var Files embed.FS
