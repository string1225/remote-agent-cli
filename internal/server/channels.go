package server

import (
	"github.com/string1225/remote-agent-cli/internal/store"
	"net/http"
)

func (s *Server) agentChannel(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "device credentials denied")
		return
	}
	var in struct {
		ServiceID   string                 `json:"serviceId"`
		Fingerprint string                 `json:"fingerprint"`
		Finalize    bool                   `json:"finalize"`
		Bindings    []store.ChannelBinding `json:"bindings"`
	}
	if !decode(w, r, &in) {
		return
	}
	if r.PathValue("sync") == "sync" {
		err = s.Store.SyncChannels(a.UserID, a.ID, in.Bindings)
	} else {
		err = s.Store.BindChannel(a.UserID, store.ChannelBinding{AgentID: a.ID, ServiceID: in.ServiceID, Fingerprint: in.Fingerprint}, in.Finalize)
	}
	if err != nil {
		fail(w, 409, "机器人已绑定其他项目或绑定信息无效，请先解绑原项目")
		return
	}
	reply(w, 200, map[string]bool{"ok": true})
}
