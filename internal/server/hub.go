package server

import (
	"encoding/json"
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
	agentID, sessionHash, userID string
	expires                      int64
}
type hub struct {
	mu       sync.Mutex
	agents   map[string]*socket
	browsers map[string]*browser
	calls    map[string]*relayCall
}

func newHub() *hub {
	return &hub{agents: map[string]*socket{}, browsers: map[string]*browser{}, calls: map[string]*relayCall{}}
}
func (h *hub) online(id string) bool { h.mu.Lock(); defer h.mu.Unlock(); return h.agents[id] != nil }
func (h *hub) closeAgent(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.calls {
		if c.agentID == id || c.sourceID == id {
			c.cancel()
		}
	}
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
func (h *hub) closeUserBrowsers(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.calls {
		if c.userID == userID {
			c.cancel()
		}
	}
	for _, b := range h.browsers {
		if b.userID == userID {
			b.close()
		}
	}
}
func (h *hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.calls {
		c.cancel()
	}
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
	conn.SetReadLimit(protocol.MaxRelayBytes)
	return &socket{conn: conn}
}
func (s *Server) agentSocket(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "agent credentials denied")
		return
	}
	if machine := r.Header.Get("X-Remote-Agent-Machine"); machine != "" && a.MachineID != "" && a.MachineID != machine {
		fail(w, 409, "machine identity differs; use the selected client's repair command to confirm replacement")
		return
	}
	if r.URL.Query().Get("transport") != protocol.RelayTransport {
		fail(w, http.StatusUpgradeRequired, "update the client to the WebSocket relay release")
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
		for _, call := range s.hub.calls {
			if call.agentID == a.ID {
				call.cancel()
			}
		}
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
			for _, call := range s.hub.calls {
				if call.agentID == a.ID {
					call.cancel()
				}
			}
			for _, b := range s.hub.browsers {
				if b.agentID == a.ID {
					b.close()
				}
			}
		}
		s.hub.mu.Unlock()
	}()
	// Re-check after registration so disable/revoke cannot race the upgrade.
	if _, err := s.authenticateAgent(r); err != nil {
		return
	}
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
		if msg.Type != "ready" && msg.Type != "event" && msg.Type != "close" {
			continue
		}
		s.hub.mu.Lock()
		b := s.hub.browsers[msg.PeerID]
		call := s.hub.calls[msg.PeerID]
		current := s.hub.agents[a.ID] == c
		s.hub.mu.Unlock()
		if current && call != nil && call.agentID == a.ID {
			select {
			case call.messages <- msg:
			default:
				call.cancel()
			}
			continue
		}
		if !current || b == nil || b.agentID != a.ID {
			continue
		}
		if time.Now().Unix() >= b.expires || msg.Type == "close" {
			b.close()
			continue
		}
		switch msg.Type {
		case "ready":
			if msg.Transport != protocol.RelayTransport {
				b.close()
				continue
			}
			if b.send(protocol.Signal{Type: "ready", ExpiresAt: b.expires, Transport: protocol.RelayTransport}) != nil {
				b.close()
			}
		case "event":
			if msg.Event != nil && b.send(protocol.Signal{Type: "event", Event: msg.Event}) != nil {
				b.close()
			}
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
	peerID := store.Token()
	expires := min(v.ExpiresAt, time.Now().Add(30*time.Minute).Unix())
	s.hub.mu.Lock()
	count = 0
	for _, b := range s.hub.browsers {
		if b.agentID == a.ID {
			count++
		}
	}
	for _, call := range s.hub.calls {
		if call.agentID == a.ID {
			count++
		}
	}
	if count >= 8 || s.hub.agents[a.ID] != agent {
		s.hub.mu.Unlock()
		c.close()
		return
	}
	s.hub.browsers[peerID] = &browser{socket: c, agentID: a.ID, sessionHash: store.Hash(token), userID: v.UserID, expires: expires}
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
	if agent.send(protocol.Signal{Type: "open", PeerID: peerID, LeaseSeconds: max(int64(1), expires-time.Now().Unix())}) != nil {
		return
	}
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
		case "request":
			if msg.Request == nil || len(msg.Request.ID) == 0 || len(msg.Request.ID) > 100 {
				return
			}
			data, err := json.Marshal(msg.Request)
			if err != nil || len(data) > protocol.MaxRequestBytes {
				return
			}
			// Never trust routing, identity or lease fields supplied by a browser.
			if agent.send(protocol.Signal{Type: "request", PeerID: peerID, Request: msg.Request}) != nil {
				return
			}
		case "close":
			return
		default:
			return
		}
	}
}
