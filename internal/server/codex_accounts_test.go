package server_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCodexAccountManagementIsAdminOnlyAndUsesDedicatedControlToken(t *testing.T) {
	f := setup(t)
	admin, member := f.admin("admin"), f.user("member")
	for _, operation := range []struct{ method, path string }{{"GET", "accounts"}, {"POST", "login"}, {"POST", "activate"}, {"POST", "cancel"}} {
		f.request(member, operation.method, "/api/admin/codex/"+operation.path, map[string]string{"name": "test"}, 403)
	}
	f.request(admin, "GET", "/api/admin/codex/accounts", nil, 200)
	var path string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if r.Header.Get("Authorization") != "Bearer control-private" || r.Header.Get("Origin") != "" {
			t.Error("wrong control credential or origin")
		}
		w.Header().Set("Set-Cookie", "must-not-forward=secret")
		io.WriteString(w, `{"enabled":true,"accounts":[]}`)
	}))
	defer up.Close()
	f.s.Config.GatewayURL = up.URL
	f.s.Config.GatewayToken = "model-private"
	f.s.Config.GatewayAdminToken = "control-private"
	data := f.request(admin, "GET", "/api/admin/codex/accounts", nil, 200)
	if !bytes.Contains(data, []byte(`"enabled":true`)) || path != "/internal/codex/accounts" {
		t.Fatal("control not forwarded safely")
	}
	f.request(admin, "POST", "/api/admin/codex/login", map[string]string{"name": "账号二"}, 200)
	if path != "/internal/codex/login" {
		t.Fatal(path)
	}
	f.request(admin, "POST", "/api/admin/codex/arbitrary", map[string]string{}, 404)
	f.request(admin, "POST", "/api/admin/codex/login", map[string]string{"name": "x", "url": "http://evil.invalid"}, 400)
	if strings.Contains(string(data), "private") {
		t.Fatal("control credential exposed")
	}
}
