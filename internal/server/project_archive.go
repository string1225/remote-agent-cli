package server

import (
	"errors"
	"net/http"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func (s *Server) projectArchive(w http.ResponseWriter, r *http.Request) {
	owner, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Archived *bool `json:"archived"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.Archived == nil {
		fail(w, 400, "archived must be a boolean")
		return
	}
	project, err := s.Store.SetProjectArchived(owner.UserID, r.PathValue("id"), r.PathValue("project"), *in.Archived)
	if errors.Is(err, store.ErrDenied) {
		fail(w, 404, "project not found")
		return
	}
	if err != nil {
		fail(w, 500, "cannot save project archive status")
		return
	}
	// Match the public catalog: never expose client paths or native IDs.
	project.Workspace = ""
	reply(w, 200, project)
}
