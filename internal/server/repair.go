package server

import (
	"net/http"
	"strings"
)

func (s *Server) repairAgent(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	a, token, err := s.Store.RepairEnrollment(v.UserID, r.PathValue("id"))
	if err != nil {
		fail(w, 404, "agent not found")
		return
	}
	base := s.Config.PublicURL
	psServer := "'" + strings.ReplaceAll(base, "'", "''") + "'"
	shQuote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
	windows := "remote-agent connect --server " + psServer
	macos := "curl -fsSL " + shQuote(base+"/install/cli-macos.sh") + " | bash && \"$HOME/.remote-agent/bin/remote-agent\" connect --server " + shQuote(base)
	linux := "curl -fsSL " + shQuote(base+"/install/cli-linux.sh") + " | bash && \"$HOME/.remote-agent/bin/remote-agent\" connect --server " + shQuote(base)
	if token != "" {
		windows = "'" + token + "' | " + windows + " --token-stdin"
		macos = "curl -fsSL " + shQuote(base+"/install/macos.sh?token="+token) + " | bash"
		linux = "curl -fsSL " + shQuote(base+"/install/linux.sh?token="+token) + " | bash"
	}
	reply(w, 200, map[string]any{"agent": a, "repair": true, "existingBinding": a.Enrolled, "windows": windows, "macos": macos, "linux": linux, "windowsMsi": base + "/downloads/remote-agent-windows-amd64.msi", "windowsArm64Msi": base + "/downloads/remote-agent-windows-arm64.msi"})
}
