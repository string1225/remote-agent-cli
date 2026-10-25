package server

import (
	"net/http"
	"strings"
	"time"
)

func (s *Server) rebindAgent(w http.ResponseWriter, r *http.Request) {
	owner, ok := s.require(w, r)
	if !ok {
		return
	}
	a, token, err := s.Store.RebindEnrollment(owner.UserID, r.PathValue("id"))
	if err != nil {
		fail(w, 404, "agent not found")
		return
	}
	base := s.Config.PublicURL
	ps := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	sh := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
	suffix := " connect --server " + sh(base) + " --rebind " + sh(a.ID) + " --token-stdin"
	unix := func(platform string) string {
		return "curl -fsSL " + sh(base+"/install/cli-"+platform+".sh") + " | bash && printf '%s' " + sh(token) + " | \"$HOME/.remote-agent/bin/remote-agent\"" + suffix
	}
	reply(w, 200, map[string]any{"agent": a, "repair": true, "rebind": true, "existingBinding": true, "expiresAt": time.Now().Add(15 * time.Minute).Unix(), "windows": ps(token) + " | remote-agent connect --server " + ps(base) + " --rebind " + ps(a.ID) + " --token-stdin", "macos": unix("macos"), "linux": unix("linux"), "windowsMsi": base + "/downloads/remote-agent-windows-amd64.msi", "windowsArm64Msi": base + "/downloads/remote-agent-windows-arm64.msi"})
}
