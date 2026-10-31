package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type relayCall struct {
	agentID, sourceID, userID string
	messages                  chan protocol.Signal
	cancel                    context.CancelFunc
}

func (s *Server) agentCatalog(w http.ResponseWriter, r *http.Request) {
	a, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "device credentials denied")
		return
	}
	list, err := s.Store.Agents(a.UserID)
	if err != nil {
		fail(w, 500, "cannot list clients")
		return
	}
	clients := []any{}
	for _, v := range list {
		if !v.Revoked {
			clients = append(clients, map[string]any{"id": v.ID, "name": v.Name, "os": v.OS, "online": s.hub.online(v.ID), "projects": catalogProjects(v.Services)})
		}
	}
	reply(w, 200, map[string]any{"clients": clients})
}

// An enrolled client can ask the server to create work on another client owned
// by the same account. No client-to-client sockets or shared credentials exist.
func (s *Server) agentToolCall(w http.ResponseWriter, r *http.Request) {
	source, err := s.authenticateAgent(r)
	if err != nil {
		fail(w, 401, "device credentials denied")
		return
	}
	var in struct {
		AgentID string          `json:"agentId"`
		Type    string          `json:"type"`
		Body    json.RawMessage `json:"body"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Type != "tasks.start" && in.Type != "sessions.spawn" && in.Type != "channels.notify" && in.Type != "sessions.control" {
		fail(w, 400, "unsupported client capability")
		return
	}
	if len(in.Body) > protocol.MaxRequestBytes-1024 {
		fail(w, 413, "request too large")
		return
	}
	if in.AgentID == "" {
		in.AgentID = source.ID
	}
	target, err := s.Store.Agent(in.AgentID)
	if err != nil || target.Revoked || target.UserID != source.UserID {
		fail(w, 404, "client not found")
		return
	}
	if in.Type == "tasks.start" || in.Type == "sessions.spawn" {
		var task protocol.TaskStart
		if json.Unmarshal(in.Body, &task) != nil {
			fail(w, 400, "invalid task")
			return
		}
		task.Notify = false
		task.NotifyAgentID, task.NotifyServiceID, task.NotifySessionID = "", "", ""
		task.ParentAgentID = source.ID
		if task.Source != "im" && task.Source != "sayso" {
			task.Source = "web"
		}
		in.Body, _ = json.Marshal(task)
	}
	if in.Type == "sessions.control" {
		var control protocol.SessionControl
		if json.Unmarshal(in.Body, &control) != nil {
			fail(w, 400, "invalid session operation")
			return
		}
		control.CallerAgentID = source.ID
		in.Body, _ = json.Marshal(control)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	call := &relayCall{agentID: target.ID, sourceID: source.ID, userID: source.UserID, messages: make(chan protocol.Signal, 4), cancel: cancel}
	peerID := store.Token()
	s.hub.mu.Lock()
	agent := s.hub.agents[target.ID]
	count := 0
	for _, b := range s.hub.browsers {
		if b.agentID == target.ID {
			count++
		}
	}
	for _, c := range s.hub.calls {
		if c.agentID == target.ID || c.sourceID == source.ID {
			count++
		}
	}
	if agent == nil || count >= 8 {
		s.hub.mu.Unlock()
		fail(w, 409, "client offline or connection limit reached")
		return
	}
	s.hub.calls[peerID] = call
	s.hub.mu.Unlock()
	defer func() {
		s.hub.mu.Lock()
		delete(s.hub.calls, peerID)
		s.hub.mu.Unlock()
		_ = agent.send(protocol.Signal{Type: "close", PeerID: peerID})
	}()
	if agent.send(protocol.Signal{Type: "open", PeerID: peerID, LeaseSeconds: 30}) != nil {
		fail(w, 503, "client disconnected")
		return
	}
	requestID := store.Token()
	sent := false
	for {
		select {
		case <-ctx.Done():
			fail(w, 503, "client operation timed out or disconnected")
			return
		case msg := <-call.messages:
			switch msg.Type {
			case "close":
				fail(w, 503, "client closed operation")
				return
			case "ready":
				if sent || msg.Transport != protocol.RelayTransport {
					continue
				}
				sent = true
				if _, err := s.authenticateAgent(r); err != nil {
					fail(w, 401, "device credentials denied")
					return
				}
				req := protocol.Request{ID: requestID, Type: in.Type, Body: in.Body}
				if agent.send(protocol.Signal{Type: "request", PeerID: peerID, Request: &req}) != nil {
					fail(w, 503, "client disconnected")
					return
				}
			case "event":
				if msg.Event == nil || msg.Event.ID != requestID {
					continue
				}
				if msg.Event.Type == "error" {
					fail(w, 409, msg.Event.Error)
					return
				}
				if msg.Event.Type == "result" {
					reply(w, 200, msg.Event.Data)
					return
				}
			}
		}
	}
}

func catalogProjects(services []protocol.Service) []protocol.Service {
	out := append([]protocol.Service{}, services...)
	for i := range out {
		out[i].Workspace = ""
	}
	return out
}
