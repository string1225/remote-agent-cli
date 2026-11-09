package server_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestTodoSessionFrameOnlyAllowsSameOriginWorkspace(t *testing.T) {
	f := setupAt(t, "/agents")
	for _, path := range []string{"/?view=todo-session&client=c&project=p&session=s", "/", "/app.js?view=todo-session", "/api/me?view=todo-session"} {
		response, err := http.Get(f.url + path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		frame := response.Header.Get("X-Frame-Options")
		policy := response.Header.Get("Content-Security-Policy")
		if strings.HasPrefix(path, "/?view=") {
			if frame != "SAMEORIGIN" || !strings.Contains(policy, "frame-ancestors 'self'") {
				t.Fatal("same-origin session dialog blocked", frame, policy)
			}
		} else if frame != "DENY" || !strings.Contains(policy, "frame-ancestors 'none'") {
			t.Fatal("frame permission escaped the workspace view", path, frame, policy)
		}
	}
}
