package personalgateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func adminRequest(handler http.Handler, cookie *http.Cookie, method, path, body, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, adminPrefix+path, strings.NewReader(body))
	if cookie != nil {
		r.AddCookie(cookie)
	}
	if method == "POST" {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Sunny-Admin", "1")
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestAdminAuthenticationAndDeviceLifecycle(t *testing.T) {
	server, deviceToken, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) { t.Error("Management request escaped upstream") })
	path := filepath.Join(t.TempDir(), "admin.json")
	password, err := CreateAdmin(path)
	if err != nil {
		t.Fatal(err)
	}
	server.Admin, err = LoadAdmin(path, "https://codex.sunny-string.cn")
	if err != nil {
		t.Fatal(err)
	}
	origin := server.Admin.Origin
	if w := adminRequest(server, nil, "GET", "/api/devices", "", ""); w.Code != 401 {
		t.Fatal("Unauthenticated device list exposed")
	}
	if w := adminRequest(server, nil, "POST", "/api/login", `{"password":"wrong"}`, origin); w.Code != 401 {
		t.Fatal("Wrong password accepted")
	}
	login, _ := json.Marshal(map[string]string{"password": password})
	if w := adminRequest(server, nil, "POST", "/api/login", string(login), "https://evil.invalid"); w.Code != 403 {
		t.Fatal("Cross-origin login accepted")
	}
	w := adminRequest(server, nil, "POST", "/api/login", string(login), origin)
	if w.Code != 200 {
		t.Fatalf("Admin login failed: %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("Missing admin cookie")
	}
	cookie := cookies[0]
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != adminPrefix {
		t.Fatal("Unsafe admin cookie")
	}
	if w := request(t, server, cookie.Value, "GET", "/models", ""); w.Code != 401 {
		t.Fatal("Admin session accepted as model token")
	}
	if w := adminRequest(server, &http.Cookie{Name: adminCookie, Value: deviceToken}, "GET", "/api/devices", "", ""); w.Code != 401 {
		t.Fatal("Device token accepted as admin session")
	}
	if w := adminRequest(server, &http.Cookie{Name: adminCookie, Value: cookie.Value + "x"}, "GET", "/api/devices", "", ""); w.Code != 401 {
		t.Fatal("Tampered admin session accepted")
	}
	w = adminRequest(server, cookie, "POST", "/api/devices", `{"name":"Laptop"}`, origin)
	if w.Code != 200 {
		t.Fatalf("Create failed: %d", w.Code)
	}
	var issued map[string]string
	if json.Unmarshal(w.Body.Bytes(), &issued) != nil || !strings.HasPrefix(issued["token"], "sunny_") {
		t.Fatal("Token was not issued")
	}
	registry, _ := LoadRegistry(server.RegistryPath)
	if !registry.Authorize(issued["token"]) {
		t.Fatal("Issued token not usable")
	}
	w = adminRequest(server, cookie, "GET", "/api/devices", "", "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "sha256") || strings.Contains(w.Body.String(), issued["token"]) || strings.Contains(w.Body.String(), deviceToken) {
		t.Fatal("Device list exposed credentials")
	}
	if w := adminRequest(server, cookie, "POST", "/api/devices/delete", `{"name":"Laptop"}`, origin); w.Code != 409 {
		t.Fatal("Active device deleted without revocation")
	}
	if w := adminRequest(server, cookie, "POST", "/api/devices/revoke", `{"name":"Laptop"}`, "https://evil.invalid"); w.Code != 403 {
		t.Fatal("CSRF revoke accepted")
	}
	if w := adminRequest(server, cookie, "POST", "/api/devices/revoke", `{"name":"Laptop"}`, origin); w.Code != 200 {
		t.Fatal("Revoke failed")
	}
	registry, _ = LoadRegistry(server.RegistryPath)
	if registry.Authorize(issued["token"]) {
		t.Fatal("Revoked token still works")
	}
	if w := adminRequest(server, cookie, "POST", "/api/devices/delete", `{"name":"Laptop"}`, origin); w.Code != 200 {
		t.Fatal("Revoked device removal failed")
	}
	w = adminRequest(server, cookie, "POST", "/api/logout", `{}`, origin)
	if w.Code != 200 || w.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("Logout did not clear cookie")
	}
}

func TestAdminPageAndRateLimit(t *testing.T) {
	server, _, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {})
	path := filepath.Join(t.TempDir(), "admin.json")
	CreateAdmin(path)
	server.Admin, _ = LoadAdmin(path, "https://codex.sunny-string.cn")
	for _, path := range []string{"", "/", "/admin.js", "/admin.css"} {
		w := adminRequest(server, nil, "GET", path, "", "")
		if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("Unsafe or unavailable admin asset %s", path)
		}
	}
	for i := 0; i < 10; i++ {
		if w := adminRequest(server, nil, "POST", "/api/login", `{"password":"wrong"}`, server.Admin.Origin); w.Code != 401 {
			t.Fatal("Unexpected login response")
		}
	}
	if w := adminRequest(server, nil, "POST", "/api/login", `{"password":"wrong"}`, server.Admin.Origin); w.Code != 429 {
		t.Fatal("Login guesses not limited")
	}
}
