package server

import "net/http"

func (s *Server) endpointGuide(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	token, secret, err := s.Store.EnsureEndpointGuideToken(actor.ID, r.PathValue("id"))
	if err != nil {
		fail(w, 403, "无法为该账号生成指引 Token，请检查账号状态")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, map[string]string{"token": secret, "tokenId": token.ID})
}
