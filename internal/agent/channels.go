package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type ChannelConfig struct {
	ServiceID string `json:"serviceId"`
	Provider  string `json:"provider"`
	AppID     string `json:"appId"`
	AppSecret string `json:"appSecret"`
	ChatID    string `json:"chatId"`
	ChatType  string `json:"chatType"`
	RobotCode string `json:"robotCode,omitempty"`
	SessionID string `json:"sessionId"`
}

func (c ChannelConfig) fingerprint() string {
	v := sha256.Sum256([]byte(c.Provider + "\x00" + strings.ToLower(c.AppID)))
	return hex.EncodeToString(v[:])
}

type channelMessage struct{ ID, Text string }
type channelAdapter interface {
	Validate(context.Context) error
	Connect(context.Context, func(channelMessage) error, func(string)) error
	Send(context.Context, string, string) error
}
type channelRuntime struct {
	a       *Agent
	config  ChannelConfig
	ctx     context.Context
	cancel  context.CancelFunc
	adapter channelAdapter
	mu      sync.Mutex
	state   string
	done    chan struct{}
	wake    chan struct{}
}

func (a *Agent) initChannels(ctx context.Context) {
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	if a.channels != nil {
		return
	}
	a.channels = map[string]*channelRuntime{}
	a.mu.Lock()
	configs := append([]ChannelConfig{}, a.Config.Channels...)
	a.mu.Unlock()
	for _, c := range configs {
		a.channels[c.ServiceID] = a.startChannel(ctx, c)
	}
}
func (a *Agent) startChannel(ctx context.Context, c ChannelConfig) *channelRuntime {
	ctx, cancel := context.WithCancel(ctx)
	factory := a.channelFactory
	if factory == nil {
		factory = newChannelAdapter
	}
	r := &channelRuntime{a: a, config: c, ctx: ctx, cancel: cancel, adapter: factory(c), state: "connecting", wake: make(chan struct{}, 1), done: make(chan struct{})}
	go r.connect()
	go r.outbox()
	return r
}
func (r *channelRuntime) setState(state string) { r.mu.Lock(); r.state = state; r.mu.Unlock() }
func (r *channelRuntime) connect() {
	defer close(r.done)
	for r.ctx.Err() == nil {
		r.setState("connecting")
		_ = r.adapter.Connect(r.ctx, r.receive, r.setState)
		if r.ctx.Err() != nil {
			r.setState("stopped")
			return
		}
		r.setState("retrying")
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
func (r *channelRuntime) receive(in channelMessage) error {
	if in.ID == "" || len(in.ID) > 200 || strings.TrimSpace(in.Text) == "" || len(in.Text) > 32000 {
		return nil
	}
	key := store.Hash(r.config.SessionID + "\x00" + in.ID)
	fresh, err := r.a.historyStore().channelIncoming(r.config.SessionID, key, in.Text)
	if err != nil || !fresh {
		return err
	}
	session, err := r.a.historyStore().get(r.config.SessionID)
	if err != nil {
		return err
	}
	result := protocol.TaskStarted{Session: session, TaskID: key}
	taskCtx, cancel := context.WithCancel(r.ctx)
	r.a.mu.Lock()
	r.a.taskCancels[key] = cancel
	r.a.mu.Unlock()
	finish := func(err error) {
		text := r.a.historyStore().turnText(session.ID, key, 12000)
		if err != nil {
			text = "渠道请求未完成，请稍后重试。"
		} else if strings.TrimSpace(text) == "" {
			text = "请求已处理。"
		}
		id := key + "-reply"
		if err == nil {
			_ = r.a.historyStore().channelQueue(r.config, id, text, false)
			select {
			case r.wake <- struct{}{}:
			default:
			}
		} else {
			_, _ = r.notify(text)
		}
	}
	select {
	case r.a.tasks <- queuedTask{ctx: taskCtx, cancel: cancel, recorded: true, start: protocol.TaskStart{ServiceID: r.config.ServiceID, Prompt: in.Text, Source: "im"}, result: result, complete: finish}:
		return nil
	default:
		cancel()
		r.a.mu.Lock()
		delete(r.a.taskCancels, key)
		r.a.mu.Unlock()
		_, _ = r.notify("客户端任务队列已满，本条消息已保存在渠道会话，请稍后重新发送。")
		return nil
	}
}
func (r *channelRuntime) notify(text string) (string, error) {
	if strings.TrimSpace(text) == "" || len(text) > 16000 {
		return "", errors.New("通知内容需为 1–16000 字节")
	}
	id := store.Token()
	err := r.a.historyStore().channelNotification(r.config, id, text)
	if err == nil {
		select {
		case r.wake <- struct{}{}:
		default:
		}
	}
	return id, err
}
func (r *channelRuntime) outbox() {
	for r.ctx.Err() == nil {
		items, err := r.a.historyStore().channelPending(r.config)
		if err == nil {
			for _, item := range items {
				if r.ctx.Err() != nil {
					return
				}
				ctx, cancel := context.WithTimeout(r.ctx, 20*time.Second)
				err = r.adapter.Send(ctx, item.ID, item.Text)
				cancel()
				if err != nil {
					break
				}
				_ = r.a.historyStore().channelDelivered(item.ID)
			}
		}
		wait := 30 * time.Second
		if err == nil {
			wait = 5 * time.Minute
		}
		select {
		case <-r.ctx.Done():
			return
		case <-r.wake:
		case <-time.After(wait):
		}
	}
}
func (a *Agent) notifyProject(serviceID, text string) (string, error) {
	a.channelMu.Lock()
	r := a.channels[serviceID]
	a.channelMu.Unlock()
	if r == nil {
		return "", errors.New("项目未绑定 IM 渠道")
	}
	return r.notify(text)
}
func (a *Agent) channelView(serviceID string) any {
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	r := a.channels[serviceID]
	if r == nil {
		return map[string]any{"configured": false}
	}
	r.mu.Lock()
	state := r.state
	r.mu.Unlock()
	c := r.config
	return map[string]any{"configured": true, "provider": c.Provider, "appId": c.AppID, "chatId": c.ChatID, "chatType": c.ChatType, "robotCode": c.RobotCode, "sessionId": c.SessionID, "state": state, "secretConfigured": c.AppSecret != ""}
}
func (a *Agent) configureChannel(ctx context.Context, in ChannelConfig) error {
	a.initTasks(ctx)
	a.initChannels(a.taskCtx)
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	a.mu.Lock()
	oldConfig := a.Config
	var project protocol.Service
	for _, v := range a.Config.Services {
		if v.ID == in.ServiceID {
			project = v
		}
	}
	var old ChannelConfig
	for _, v := range a.Config.Channels {
		if v.ServiceID == in.ServiceID {
			old = v
		}
	}
	a.mu.Unlock()
	if project.ID == "" {
		return errors.New("项目不存在")
	}
	if in.Provider != "" {
		if in.Provider != "feishu" && in.Provider != "dingtalk" {
			return errors.New("不支持的渠道")
		}
		in.AppID = strings.TrimSpace(in.AppID)
		in.ChatID = strings.TrimSpace(in.ChatID)
		if in.AppSecret == "" && in.Provider == old.Provider && in.AppID == old.AppID {
			in.AppSecret = old.AppSecret
		}
		if in.AppID == "" || len(in.AppID) > 200 || in.AppSecret == "" || len(in.AppSecret) > 500 || in.ChatID == "" || len(in.ChatID) > 500 || len(in.RobotCode) > 200 {
			return errors.New("请填写应用 ID、Secret 和允许接收消息的会话/用户 ID")
		}
		if in.ChatType != "group" && in.ChatType != "user" {
			return errors.New("请选择群聊或单聊")
		}
		factory := a.channelFactory
		if factory == nil {
			factory = newChannelAdapter
		}
		if err := factory(in).Validate(ctx); err != nil {
			return errors.New("机器人凭据验证失败，请检查应用 ID、Secret、权限与网络")
		}
	}
	fingerprint := ""
	if in.Provider != "" {
		fingerprint = in.fingerprint()
	}
	restoreRuntime := func() {
		if old.Provider != "" {
			a.channels[in.ServiceID] = a.startChannel(a.taskCtx, old)
		}
	}
	oldRuntime := a.channels[in.ServiceID]
	if oldRuntime != nil {
		oldRuntime.cancel()
		select {
		case <-oldRuntime.done:
		case <-ctx.Done():
			restoreRuntime()
			return ctx.Err()
		}
	}
	if _, err := ClientAPI(ctx, oldConfig, "POST", "/api/agent/channel", map[string]any{"serviceId": in.ServiceID, "fingerprint": fingerprint}); err != nil {
		restoreRuntime()
		return err
	}

	rollback := func() {
		restoreRuntime()
		fp := ""
		if old.Provider != "" {
			fp = old.fingerprint()
		}
		restore, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = ClientAPI(restore, oldConfig, "POST", "/api/agent/channel", map[string]any{"serviceId": in.ServiceID, "fingerprint": fp, "finalize": true})
	}
	if in.Provider != "" {
		in.SessionID = old.SessionID
		if in.SessionID == "" {
			session, err := a.historyStore().create(project, "IM · "+in.Provider)
			if err != nil {
				rollback()
				return err
			}
			in.SessionID = session.ID
		}
		if err := a.historyStore().update(in.SessionID, func(v *protocol.Conversation) { v.Kind = "im"; v.Source = "im" }); err != nil {
			rollback()
			return err
		}
	}
	a.mu.Lock()
	next := a.Config
	found := false
	for _, project := range next.Services {
		if project.ID == in.ServiceID {
			found = true
		}
	}
	if !found {
		a.mu.Unlock()
		rollback()
		return errors.New("项目已移除")
	}
	next.Services = append([]protocol.Service{}, a.Config.Services...)
	next.Channels = []ChannelConfig{}
	for _, v := range a.Config.Channels {
		if v.ServiceID != in.ServiceID {
			next.Channels = append(next.Channels, v)
		}
	}
	if in.Provider != "" {
		next.Channels = append(next.Channels, in)
	}
	for i := range next.Services {
		if next.Services[i].ID == in.ServiceID {
			next.Services[i].Channel = nil
			if in.Provider != "" {
				next.Services[i].Channel = &protocol.ChannelInfo{Provider: in.Provider, SessionID: in.SessionID}
			}
		}
	}
	err := next.Save()
	if err == nil {
		a.Config.Channels = next.Channels
		a.Config.Services = next.Services
	}
	a.mu.Unlock()
	if err != nil {
		rollback()
		return errors.New("无法保存客户端配置，原渠道已保留")
	}
	if oldRuntime := a.channels[in.ServiceID]; oldRuntime != nil {
		oldRuntime.cancel()
		delete(a.channels, in.ServiceID)
	}
	if in.Provider != "" {
		a.channels[in.ServiceID] = a.startChannel(a.taskCtx, in)
	}
	_, _ = ClientAPI(ctx, oldConfig, "POST", "/api/agent/channel", map[string]any{"serviceId": in.ServiceID, "fingerprint": fingerprint, "finalize": true})
	return nil
}
func (a *Agent) handleChannel(p *peer, req protocol.Request, send func(protocol.Event) error) {
	a.initTasks(p.ctx)
	a.initChannels(a.taskCtx)
	fail := func(err error) { _ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()}) }
	switch req.Type {
	case "channels.get":
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: a.channelView(req.ServiceID)})
	case "channels.configure":
		var in ChannelConfig
		if json.Unmarshal(req.Body, &in) != nil {
			fail(errors.New("invalid channel config"))
			return
		}
		if err := a.configureChannel(p.ctx, in); err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: a.channelView(in.ServiceID)})
	case "channels.notify":
		var in struct {
			ServiceID string `json:"serviceId"`
			SessionID string `json:"sessionId"`
			Message   string `json:"message"`
		}
		if json.Unmarshal(req.Body, &in) != nil {
			fail(errors.New("invalid notification"))
			return
		}
		id, err := a.notifyProject(in.ServiceID, in.Message)
		if err != nil {
			fail(err)
			return
		}
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: map[string]any{"queued": true, "messageId": id}})
	}
}
func (a *Agent) notifyTask(task queuedTask, err error) {
	if !task.start.Notify {
		return
	}
	status := "完成"
	if err != nil {
		status = "失败"
	}
	link := fmt.Sprintf("%s/#client=%s&project=%s&session=%s", a.Config.Server, a.Config.ID, task.start.ServiceID, task.result.Session.ID)
	text := fmt.Sprintf("任务%s：%s\n%s", status, task.result.Session.Title, link)
	target, project := task.start.NotifyAgentID, task.start.NotifyServiceID
	if target == "" {
		target = a.Config.ID
	}
	if project == "" {
		project = task.start.ServiceID
	}
	if target == a.Config.ID {
		_, _ = a.notifyProject(project, text)
	} else {
		_ = a.queueRemoteNotification(remoteNotification{ID: task.result.TaskID + "-completion", AgentID: target, ServiceID: project, Text: text})
	}
}

func (a *Agent) syncChannels(ctx context.Context) {
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	a.mu.Lock()
	config := a.Config
	bindings := []map[string]string{}
	for _, c := range a.Config.Channels {
		bindings = append(bindings, map[string]string{"serviceId": c.ServiceID, "fingerprint": c.fingerprint()})
	}
	a.mu.Unlock()
	_, _ = ClientAPI(ctx, config, "POST", "/api/agent/channels/sync", map[string]any{"bindings": bindings})
}
