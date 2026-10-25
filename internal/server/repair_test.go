package server_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestRepairPreservesDeviceAndPendingEnrollment(t *testing.T) {
	f := setup(t)
	alice, bob := f.user("alice"), f.user("bob")
	a, _ := f.enroll(alice)
	f.request(bob, "POST", "/api/agents/"+a.ID+"/repair", map[string]string{}, 404)
	data := f.request(alice, "POST", "/api/agents/"+a.ID+"/repair", map[string]string{}, 200)
	var commands struct {
		Windows string `json:"windows"`
	}
	if err := json.Unmarshal(data, &commands); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(commands.Windows, "'")
	if len(parts) < 3 {
		t.Fatal("missing renewed pending token")
	}
	f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": parts[1], "os": "windows", "arch": "amd64"}, 200)
	data = f.request(alice, "POST", "/api/agents/"+a.ID+"/repair", map[string]string{}, 200)
	if err := json.Unmarshal(data, &commands); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(commands.Windows, "--token-stdin") || !strings.Contains(commands.Windows, "--client-id '"+a.ID+"'") {
		t.Fatal("repair must carry the selected client ID and scoped recovery capability")
	}
	list := f.request(alice, "GET", "/api/agents", nil, 200)
	if strings.Count(string(list), `"id":`) != 1 || !strings.Contains(string(list), a.ID) {
		t.Fatal("repair created a different device")
	}
}
