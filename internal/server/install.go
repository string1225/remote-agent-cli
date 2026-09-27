package server

import (
	"embed"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed installers/*
var installers embed.FS
var tokenRE = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var artifactRE = regexp.MustCompile(`^remote-agent-(windows|darwin|linux)-(amd64|arm64)(\.exe)?(\.sha256)?$`)

func (s *Server) installer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("script")
	if name != "windows.ps1" && name != "macos.sh" {
		http.NotFound(w, r)
		return
	}
	token := r.URL.Query().Get("token")
	if !tokenRE.MatchString(token) {
		fail(w, 400, "invalid enrollment token")
		return
	}
	data, _ := installers.ReadFile("installers/" + name)
	server := s.Config.PublicURL
	if name == "windows.ps1" {
		server = strings.ReplaceAll(server, "'", "''")
	} else {
		server = strings.ReplaceAll(server, "'", "'\"'\"'")
	}
	body := strings.ReplaceAll(strings.ReplaceAll(string(data), "__SERVER__", server), "__TOKEN__", token)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(body))
}
func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("file")
	if !artifactRE.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if strings.HasSuffix(name, ".sha256") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	http.ServeFile(w, r, filepath.Join(s.Config.Downloads, name))
}
