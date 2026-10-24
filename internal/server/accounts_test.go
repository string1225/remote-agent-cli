package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestPublicRegistrationAlwaysDisabled(t *testing.T) {
	f := setup(t)
	f.request(&http.Client{}, "POST", "/api/register", map[string]string{"username": "firstadmin", "password": "secret-password"}, 403)
	res, err := http.Post(f.url+"/api/register", "text/plain", strings.NewReader("ignored"))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 403 {
		t.Fatalf("registration not always disabled: %d", res.StatusCode)
	}
}

func (f *fixture) admin(name string) *http.Client {
	f.t.Helper()
	c := f.user(name)
	u, err := f.db.User(name)
	if err != nil {
		f.t.Fatal(err)
	}
	yes := true
	if _, err := f.db.UpdateUser("", u.ID, &yes, nil, nil); err != nil {
		f.t.Fatal(err)
	}
	f.request(c, "POST", "/api/login", map[string]string{"username": name, "password": "a-long-test-password"}, 200)
	return c
}

func TestAccountRenameAndPasswordChangesRequireCurrentPassword(t *testing.T) {
	f := setupAt(t, "/agents")
	owner := f.user("Alice")
	f.user("Bob")
	a, _ := f.enroll(owner)
	original, _ := f.db.User("Alice")
	profile := f.request(owner, "GET", "/api/account", nil, 200)
	if !bytes.Contains(profile, []byte(`"username":"Alice"`)) || !bytes.Contains(profile, []byte(`"admin":false`)) {
		t.Fatalf("profile missing fields: %s", profile)
	}
	f.request(owner, "PATCH", "/api/account", map[string]string{"username": "newname"}, 403)
	for _, name := range []string{"", strings.Repeat("名", 65), "bad\nname"} {
		f.request(owner, "PATCH", "/api/account", map[string]string{"username": name, "currentPassword": "a-long-test-password"}, 400)
	}
	f.request(owner, "PATCH", "/api/account", map[string]string{"username": " bOB ", "currentPassword": "a-long-test-password"}, 409)
	f.request(owner, "PATCH", "/api/account", map[string]string{"username": "  新用户 Alice ", "currentPassword": "a-long-test-password"}, 200)
	f.request(owner, "GET", "/api/me", nil, 401)
	endpoint, _ := url.Parse(f.url + "/api/me")
	if len(owner.Jar.Cookies(endpoint)) != 0 {
		t.Fatal("rename did not clear cookie")
	}
	f.request(owner, "POST", "/api/login", map[string]string{"username": "Alice", "password": "a-long-test-password"}, 401)
	f.request(owner, "POST", "/api/login", map[string]string{"username": "新用户 ALICE", "password": "a-long-test-password"}, 200)
	renamed, _ := f.db.User("新用户 Alice")
	if renamed.ID != original.ID {
		t.Fatal("rename replaced immutable user ID")
	}
	devices := f.request(owner, "GET", "/api/agents", nil, 200)
	if !bytes.Contains(devices, []byte(a.ID)) {
		t.Fatal("rename lost owned devices")
	}
	for _, password := range []string{"short", strings.Repeat("x", 73)} {
		f.request(owner, "PATCH", "/api/account", map[string]string{"password": password, "currentPassword": "a-long-test-password"}, 400)
	}
	f.request(owner, "PATCH", "/api/account", map[string]string{"password": "new-secure-password", "currentPassword": "a-long-test-password"}, 200)
	f.request(owner, "GET", "/api/me", nil, 401)
	f.request(owner, "POST", "/api/login", map[string]string{"username": "新用户 Alice", "password": "a-long-test-password"}, 401)
	f.request(owner, "POST", "/api/login", map[string]string{"username": "新用户 Alice", "password": "new-secure-password"}, 200)
}

func TestAdministrationAuthorizationAndLastAdmin(t *testing.T) {
	f := setup(t)
	member := f.user("member")
	for _, tc := range []struct {
		method, path string
		body         any
	}{{"GET", "/api/admin/users", nil}, {"POST", "/api/admin/users", map[string]any{"username": "new", "password": "secret-password"}}, {"PATCH", "/api/admin/users/missing", map[string]bool{"disabled": true}}, {"GET", "/api/admin/tokens", nil}, {"DELETE", "/api/admin/tokens/missing", nil}} {
		f.request(&http.Client{}, tc.method, tc.path, tc.body, 401)
		f.request(member, tc.method, tc.path, tc.body, 403)
	}
	admin := f.admin("administrator")
	a, _ := f.db.User("administrator")
	f.request(admin, "PATCH", "/api/admin/users/"+a.ID, map[string]bool{"admin": false}, 409)
	f.request(admin, "PATCH", "/api/admin/users/"+a.ID, map[string]bool{"disabled": true}, 409)
	body := f.request(admin, "POST", "/api/admin/users", map[string]any{"username": " 第二管理员 ", "password": "secret-password", "admin": true}, 201)
	if bytes.Contains(body, []byte("password")) || bytes.Contains(body, []byte("$2")) {
		t.Fatal("password hash leaked", string(body))
	}
	var created store.User
	if err := json.Unmarshal(body, &created); err != nil || !created.Admin || created.Username != "第二管理员" {
		t.Fatalf("bad account response %s", body)
	}
	f.request(admin, "POST", "/api/admin/users", map[string]any{"username": "MEMBER", "password": "secret-password"}, 409)
	users := f.request(admin, "GET", "/api/admin/users?search=member", nil, 200)
	var list []store.User
	if err := json.Unmarshal(users, &list); err != nil || len(list) != 1 || list[0].Username != "member" {
		t.Fatalf("search: %s", users)
	}
	f.request(admin, "PATCH", "/api/admin/users/"+created.ID, map[string]bool{"admin": false}, 200)
	f.request(admin, "PATCH", "/api/admin/users/"+a.ID, map[string]bool{"disabled": true}, 409)
	memberUser, _ := f.db.User("member")
	f.request(admin, "PATCH", "/api/admin/users/"+memberUser.ID, map[string]string{"password": "reset-password"}, 200)
	f.request(member, "GET", "/api/me", nil, 401)
	f.request(member, "POST", "/api/login", map[string]string{"username": "member", "password": "a-long-test-password"}, 401)
	f.request(member, "POST", "/api/login", map[string]string{"username": "member", "password": "reset-password"}, 200)
}

func (f *fixture) connectedDevice(c *http.Client) (store.Agent, string, *websocket.Conn, *websocket.Conn) {
	f.t.Helper()
	a, invite := f.enroll(c)
	body := f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": invite, "os": "linux", "arch": "amd64"}, 200)
	var enrollment struct {
		Credential string `json:"credential"`
	}
	if err := json.Unmarshal(body, &enrollment); err != nil {
		f.t.Fatal(err)
	}
	device, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(f.url, "http")+"/api/agent/connect?transport=relay-v1&id="+a.ID, http.Header{"Authorization": []string{"Bearer " + enrollment.Credential}})
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { device.Close() })
	// Completing the HTTP upgrade does not guarantee the server goroutine has
	// published the device in its hub yet. Wait for a real heartbeat round trip.
	if err := device.WriteJSON(protocol.Signal{Type: "heartbeat"}); err != nil {
		f.t.Fatal(err)
	}
	_ = device.SetReadDeadline(time.Now().Add(3 * time.Second))
	var heartbeat protocol.Signal
	if err := device.ReadJSON(&heartbeat); err != nil || heartbeat.Type != "heartbeat" {
		f.t.Fatal("device heartbeat not ready", err)
	}
	_ = device.SetReadDeadline(time.Time{})
	go func() {
		for {
			var msg protocol.Signal
			if device.ReadJSON(&msg) != nil {
				return
			}
			if msg.Type == "open" {
				_ = device.WriteJSON(protocol.Signal{Type: "ready", PeerID: msg.PeerID, Transport: protocol.RelayTransport})
				return
			}
		}
	}()
	browser, _, err := f.socket(c, a.ID)
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { browser.Close() })
	browser.SetReadDeadline(time.Now().Add(3 * time.Second))
	var ready protocol.Signal
	if err := browser.ReadJSON(&ready); err != nil || ready.Type != "ready" {
		f.t.Fatal("browser not ready", err)
	}
	return a, enrollment.Credential, device, browser
}

func assertSocketClosed(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(time.Second))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if timeout, ok := err.(interface{ Timeout() bool }); ok && timeout.Timeout() {
				t.Fatal("socket remained connected until read timeout")
			}
			return
		}
	}
}

func TestDisableAccountClosesItsConnectionsAndBlocksEveryCredential(t *testing.T) {
	f := setup(t)
	admin := f.admin("administrator")
	alice, bob := f.user("alice"), f.user("bob")
	a, credential, device, browser := f.connectedDevice(alice)
	_, _, otherDevice, otherBrowser := f.connectedDevice(bob)
	owner, _ := f.db.User("alice")
	_, pending, _ := f.db.Enroll(owner.ID, "pending")
	_, personal, _ := f.db.NewCodexToken(owner.ID, "personal", "")
	f.request(admin, "PATCH", "/api/admin/users/"+owner.ID, map[string]bool{"disabled": true}, 200)
	assertSocketClosed(t, browser)
	// Browser teardown may have queued one final peer-close signal.
	assertSocketClosed(t, device)
	f.request(alice, "GET", "/api/me", nil, 401)
	f.request(alice, "POST", "/api/login", map[string]string{"username": "alice", "password": "a-long-test-password"}, 401)
	f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": pending, "os": "linux", "arch": "amd64"}, 401)
	req, _ := http.NewRequest("POST", f.url+"/api/agent/codex?id="+a.ID, strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+credential)
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatal("disabled device set up Codex", response.StatusCode)
	}
	modelDo(t, modelRequest(t, f.url, personal, "GET", "models", ""), 401)
	otherDevice.SetReadDeadline(time.Now().Add(time.Second))
	if err := otherDevice.WriteJSON(protocol.Signal{Type: "heartbeat"}); err != nil {
		t.Fatal(err)
	}
	var ack protocol.Signal
	if err := otherDevice.ReadJSON(&ack); err != nil || ack.Type != "heartbeat" {
		t.Fatal("other account device disconnected", err)
	}
	otherBrowser.SetReadDeadline(time.Now().Add(time.Second))
	if err := otherBrowser.WriteJSON(protocol.Signal{Type: "ping"}); err != nil {
		t.Fatal(err)
	}
	if err := otherBrowser.ReadJSON(&ack); err != nil || ack.Type != "pong" {
		t.Fatal("other account browser disconnected", err)
	}
}

func TestAccountChangesCloseAssociatedBrowserSockets(t *testing.T) {
	f := setup(t)
	owner := f.user("owner")
	_, _, _, browser := f.connectedDevice(owner)
	f.request(owner, "PATCH", "/api/account", map[string]string{"username": "renamed", "currentPassword": "a-long-test-password"}, 200)
	assertSocketClosed(t, browser)
}
