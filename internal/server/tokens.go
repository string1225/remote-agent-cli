package server

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func bearer(r *http.Request) string {
	if len(r.Header.Values("Authorization")) != 1 {
		return ""
	}
	fields := strings.Fields(r.Header.Get("Authorization"))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return ""
	}
	return fields[1]
}

func (s *Server) tokens(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	tokens, err := s.Store.CodexTokens(v.UserID)
	if err != nil {
		fail(w, 500, "cannot list tokens")
		return
	}
	reply(w, 200, tokens)
}

func (s *Server) createToken(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	info, token, err := s.Store.NewCodexToken(v.UserID, in.Name, "")
	if err != nil {
		if errors.Is(err, store.ErrInvalid) {
			fail(w, 400, "token name must contain 1–80 characters and no control characters")
		} else {
			fail(w, 500, "cannot create token")
		}
		return
	}
	reply(w, 201, struct {
		store.CodexToken
		Token string `json:"token"`
	}{info, token})
}

func (s *Server) revokeToken(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	if err := s.Store.RevokeCodexToken(v.UserID, r.PathValue("id"), false); err != nil {
		fail(w, 404, "token not found")
		return
	}
	reply(w, 200, map[string]bool{"ok": true})
}

func (s *Server) adminTokens(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	tokens, err := s.Store.CodexTokens(r.URL.Query().Get("userId"))
	if err != nil {
		fail(w, 500, "cannot list tokens")
		return
	}
	reply(w, 200, tokens)
}

func (s *Server) adminRevokeToken(w http.ResponseWriter, r *http.Request) {
	u, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	if err := s.Store.RevokeCodexToken(u.ID, r.PathValue("id"), true); err != nil {
		fail(w, 404, "token not found")
		return
	}
	reply(w, 200, map[string]bool{"ok": true})
}

func (s *Server) codex(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.require(w, r); !ok {
		return
	}
	reply(w, 200, map[string]any{"enabled": s.Config.GatewayURL != "", "baseUrl": s.Config.PublicURL + "/codex/v1", "model": s.Config.GatewayModel})
}

func (s *Server) agentCodex(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "agent credentials denied")
		return
	}
	if s.Config.GatewayURL == "" {
		fail(w, 503, "Codex gateway is not configured")
		return
	}
	name := []rune("Codex on " + strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, a.Name))
	if len(name) > 80 {
		name = name[:80]
	}
	_, token, err := s.Store.NewCodexToken(a.UserID, string(name), a.ID)
	if err != nil {
		fail(w, 403, "cannot issue device token")
		return
	}
	reply(w, 200, map[string]string{"baseUrl": s.Config.PublicURL + "/codex/v1", "token": token, "model": s.Config.GatewayModel})
}
