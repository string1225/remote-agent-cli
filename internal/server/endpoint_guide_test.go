package server_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEndpointGuideRequiresAdminAndIssuesSeparateReusableRevocableToken(t *testing.T) {
	f := setup(t)
	owner := f.user("owner")
	admin := f.admin("admin")
	u, _ := f.db.User("owner")
	path := "/api/admin/users/" + u.ID + "/endpoint-guide"
	req, _ := http.NewRequest("POST", f.url+path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://other.example")
	res, err := admin.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 403 {
		t.Fatal("cross-origin issuance accepted")
	}
	f.request(&http.Client{}, "POST", path, map[string]any{}, 401)
	f.request(owner, "POST", path, map[string]any{}, 403)
	_, original, _ := f.db.NewCodexToken(u.ID, "private existing token", "")
	var first, second struct {
		Token string `json:"token"`
		ID    string `json:"tokenId"`
	}
	if err := json.Unmarshal(f.request(admin, "POST", path, map[string]any{}, 200), &first); err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(f.request(admin, "POST", path, map[string]any{}, 200), &second)
	if first.Token == "" || first.Token == original || first != second {
		t.Fatal("guide token must be distinct and stable")
	}
	info, err := f.db.AuthenticateCodexToken(first.Token)
	if err != nil || info.UserID != u.ID || info.AgentID != "" || info.Purpose != "endpoint-guide" {
		t.Fatal("incorrect guide ownership/purpose", err)
	}
	list := f.request(admin, "GET", "/api/admin/tokens", nil, 200)
	if strings.Contains(string(list), first.Token) || strings.Contains(string(list), original) {
		t.Fatal("list exposed a secret")
	}
	f.request(admin, "POST", "/api/tokens/"+info.ID+"/reveal", map[string]any{}, 404)
	f.request(owner, "DELETE", "/api/tokens/"+info.ID, nil, 200)
	if _, err = f.db.AuthenticateCodexToken(first.Token); err == nil {
		t.Fatal("revoked guide still accepted")
	}
	_ = json.Unmarshal(f.request(admin, "POST", path, map[string]any{}, 200), &second)
	if first.Token == second.Token {
		t.Fatal("revocation did not rotate guide")
	}
	if _, err = f.db.AuthenticateCodexToken(original); err != nil {
		t.Fatal("existing personal token changed", err)
	}
	f.request(admin, "PATCH", "/api/admin/users/"+u.ID, map[string]bool{"disabled": true}, 200)
	f.request(admin, "POST", path, map[string]any{}, 403)
	f.request(admin, "POST", "/api/admin/users/missing/endpoint-guide", map[string]any{}, 403)
}
