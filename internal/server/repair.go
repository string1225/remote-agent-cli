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
	if a.Enrolled {
		a, token, err = s.Store.RebindEnrollment(v.UserID, a.ID)
		if err != nil {
			fail(w, 404, "agent not found")
			return
		}
	}
	base := s.Config.PublicURL
	psServer := "'" + strings.ReplaceAll(base, "'", "''") + "'"
	shQuote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
	windows := "'" + token + "' | remote-agent connect --server " + psServer + " --client-id '" + a.ID + "' --token-stdin"
	unix := func(platform string) string {
		return "curl -fsSL " + shQuote(base+"/install/cli-"+platform+".sh") + " | bash && printf '%s' " + shQuote(token) + " | \"$HOME/.remote-agent/bin/remote-agent\" connect --server " + shQuote(base) + " --client-id " + shQuote(a.ID) + " --token-stdin"
	}
	macos, linux := unix("macos"), unix("linux")
	reply(w, 200, map[string]any{"agent": a, "repair": true, "existingBinding": a.Enrolled, "windows": windows, "macos": macos, "linux": linux, "windowsMsi": base + "/downloads/remote-agent-windows-amd64.msi", "windowsArm64Msi": base + "/downloads/remote-agent-windows-arm64.msi"})
}
