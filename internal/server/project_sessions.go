package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func (s *Server) projectSessionToken(w http.ResponseWriter, r *http.Request) {
	owner, ok := s.require(w, r)
	if !ok {
		return
	}
	agentID, projectID := r.PathValue("id"), r.PathValue("project")
	info, err := s.Store.ProjectSessionToken(owner.UserID, agentID, projectID)
	if err != nil {
		fail(w, 404, "project not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		reply(w, 200, info)
	case http.MethodPut:
		var in struct {
			Token string `json:"token"`
		}
		if !decode(w, r, &in) {
			return
		}
		updated, token, err := s.Store.SetProjectSessionToken(owner.UserID, agentID, projectID, in.Token)
		if errors.Is(err, store.ErrProjectSessionToken) {
			fail(w, 400, err.Error())
			return
		}
		if err != nil {
			fail(w, 500, "cannot save project token")
			return
		}
		reply(w, 200, map[string]any{"configured": updated.Configured, "updatedAt": updated.UpdatedAt, "token": token})
	case http.MethodPost:
		token, err := s.Store.RevealProjectSessionToken(owner.UserID, agentID, projectID)
		if err != nil {
			fail(w, 404, "project token not configured")
			return
		}
		reply(w, 200, map[string]string{"token": token})
	case http.MethodDelete:
		if s.Store.DeleteProjectSessionToken(owner.UserID, agentID, projectID) != nil {
			fail(w, 500, "cannot revoke project token")
			return
		}
		reply(w, 200, map[string]bool{"configured": false})
	}
}

func (s *Server) createProjectSession(w http.ResponseWriter, r *http.Request) {
	agentID, projectID := r.PathValue("id"), r.PathValue("project")
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		fail(w, 401, "project token denied")
		return
	}
	secret := parts[1]
	info, err := s.Store.AuthenticateProjectSessionToken(agentID, projectID, secret)
	if err != nil {
		fail(w, 401, "project token denied")
		return
	}
	var in protocol.ProjectSessionCreate
	if !decodeLimit(w, r, &in, 256*1024) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if !utf8.ValidString(in.Title) || !utf8.ValidString(in.Message) || utf8.RuneCountInString(in.Title) > 80 || len(in.Message) > 32000 || in.Execute && strings.TrimSpace(in.Message) == "" {
		fail(w, 400, "title must be at most 80 characters; message at most 32000 bytes and required when execute=true")
		return
	}
	check := func() bool {
		current, err := s.Store.AuthenticateProjectSessionToken(agentID, projectID, secret)
		return err == nil && current.UserID == info.UserID
	}
	body, _ := json.Marshal(protocol.ProjectSessionRelay{Title: in.Title, Message: []byte(in.Message), Execute: in.Execute})
	raw, err := s.relayRequest(r.Context(), agentID, info.UserID, protocol.Request{Type: "sessions.create-external", ServiceID: projectID, Body: body}, check, false)
	if errors.Is(err, store.ErrDenied) {
		fail(w, 401, "project token denied")
		return
	}
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	var result protocol.TaskStarted
	if json.Unmarshal(raw, &result) != nil || result.Session.ID == "" || result.Session.ServiceID != projectID {
		fail(w, 502, "invalid session response from client")
		return
	}
	// The integration needs session identity, not local paths or native IDs.
	result.Session.Workspace = ""
	result.Session.NativeID = ""
	link := s.Config.PublicURL + "/#" + url.Values{"client": {agentID}, "project": {projectID}, "session": {result.Session.ID}}.Encode()
	reply(w, 201, map[string]any{"agentId": agentID, "projectId": projectID, "sessionId": result.Session.ID, "session": result.Session, "taskId": result.TaskID, "executed": in.Execute, "url": link})
}
