package server

import (
	"context"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/githubproxy"
)

func (s *Server) githubProxy(w http.ResponseWriter, r *http.Request) {
	s.serveGitHubProxy(w, r, githubproxy.Dial)
}

func (s *Server) serveGitHubProxy(w http.ResponseWriter, r *http.Request, dial func(context.Context, string) (net.Conn, error)) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "device credentials denied")
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != s.origin() {
		fail(w, http.StatusForbidden, "origin denied")
		return
	}
	target, err := githubproxy.Target(r.URL.Query().Get("target"))
	if err != nil {
		fail(w, http.StatusForbidden, err.Error())
		return
	}
	if !websocket.IsWebSocketUpgrade(r) {
		fail(w, http.StatusUpgradeRequired, "WebSocket required")
		return
	}
	ctx, release, ok := s.githubProxyPool.Open(r.Context(), a.ID)
	if !ok {
		fail(w, http.StatusTooManyRequests, "GitHub proxy connection limit reached")
		return
	}
	defer release()
	upstream, err := dial(ctx, target)
	if err != nil {
		fail(w, http.StatusBadGateway, "cannot connect to GitHub through the relay")
		return
	}
	defer upstream.Close()
	// Recheck after the network dial to close the credential-rotation race.
	if _, err := s.authenticateAgent(r); err != nil {
		fail(w, http.StatusUnauthorized, "device credentials denied")
		return
	}
	if ctx.Err() != nil {
		fail(w, http.StatusServiceUnavailable, "GitHub proxy is stopping")
		return
	}
	c := upgrade(w, r, s.origin())
	if c == nil {
		return
	}
	githubproxy.Tunnel(ctx, c.conn, upstream, upstream, func() bool {
		_, err := s.authenticateAgent(r)
		return err == nil
	})
}
