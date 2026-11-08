package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func externalSessionRequest(t *testing.T, endpoint, token, body string) (int, []byte) {
	t.Helper()
	r, _ := http.NewRequest("POST", endpoint, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(r)
	if err != nil {
		t.Error(err)
		return 0, nil
	}
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	return response.StatusCode, data
}

func TestProjectSessionAPIRealClient(t *testing.T) {
	f := setupAt(t, "/agents")
	owner := f.user("owner")
	a, invite := f.enroll(owner, runtime.GOOS)
	root := t.TempDir()
	cfg, err := agent.Enroll(f.url, invite, filepath.Join(root, "config.json"), []string{root}, true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Services = []protocol.Service{{ID: "project", Name: "test", Provider: "codex", Workspace: root}}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	started := make(chan string, 4)
	daemon := &agent.Agent{Config: cfg, Providers: map[string]agent.Provider{"codex": {Name: "codex", Available: true}}, Execute: func(_ context.Context, _ agent.Provider, _ protocol.Service, prompt string, _ bool, _ func(protocol.Event) error, _ string) error {
		started <- prompt
		return nil
	}}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = daemon.Run(ctx); close(done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("daemon did not stop")
		}
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		row, _ := f.db.Agent(a.ID)
		if len(row.Services) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("client did not register projects")
		}
		time.Sleep(20 * time.Millisecond)
	}
	path := "/api/agents/" + a.ID + "/projects/project"
	raw := f.request(owner, "PUT", path+"/session-token", map[string]string{}, 200)
	var credential struct {
		Token string `json:"token"`
	}
	json.Unmarshal(raw, &credential)
	p := f.peer(owner, a.ID)
	for index, in := range []protocol.ProjectSessionCreate{{Title: "Empty"}, {Title: "Deferred", Message: "save only"}, {Title: "Immediate", Message: "execute now", Execute: true}, {Title: "HTML", Message: strings.Repeat("<>&", 10666)}, {Title: "Escaped", Message: strings.Repeat("\x01", 32000)}} {
		body, _ := json.Marshal(in)
		status, raw := externalSessionRequest(t, f.url+path+"/sessions", credential.Token, string(body))
		if status != 201 {
			t.Fatalf("create mode %d: %d %s", index, status, raw)
		}
		var result struct {
			SessionID string                `json:"sessionId"`
			TaskID    string                `json:"taskId"`
			Session   protocol.Conversation `json:"session"`
		}
		json.Unmarshal(raw, &result)
		var page protocol.HistoryPage
		if len(in.Message) < 8000 {
			p.send(t, protocol.Request{ID: "read", Type: "sessions.get", SessionID: result.SessionID})
			event := p.next(t, "read", "result")
			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &page)
		} else {
			page.Session = result.Session
		}
		if page.Session.Title != in.Title || page.Session.ServiceID != "project" {
			t.Fatal("wrong persisted session", page)
		}
		if in.Message != "" && !in.Execute {
			if len(page.Session.QueuedMessages) != 1 || page.Session.QueuedMessages[0].Prompt != in.Message || page.Session.QueuedMessages[0].Status != "paused" {
				t.Fatal("deferred message lost", page)
			}
			select {
			case <-started:
				t.Fatal("saved message executed")
			default:
			}
			p.send(t, protocol.Request{ID: "promote", Type: "tasks.promote", ServiceID: "project", SessionID: result.SessionID, RunID: result.TaskID})
			p.next(t, "promote", "result")
		}
		if index > 0 {
			select {
			case message := <-started:
				if message != in.Message {
					t.Fatal(message)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("message not executed")
			}
		}
	}
}

func TestProjectSessionAPITokenRevokedBeforeReady(t *testing.T) {
	f := setup(t)
	owner := f.user("owner")
	a, _, device, _ := f.connectedDevice(owner)
	if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: []protocol.Service{{ID: "project"}}}); err != nil {
		t.Fatal(err)
	}
	_, token, err := f.db.SetProjectSessionToken(a.UserID, a.ID, "project", "")
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan int, 1)
	go func() {
		status, _ := externalSessionRequest(t, f.url+"/api/agents/"+a.ID+"/projects/project/sessions", token, `{}`)
		result <- status
	}()
	device.SetReadDeadline(time.Now().Add(3 * time.Second))
	var open protocol.Signal
	if err := device.ReadJSON(&open); err != nil {
		t.Fatal(err)
	}
	if err := f.db.DeleteProjectSessionToken(a.UserID, a.ID, "project"); err != nil {
		t.Fatal(err)
	}
	if err := device.WriteJSON(protocol.Signal{Type: "ready", PeerID: open.PeerID, Transport: protocol.RelayTransport}); err != nil {
		t.Fatal(err)
	}
	select {
	case status := <-result:
		if status != 401 {
			t.Fatal(status)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("revocation not checked")
	}
}

func TestProjectSessionAPIAuthAndRelay(t *testing.T) {
	for _, base := range []string{"", "/agents"} {
		t.Run(base, func(t *testing.T) {
			f := setupAt(t, base)
			alice, bob := f.user("alice"), f.user("bob")
			a, credential, device, _ := f.connectedDevice(alice)
			if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: []protocol.Service{{ID: "project", Name: "Project"}, {ID: "other"}}}); err != nil {
				t.Fatal(err)
			}
			path := "/api/agents/" + a.ID + "/projects/project"
			f.request(bob, "GET", path+"/session-token", nil, 404)
			f.request(alice, "GET", path+"/session-token", nil, 200)
			f.request(alice, "PUT", path+"/session-token", map[string]string{"token": "too-short"}, 400)
			raw := f.request(alice, "PUT", path+"/session-token", map[string]string{}, 200)
			var created struct {
				Token      string `json:"token"`
				Configured bool   `json:"configured"`
			}
			if json.Unmarshal(raw, &created) != nil || created.Token == "" || !created.Configured {
				t.Fatal("missing generated token", string(raw))
			}
			if raw := f.request(alice, "GET", path+"/session-token", nil, 200); bytes.Contains(raw, []byte(created.Token)) {
				t.Fatal("metadata exposed token")
			}
			f.request(bob, "POST", path+"/session-token/reveal", map[string]string{}, 404)
			f.request(alice, "POST", path+"/session-token/reveal", map[string]string{}, 200)
			endpoint := f.url + path + "/sessions"
			for _, token := range []string{"", "wrong", credential} {
				if status, _ := externalSessionRequest(t, endpoint, token, `{}`); status != 401 {
					t.Fatal("accepted invalid token", status)
				}
			}
			if status, _ := externalSessionRequest(t, strings.Replace(endpoint, "/project/", "/other/", 1), created.Token, `{}`); status != 401 {
				t.Fatal("cross-project token", status)
			}
			for _, body := range []string{`{"serviceId":"other"}`, `{"prompt":"run"}`, `{"execute":true}`, `{"title":"` + strings.Repeat("x", 81) + `"}`, `{} {}`} {
				if status, _ := externalSessionRequest(t, endpoint, created.Token, body); status != 400 {
					t.Fatal("accepted invalid input", status, body)
				}
			}
			f.request(alice, "POST", path+"/sessions", map[string]string{}, 401)
			results := make(chan []byte, 1)
			go func() {
				status, raw := externalSessionRequest(t, endpoint, created.Token, `{"title":"From external"}`)
				if status != 201 {
					t.Errorf("create status %d: %s", status, raw)
				}
				results <- raw
			}()
			device.SetReadDeadline(time.Now().Add(3 * time.Second))
			var open, request protocol.Signal
			if err := device.ReadJSON(&open); err != nil || open.Type != "open" {
				t.Fatal("missing relay open", err)
			}
			if open.AccountTools {
				t.Fatal("granted account-wide tools to project token")
			}
			if err := device.WriteJSON(protocol.Signal{Type: "ready", PeerID: open.PeerID, Transport: protocol.RelayTransport}); err != nil {
				t.Fatal(err)
			}
			if err := device.ReadJSON(&request); err != nil || request.Request == nil || request.Request.Type != "sessions.create-external" || request.Request.ServiceID != "project" || !bytes.Contains(request.Request.Body, []byte("From external")) {
				t.Fatal("wrong fixed project relay", request, err)
			}
			conversation := protocol.Conversation{ID: "external-session", ServiceID: "project", Title: "From external", Workspace: "private-workspace"}
			if err := device.WriteJSON(protocol.Signal{Type: "event", PeerID: open.PeerID, Event: &protocol.Event{ID: request.Request.ID, Type: "result", Data: protocol.TaskStarted{Session: conversation}}}); err != nil {
				t.Fatal(err)
			}
			select {
			case raw := <-results:
				var out struct {
					SessionID string                `json:"sessionId"`
					URL       string                `json:"url"`
					Session   protocol.Conversation `json:"session"`
				}
				if json.Unmarshal(raw, &out) != nil || out.SessionID != conversation.ID || out.Session.ServiceID != "project" || !strings.Contains(out.URL, base+"/#client=") || !strings.Contains(out.URL, "session=external-session") {
					t.Fatal("invalid response", string(raw))
				}
				if bytes.Contains(raw, []byte("private-workspace")) {
					t.Fatal("local path exposed")
				}
			case <-time.After(4 * time.Second):
				t.Fatal("relay hung")
			}
			rotated := f.request(alice, "PUT", path+"/session-token", map[string]string{}, 200)
			if status, _ := externalSessionRequest(t, endpoint, created.Token, `{}`); status != 401 {
				t.Fatal("old token still accepted", status)
			}
			if json.Unmarshal(rotated, &created) != nil {
				t.Fatal("invalid rotation response")
			}
			f.request(alice, "DELETE", path+"/session-token", nil, 200)
			if status, _ := externalSessionRequest(t, endpoint, created.Token, `{}`); status != 401 {
				t.Fatal("revoked token accepted", status)
			}
		})
	}
}

func TestProjectSessionAPIOffline(t *testing.T) {
	f := setup(t)
	alice := f.user("alice")
	a, invite := f.enroll(alice, "linux")
	f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": invite, "os": "linux", "arch": "amd64"}, 200)
	if err := f.db.Heartbeat(a.ID, protocol.Signal{Services: []protocol.Service{{ID: "project"}}}); err != nil {
		t.Fatal(err)
	}
	path := "/api/agents/" + a.ID + "/projects/project"
	raw := f.request(alice, "PUT", path+"/session-token", map[string]string{}, 200)
	var created struct {
		Token string `json:"token"`
	}
	json.Unmarshal(raw, &created)
	if status, _ := externalSessionRequest(t, f.url+path+"/sessions", created.Token, `{}`); status != 503 {
		t.Fatal("offline status", status)
	}
}
