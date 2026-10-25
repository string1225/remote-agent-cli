package server_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestDeviceRenamePreservesBindingAndRequiresOwner(t *testing.T) {
	f := setup(t)
	owner, other := f.user("rename-owner"), f.user("rename-other")
	a, _ := f.enroll(owner)
	path := "/api/agents/" + a.ID
	f.request(other, "PATCH", path, map[string]string{"name": "other"}, 404)
	for _, name := range []string{"", " \t ", "bad\nname", strings.Repeat("x", 81)} {
		f.request(owner, "PATCH", path, map[string]string{"name": name}, 400)
	}
	f.request(owner, "PATCH", path, map[string]string{"name": "  工作电脑  "}, 200)
	var devices []store.Agent
	if err := json.Unmarshal(f.request(owner, "GET", "/api/agents", nil, 200), &devices); err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0].ID != a.ID || devices[0].Name != "工作电脑" {
		t.Fatalf("rename did not preserve identity: %+v", devices)
	}
	f.request(owner, "DELETE", path, nil, 200)
	f.request(owner, "PATCH", path, map[string]string{"name": "revoked"}, 404)
}
