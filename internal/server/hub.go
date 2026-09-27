package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type socket struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *socket) send(v protocol.Signal) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(v)
}
func (c *socket) close() { _ = c.conn.Close() }

type browser struct {
	*socket
	agentID, sessionHash string
}
type hub struct {
	mu       sync.Mutex
	agents   map[string]*socket
	browsers map[string]*browser
}

func newHub() *hub                   { return &hub{agents: map[string]*socket{}, browsers: map[string]*browser{}} }
func (h *hub) online(id string) bool { h.mu.Lock(); defer h.mu.Unlock(); return h.agents[id] != nil }
func (h *hub) closeAgent(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if a := h.agents[id]; a != nil {
		a.close()
		delete(h.agents, id)
	}
	for _, b := range h.browsers {
		if b.agentID == id {
			b.close()
		}
	}
}
func (h *hub) closeSession(hash string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, b := range h.browsers {
		if b.sessionHash == hash {
			b.close()
		}
	}
}
func (h *hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, a := range h.agents {
		a.close()
	}
	for _, b := range h.browsers {
		b.close()
	}
}
func upgrade(w http.ResponseWriter, r *http.Request, origin string) *socket {
	u := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == origin || r.Header.Get("Origin") == "" }, HandshakeTimeout: 10 * time.Second}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
	conn.SetReadLimit(128 * 1024)
	return &socket{conn: conn}
}
func (s *Server) agentSocket(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "agent credentials denied")
		return
	}
	c := upgrade(w, r, s.origin())
	if c == nil {
		return
	}
	s.hub.mu.Lock()
	old := s.hub.agents[a.ID]
	s.hub.agents[a.ID] = c
	// Replacement invalidates all sessions tied to the old agent transport.
	if old != nil {
		old.close()
		for _, b := range s.hub.browsers {
			if b.agentID == a.ID {
				b.close()
			}
		}
	}
	s.hub.mu.Unlock()
	defer func() {
		c.close()
		s.hub.mu.Lock()
		if s.hub.agents[a.ID] == c {
			delete(s.hub.agents, a.ID)
			for _, b := range s.hub.browsers {
				if b.agentID == a.ID {
					b.close()
				}
			}
		}
		s.hub.mu.Unlock()
	}()
	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(65 * time.Second))
		var msg protocol.Signal
		if err := c.conn.ReadJSON(&msg); err != nil {
			return
		}
		if _, err := s.authenticateAgent(r); err != nil {
			return
		}
		if msg.Type == "heartbeat" {
			if len(msg.Services) > 100 {
				return
			}
			if err := s.Store.Heartbeat(a.ID, msg); err != nil {
				return
			}
			if c.send(protocol.Signal{Type: "heartbeat"}) != nil {
				return
			}
			continue
		}
		if msg.Type != "answer" && msg.Type != "candidate" && msg.Type != "error" {
			continue
		}
		s.hub.mu.Lock()
		b := s.hub.browsers[msg.PeerID]
		current := s.hub.agents[a.ID] == c
		s.hub.mu.Unlock()
		if current && b != nil && b.agentID == a.ID {
			_ = b.send(protocol.Signal{Type: msg.Type, SDP: msg.SDP, Candidate: msg.Candidate, Error: msg.Error})
		}
	}
}
func (s *Server) browserSocket(w http.ResponseWriter, r *http.Request) {
	v, token, err := s.session(r)
	if err != nil {
		fail(w, 401, "please sign in")
		return
	}
	a, err := s.Store.Agent(r.PathValue("id"))
	if err != nil || a.UserID != v.UserID || a.Revoked {
		fail(w, 404, "agent not found")
		return
	}
	s.hub.mu.Lock()
	agent := s.hub.agents[a.ID]
	count := 0
	for _, b := range s.hub.browsers {
		if b.agentID == a.ID {
			count++
		}
	}
	s.hub.mu.Unlock()
	if agent == nil {
		fail(w, 409, "agent is offline")
		return
	}
	if count >= 8 {
		fail(w, 429, "too many device connections")
		return
	}
	c := upgrade(w, r, s.origin())
	if c == nil {
		return
	}
	peerID, secret := store.Token(), store.Token()
	expires := min(v.ExpiresAt, time.Now().Add(30*time.Minute).Unix())
	ice := s.ice(v.UserID)
	s.hub.mu.Lock()
	count = 0
	for _, b := range s.hub.browsers {
		if b.agentID == a.ID {
			count++
		}
	}
	if count >= 8 || s.hub.agents[a.ID] != agent {
		s.hub.mu.Unlock()
		c.close()
		return
	}
	s.hub.browsers[peerID] = &browser{c, a.ID, store.Hash(token)}
	s.hub.mu.Unlock()
	defer func() {
		c.close()
		s.hub.mu.Lock()
		delete(s.hub.browsers, peerID)
		s.hub.mu.Unlock()
		_ = agent.send(protocol.Signal{Type: "close", PeerID: peerID})
	}()
	// A second check closes the upgrade/revocation race.
	if _, err := s.Store.Session(token); err != nil {
		return
	}
	fresh, err := s.Store.Agent(a.ID)
	if err != nil || fresh.Revoked {
		return
	}
	if c.send(protocol.Signal{Type: "ready", PeerID: peerID, Secret: secret, ExpiresAt: expires, ICE: ice}) != nil {
		return
	}
	offered := false
	candidateCount := 0
	for {
		deadline := time.Now().Add(65 * time.Second)
		if end := time.Unix(expires, 0); end.Before(deadline) {
			deadline = end
		}
		_ = c.conn.SetReadDeadline(deadline)
		var msg protocol.Signal
		if c.conn.ReadJSON(&msg) != nil {
			return
		}
		if _, err := s.Store.Session(token); err != nil || time.Now().Unix() >= expires {
			return
		}
		switch msg.Type {
		case "ping":
			if c.send(protocol.Signal{Type: "pong"}) != nil {
				return
			}
		case "offer":
			if offered {
				return
			}
			offered = true
			// Keep expiry enforcement on the server's clock. The Agent uses a bounded
			// relative lifetime because its wall clock may differ from this host.
			remaining := max(int64(1), expires-time.Now().Unix())
			if agent.send(protocol.Signal{Type: "offer", PeerID: peerID, Secret: secret, ExpiresAt: expires, LeaseSeconds: remaining, SDP: msg.SDP, ICE: ice}) != nil {
				return
			}
		case "candidate":
			candidateCount++
			if !offered || candidateCount > 128 {
				return
			}
			if agent.send(protocol.Signal{Type: "candidate", PeerID: peerID, Candidate: msg.Candidate}) != nil {
				return
			}
		case "close":
			return
		}
	}
}
