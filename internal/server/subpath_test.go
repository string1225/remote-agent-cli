package server_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSubpathMountAndInstallers(t *testing.T) {
	f := setupAt(t, "/agents")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Get(f.url)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 308 || res.Header.Get("Location") != "/agents/" {
		t.Fatal("mount URL did not redirect to a trailing slash")
	}
	for _, path := range []string{"/", "/app.js", "/style.css", "/healthz"} {
		res, err := client.Get(f.url + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || res.StatusCode != 200 {
			t.Fatalf("%s: %d / %v", path, res.StatusCode, err)
		}
		if path == "/" && (!bytes.Contains(body, []byte(`href="./style.css"`)) || !bytes.Contains(body, []byte(`src="./app.js"`))) {
			t.Fatal("assets are not relative to the mount")
		}
	}
	owner := f.user("alice")
	f.request(owner, "GET", "/api/me", nil, 200)
	inside, _ := url.Parse(f.url + "/api/me")
	outside, _ := url.Parse(strings.TrimSuffix(f.url, "/agents") + "/other/")
	if len(owner.Jar.Cookies(inside)) != 1 || len(owner.Jar.Cookies(outside)) != 0 {
		t.Fatal("session cookie escaped its application path")
	}
	_, token := f.enroll(owner)
	for _, script := range []string{"windows.ps1", "macos.sh"} {
		res, err := client.Get(f.url + "/install/" + script + "?token=" + token)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || res.StatusCode != 200 || !bytes.Contains(body, []byte(f.url)) {
			t.Fatalf("installer lost application mount: %s", script)
		}
	}
	res, err = client.Get(strings.TrimSuffix(f.url, "/agents") + "/api/me")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatal("unprefixed API exposed")
	}
	f.request(owner, "POST", "/api/logout", map[string]any{}, 200)
	if len(owner.Jar.Cookies(inside)) != 0 {
		t.Fatal("logout did not expire the scoped cookie")
	}
}
