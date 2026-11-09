package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestProjectArchiveOfflineOwnershipAndHeartbeat(t *testing.T) {
	f := setupAt(t, "/agents")
	owner, other := f.user("archive-owner"), f.user("archive-other")
	a, _ := f.enroll(owner)
	projects := []protocol.Service{{ID: "project", Name: "Keep", Provider: "codex", Workspace: "private-path"}, {ID: "manager", Scope: "client"}}
	if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: projects}); err != nil {
		t.Fatal(err)
	}
	path := "/api/agents/" + a.ID + "/projects/project/archive"
	f.request(http.DefaultClient, "PATCH", path, map[string]bool{"archived": true}, 401)
	f.request(other, "PATCH", path, map[string]bool{"archived": true}, 404)
	f.request(owner, "PATCH", path, map[string]any{}, 400)
	f.request(owner, "PATCH", path, map[string]any{"archived": "yes"}, 400)
	f.request(owner, "PATCH", "/api/agents/"+a.ID+"/projects/manager/archive", map[string]bool{"archived": true}, 404)
	var archived protocol.Service
	if err := json.Unmarshal(f.request(owner, "PATCH", path, map[string]bool{"archived": true}, 200), &archived); err != nil {
		t.Fatal(err)
	}
	if archived.ID != "project" || archived.ArchivedAt == 0 || archived.Workspace != "" {
		t.Fatal("invalid public project", archived)
	}
	// A delayed heartbeat or old client must not undo the account's choice.
	projects[0].Name = "Renamed"
	projects[1].ArchivedAt = 123 // Clients cannot set server-owned flags.
	if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: projects}); err != nil {
		t.Fatal(err)
	}
	var agents []store.Agent
	json.Unmarshal(f.request(owner, "GET", "/api/agents", nil, 200), &agents)
	if len(agents) != 1 || agents[0].Services[0].ArchivedAt != archived.ArchivedAt || agents[0].Services[0].Name != "Renamed" || agents[0].Services[1].ArchivedAt != 0 {
		t.Fatal("archive lost or metadata overwritten", agents)
	}
	f.request(owner, "PATCH", path, map[string]bool{"archived": true}, 200) // Idempotent retry.
	f.request(owner, "PATCH", path, map[string]bool{"archived": false}, 200)
	row, err := f.db.Agent(a.ID)
	if err != nil || row.Services[0].ArchivedAt != 0 || row.Services[0].ID != "project" || row.Services[0].Workspace != "private-path" {
		t.Fatal("restore changed identity or registration", row, err)
	}
	if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: projects}); err != nil {
		t.Fatal(err)
	}
	row, _ = f.db.Agent(a.ID)
	if row.Services[0].ArchivedAt != 0 {
		t.Fatal("restore lost on heartbeat")
	}
	f.request(owner, "DELETE", "/api/agents/"+a.ID, nil, 200)
	f.request(owner, "PATCH", path, map[string]bool{"archived": true}, 404)
}
