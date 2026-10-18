package web

import "embed"

//go:embed index.html app.js platform.js style.css sayso
var Files embed.FS
