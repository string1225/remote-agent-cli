package server

import (
	"errors"
	"net/http"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func (s *Server) renameAgent(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Name string `json:"name"`
		OS   string `json:"os"`
	}
	if !decode(w, r, &in) {
		return
	}
	if err := s.Store.UpdateAgentSettings(v.UserID, r.PathValue("id"), in.Name, in.OS); err != nil {
		if errors.Is(err, store.ErrInvalid) {
			fail(w, 400, "设备名称应为 1–80 个字符，不能包含控制字符")
		} else {
			fail(w, 404, "agent not found")
		}
		return
	}
	reply(w, 200, map[string]bool{"ok": true})
}
