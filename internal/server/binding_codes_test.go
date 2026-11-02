package server_test

import (
	"encoding/json"
	"github.com/string1225/remote-agent-cli/internal/store"
	"net/http"
	"strings"
	"testing"
)

func TestOwnerCanReopenBindingCodeAndActivationIsIdempotent(t *testing.T) {
	f := setupAt(t, "/agents")
	owner, other := f.user("owner"), f.user("other")
	a, code := f.enroll(owner, "windows")
	var result struct {
		BindingCode    string
		WindowsInstall string
	}
	body := f.request(owner, "POST", "/api/agents/"+a.ID+"/binding-code", map[string]bool{}, 200)
	if json.Unmarshal(body, &result) != nil || result.BindingCode != code || !strings.Contains(result.WindowsInstall, "BINDING_CODE="+code) {
		t.Fatal("missing installation code")
	}
	f.request(other, "POST", "/api/agents/"+a.ID+"/binding-code", map[string]bool{}, 409)
	f.request(&http.Client{}, "POST", "/api/agents/"+a.ID+"/binding-code", map[string]bool{}, 401)
	credential := store.Token()
	request := map[string]string{"token": code, "credential": credential, "os": "windows", "arch": "amd64"}
	for range 2 {
		var bound struct{ ID, Credential string }
		body = f.request(&http.Client{}, "POST", "/api/enroll", request, 200)
		if json.Unmarshal(body, &bound) != nil || bound.ID != a.ID || bound.Credential != credential {
			t.Fatal("activation identity/credential changed")
		}
	}
	request["credential"] = store.Token()
	f.request(&http.Client{}, "POST", "/api/enroll", request, 401)
	f.request(owner, "POST", "/api/agents/"+a.ID+"/binding-code", map[string]bool{}, 409)
}
