package server

import (
	"github.com/string1225/remote-agent-cli/internal/store"
	"net/http"
)

func (s *Server) inspectClientRepair(w http.ResponseWriter, r *http.Request) {
	if s.limited(r) {
		fail(w, 429, "too many attempts")
		return
	}
	var in struct {
		Token     string `json:"token"`
		ID        string `json:"id"`
		MachineID string `json:"machineId"`
	}
	if !decode(w, r, &in) {
		return
	}
	a, valid, err := s.Store.InspectRepair(in.Token, in.ID, bearer(r), in.MachineID)
	if err != nil {
		fail(w, 401, "repair authorization invalid or expired")
		return
	}
	reply(w, 200, map[string]any{"id": a.ID, "name": a.Name, "os": a.OS, "enrolled": a.Enrolled, "machineId": a.MachineID, "hostname": a.Hostname, "bindingValid": valid, "machineMatch": a.MachineID == "" || a.MachineID == in.MachineID, "repairSourceId": a.RepairSourceID, "services": a.Services})
}
func (s *Server) claimClientRepair(w http.ResponseWriter, r *http.Request) {
	if s.limited(r) {
		fail(w, 429, "too many attempts")
		return
	}
	var in struct {
		Token            string `json:"token"`
		ID               string `json:"id"`
		Credential       string `json:"credential"`
		MachineID        string `json:"machineId"`
		Hostname         string `json:"hostname"`
		OS               string `json:"os"`
		Arch             string `json:"arch"`
		SourceID         string `json:"sourceId"`
		SourceCredential string `json:"sourceCredential"`
		Force            bool   `json:"force"`
	}
	if !decode(w, r, &in) {
		return
	}
	if (in.OS != "windows" && in.OS != "darwin" && in.OS != "linux") || (in.Arch != "amd64" && in.Arch != "arm64") {
		fail(w, 400, "unsupported platform")
		return
	}
	a, moved, err := s.Store.ClaimRepair(store.RepairClaim{Token: in.Token, ID: in.ID, Credential: in.Credential, MachineID: in.MachineID, Hostname: in.Hostname, OS: in.OS, Arch: in.Arch, SourceID: in.SourceID, SourceCredential: in.SourceCredential, Force: in.Force})
	if err != nil {
		fail(w, 401, "repair authorization invalid, expired, or platform mismatch")
		return
	}
	s.hub.closeAgent(a.ID)
	if moved && in.SourceID != "" {
		s.hub.closeAgent(in.SourceID)
	}
	reply(w, 200, map[string]any{"id": a.ID, "name": a.Name, "credential": in.Credential, "services": a.Services, "channelsTransferred": moved})
}
