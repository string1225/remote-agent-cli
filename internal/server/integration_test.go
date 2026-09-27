package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/server"
	"github.com/string1225/remote-agent-cli/internal/store"
	"github.com/string1225/remote-agent-cli/web"
)

type fixture struct {
	t   *testing.T
	db  *store.Store
	s   *server.Server
	url string
}

func setup(t *testing.T) *fixture {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	s, handler, err := server.New(db, server.Config{PublicURL: "http://localhost:8080", AllowSignup: true}, web.Files)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(handler)
	s.Config.PublicURL = ts.URL
	t.Cleanup(func() { s.Close(); ts.Close(); db.Close() })
	return &fixture{t, db, s, ts.URL}
}
func (f *fixture) request(c *http.Client, method, path string, body any, status int) []byte {
	f.t.Helper()
	var b bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			f.t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, f.url+path, &b)
	if err != nil {
		f.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		f.t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		f.t.Fatal(err)
	}
	if res.StatusCode != status {
		f.t.Fatalf("%s %s: got %d, want %d: %s", method, path, res.StatusCode, status, data)
	}
	return data
}
func (f *fixture) user(name string) *http.Client {
	f.t.Helper()
	jar, _ := cookiejar.New(nil)
	c := &http.Client{Jar: jar, Timeout: 10 * time.Second}
	f.request(c, "POST", "/api/register", map[string]string{"username": name, "password": "a-long-test-password"}, 200)
	return c
}
func (f *fixture) enroll(c *http.Client) (store.Agent, string) {
	f.t.Helper()
	data := f.request(c, "POST", "/api/agents", map[string]string{"name": "test workstation"}, 201)
	var result struct {
		Agent store.Agent `json:"agent"`
		MacOS string      `json:"macos"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		f.t.Fatal(err)
	}
	parts := strings.Split(result.MacOS, "?token=")
	if len(parts) != 2 {
		f.t.Fatal("missing installer token")
	}
	return result.Agent, strings.Split(parts[1], "'")[0]
}
func (f *fixture) socket(c *http.Client, id string) (*websocket.Conn, *http.Response, error) {
	u, _ := url.Parse(f.url)
	header := http.Header{"Origin": []string{f.url}}
	for _, cookie := range c.Jar.Cookies(u) {
		header.Add("Cookie", cookie.String())
	}
	return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(f.url, "http")+"/api/agents/"+id+"/signal", header)
}

func TestHTTPAuthenticationAndTenantIsolation(t *testing.T) {
	f := setup(t)
	alice, bob := f.user("alice"), f.user("bob")
	a, token := f.enroll(alice)
	data := f.request(bob, "GET", "/api/agents", nil, 200)
	if string(bytes.TrimSpace(data)) != "[]" {
		t.Fatal("cross-account device visible")
	}
	f.request(bob, "DELETE", "/api/agents/"+a.ID, nil, 404)
	_, res, err := f.socket(bob, a.ID)
	if err == nil || res.StatusCode != 404 {
		t.Fatal("cross-account signal accepted")
	}
	res.Body.Close()
	request, _ := http.NewRequest("POST", f.url+"/api/agents", strings.NewReader(`{"name":"evil"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://untrusted.example")
	response, err := alice.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 403 {
		t.Fatal("cross-origin mutation accepted")
	}
	bound := f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": token, "os": "windows", "arch": "amd64"}, 200)
	if !bytes.Contains(bound, []byte("credential")) {
		t.Fatal("no device credential")
	}
	f.request(&http.Client{}, "POST", "/api/enroll", map[string]string{"token": token, "os": "windows", "arch": "amd64"}, 401)
	f.request(alice, "POST", "/api/logout", map[string]any{}, 200)
	f.request(alice, "GET", "/api/agents", nil, 401)
	f.request(alice, "POST", "/api/login", map[string]string{"username": "alice", "password": "a-long-test-password"}, 200)
	f.request(alice, "GET", "/api/agents", nil, 200)
}

type testPeer struct {
	conn   *websocket.Conn
	pc     *webrtc.PeerConnection
	dc     *webrtc.DataChannel
	events chan protocol.Event
	closed chan struct{}
	secret string
}

func (f *fixture) peer(c *http.Client, id string) *testPeer {
	f.t.Helper()
	conn, _, err := f.socket(c, id)
	if err != nil {
		f.t.Fatal(err)
	}
	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	var ready protocol.Signal
	if err := conn.ReadJSON(&ready); err != nil {
		f.t.Fatal(err)
	}
	if ready.Type != "ready" {
		f.t.Fatal(ready.Type)
	}
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		f.t.Fatal(err)
	}
	dc, err := pc.CreateDataChannel("remote-agent", nil)
	if err != nil {
		f.t.Fatal(err)
	}
	p := &testPeer{conn: conn, pc: pc, dc: dc, events: make(chan protocol.Event, 100), closed: make(chan struct{}), secret: ready.Secret}
	f.t.Cleanup(func() { conn.Close(); pc.Close() })
	var once sync.Once
	dc.OnClose(func() { once.Do(func() { close(p.closed) }) })
	dc.OnMessage(func(message webrtc.DataChannelMessage) {
		var event protocol.Event
		if json.Unmarshal(message.Data, &event) == nil {
			p.events <- event
		}
	})
	opened := make(chan struct{})
	dc.OnOpen(func() { close(opened) })
	go func() {
		pending := []webrtc.ICECandidateInit{}
		for {
			var msg protocol.Signal
			if conn.ReadJSON(&msg) != nil {
				return
			}
			switch msg.Type {
			case "answer":
				var answer webrtc.SessionDescription
				if json.Unmarshal(msg.SDP, &answer) == nil {
					if pc.SetRemoteDescription(answer) != nil {
						return
					}
					for _, c := range pending {
						_ = pc.AddICECandidate(c)
					}
					pending = nil
				}
			case "candidate":
				var c webrtc.ICECandidateInit
				if json.Unmarshal(msg.Candidate, &c) == nil {
					if pc.RemoteDescription() == nil {
						pending = append(pending, c)
					} else {
						_ = pc.AddICECandidate(c)
					}
				}
			}
		}
	}()
	complete := webrtc.GatheringCompletePromise(pc)
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		f.t.Fatal(err)
	}
	if err = pc.SetLocalDescription(offer); err != nil {
		f.t.Fatal(err)
	}
	select {
	case <-complete:
	case <-time.After(8 * time.Second):
		f.t.Fatal("ICE gathering timeout")
	}
	sdp, _ := json.Marshal(pc.LocalDescription())
	if err := conn.WriteJSON(protocol.Signal{Type: "offer", SDP: sdp}); err != nil {
		f.t.Fatal(err)
	}
	select {
	case <-opened:
	case <-time.After(10 * time.Second):
		f.t.Fatal("WebRTC DataChannel did not open")
	}
	return p
}
func (p *testPeer) send(t *testing.T, r protocol.Request) {
	t.Helper()
	b, _ := json.Marshal(r)
	if err := p.dc.SendText(string(b)); err != nil {
		t.Fatal(err)
	}
}
func (p *testPeer) next(t *testing.T, id, kind string) protocol.Event {
	t.Helper()
	timeout := time.NewTimer(8 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case e := <-p.events:
			if e.ID == id && e.Type == kind {
				return e
			}
			if e.ID == id && e.Type == "error" {
				t.Fatalf("request %s: %s", id, e.Error)
			}
		case <-timeout.C:
			t.Fatalf("timed out waiting for %s/%s", id, kind)
		}
	}
}

func TestWebRTCDirectRPCAndRevocation(t *testing.T) {
	for _, revocation := range []string{"device", "logout"} {
		t.Run(revocation, func(t *testing.T) {
			f := setup(t)
			owner := f.user("alice")
			row, token := f.enroll(owner)
			root := t.TempDir()
			cfg, err := agent.Enroll(f.url, token, filepath.Join(root, "config.json"), []string{root}, false)
			if err != nil {
				t.Fatal(err)
			}
			ran := make(chan string, 1)
			stopped := make(chan struct{})
			a := &agent.Agent{Config: cfg, Providers: map[string]agent.Provider{"codex": {Name: "codex", Available: true}}, Execute: func(ctx context.Context, p agent.Provider, s protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
				ran <- prompt
				if write {
					t.Error("read-only policy lost")
				}
				if err := send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "direct-output-中文"}}); err != nil {
					return err
				}
				<-ctx.Done()
				close(stopped)
				return ctx.Err()
			}}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { _ = a.Run(ctx); close(done) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Error("agent shutdown timed out")
				}
			})
			deadline := time.Now().Add(5 * time.Second)
			for {
				data := f.request(owner, "GET", "/api/agents", nil, 200)
				if bytes.Contains(data, []byte(`"online":true`)) {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("agent did not become online")
				}
				time.Sleep(20 * time.Millisecond)
			}
			p := f.peer(owner, row.ID)
			p.send(t, protocol.Request{ID: "auth", Type: "authenticate", Secret: p.secret})
			p.next(t, "auth", "result")
			p.send(t, protocol.Request{ID: "add", Type: "services.add", Service: protocol.Service{Name: "test", Provider: "codex", Workspace: root}})
			result := p.next(t, "add", "result")
			data, _ := json.Marshal(result.Data)
			var service protocol.Service
			if err := json.Unmarshal(data, &service); err != nil {
				t.Fatal(err)
			}
			if service.ID == "" {
				t.Fatal("missing registered service")
			}
			p.send(t, protocol.Request{ID: "run", Type: "run", ServiceID: service.ID, Prompt: "a private prompt"})
			p.next(t, "run", "started")
			output := p.next(t, "run", "output")
			encoded, _ := json.Marshal(output.Data)
			if !bytes.Contains(encoded, []byte("direct-output-中文")) {
				t.Fatal("output did not traverse datachannel")
			}
			select {
			case prompt := <-ran:
				if prompt != "a private prompt" {
					t.Fatal(prompt)
				}
			case <-time.After(time.Second):
				t.Fatal("runner not invoked")
			}
			if revocation == "device" {
				f.request(owner, "DELETE", "/api/agents/"+row.ID, nil, 200)
			} else {
				f.request(owner, "POST", "/api/logout", map[string]any{}, 200)
			}
			select {
			case <-stopped:
			case <-time.After(5 * time.Second):
				t.Fatal("revocation did not cancel active job")
			}
			select {
			case <-p.closed:
			case <-time.After(5 * time.Second):
				t.Fatal("revocation did not close P2P connection")
			}
		})
	}
}
