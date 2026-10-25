package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type imAdapter struct {
	c       ChannelConfig
	client  *http.Client
	mu      sync.Mutex
	token   string
	until   time.Time
	botID   string
	botName string
	names   map[string]channelIdentity
}
type channelIdentity struct {
	Name, PeerName string
	Until          time.Time
}

type imAPIError struct {
	HTTPStatus int
	Code       string
}

func (e *imAPIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("IM 服务拒绝请求（API %s），请检查应用权限和事件订阅", e.Code)
	}
	return fmt.Sprintf("IM 服务返回 HTTP %d，请检查网络、应用凭据与权限", e.HTTPStatus)
}

func newChannelAdapter(c ChannelConfig) channelAdapter {
	return &imAdapter{c: c, client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (m *imAdapter) request(ctx context.Context, method, target, token string, body any) (json.RawMessage, error) {
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("IM 请求无效")
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		if m.c.Provider == "feishu" {
			req.Header.Set("Authorization", "Bearer "+token)
		} else {
			req.Header.Set("x-acs-dingtalk-access-token", token)
		}
	}
	res, err := m.client.Do(req)
	if err != nil {
		return nil, errors.New("无法连接 IM 服务")
	}
	defer res.Body.Close()
	data, err = io.ReadAll(io.LimitReader(res.Body, 1<<20+1))
	if err != nil || len(data) > 1<<20 {
		return nil, errors.New("IM 响应无效")
	}
	var status struct {
		Code json.RawMessage `json:"code"`
	}
	validJSON := json.Unmarshal(data, &status) == nil
	code := strings.Trim(string(status.Code), `"`)
	if len(code) > 24 || strings.Trim(code, "0123456789") != "" {
		code = "unknown"
	}
	if code == "0" {
		code = ""
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if res.StatusCode == 401 {
			m.mu.Lock()
			m.token = ""
			m.mu.Unlock()
		}
		return nil, &imAPIError{HTTPStatus: res.StatusCode, Code: code}
	}
	if !validJSON {
		return nil, errors.New("IM 响应无效")
	}
	if len(status.Code) > 0 && string(status.Code) != "0" && string(status.Code) != `"0"` && string(status.Code) != `""` {
		return nil, &imAPIError{HTTPStatus: res.StatusCode, Code: code}
	}
	return data, nil
}
func (m *imAdapter) accessToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	token, valid := m.token, time.Now().Before(m.until)
	m.mu.Unlock()
	if token != "" && valid {
		return token, nil
	}
	target := "https://api.dingtalk.com/v1.0/oauth2/accessToken"
	body := map[string]string{"appKey": m.c.AppID, "appSecret": m.c.AppSecret}
	if m.c.Provider == "feishu" {
		target = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
		body = map[string]string{"app_id": m.c.AppID, "app_secret": m.c.AppSecret}
	}
	data, err := m.request(ctx, "POST", target, "", body)
	if err != nil {
		return "", err
	}
	var v struct {
		Tenant   string `json:"tenant_access_token"`
		Access   string `json:"accessToken"`
		Expire   int    `json:"expire"`
		ExpireIn int    `json:"expireIn"`
	}
	if json.Unmarshal(data, &v) != nil {
		return "", errors.New("IM 授权响应无效")
	}
	token = v.Access
	expires := v.ExpireIn
	if m.c.Provider == "feishu" {
		token = v.Tenant
		expires = v.Expire
	}
	if token == "" || expires < 60 {
		return "", errors.New("IM 授权失败")
	}
	m.mu.Lock()
	m.token = token
	m.until = time.Now().Add(time.Duration(expires-30) * time.Second)
	m.mu.Unlock()
	return token, nil
}
func (m *imAdapter) Validate(ctx context.Context) error {
	token, err := m.accessToken(ctx)
	if err != nil || m.c.Provider != "feishu" {
		return err
	}
	data, err := m.request(ctx, "GET", "https://open.feishu.cn/open-apis/bot/v3/info", token, nil)
	if err != nil {
		return err
	}
	var info struct {
		Bot struct {
			OpenID  string `json:"open_id"`
			AppName string `json:"app_name"`
		} `json:"bot"`
	}
	if json.Unmarshal(data, &info) != nil || info.Bot.OpenID == "" {
		return errors.New("无法确认飞书机器人身份，请启用并发布应用机器人能力")
	}
	m.mu.Lock()
	m.botID = info.Bot.OpenID
	m.botName = info.Bot.AppName
	m.mu.Unlock()
	return nil
}

// Names improve the session list, while a missing optional directory permission
// must never prevent a user from talking to the bot.
func (m *imAdapter) ResolveConversation(ctx context.Context, in channelMessage) channelMessage {
	if m.c.Provider != "feishu" || in.ChatID == "" {
		return in
	}
	m.mu.Lock()
	cached := m.names[in.ChatID]
	botName := m.botName
	m.mu.Unlock()
	if time.Now().Before(cached.Until) {
		in.Name, in.PeerName = cached.Name, cached.PeerName
		return in
	}
	in.Name = m.c.Name
	if in.Name == "" {
		in.Name = botName
	}
	token, err := m.accessToken(ctx)
	if err != nil {
		return in
	}
	data, err := m.request(ctx, "GET", "https://open.feishu.cn/open-apis/im/v1/chats/"+url.PathEscape(in.ChatID), token, nil)
	if err == nil {
		var info struct {
			Data struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		if json.Unmarshal(data, &info) == nil {
			in.PeerName = info.Data.Name
		}
	}
	if in.ChatType == "user" && in.PeerID != "" {
		data, err = m.request(ctx, "GET", "https://open.feishu.cn/open-apis/contact/v3/users/"+url.PathEscape(in.PeerID)+"?user_id_type=open_id", token, nil)
		if err == nil {
			var info struct {
				Data struct {
					User struct {
						Name string `json:"name"`
					} `json:"user"`
				} `json:"data"`
			}
			if json.Unmarshal(data, &info) == nil && info.Data.User.Name != "" {
				in.PeerName = info.Data.User.Name
			}
		}
	}
	m.mu.Lock()
	if m.names == nil {
		m.names = map[string]channelIdentity{}
	}
	if len(m.names) > 2000 {
		m.names = map[string]channelIdentity{}
	}
	m.names[in.ChatID] = channelIdentity{Name: in.Name, PeerName: in.PeerName, Until: time.Now().Add(15 * time.Minute)}
	m.mu.Unlock()
	return in
}
func (m *imAdapter) Send(ctx context.Context, id, text string) error {
	return m.SendTo(ctx, m.c, id, text)
}
func (m *imAdapter) SendTo(ctx context.Context, route ChannelConfig, id, text string) error {
	if route.ChatID == "" {
		return errors.New("等待 IM 会话来信")
	}
	token, err := m.accessToken(ctx)
	if err != nil {
		return err
	}
	runes := []rune(text)
	for i := 0; len(runes) > 0; i++ {
		n := min(len(runes), 3000)
		part := string(runes[:n])
		runes = runes[n:]
		if m.c.Provider == "feishu" {
			content, _ := json.Marshal(map[string]any{"config": map[string]bool{"wide_screen_mode": true}, "elements": []any{map[string]any{"tag": "div", "text": map[string]string{"tag": "lark_md", "content": part}}}})
			uuid := store.Hash(fmt.Sprintf("%s-%d", id, i))[:32]
			_, err = m.request(ctx, "POST", "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id", token, map[string]string{"receive_id": route.ChatID, "msg_type": "interactive", "content": string(content), "uuid": uuid})
		} else {
			content, _ := json.Marshal(map[string]string{"title": "Remote Agent", "text": part})
			robot := m.c.RobotCode
			if robot == "" {
				robot = m.c.AppID
			}
			body := map[string]any{"robotCode": robot, "msgKey": "sampleMarkdown", "msgParam": string(content)}
			target := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
			if route.ChatType == "user" {
				target = "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
				body["userIds"] = []string{route.ChatID}
			} else {
				body["openConversationId"] = route.ChatID
			}
			_, err = m.request(ctx, "POST", target, token, body)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Feishu exposes message reactions to application bots. DingTalk's robot
// message APIs do not expose an equivalent reaction operation.
func (m *imAdapter) Thinking(ctx context.Context, in channelMessage) (string, error) {
	if m.c.Provider != "feishu" {
		return "", nil
	}
	token, err := m.accessToken(ctx)
	if err != nil {
		return "", err
	}
	data, err := m.request(ctx, "POST", "https://open.feishu.cn/open-apis/im/v1/messages/"+url.PathEscape(in.ID)+"/reactions", token, map[string]any{"reaction_type": map[string]string{"emoji_type": "THINKING"}})
	if err != nil {
		return "", err
	}
	var reply struct {
		Data struct {
			ReactionID string `json:"reaction_id"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &reply) != nil || reply.Data.ReactionID == "" {
		return "", errors.New("IM 表情回复响应无效")
	}
	return reply.Data.ReactionID, nil
}
func (m *imAdapter) ClearThinking(ctx context.Context, messageID, reactionID string) error {
	if m.c.Provider != "feishu" || reactionID == "" {
		return nil
	}
	token, err := m.accessToken(ctx)
	if err != nil {
		return err
	}
	_, err = m.request(ctx, "DELETE", "https://open.feishu.cn/open-apis/im/v1/messages/"+url.PathEscape(messageID)+"/reactions/"+url.PathEscape(reactionID), token, nil)
	return err
}

type silentIMLogger struct{}

func (silentIMLogger) Debug(context.Context, ...interface{}) {}
func (silentIMLogger) Info(context.Context, ...interface{})  {}
func (silentIMLogger) Warn(context.Context, ...interface{})  {}
func (silentIMLogger) Error(context.Context, ...interface{}) {}
func str(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func (m *imAdapter) feishuMessage(event *larkim.P2MessageReceiveV1) (channelMessage, bool) {
	if event == nil || event.Event == nil || event.Event.Message == nil || event.Event.Sender == nil || str(event.Event.Sender.SenderType) != "user" {
		return channelMessage{}, false
	}
	v := event.Event.Message
	if str(v.ChatId) == "" || str(v.MessageType) != "text" || str(v.ChatType) != "group" && str(v.ChatType) != "p2p" {
		return channelMessage{}, false
	}
	if str(v.ChatType) == "group" {
		m.mu.Lock()
		botID := m.botID
		m.mu.Unlock()
		mentioned := false
		for _, mention := range v.Mentions {
			if mention != nil && mention.Id != nil && botID != "" && str(mention.Id.OpenId) == botID {
				mentioned = true
			}
		}
		if !mentioned {
			return channelMessage{}, false
		}
	}
	var content struct {
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(str(v.Content)), &content) != nil {
		return channelMessage{}, false
	}
	for _, mention := range v.Mentions {
		if mention != nil {
			content.Text = strings.ReplaceAll(content.Text, str(mention.Key), "")
		}
	}
	kind := "user"
	if str(v.ChatType) == "group" {
		kind = "group"
	}
	peerID := ""
	if event.Event.Sender.SenderId != nil {
		peerID = str(event.Event.Sender.SenderId.OpenId)
	}
	return channelMessage{ID: str(v.MessageId), Text: strings.TrimSpace(content.Text), ChatID: str(v.ChatId), ChatType: kind, PeerID: peerID}, true
}
func (m *imAdapter) Connect(ctx context.Context, receive func(channelMessage) error, state func(string)) error {
	if m.c.Provider == "dingtalk" {
		return m.dingConnect(ctx, receive, state)
	}
	if err := m.Validate(ctx); err != nil {
		return err
	}
	handler := dispatcher.NewEventDispatcher("", "").OnP2MessageReceiveV1(func(_ context.Context, event *larkim.P2MessageReceiveV1) error {
		msg, ok := m.feishuMessage(event)
		if !ok {
			return nil
		}
		return receive(msg)
	})
	client := larkws.NewClient(m.c.AppID, m.c.AppSecret, larkws.WithEventHandler(handler), larkws.WithHttpClient(m.client), larkws.WithLogger(silentIMLogger{}), larkws.WithLogLevel(larkcore.LogLevelError), larkws.WithOnReady(func() { state("connected") }), larkws.WithOnError(func(error) { state("error") }), larkws.WithOnReconnecting(func() { state("retrying") }), larkws.WithOnReconnected(func() { state("connected") }), larkws.WithOnDisconnected(func() { state("retrying") }))
	return client.Start(ctx)
}
func (m *imAdapter) dingMessage(data string) (channelMessage, bool) {
	var v chatbot.BotCallbackDataModel
	if json.Unmarshal([]byte(data), &v) != nil || v.Msgtype != "text" {
		return channelMessage{}, false
	}
	chatID, kind, name := v.SenderStaffId, "user", v.SenderNick
	if v.ConversationType == "2" {
		if v.ConversationId == "" || !v.IsInAtList {
			return channelMessage{}, false
		}
		chatID, kind, name = v.ConversationId, "group", v.ConversationTitle
	} else if v.ConversationType != "1" || chatID == "" {
		return channelMessage{}, false
	}
	return channelMessage{ID: v.MsgId, Text: strings.TrimSpace(v.Text.Content), ChatID: chatID, ChatType: kind, PeerName: name}, true
}
func (m *imAdapter) dingConnect(ctx context.Context, receive func(channelMessage) error, state func(string)) error {
	data, err := m.request(ctx, "POST", "https://api.dingtalk.com/v1.0/gateway/connections/open", "", map[string]any{"clientId": m.c.AppID, "clientSecret": m.c.AppSecret, "userAgent": "remote-agent/1.0", "subscriptions": []any{map[string]string{"type": "CALLBACK", "topic": payload.BotMessageCallbackTopic}}})
	if err != nil {
		return err
	}
	var endpoint struct {
		Endpoint string `json:"endpoint"`
		Ticket   string `json:"ticket"`
	}
	if json.Unmarshal(data, &endpoint) != nil || endpoint.Ticket == "" {
		return errors.New("钉钉长连接地址无效")
	}
	u, err := url.Parse(endpoint.Endpoint)
	if err != nil || u.Scheme != "wss" || u.User != nil || !(u.Hostname() == "dingtalk.com" || strings.HasSuffix(u.Hostname(), ".dingtalk.com")) {
		return errors.New("钉钉长连接地址无效")
	}
	q := u.Query()
	q.Set("ticket", endpoint.Ticket)
	u.RawQuery = q.Encode()
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second, Proxy: http.ProxyFromEnvironment}
	conn, res, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return errors.New("无法建立钉钉长连接")
	}
	defer conn.Close()
	conn.SetReadLimit(256 << 10)
	connectionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() { <-connectionCtx.Done(); conn.Close() }()
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-connectionCtx.Done():
				return
			case <-ticker.C:
				if conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)) != nil {
					conn.Close()
					return
				}
			}
		}
	}()
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(65 * time.Second)) })
	state("connected")
	for {
		_ = conn.SetReadDeadline(time.Now().Add(65 * time.Second))
		var frame payload.DataFrame
		if conn.ReadJSON(&frame) != nil {
			return errors.New("钉钉长连接已断开")
		}
		ack := payload.NewSuccessDataFrameResponse()
		ack.SetHeader(payload.DataFrameHeaderKMessageId, frame.GetMessageId())
		ack.SetHeader(payload.DataFrameHeaderKContentType, payload.DataFrameContentTypeKJson)
		if frame.Type == "SYSTEM" && frame.GetTopic() == "ping" {
			ack = payload.NewDataFrameAckPong(frame.GetMessageId())
			ack.Data = frame.Data
		}
		if frame.Type == "CALLBACK" && frame.GetTopic() == payload.BotMessageCallbackTopic {
			if msg, ok := m.dingMessage(frame.Data); ok {
				if err := receive(msg); err != nil {
					ack = payload.NewDataFrameResponse(500)
					ack.SetHeader(payload.DataFrameHeaderKMessageId, frame.GetMessageId())
				}
			}
		}
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if conn.WriteJSON(ack) != nil {
			return errors.New("钉钉长连接应答失败")
		}
		if frame.Type == "SYSTEM" && frame.GetTopic() == "disconnect" {
			return errors.New("钉钉要求重新连接")
		}
	}
}
