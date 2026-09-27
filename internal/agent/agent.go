package agent

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type Runner func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error
type Agent struct {
	mu        sync.Mutex
	Config    Config
	Providers map[string]Provider
	Execute   Runner
	busy      bool
}

func New(c Config) *Agent { return &Agent{Config: c, Providers: Detect(), Execute: runProcess} }
func (a *Agent) services() []protocol.Service {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]protocol.Service{}, a.Config.Services...)
}
func (a *Agent) Run(ctx context.Context) error {
	backoff := time.Second
	for {
		start := time.Now()
		err := a.connect(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if errors.Is(err, store.ErrDenied) {
			return errors.New("device credential revoked or invalid; enroll again from the web console")
		}
		log.Printf("control connection unavailable; retrying in %s", backoff)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff + time.Duration(rand.IntN(1000))*time.Millisecond):
		}
		if time.Since(start) > time.Minute {
			backoff = time.Second
		} else {
			backoff = min(backoff*2, 30*time.Second)
		}
	}
}

type control struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *control) send(msg protocol.Signal) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(msg)
}
func (a *Agent) connect(ctx context.Context) error {
	u, _ := url.Parse(a.Config.Server)
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/api/agent/connect"
	q := u.Query()
	q.Set("id", a.Config.ID)
	u.RawQuery = q.Encode()
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, res, err := dialer.DialContext(ctx, u.String(), http.Header{"Authorization": []string{"Bearer " + a.Config.Credential}})
	if err != nil {
		if res != nil {
			res.Body.Close()
			if res.StatusCode == 401 {
				return store.ErrDenied
			}
		}
		return err
	}
	defer conn.Close()
	conn.SetReadLimit(128 * 1024)
	c := &control{conn: conn}
	connectionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	peers := map[string]*peer{}
	defer func() {
		for _, p := range peers {
			p.close()
		}
	}()
	go func() { <-connectionCtx.Done(); conn.Close() }()
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			if c.send(protocol.Signal{Type: "heartbeat", Services: a.services()}) != nil {
				conn.Close()
				return
			}
			select {
			case <-connectionCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	log.Print("agent connected")
	for {
		_ = conn.SetReadDeadline(time.Now().Add(65 * time.Second))
		var msg protocol.Signal
		if err := conn.ReadJSON(&msg); err != nil {
			return err
		}
		switch msg.Type {
		case "offer":
			for id, p := range peers {
				if p.ctx.Err() != nil {
					delete(peers, id)
				}
			}
			if peers[msg.PeerID] != nil || len(peers) >= 8 {
				_ = c.send(protocol.Signal{Type: "error", PeerID: msg.PeerID, Error: "connection limit reached"})
				continue
			}
			p, err := a.newPeer(connectionCtx, msg, c.send)
			if err != nil {
				_ = c.send(protocol.Signal{Type: "error", PeerID: msg.PeerID, Error: "cannot negotiate peer connection"})
				continue
			}
			peers[msg.PeerID] = p
		case "candidate":
			if p := peers[msg.PeerID]; p != nil {
				var candidate webrtc.ICECandidateInit
				if json.Unmarshal(msg.Candidate, &candidate) == nil {
					_ = p.pc.AddICECandidate(candidate)
				}
			}
		case "close":
			if p := peers[msg.PeerID]; p != nil {
				p.close()
				delete(peers, msg.PeerID)
			}
		}
	}
}

type peer struct {
	pc              *webrtc.PeerConnection
	ctx             context.Context
	cancel          context.CancelFunc
	once            sync.Once
	mu              sync.Mutex
	authenticated   bool
	channelAccepted bool
	runID           string
	runCancel       context.CancelFunc
}

func (p *peer) close() { p.once.Do(func() { p.cancel(); _ = p.pc.Close() }) }
func peerLifetime(msg protocol.Signal, now time.Time) (time.Duration, error) {
	if msg.LeaseSeconds != 0 {
		if msg.LeaseSeconds < 0 || msg.LeaseSeconds > 30*60 {
			return 0, errors.New("invalid peer lease duration")
		}
		return time.Duration(msg.LeaseSeconds) * time.Second, nil
	}
	// Compatibility with older control servers that only send an absolute expiry.
	lifetime := time.Unix(msg.ExpiresAt, 0).Sub(now)
	if lifetime <= 0 || lifetime > 31*time.Minute {
		return 0, errors.New("invalid peer lease expiry; check clock synchronization or update the server")
	}
	return lifetime, nil
}
func (a *Agent) newPeer(parent context.Context, msg protocol.Signal, signal func(protocol.Signal) error) (*peer, error) {
	lifetime, err := peerLifetime(msg, time.Now())
	if err != nil {
		log.Printf("peer negotiation rejected: %v", err)
		return nil, err
	}
	if len(msg.Secret) < 32 {
		return nil, errors.New("invalid peer secret")
	}
	config := webrtc.Configuration{}
	for _, ice := range msg.ICE {
		config.ICEServers = append(config.ICEServers, webrtc.ICEServer{URLs: ice.URLs, Username: ice.Username, Credential: ice.Credential})
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(parent, lifetime)
	p := &peer{pc: pc, ctx: ctx, cancel: cancel}
	go func() { <-ctx.Done(); p.close() }()
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateFailed {
			go p.close()
		}
	})
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			b, _ := json.Marshal(candidate.ToJSON())
			if signal(protocol.Signal{Type: "candidate", PeerID: msg.PeerID, Candidate: b}) != nil {
				go p.close()
			}
		}
	})
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		p.mu.Lock()
		if p.channelAccepted || dc.Label() != "remote-agent" {
			p.mu.Unlock()
			_ = dc.Close()
			return
		}
		p.channelAccepted = true
		p.mu.Unlock()
		var sendMu sync.Mutex
		send := func(event protocol.Event) error {
			sendMu.Lock()
			defer sendMu.Unlock()
			b, err := json.Marshal(event)
			if err != nil {
				return err
			}
			deadline := time.NewTimer(10 * time.Second)
			defer deadline.Stop()
			for dc.BufferedAmount() > 1024*1024 {
				select {
				case <-p.ctx.Done():
					return p.ctx.Err()
				case <-deadline.C:
					return errors.New("client output backpressure timeout")
				case <-time.After(20 * time.Millisecond):
				}
			}
			return dc.SendText(string(b))
		}
		dc.OnClose(func() { go p.close() })
		dc.OnMessage(func(message webrtc.DataChannelMessage) {
			if !message.IsString || len(message.Data) > 64*1024 {
				go p.close()
				return
			}
			var req protocol.Request
			if json.Unmarshal(message.Data, &req) != nil || len(req.ID) == 0 || len(req.ID) > 100 {
				go p.close()
				return
			}
			p.mu.Lock()
			authenticated := p.authenticated
			if !authenticated && req.Type == "authenticate" && subtle.ConstantTimeCompare([]byte(req.Secret), []byte(msg.Secret)) == 1 {
				p.authenticated = true
				authenticated = true
			}
			p.mu.Unlock()
			if !authenticated {
				_ = send(protocol.Event{ID: req.ID, Type: "error", Error: "peer authentication required"})
				go p.close()
				return
			}
			a.handle(p, req, send)
		})
	})
	var offer webrtc.SessionDescription
	if err = json.Unmarshal(msg.SDP, &offer); err != nil || offer.Type != webrtc.SDPTypeOffer {
		p.close()
		return nil, errors.New("invalid offer")
	}
	if err = pc.SetRemoteDescription(offer); err != nil {
		p.close()
		return nil, err
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		p.close()
		return nil, err
	}
	if err = pc.SetLocalDescription(answer); err != nil {
		p.close()
		return nil, err
	}
	sdp, _ := json.Marshal(answer)
	if err = signal(protocol.Signal{Type: "answer", PeerID: msg.PeerID, SDP: sdp}); err != nil {
		p.close()
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
			p.mu.Lock()
			authenticated := p.authenticated
			p.mu.Unlock()
			if !authenticated {
				p.close()
			}
		}
	}()
	return p, nil
}
func (a *Agent) handle(p *peer, req protocol.Request, send func(protocol.Event) error) {
	fail := func(err error) { _ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()}) }
	switch req.Type {
	case "sessions.list":
		page, err := a.historyStore().list(req.ServiceID, req.Search, req.Cursor)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: page})
	case "sessions.get":
		page, err := a.historyStore().read(req.SessionID, req.Cursor)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: page})
	case "sessions.create":
		var service protocol.Service
		for _, s := range a.services() {
			if s.ID == req.ServiceID {
				service = s
				break
			}
		}
		if service.ID == "" {
			fail(errors.New("unknown service"))
			return
		}
		conversation, err := a.historyStore().create(service, req.Title)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: conversation})
	case "authenticate", "services.list":
		a.mu.Lock()
		data := map[string]any{"services": append([]protocol.Service{}, a.Config.Services...), "providers": a.Providers, "allowedRoots": a.Config.AllowedRoots, "allowWrite": a.Config.AllowWrite}
		a.mu.Unlock()
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: data})
	case "services.add":
		a.mu.Lock()
		defer a.mu.Unlock()
		if len(a.Config.Services) >= 100 {
			fail(errors.New("100 service limit"))
			return
		}
		if !a.Providers[req.Service.Provider].Available {
			fail(errors.New("install and sign in to a compatible provider CLI locally first"))
			return
		}
		if len(strings.TrimSpace(req.Service.Name)) == 0 || len(req.Service.Name) > 80 {
			fail(errors.New("service name must be 1–80 bytes"))
			return
		}
		workspace, err := a.Config.Workspace(req.Service.Workspace)
		if err != nil {
			fail(err)
			return
		}
		s := protocol.Service{ID: store.Token(), Name: strings.TrimSpace(req.Service.Name), Provider: req.Service.Provider, Workspace: workspace}
		a.Config.Services = append(a.Config.Services, s)
		if err = a.Config.Save(); err != nil {
			a.Config.Services = a.Config.Services[:len(a.Config.Services)-1]
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: s})
	case "services.remove":
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.busy {
			fail(errors.New("stop the running task before removing a service"))
			return
		}
		old := a.Config.Services
		list := []protocol.Service{}
		for _, s := range old {
			if s.ID != req.ServiceID {
				list = append(list, s)
			}
		}
		a.Config.Services = list
		if err := a.Config.Save(); err != nil {
			a.Config.Services = old
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: true})
	case "run":
		if len(strings.TrimSpace(req.Prompt)) == 0 || len(req.Prompt) > 32000 {
			fail(errors.New("prompt must be 1–32000 bytes"))
			return
		}
		a.mu.Lock()
		if a.busy {
			a.mu.Unlock()
			fail(errors.New("this device is already running a task"))
			return
		}
		var service protocol.Service
		for _, s := range a.Config.Services {
			if s.ID == req.ServiceID {
				service = s
				break
			}
		}
		if service.ID == "" {
			a.mu.Unlock()
			fail(errors.New("unknown service"))
			return
		}
		workspace, err := a.Config.Workspace(service.Workspace)
		if err != nil {
			a.mu.Unlock()
			fail(err)
			return
		}
		service.Workspace = workspace
		history := a.historyStore()
		var conversation protocol.Conversation
		if req.SessionID != "" {
			conversation, err = history.get(req.SessionID)
			if err == nil && (conversation.ServiceID != service.ID || conversation.Provider != service.Provider || conversation.Workspace != service.Workspace) {
				err = errors.New("session does not belong to this project")
			}
		} else {
			conversation, err = history.create(service, req.Prompt)
		}
		if err != nil {
			a.mu.Unlock()
			fail(err)
			return
		}
		service.ResumeID = conversation.NativeID
		if _, err = history.append(conversation.ID, protocol.ChatEntry{TurnID: req.ID, Role: "user", Text: req.Prompt, Kind: "message"}); err != nil {
			a.mu.Unlock()
			fail(err)
			return
		}
		provider := a.Providers[service.Provider]
		write := a.Config.AllowWrite
		a.busy = true
		a.mu.Unlock()
		ctx, cancel := context.WithTimeout(p.ctx, 20*time.Minute)
		p.mu.Lock()
		p.runID = req.ID
		p.runCancel = cancel
		p.mu.Unlock()
		go func() {
			defer cancel()
			_ = send(protocol.Event{ID: req.ID, Type: "started", Data: map[string]string{"sessionId": conversation.ID}})
			decoder := &transcript{
				native: func(id string) error { return history.nativeID(conversation.ID, id) },
				text: func(stream, text string) error {
					role := "assistant"
					if stream != "stdout" {
						role = "system"
					}
					// Bound each SCTP message, including both display text and history data.
					runes := []rune(text)
					for len(runes) > 0 {
						n := min(len(runes), 2000)
						chunk := string(runes[:n])
						runes = runes[n:]
						entries, err := history.append(conversation.ID, protocol.ChatEntry{TurnID: req.ID, Role: role, Text: chunk, Kind: stream})
						if err != nil {
							return err
						}
						if err = send(protocol.Event{ID: req.ID, Type: "output", Data: map[string]any{"stream": stream, "text": chunk, "normalized": true, "entries": entries}}); err != nil {
							return err
						}
					}
					return nil
				},
			}
			err := a.Execute(ctx, provider, service, req.Prompt, write, func(event protocol.Event) error {
				if event.Type != "output" {
					return send(event)
				}
				data, ok := event.Data.(map[string]string)
				if !ok {
					return errors.New("invalid provider output")
				}
				return decoder.write(data["stream"], data["text"])
			}, req.ID)
			if flushErr := decoder.flush(); err == nil {
				err = flushErr
			}
			ending := protocol.ChatEntry{TurnID: req.ID, Role: "system", Kind: "done"}
			if err != nil {
				ending.Kind = "error"
				ending.Text = err.Error()
			}
			if _, saveErr := history.append(conversation.ID, ending); saveErr != nil && err == nil {
				err = saveErr
			}
			p.mu.Lock()
			p.runCancel = nil
			p.runID = ""
			p.mu.Unlock()
			a.mu.Lock()
			a.busy = false
			a.mu.Unlock()
			event := protocol.Event{ID: req.ID, Type: "done", Data: map[string]bool{"ok": err == nil}}
			if err != nil {
				event.Error = err.Error()
			}
			_ = send(event)
		}()
	case "cancel":
		p.mu.Lock()
		if p.runID == req.RunID && p.runCancel != nil {
			p.runCancel()
		}
		p.mu.Unlock()
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: true})
	default:
		fail(errors.New("unknown request type"))
	}
}
