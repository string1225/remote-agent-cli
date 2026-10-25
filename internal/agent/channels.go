package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type ChannelConfig struct {
	ServiceID string `json:"serviceId"`
	Name      string `json:"name,omitempty"`
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

type channelMessage struct {
	ID, Text                         string
	ChatID, ChatType, Name, PeerName string
	PeerID                           string
}
type channelAdapter interface {
	Validate(context.Context) error
	Connect(context.Context, func(channelMessage) error, func(string)) error
	Send(context.Context, string, string) error
}
type channelRuntime struct {
	a                *Agent
	config           ChannelConfig
	ctx              context.Context
	cancel           context.CancelFunc
	adapter          channelAdapter
	mu               sync.Mutex
	state            string
	lastError        string
	operationErrors  map[string]string
	reactionError    string
	updatedAt        int64
	reactions        map[string]string
	reactionMu       sync.Mutex
	clearedReactions map[string]bool
	done             chan struct{}
	wake             chan struct{}
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
	go func() {
		defer close(r.done)
		r.recoverReactions()
		var workers sync.WaitGroup
		workers.Add(2)
		go func() { defer workers.Done(); r.connect() }()
		go func() { defer workers.Done(); r.outbox() }()
		<-ctx.Done()
		workers.Wait()
		r.clearReactions()
		r.setState("stopped")
	}()
	return r
}
func (r *channelRuntime) setState(state string) {
	r.mu.Lock()
	changed := r.state != state
	r.state, r.updatedAt = state, time.Now().UnixMilli()
	if state == "connected" {
		r.lastError = ""
	}
	if state == "error" && r.lastError == "" {
		r.lastError = "机器人连接失败，请检查网络、应用发布状态及事件订阅权限"
	}
	r.mu.Unlock()
	if changed {
		log.Printf("IM channel %s state: %s", r.config.Provider, state)
	}
}
func channelPublicError(err error) string {
	message := "机器人连接或消息发送失败，请检查网络、应用权限和事件订阅"
	var api *imAPIError
	if errors.As(err, &api) {
		message = api.Error()
	}
	return message
}
func (r *channelRuntime) setError(err error) {
	log.Printf("IM channel %s error: %s", r.config.Provider, channelPublicError(err))
	r.mu.Lock()
	r.lastError = channelPublicError(err)
	r.mu.Unlock()
	r.setState("error")
}
func (r *channelRuntime) setOperationError(key string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.operationErrors == nil {
		r.operationErrors = map[string]string{}
	}
	message := channelPublicError(err)
	if r.operationErrors[key] != message {
		log.Printf("IM channel %s operation failed: %s", r.config.Provider, message)
	}
	r.operationErrors[key] = message
	r.updatedAt = time.Now().UnixMilli()
}
func (r *channelRuntime) clearOperationError(key string) {
	r.mu.Lock()
	delete(r.operationErrors, key)
	r.updatedAt = time.Now().UnixMilli()
	r.mu.Unlock()
}
func (r *channelRuntime) connect() {
	for r.ctx.Err() == nil {
		r.setState("connecting")
		err := r.adapter.Connect(r.ctx, r.receive, r.setState)
		if r.ctx.Err() != nil {
			r.setState("stopped")
			return
		}
		if err != nil {
			r.setError(err)
		} else {
			r.setState("retrying")
		}
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
func (r *channelRuntime) receive(in channelMessage) error {
	if r.ctx.Err() != nil {
		return r.ctx.Err()
	}
	if in.ID == "" || len(in.ID) > 200 || strings.TrimSpace(in.Text) == "" || len(in.Text) > 32000 {
		return nil
	}
	var project protocol.Service
	r.a.mu.Lock()
	if r.a.removingProjects[r.config.ServiceID] {
		r.a.mu.Unlock()
		return errors.New("项目正在删除")
	}
	for _, s := range r.a.Config.Services {
		if s.ID == r.config.ServiceID {
			project = s
		}
	}
	r.a.mu.Unlock()
	session, route, err := r.a.historyStore().channelSession(r.config, project, in)
	if err != nil {
		return err
	}
	key := store.Hash(session.ID + "\x00" + in.ID)
	fresh, err := r.a.historyStore().channelIncoming(session.ID, key, in.Text)
	if err != nil || !fresh {
		return err
	}
	log.Printf("IM message accepted provider=%s session=%s", r.config.Provider, session.ID)
	reactionID := ""
	if reactions, ok := r.adapter.(interface {
		Thinking(context.Context, channelMessage) (string, error)
	}); ok {
		reactionCtx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
		reactionID, err = reactions.Thinking(reactionCtx, in)
		cancel()
		if err == nil && reactionID != "" {
			if trackErr := r.a.historyStore().channelTrackReaction(r.config, in.ID, reactionID); trackErr != nil {
				r.setOperationError("storage", trackErr)
				if r.clearThinking(in.ID, reactionID) == nil {
					reactionID = ""
				}
			}
			r.mu.Lock()
			if r.reactions == nil {
				r.reactions = map[string]string{}
			}
			if reactionID != "" {
				r.reactions[in.ID] = reactionID
			}
			r.mu.Unlock()
		}
		if err != nil {
			r.mu.Lock()
			r.reactionError = "无法添加思考表情，请检查机器人的消息表情权限"
			r.mu.Unlock()
		} else {
			r.mu.Lock()
			r.reactionError = ""
			r.mu.Unlock()
		}
	}
	if resolver, ok := r.adapter.(interface {
		ResolveConversation(context.Context, channelMessage) channelMessage
	}); ok {
		ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
		resolved := resolver.ResolveConversation(ctx, in)
		cancel()
		in.Name, in.PeerName = resolved.Name, resolved.PeerName
		if enriched, _, err := r.a.historyStore().channelSession(r.config, project, in); err == nil {
			session = enriched
		}
	}
	result := protocol.TaskStarted{Session: session, TaskID: key}
	taskCtx, cancel := context.WithCancel(r.ctx)
	r.a.mu.Lock()
	r.a.taskCancels[key] = cancel
	r.a.mu.Unlock()
	finish := func(err error) {
		text := r.a.historyStore().turnText(session.ID, key, 1<<20)
		if err != nil {
			text = "渠道请求未完成，请稍后重试。"
		} else if strings.TrimSpace(text) == "" {
			text = "请求已处理。"
		}
		id := key + "-reply"
		if err == nil {
			if queueErr := r.a.historyStore().channelQueueReply(route, id, text, false, in.ID, reactionID); queueErr != nil {
				r.clearThinking(in.ID, reactionID)
				r.setOperationError("storage", queueErr)
				return
			}
			select {
			case r.wake <- struct{}{}:
			default:
			}
		} else {
			r.clearThinking(in.ID, reactionID)
			_, _ = r.notifyTo(route, text)
		}
	}
	select {
	case r.a.tasks <- queuedTask{ctx: taskCtx, cancel: cancel, recorded: true, start: protocol.TaskStart{ServiceID: r.config.ServiceID, Prompt: in.Text, Source: "im"}, result: result, complete: finish}:
		return nil
	default:
		cancel()
		r.clearThinking(in.ID, reactionID)
		r.a.mu.Lock()
		delete(r.a.taskCancels, key)
		r.a.mu.Unlock()
		_, _ = r.notifyTo(route, "客户端任务队列已满，本条消息已保存在渠道会话，请稍后重新发送。")
		return nil
	}
}
func (r *channelRuntime) clearThinking(messageID, reactionID string) error {
	if reactionID == "" {
		return nil
	}
	r.reactionMu.Lock()
	defer r.reactionMu.Unlock()
	key := messageID + "\x00" + reactionID
	if r.clearedReactions[key] {
		err := r.a.historyStore().channelReactionCleared(r.config, messageID, reactionID)
		if err == nil {
			r.clearOperationError("reaction:" + messageID)
		}
		return err
	}
	if err := r.a.historyStore().channelRetireReaction(r.config, messageID); err != nil {
		r.setOperationError("reaction:"+messageID, err)
		return err
	}
	if reactions, ok := r.adapter.(interface {
		ClearThinking(context.Context, string, string) error
	}); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := reactions.ClearThinking(ctx, messageID, reactionID)
		cancel()
		if err != nil {
			r.setOperationError("reaction:"+messageID, err)
			return err
		}
	}
	r.mu.Lock()
	delete(r.reactions, messageID)
	r.mu.Unlock()
	if len(r.clearedReactions) >= 4096 {
		r.clearedReactions = map[string]bool{}
	}
	if r.clearedReactions == nil {
		r.clearedReactions = map[string]bool{}
	}
	r.clearedReactions[key] = true
	err := r.a.historyStore().channelReactionCleared(r.config, messageID, reactionID)
	if err == nil {
		r.clearOperationError("reaction:" + messageID)
	} else {
		r.setOperationError("reaction:"+messageID, err)
	}
	return err
}
func (r *channelRuntime) recoverReactions() {
	pending, err := r.a.historyStore().channelPendingReactions(r.config, true)
	if err != nil {
		r.setOperationError("storage", err)
		return
	}
	for _, item := range pending {
		r.clearThinking(item.MessageID, item.ReactionID)
	}
}
func (r *channelRuntime) clearReactions() {
	r.mu.Lock()
	pending := map[string]string{}
	for id, reaction := range r.reactions {
		pending[id] = reaction
	}
	r.mu.Unlock()
	for id, reaction := range pending {
		r.clearThinking(id, reaction)
	}
}
func (r *channelRuntime) notify(text string) (string, error) {
	route, err := r.a.historyStore().channelTarget(r.config, "")
	if err != nil {
		return "", err
	}
	return r.notifyTo(route, text)
}
func (r *channelRuntime) notifyTo(route ChannelConfig, text string) (string, error) {
	if strings.TrimSpace(text) == "" || len(text) > 16000 {
		return "", errors.New("通知内容需为 1–16000 字节")
	}
	id := store.Token()
	err := r.a.historyStore().channelNotification(route, id, text)
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
		var cleanupErr error
		if pending, err := r.a.historyStore().channelPendingReactions(r.config, false); err == nil {
			for _, item := range pending {
				if r.ctx.Err() != nil {
					return
				}
				if err := r.clearThinking(item.MessageID, item.ReactionID); err != nil {
					cleanupErr = err
				}
			}
		} else {
			cleanupErr = err
			r.setOperationError("storage", err)
		}
		items, err := r.a.historyStore().channelPendingItems(r.config, true)
		var retryErr error
		blockedDestinations := map[string]bool{}
		if err != nil {
			r.setOperationError("storage", err)
		}
		if err == nil {
			for _, item := range items {
				if r.ctx.Err() != nil {
					return
				}
				if blockedDestinations[item.Destination] {
					continue
				}
				if !item.Delivered {
					ctx, cancel := context.WithTimeout(r.ctx, 20*time.Second)
					if sender, ok := r.adapter.(interface {
						SendTo(context.Context, ChannelConfig, string, string) error
					}); ok {
						route := r.config
						if item.ChatID != "" {
							route.ChatID, route.ChatType, route.SessionID = item.ChatID, item.ChatType, item.SessionID
						}
						err = sender.SendTo(ctx, route, item.ID, item.Text)
					} else {
						err = r.adapter.Send(ctx, item.ID, item.Text)
					}
					cancel()
					if err != nil {
						r.clearThinking(item.MessageID, item.ReactionID)
						r.setOperationError("send:"+item.ID, err)
						retryErr = err
						err = nil
						blockedDestinations[item.Destination] = true
						continue
					}
					if err = r.a.historyStore().channelSent(item.ID); err != nil {
						r.setOperationError("storage", err)
						break
					}
					r.clearOperationError("send:" + item.ID)
				}
				if err = r.clearThinking(item.MessageID, item.ReactionID); err != nil {
					retryErr = err
					err = nil
					continue
				}
				if err = r.a.historyStore().channelDelivered(item.ID); err != nil {
					r.setOperationError("storage", err)
					break
				}
			}
		}
		if err == nil {
			if retryErr != nil {
				err = retryErr
			} else if cleanupErr != nil {
				err = cleanupErr
			} else {
				r.clearOperationError("storage")
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
	return a.notifyProjectSession(serviceID, "", text)
}
func (a *Agent) notifyProjectSession(serviceID, sessionID, text string) (string, error) {
	a.channelMu.Lock()
	r := a.channels[serviceID]
	a.channelMu.Unlock()
	if r == nil {
		return "", errors.New("项目未绑定 IM 渠道")
	}
	route, err := a.historyStore().channelTarget(r.config, sessionID)
	if err != nil {
		return "", err
	}
	return r.notifyTo(route, text)
}
func (a *Agent) channelView(serviceID string) any {
	a.channelMu.Lock()
	defer a.channelMu.Unlock()
	r := a.channels[serviceID]
	if r == nil {
		return map[string]any{"configured": false}
	}
	r.mu.Lock()
	state, lastError, reactionError, updatedAt := r.state, r.lastError, r.reactionError, r.updatedAt
	if len(r.operationErrors) > 0 {
		keys := make([]string, 0, len(r.operationErrors))
		for key := range r.operationErrors {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if lastError == "" {
			lastError = r.operationErrors[keys[0]]
		}
		if state == "connected" {
			state = "error"
		}
	}
	r.mu.Unlock()
	c := r.config
	stage := "starting"
	if state == "connected" {
		stage = "success"
	} else if state == "error" {
		stage = "error"
	} else if state == "stopped" {
		stage = "configured"
	}
	conversations := []map[string]string{}
	all, _ := a.historyStore().channelConversations(c)
	for _, s := range all {
		conversations = append(conversations, map[string]string{"sessionId": s.ID, "chatType": s.ChannelChatType, "name": s.ChannelName, "peerName": s.ChannelPeerName})
	}
	return map[string]any{"configured": true, "provider": c.Provider, "name": c.Name, "accessPolicy": "everyone", "appId": c.AppID, "chatId": c.ChatID, "chatType": c.ChatType, "robotCode": c.RobotCode, "sessionId": c.SessionID, "state": state, "stateStage": stage, "error": lastError, "updatedAt": updatedAt, "conversations": conversations, "waitingForMessage": len(conversations) == 0, "thinkingReactionSupported": c.Provider == "feishu", "reactionError": reactionError, "secretConfigured": c.AppSecret != ""}
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
		in.Name = strings.TrimSpace(in.Name)
		in.ChatID = strings.TrimSpace(in.ChatID)
		// Keep an existing notification destination for pre-upgrade durable
		// deliveries. It is never used as an incoming-message allowlist.
		if in.ChatID == "" && in.Provider == old.Provider && in.AppID == old.AppID {
			in.ChatID, in.ChatType = old.ChatID, old.ChatType
		}
		if in.AppSecret == "" && in.Provider == old.Provider && in.AppID == old.AppID {
			in.AppSecret = old.AppSecret
		}
		if in.AppID == "" || len(in.AppID) > 200 || in.AppSecret == "" || len(in.AppSecret) > 500 || len(in.ChatID) > 500 || len(in.RobotCode) > 200 || len([]rune(in.Name)) > 80 {
			return errors.New("请填写应用 ID 和 Secret，渠道名称最多 80 字")
		}
		if in.ChatType != "" && in.ChatType != "group" && in.ChatType != "user" {
			return errors.New("无效的旧渠道会话类型")
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
	if old.Provider != "" {
		if err := a.historyStore().channelMigrateLegacy(old, project); err != nil {
			rollback()
			return errors.New("无法保留原渠道历史和待发通知，原配置已保留")
		}
	}
	if in.Provider != "" {
		in.SessionID = old.SessionID
		if old.fingerprint() != in.fingerprint() || old.ChatID != in.ChatID || old.ChatType != in.ChatType {
			in.SessionID = ""
		}
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
		id, err := a.notifyProjectSession(in.ServiceID, in.SessionID, in.Message)
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
	notifySession := task.start.NotifySessionID
	if notifySession == "" && task.start.ParentAgentID == target && task.start.Source == "im" {
		notifySession = task.start.ParentSessionID
	}
	if target == a.Config.ID {
		_, _ = a.notifyProjectSession(project, notifySession, text)
	} else {
		_ = a.queueRemoteNotification(remoteNotification{ID: task.result.TaskID + "-completion", AgentID: target, ServiceID: project, SessionID: notifySession, Text: text})
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
