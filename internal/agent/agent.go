package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type Runner func(context.Context, Provider, protocol.Service, string, bool, func(protocol.Event) error, string) error
type Agent struct {
	notifyWake     chan struct{}
	notifyDone     chan struct{}
	channelMu      sync.Mutex
	channels       map[string]*channelRuntime
	channelFactory func(ChannelConfig) channelAdapter
	mu             sync.Mutex
	Config         Config
	Providers      map[string]Provider
	Execute        Runner
	busy           bool
	saysoOnce      sync.Once
	sayso          *saysoStore
	saysoErr       error
	tasksOnce      sync.Once
	taskCtx        context.Context
	tasks          chan queuedTask
	taskCancels    map[string]context.CancelFunc
	taskDone       chan struct{}
}

func New(c Config) *Agent {
	c.PermissionMode, c.AllowWrite = "full", true
	return &Agent{Config: c, Providers: Detect(), Execute: runProcess}
}
func (a *Agent) services() []protocol.Service {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]protocol.Service{}, a.Config.Services...)
}
func (a *Agent) catalogServices() []protocol.Service {
	services := a.services()
	for i := range services {
		services[i].Workspace = ""
	}
	return services
}
func (a *Agent) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		timer := time.NewTimer(8 * time.Second)
		defer timer.Stop()
		for _, done := range []chan struct{}{a.taskDone, a.notifyDone} {
			select {
			case <-done:
			case <-timer.C:
				return
			}
		}
	}()
	a.initTasks(ctx)
	a.initChannels(ctx)
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
		log.Printf("control connection unavailable (%v); retrying in %s", err, backoff)
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
	q.Set("transport", protocol.RelayTransport)
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
	conn.SetReadLimit(protocol.MaxRelayBytes)
	c := &control{conn: conn}
	connectionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.syncChannels(connectionCtx)
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
			if c.send(protocol.Signal{Type: "heartbeat", Services: a.catalogServices()}) != nil {
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
		case "open":
			for id, p := range peers {
				if p.ctx.Err() != nil {
					delete(peers, id)
				}
			}
			if msg.PeerID == "" || peers[msg.PeerID] != nil || len(peers) >= 8 {
				_ = c.send(protocol.Signal{Type: "close", PeerID: msg.PeerID})
				continue
			}
			p, err := a.newPeer(connectionCtx, msg, c.send)
			if err != nil {
				_ = c.send(protocol.Signal{Type: "close", PeerID: msg.PeerID})
				continue
			}
			peers[msg.PeerID] = p
			if err := c.send(protocol.Signal{Type: "ready", PeerID: msg.PeerID, Transport: protocol.RelayTransport}); err != nil {
				return err
			}
		case "request":
			p := peers[msg.PeerID]
			if p == nil || p.ctx.Err() != nil || msg.Request == nil {
				continue
			}
			data, err := json.Marshal(msg.Request)
			if err != nil || len(data) > protocol.MaxRequestBytes || len(msg.Request.ID) == 0 || len(msg.Request.ID) > 100 {
				p.close()
				continue
			}
			// Ordered bounded workers keep disk/CLI work off the heartbeat reader.
			select {
			case p.requests <- *msg.Request:
			default:
				p.close()
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
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
	mu        sync.Mutex
	runID     string
	runCancel context.CancelFunc
	requests  chan protocol.Request
}

func (p *peer) close() { p.once.Do(p.cancel) }
func peerLifetime(msg protocol.Signal, _ time.Time) (time.Duration, error) {
	if msg.LeaseSeconds <= 0 || msg.LeaseSeconds > 30*60 {
		return 0, errors.New("invalid relay lease duration")
	}
	return time.Duration(msg.LeaseSeconds) * time.Second, nil
}
func (a *Agent) newPeer(parent context.Context, msg protocol.Signal, signal func(protocol.Signal) error) (*peer, error) {
	lifetime, err := peerLifetime(msg, time.Now())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(parent, lifetime)
	p := &peer{ctx: ctx, cancel: cancel, requests: make(chan protocol.Request, 16)}
	send := func(event protocol.Event) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if event.Type == "result" {
			data, err := json.Marshal(event.Data)
			if err != nil {
				return err
			}
			if len(data) > 48000 {
				for len(data) > 0 {
					n := min(len(data), 16000)
					chunk := protocol.Event{ID: event.ID, Type: "chunk", Data: base64.StdEncoding.EncodeToString(data[:n])}
					if err := signal(protocol.Signal{Type: "event", PeerID: msg.PeerID, Event: &chunk}); err != nil {
						p.close()
						return err
					}
					data = data[n:]
				}
				event.Data = map[string]bool{"chunked": true}
			}
		}
		err := signal(protocol.Signal{Type: "event", PeerID: msg.PeerID, Event: &event})
		if err != nil {
			p.close()
		}
		return err
	}
	go func() {
		defer p.close()
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-p.requests:
				if ctx.Err() != nil {
					return
				}
				a.handle(p, req, send)
			}
		}
	}()
	go func() { <-ctx.Done(); _ = signal(protocol.Signal{Type: "close", PeerID: msg.PeerID}) }()
	return p, nil
}
func (a *Agent) handle(p *peer, req protocol.Request, send func(protocol.Event) error) {
	fail := func(err error) { _ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()}) }
	switch req.Type {
	case "channels.get", "channels.configure", "channels.notify":
		a.handleChannel(p, req, send)
	case "tasks.cancel":
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: a.cancelTask(req.RunID)})
	case "tasks.start":
		a.handleTask(p, req, send)
	case "directories.search":
		page, err := a.directories(p.ctx, req.Path, req.Search)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: page})
	case "sayso.open":
		s, err := a.saysoStore()
		var project *soProject
		if err == nil {
			project, err = s.bindProject(req.ServiceID)
		}
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: project})
	case "sayso.call":
		a.handleSayso(p, req, send)
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
		data := map[string]any{"services": append([]protocol.Service{}, a.Config.Services...), "providers": a.Providers, "allowedRoots": a.Config.AllowedRoots, "allowWrite": true, "permissionMode": "full", "sayso": true, "directorySearch": true}
		a.mu.Unlock()
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: data})
	case "services.add":
		result, err := a.registerProject(req.Service)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: result})
	case "services.remove":
		a.mu.Lock()
		defer a.mu.Unlock()
		if a.busy {
			fail(errors.New("stop the running task before removing a service"))
			return
		}
		for _, s := range a.Config.Services {
			if s.ID == req.ServiceID && s.Channel != nil {
				fail(errors.New("请先解绑项目的 IM 渠道"))
				return
			}
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
			if err == nil {
				// Existing histories may retain a symlink spelling (notably
				// macOS /var -> /private/var). Authorize and compare real roots.
				root, rootErr := a.Config.Workspace(conversation.Workspace)
				if rootErr != nil || conversation.ServiceID != service.ID || conversation.Provider != service.Provider || root != service.Workspace {
					err = errors.New("session does not belong to this project")
				}
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
		service.Model = req.Service.Model
		service.ToolBinary, _ = os.Executable()
		service.ToolConfig = a.Config.Path
		service.ToolSession = conversation.ID
		service.ToolSource = conversation.Kind
		service.ToolNotifyAgent = conversation.NotifyAgentID
		service.ToolNotifyProject = conversation.NotifyServiceID
		service.ToolNotifySession = conversation.NotifySessionID
		if conversation.Kind == "im" {
			service.ToolNotifyAgent = a.Config.ID
			service.ToolNotifyProject = service.ID
			service.ToolNotifySession = conversation.ID
		}
		if !req.Recorded {
			if _, err = history.append(conversation.ID, protocol.ChatEntry{TurnID: req.ID, Role: "user", Text: req.Prompt, Kind: "message"}); err != nil {
				a.mu.Unlock()
				fail(err)
				return
			}
		}
		provider := a.Providers[service.Provider]
		write := true
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
					// Bound each relay event, including both display text and history data.
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
			prompt := req.Prompt
			if conversation.Kind == "im" {
				prompt = "这是固定的 IM 渠道会话。需要执行具体工作时，请使用 remote_agent 的 start_web_session 工具创建项目下的新 Web 会话，及时回复创建结果，并启用完成通知；也可以通过 list_clients_projects 查看同账号其他客户端项目。不要等待新任务结束才回复用户。\n\n" + prompt
			}
			err := a.Execute(ctx, provider, service, prompt, write, func(event protocol.Event) error {
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
