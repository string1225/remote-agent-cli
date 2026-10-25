package server_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestRebindAPIIsOwnerOnlyAndChecksTargetBeforeConsumption(t *testing.T) {
	f := setup(t)
	alice, bob := f.user("alice"), f.user("bob")
	a, invite := f.enroll(alice, "linux")
	enrollment := f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": invite, "os": "linux", "arch": "amd64"}, 200)
	var old struct {
		Credential string `json:"credential"`
	}
	if err := json.Unmarshal(enrollment, &old); err != nil {
		t.Fatal(err)
	}
	f.request(bob, "POST", "/api/agents/"+a.ID+"/rebind", map[string]string{}, 404)
	f.request(f.admin("administrator"), "POST", "/api/agents/"+a.ID+"/rebind", map[string]string{}, 404)
	data := f.request(alice, "POST", "/api/agents/"+a.ID+"/rebind", map[string]string{}, 200)
	var result struct {
		Windows string `json:"windows"`
		Rebind  bool   `json:"rebind"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if !result.Rebind || !strings.Contains(result.Windows, "--rebind '"+a.ID+"'") {
		t.Fatal("wrong recovery command", result.Windows)
	}
	token := strings.Split(result.Windows, "'")[1]
	if _, err := f.db.AuthenticateAgent(a.ID, old.Credential); err != nil {
		t.Fatal("issuance invalidated active binding")
	}
	next := store.Token()
	body := map[string]any{"token": token, "os": "linux", "arch": "amd64", "rebind": true, "expectedId": "wrong", "credential": next}
	f.request(&http.Client{}, "POST", "/api/enroll", body, 401)
	body["expectedId"] = a.ID
	f.request(&http.Client{}, "POST", "/api/enroll", body, 200)
	f.request(&http.Client{}, "POST", "/api/enroll", body, 401)
	if _, err := f.db.AuthenticateAgent(a.ID, next); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.AuthenticateAgent(a.ID, old.Credential); err == nil {
		t.Fatal("old binding remained valid")
	}
	// A mistaken --rebind on a normal invite must not consume that invite.
	_, normal := f.enroll(alice, "linux")
	body["token"] = normal
	f.request(&http.Client{}, "POST", "/api/enroll", body, 401)
	f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": normal, "os": "linux", "arch": "amd64"}, 200)
}
