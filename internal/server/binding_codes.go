package server

import (
	"net/http"
	"strings"
)

func (s *Server) activationCapabilities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, map[string]bool{"proposedCredential": true})
}

func windowsActivationCommand(base, code string) string {
	base = strings.ReplaceAll(base, "'", "''")
	return "$raArch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') { 'arm64' } else { 'amd64' }; $raMsi = Join-Path $env:TEMP 'remote-agent-setup.msi'; Invoke-WebRequest -UseBasicParsing -ErrorAction Stop ('" + base + "/downloads/remote-agent-windows-' + $raArch + '.msi') -OutFile $raMsi; Start-Process msiexec.exe -Wait -ArgumentList @('/i', ('\"' + $raMsi + '\"'), 'BINDING_CODE=" + code + "', 'AGENT_SERVER=" + base + "')"
}

func (s *Server) activationStatus(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "invalid device credential")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, map[string]any{"online": s.hub.online(a.ID), "name": a.Name})
}

func (s *Server) bindingCode(w http.ResponseWriter, r *http.Request) {
	owner, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Renew bool `json:"renew"`
	}
	if !decode(w, r, &in) {
		return
	}
	a, code, expires, err := s.Store.BindingCode(owner.UserID, r.PathValue("id"), in.Renew)
	if err != nil {
		fail(w, 409, "绑定码不可用：请确认设备尚未绑定且属于当前账号")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, map[string]any{"agent": a, "bindingCode": code, "expiresAt": expires, "windowsInstall": windowsActivationCommand(s.Config.PublicURL, code), "windowsMsi": s.Config.PublicURL + "/downloads/remote-agent-windows-amd64.msi", "windowsArm64Msi": s.Config.PublicURL + "/downloads/remote-agent-windows-arm64.msi"})
}
