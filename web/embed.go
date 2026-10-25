package web

import "embed"

//go:embed index.html app.js platform.js channels.js markdown.js device-settings.js style.css workspace.css vendor icons sayso
var Files embed.FS
