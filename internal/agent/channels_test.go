package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

type fakeIM struct {
	sent chan string
	fail *atomic.Bool
}

func (f *fakeIM) Validate(context.Context) error {
	if f.fail.Load() {
		return errors.New("private-secret-error")
	}
	return nil
}
func (f *fakeIM) Connect(ctx context.Context, _ func(channelMessage) error, state func(string)) error {
	state("connected")
	<-ctx.Done()
	return ctx.Err()
}
func (f *fakeIM) Send(ctx context.Context, _ string, text string) error {
	select {
	case f.sent <- text:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func TestIMConversationDelegatesWebTaskDeduplicatesAndPreservesConfiguration(t *testing.T) {
	root := t.TempDir()
	var fail atomic.Bool
	sent := make(chan string, 20)
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		if strings.Contains(string(data), "secret-fixture") {
			t.Error("robot secret sent to binding registry")
		}
		if r.Header.Get("Authorization") != "Bearer device-fixture" {
			t.Error("missing device identity")
		}
		io.WriteString(w, `{"ok":true}`)
	}))
	defer site.Close()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), ID: "client", Server: site.URL, Credential: "device-fixture", AllowedRoots: []string{root}, Services: []protocol.Service{{ID: "project", Name: "Demo", Provider: "codex", Workspace: root}}}, Providers: map[string]Provider{"codex": {Available: true}}}
	ctx, cancel := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() {
		cancel()
		select {
		case <-a.taskDone:
		case <-time.After(10 * time.Second):
			t.Error("task worker did not stop")
		}
		a.channelMu.Lock()
		runtimes := []*channelRuntime{}
		for _, r := range a.channels {
			runtimes = append(runtimes, r)
		}
		a.channelMu.Unlock()
		for _, r := range runtimes {
			select {
			case <-r.done:
			case <-time.After(time.Second):
				t.Error("channel did not stop")
			}
		}
	})
	a.channelFactory = func(ChannelConfig) channelAdapter { return &fakeIM{sent: sent, fail: &fail} }
	var executions atomic.Int32
	allowCompletion := make(chan struct{})
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		if !write {
			t.Error("channel did not use the configured full-access project permission")
		}
		text := "web-complete"
		if s.ToolSource == "im" {
			if s.ToolNotifySession != s.ToolSession {
				t.Error("IM tool context lost its original reply conversation")
			}
			text = "request-accepted " + a.Config.Server + "/#client=c&project=p&session=s"
			if !strings.Contains(prompt, "固定的 IM 渠道会话") || !strings.Contains(prompt, IMDeliveryInstructions) {
				t.Error("missing channel context")
			}
			if _, err := a.startTask(ctx, protocol.TaskStart{ServiceID: s.ID, Prompt: "delegated work", Title: "web-from-im", Source: "im", ParentAgentID: "client", ParentSessionID: s.ToolSession, Notify: true}, nil); err != nil {
				return err
			}
		} else if s.ToolNotifySession == "" || s.ToolNotifySession == s.ToolSession {
			t.Error("delegated Web task lost its origin IM conversation")
		} else {
			if !strings.Contains(prompt, IMDeliveryInstructions) {
				t.Error("delegated task missed the IM delivery contract")
			}
			select {
			case <-allowCompletion:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		data := `{"type":"thread.started","thread_id":"native-thread"}` + "\n" + `{"type":"item.completed","item":{"type":"agent_message","text":"` + text + `"}}` + "\n"
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": data}})
	}
	config := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app-public", AppSecret: "secret-fixture"}
	if err := a.configureChannel(ctx, config); err != nil {
		t.Fatal(err)
	}
	a.channelMu.Lock()
	r := a.channels["project"]
	a.channelMu.Unlock()
	if err := r.receive(channelMessage{ID: "message-1", Text: "start work", ChatID: "chat", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if err := r.receive(channelMessage{ID: "message-1", Text: "start work", ChatID: "chat", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	gotReply, gotCompletion := false, false
	for deadline := time.After(10 * time.Second); !gotReply || !gotCompletion; {
		select {
		case text := <-sent:
			if strings.Contains(text, "#client=") {
				t.Error("IM creation/completion included a workspace address")
			}
			if strings.Contains(text, "任务完成：web-from-im") && (!strings.Contains(text, "web-complete") || strings.Contains(text, "#client=")) {
				t.Error("completion did not include the actual result")
			}
			if strings.Contains(text, "request-accepted") && !gotReply {
				gotReply = true
				close(allowCompletion) // Initial confirmation must arrive before work finishes.
			}
			gotCompletion = gotCompletion || strings.Contains(text, "任务完成：web-from-im")
		case <-deadline:
			t.Fatal("channel reply or task notification missing")
		}
	}
	if executions.Load() != 2 {
		t.Fatal("duplicate message executed or web task not delegated", executions.Load())
	}
	page, err := a.historyStore().list("project", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	var channel, web protocol.Conversation
	for _, s := range page.Sessions {
		if s.Kind == "im" && s.ChannelChatID != "" {
			channel = s
		} else if s.Kind == "web" {
			web = s
		}
	}
	if channel.ID == "" || web.ID == "" || channel.ID == web.ID || web.ParentSessionID != channel.ID || web.NotifySessionID != channel.ID || web.Source != "im" || web.TaskStatus != "done" {
		t.Fatal("incorrect project/session relationships")
	}
	private, err := a.historyStore().get(channel.ID)
	if err != nil || private.NativeID != "native-thread" {
		t.Fatal("fixed native channel session not retained", err)
	}
	history, err := a.historyStore().read(channel.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	retainedLink := false
	for _, entry := range history.Entries {
		retainedLink = retainedLink || strings.Contains(entry.Text, a.Config.Server+"/#client=")
	}
	if !retainedLink {
		t.Fatal("IM filtering changed original web history")
	}
	public, _ := json.Marshal(a.channelView("project"))
	if strings.Contains(string(public), "secret-fixture") {
		t.Fatal("channel view leaked secret")
	}
	saved, _ := os.ReadFile(a.Config.Path)
	fail.Store(true)
	if err := a.configureChannel(ctx, config); err == nil {
		t.Fatal("invalid credentials replaced working channel")
	}
	fail.Store(false)
	after, _ := os.ReadFile(a.Config.Path)
	if string(saved) != string(after) {
		t.Fatal("credential failure changed saved config")
	}
	// A filesystem failure after reserving a robot must restore the old runtime
	// and retain the original on-disk configuration, project and native session.
	goodPath := a.Config.Path
	backup := filepath.Join(root, "retained-config.json")
	if err := os.Rename(goodPath, backup); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(goodPath, 0700); err != nil {
		t.Fatal(err)
	}
	if err := a.configureChannel(ctx, config); err == nil {
		t.Fatal("accepted failed config publication")
	}
	if err := os.Remove(goodPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(backup, goodPath); err != nil {
		t.Fatal(err)
	}
	after, _ = os.ReadFile(goodPath)
	if string(saved) != string(after) {
		t.Fatal("failed save changed original config")
	}
}
func TestChannelOutboxPersistenceAndRecipientIsolation(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	session, err := h.create(protocol.Service{ID: "project"}, "IM")
	if err != nil {
		t.Fatal(err)
	}
	c := ChannelConfig{Provider: "feishu", AppID: "app", ChatType: "group", ChatID: "approved", SessionID: session.ID}
	for _, id := range []string{"z-first", "a-second"} {
		if err := h.channelNotification(c, id, "notice"); err != nil {
			t.Fatal(err)
		}
	}
	reopened := historyStore{h.path}
	pending, err := reopened.channelPending(c)
	if err != nil || len(pending) != 2 || pending[0].ID != "z-first" {
		t.Fatal("outbox lost durable FIFO order", err)
	}
	changed := c
	changed.ChatID = "different"
	other, err := reopened.channelPending(changed)
	if err != nil || len(other) != 0 {
		t.Fatal("old notification routed to new recipient")
	}
	if err := reopened.channelDelivered("z-first"); err != nil {
		t.Fatal(err)
	}
	pending, err = reopened.channelPending(c)
	if err != nil || len(pending) != 1 || pending[0].ID != "a-second" {
		t.Fatal("delivery acknowledgement lost")
	}
}
func TestIMAdaptersOnlyAcceptApprovedTextAndIgnoreBotMessages(t *testing.T) {
	feishu := &imAdapter{c: ChannelConfig{Provider: "feishu", ChatID: "approved", ChatType: "group"}, botID: "own-bot"}
	decode := func(sender, chat, kind, mentions string) *larkim.P2MessageReceiveV1 {
		var event larkim.P2MessageReceiveV1
		data := `{"event":{"sender":{"sender_type":"` + sender + `"},"message":{"message_id":"id","chat_id":"` + chat + `","chat_type":"group","message_type":"` + kind + `","content":"{\"text\":\"@bot work\"}","mentions":` + mentions + `}}}`
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			t.Fatal(err)
		}
		return &event
	}
	mention := `[{"key":"@bot","id":{"open_id":"own-bot"}}]`
	if msg, ok := feishu.feishuMessage(decode("user", "approved", "text", mention)); !ok || msg.Text != "work" {
		t.Fatal("valid Feishu text rejected")
	}
	for _, e := range []*larkim.P2MessageReceiveV1{decode("app", "approved", "text", mention), decode("user", "approved", "image", mention), decode("user", "approved", "text", `[]`), decode("user", "approved", "text", `[{"key":"@bot","id":{"open_id":"another-bot"}}]`)} {
		if _, ok := feishu.feishuMessage(e); ok {
			t.Fatal("unsafe Feishu message accepted")
		}
	}
	ding := &imAdapter{c: ChannelConfig{Provider: "dingtalk", ChatID: "approved", ChatType: "group"}}
	if _, ok := ding.dingMessage(`{"msgId":"id","msgtype":"text","conversationId":"approved","conversationType":"2","isInAtList":true,"text":{"content":"work"}}`); !ok {
		t.Fatal("valid DingTalk text rejected")
	}
	if _, ok := ding.dingMessage(`{"msgId":"id","msgtype":"text","conversationId":"other","conversationType":"2","isInAtList":true,"text":{"content":"work"}}`); !ok {
		t.Fatal("dynamic DingTalk destination rejected")
	}
}

func TestIMAdaptersAcceptEveryConversationWithoutConfiguredTarget(t *testing.T) {
	m := &imAdapter{c: ChannelConfig{Provider: "feishu"}, botID: "bot"}
	for _, kind := range []string{"p2p", "group"} {
		var event larkim.P2MessageReceiveV1
		data := `{"event":{"sender":{"sender_type":"user","sender_id":{"open_id":"person"}},"message":{"message_id":"m","chat_id":"oc_dynamic","chat_type":"` + kind + `","message_type":"text","content":"{\"text\":\"@bot hello\"}","mentions":[{"key":"@bot","id":{"open_id":"bot"}}]}}}`
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			t.Fatal(err)
		}
		if _, ok := m.feishuMessage(&event); !ok {
			t.Fatalf("default channel rejected %s", kind)
		}
	}
	d := &imAdapter{c: ChannelConfig{Provider: "dingtalk"}}
	if _, ok := d.dingMessage(`{"msgId":"m","msgtype":"text","conversationId":"private","conversationType":"1","senderStaffId":"user","senderNick":"Alice","text":{"content":"hello"}}`); !ok {
		t.Fatal("default channel rejected direct message")
	}
}

func TestIMIncomingKeepsConversationIdentityAcrossRestart(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Name: "Demo", Provider: "codex", Workspace: root}}}, tasks: make(chan queuedTask, 10), taskCancels: map[string]context.CancelFunc{}}
	legacy, err := a.historyStore().create(a.Config.Services[0], "old IM history")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.historyStore().append(legacy.ID, protocol.ChatEntry{Role: "user", Text: "retained"}); err != nil {
		t.Fatal(err)
	}
	c := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app", Name: "Assistant", SessionID: legacy.ID}
	r := &channelRuntime{a: a, config: c, ctx: context.Background(), wake: make(chan struct{}, 1)}
	for i, chat := range []string{"oc_alice", "oc_group", "oc_alice"} {
		kind, peer := "user", "Alice"
		if chat == "oc_group" {
			kind, peer = "group", "Team"
		}
		if err := r.receive(channelMessage{ID: fmt.Sprint(i), Text: "hello", ChatID: chat, ChatType: kind, PeerName: peer}); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			if _, err := a.historyStore().channelTarget(c, legacy.ID); err == nil {
				t.Fatal("legacy placeholder was silently routed to a new peer")
			}
		}
		if i == 1 {
			r = &channelRuntime{a: a, config: c, ctx: context.Background(), wake: make(chan struct{}, 1)}
		}
	}
	first, second, third := <-a.tasks, <-a.tasks, <-a.tasks
	defer first.cancel()
	defer second.cancel()
	defer third.cancel()
	if first.result.Session.ID == second.result.Session.ID || first.result.Session.ID != third.result.Session.ID {
		t.Fatal("distinct IM conversations shared a native session or restart lost identity")
	}
	if _, err := a.historyStore().channelTarget(c, ""); err == nil {
		t.Fatal("ambiguous notification destination accepted")
	}
	route, err := a.historyStore().channelTarget(c, first.result.Session.ID)
	if err != nil || route.ChatID != "oc_alice" {
		t.Fatal("explicit session target lost recipient", err)
	}
	if _, err := a.historyStore().channelTarget(c, "foreign-session"); err == nil {
		t.Fatal("unbound notification session accepted")
	}
	page, err := a.historyStore().read(legacy.ID, 0)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Text != "retained" {
		t.Fatal("legacy history changed", err)
	}
}

func TestChannelLegacySessionRetainsHistoryButRebindingBotKeepsContextsSeparate(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	project := protocol.Service{ID: "project", Provider: "codex"}
	old, err := h.create(project, "old channel")
	if err != nil {
		t.Fatal(err)
	}
	if err = h.nativeID(old.ID, "native-original"); err != nil {
		t.Fatal(err)
	}
	if _, err = h.append(old.ID, protocol.ChatEntry{Role: "user", Text: "preserved"}); err != nil {
		t.Fatal(err)
	}
	c := ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "old-app", ChatID: "oc_old", ChatType: "user", SessionID: old.ID}
	s, _, err := h.channelSession(c, project, channelMessage{ChatID: "oc_old", ChatType: "user", PeerName: "Alice"})
	if err != nil || s.ID != old.ID || s.NativeID != "native-original" {
		t.Fatal("legacy identity was not retained", err)
	}
	c.AppID = "new-app"
	s, _, err = h.channelSession(c, project, channelMessage{ChatID: "oc_old", ChatType: "user", PeerName: "Alice"})
	if err != nil || s.ID == old.ID || s.NativeID != "" {
		t.Fatal("rebound bot inherited another bot context", err)
	}
	page, err := h.read(old.ID, 0)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Text != "preserved" {
		t.Fatal("legacy history changed", err)
	}
}

func TestChannelReconfigureDoesNotGiveUnmigratedHistoryToNewDestination(t *testing.T) {
	root := t.TempDir()
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, `{"ok":true}`) }))
	defer site.Close()
	project := protocol.Service{ID: "project", Provider: "codex", Workspace: root}
	var fail atomic.Bool
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{project}, Server: site.URL, ID: "client", Credential: "mock"}}
	a.channelFactory = func(ChannelConfig) channelAdapter { return &fakeIM{sent: make(chan string, 10), fail: &fail} }
	old, err := a.historyStore().create(project, "legacy Alice")
	if err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().nativeID(old.ID, "alice-native"); err != nil {
		t.Fatal(err)
	}
	if _, err = a.historyStore().append(old.ID, protocol.ChatEntry{Role: "user", Text: "Alice private history"}); err != nil {
		t.Fatal(err)
	}
	legacy := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "old-app", AppSecret: "mock", ChatID: "oc_alice", ChatType: "user", SessionID: old.ID}
	a.Config.Channels = []ChannelConfig{legacy}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		cancel()
		if a.taskDone != nil {
			<-a.taskDone
		}
		for _, r := range a.channels {
			<-r.done
		}
	}()
	next := legacy
	next.ChatID = "oc_bob"
	if err = a.configureChannel(ctx, next); err != nil {
		t.Fatal(err)
	}
	configured := a.Config.Channels[0]
	s, _, err := a.historyStore().channelSession(configured, project, channelMessage{ChatID: "oc_bob", ChatType: "user", PeerName: "Bob"})
	if err != nil || s.ID == old.ID || s.NativeID != "" {
		t.Fatal("new destination inherited unmigrated private history", err)
	}
	retained, err := a.historyStore().get(old.ID)
	if err != nil || retained.NativeID != "alice-native" {
		t.Fatal("legacy context changed", err)
	}
}

func TestChannelLegacyOutboxKeepsOriginalRecipientWhenDefaultChanges(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	project := protocol.Service{ID: "project", Provider: "codex"}
	s, err := h.create(project, "old IM")
	if err != nil {
		t.Fatal(err)
	}
	c := ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "app", ChatID: "oc_alice", ChatType: "user", SessionID: s.ID}
	if err = h.channelNotification(c, "legacy-notice", "preserved"); err != nil {
		t.Fatal(err)
	}
	if err = h.transact(true, func(tx *bolt.Tx) error {
		data, e := json.Marshal(channelDelivery{ID: "legacy-notice", Destination: c.destination(), SessionID: s.ID, Text: "preserved", Sequence: 1})
		if e != nil {
			return e
		}
		return tx.Bucket([]byte("channel_outbox")).Put([]byte("legacy-notice"), data)
	}); err != nil {
		t.Fatal(err)
	}
	if err = h.channelMigrateLegacy(c, project); err != nil {
		t.Fatal(err)
	}
	next := c
	next.ChatID = "oc_bob"
	next.SessionID = "new-default"
	items, err := h.channelPendingItems(next, true)
	if err != nil || len(items) != 1 || items[0].ChatID != "oc_alice" || items[0].SessionID != s.ID {
		t.Fatal("old outbox lost its original recipient", err)
	}
	page, err := h.read(s.ID, 0)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Text != "preserved" {
		t.Fatal("migration changed original history", err)
	}
}

type imRoundTripper func(*http.Request) (*http.Response, error)

func (f imRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestIMRepliesUseMarkdown(t *testing.T) {
	for _, provider := range []string{"feishu", "dingtalk"} {
		t.Run(provider, func(t *testing.T) {
			var got map[string]any
			m := &imAdapter{c: ChannelConfig{Provider: provider, AppID: "app", ChatID: "chat", ChatType: "group"}, token: "mock", until: time.Now().Add(time.Hour), client: &http.Client{Transport: imRoundTripper(func(r *http.Request) (*http.Response, error) {
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":0}`)), Header: make(http.Header)}, nil
			})}}
			if err := m.Send(context.Background(), "reply", "**ready**\n- item"); err != nil {
				t.Fatal(err)
			}
			if provider == "feishu" && got["msg_type"] != "interactive" {
				t.Fatal("Feishu reply is not a Markdown card", got["msg_type"])
			}
			if provider == "feishu" && !strings.Contains(fmt.Sprint(got["content"]), `"tag":"lark_md"`) {
				t.Fatal("card did not use the supported Markdown text element")
			}
			if provider == "dingtalk" && got["msgKey"] != "sampleMarkdown" {
				t.Fatal("DingTalk reply is not Markdown", got["msgKey"])
			}
		})
	}
}

type reactionIM struct {
	fakeIM
	events     chan string
	sendErr    error
	clearFails int
}

type routingIM struct {
	fakeIM
	routes chan ChannelConfig
}

func (f *routingIM) SendTo(ctx context.Context, c ChannelConfig, _ string, _ string) error {
	select {
	case f.routes <- c:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func TestChannelOutboxRoutesEveryReplyToItsPersistedConversation(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	base := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app", ChatID: "old-default", ChatType: "user"}
	for i, chat := range []string{"oc_alice", "oc_team"} {
		c := base
		c.ChatID = chat
		c.SessionID = fmt.Sprint(i)
		if i == 1 {
			c.ChatType = "group"
		}
		if err := a.historyStore().channelQueue(c, fmt.Sprint(i), "reply", false); err != nil {
			t.Fatal(err)
		}
	}
	other := base
	other.AppID = "other-bot"
	if err := a.historyStore().channelQueue(other, "foreign", "private", false); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := &routingIM{routes: make(chan ChannelConfig, 5)}
	r := &channelRuntime{a: a, config: base, adapter: f, ctx: ctx, wake: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() { defer close(done); r.outbox() }()
	defer func() { cancel(); <-done }()
	for i, chat := range []string{"oc_alice", "oc_team"} {
		select {
		case got := <-f.routes:
			if got.ChatID != chat || got.SessionID != fmt.Sprint(i) {
				t.Fatal("reply changed conversation", got.ChatID, got.SessionID)
			}
		case <-time.After(time.Second):
			t.Fatal("reply not delivered")
		}
	}
	select {
	case <-f.routes:
		t.Fatal("outbox leaked to another bot")
	default:
	}
}

type isolatedIM struct {
	fakeIM
	attempts chan string
}

func (f *isolatedIM) SendTo(ctx context.Context, c ChannelConfig, _ string, _ string) error {
	select {
	case f.attempts <- c.ChatID:
	case <-ctx.Done():
		return ctx.Err()
	}
	if c.ChatID == "oc_blocked" {
		return errors.New("recipient unavailable")
	}
	return nil
}
func TestChannelFailedConversationDoesNotBlockOtherRecipients(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	base := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}
	for i, chat := range []string{"oc_blocked", "oc_blocked", "oc_ready"} {
		c := base
		c.ChatID, c.ChatType, c.SessionID = chat, "user", chat
		if err := a.historyStore().channelQueue(c, fmt.Sprint(i), "reply", false); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := &isolatedIM{attempts: make(chan string, 10)}
	r := &channelRuntime{a: a, config: base, adapter: f, ctx: ctx, wake: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() { defer close(done); r.outbox() }()
	defer func() { cancel(); <-done }()
	for _, want := range []string{"oc_blocked", "oc_ready"} {
		select {
		case got := <-f.attempts:
			if got != want {
				t.Fatal("failed recipient lost FIFO or blocked another conversation", got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("failed conversation prevented another reply")
		}
	}
}

func (f *reactionIM) Thinking(context.Context, channelMessage) (string, error) {
	f.events <- "thinking"
	return "reaction", nil
}
func (f *reactionIM) ClearThinking(context.Context, string, string) error {
	f.events <- "clear"
	if f.clearFails > 0 {
		f.clearFails--
		return errors.New("cleanup unavailable")
	}
	return nil
}

func TestIMReactionCleanupRetryDoesNotResendDeliveredReply(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	c := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app", ChatID: "oc_chat", ChatType: "user", SessionID: "session"}
	if err := a.historyStore().channelQueueReply(c, "reply", "result", false, "message", "reaction"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := &reactionIM{events: make(chan string, 10), clearFails: 1}
	r := &channelRuntime{a: a, config: c, ctx: ctx, adapter: f, wake: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() { defer close(done); r.outbox() }()
	defer func() { cancel(); <-done }()
	for _, want := range []string{"send", "clear"} {
		select {
		case got := <-f.events:
			if got != want {
				t.Fatal(got, want)
			}
		case <-time.After(time.Second):
			t.Fatal("missing initial delivery")
		}
	}
	r.wake <- struct{}{}
	select {
	case got := <-f.events:
		if got != "clear" {
			t.Fatal("cleanup retry resent answer", got)
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup was not retried")
	}
}

func TestIMRestartClearsThinkingFromInterruptedTurn(t *testing.T) {
	root := t.TempDir()
	c := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex"}}}, tasks: make(chan queuedTask, 1), taskCancels: map[string]context.CancelFunc{}}
	f := &reactionIM{events: make(chan string, 10)}
	r := &channelRuntime{a: a, config: c, ctx: context.Background(), adapter: f, wake: make(chan struct{}, 1)}
	if err := r.receive(channelMessage{ID: "interrupted", Text: "hello", ChatID: "oc_chat", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if <-f.events != "thinking" {
		t.Fatal("missing reaction")
	}
	task := <-a.tasks
	task.cancel()
	reopened := &Agent{Config: a.Config}
	next := &reactionIM{events: make(chan string, 10)}
	reopened.channelFactory = func(ChannelConfig) channelAdapter { return next }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime := reopened.startChannel(ctx, c)
	defer func() { cancel(); <-runtime.done }()
	select {
	case event := <-next.events:
		if event != "clear" {
			t.Fatal("restart did not clear interrupted reaction", event)
		}
	case <-time.After(time.Second):
		t.Fatal("interrupted thinking reaction lost across restart")
	}
}

type blockingReactionIM struct {
	reactionIM
	release chan struct{}
}

func (f *blockingReactionIM) ClearThinking(context.Context, string, string) error {
	f.events <- "clear-start"
	<-f.release
	return nil
}
func TestChannelStopWaitsForReactionCleanup(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex"}}}, tasks: make(chan queuedTask, 1), taskCancels: map[string]context.CancelFunc{}}
	f := &blockingReactionIM{reactionIM: reactionIM{events: make(chan string, 10)}, release: make(chan struct{})}
	a.channelFactory = func(ChannelConfig) channelAdapter { return f }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := a.startChannel(ctx, ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"})
	for deadline := time.Now().Add(time.Second); ; {
		r.mu.Lock()
		state := r.state
		r.mu.Unlock()
		if state == "connected" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("channel did not connect")
		}
		time.Sleep(time.Millisecond)
	}
	if err := r.receive(channelMessage{ID: "message", Text: "work", ChatID: "oc_chat", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if <-f.events != "thinking" {
		t.Fatal("missing reaction")
	}
	task := <-a.tasks
	defer task.cancel()
	cancel()
	select {
	case event := <-f.events:
		if event != "clear-start" {
			t.Fatal(event)
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup did not start")
	}
	select {
	case <-r.done:
		close(f.release)
		t.Fatal("runtime stopped before finishing reaction cleanup")
	case <-time.After(50 * time.Millisecond):
	}
	close(f.release)
	select {
	case <-r.done:
	case <-time.After(time.Second):
		t.Fatal("runtime did not stop after cleanup")
	}
}

func TestIMReactionCleanupIsIdempotentAcrossStopAndCompletion(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	f := &reactionIM{events: make(chan string, 5)}
	r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}, adapter: f}
	for range 2 {
		if err := r.clearThinking("message", "reaction"); err != nil {
			t.Fatal(err)
		}
	}
	if <-f.events != "clear" {
		t.Fatal("missing cleanup")
	}
	select {
	case <-f.events:
		t.Fatal("completion repeated a reaction deletion after shutdown")
	default:
	}
}

func TestIMHTTPErrorRetainsSafeAPICodeWithoutProviderResponse(t *testing.T) {
	m := &imAdapter{c: ChannelConfig{Provider: "feishu"}, client: &http.Client{Transport: imRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"code":99992356,"msg":"private-secret-fixture"}`)), Header: make(http.Header)}, nil
	})}}
	_, err := m.request(context.Background(), "GET", "https://open.feishu.cn/open-apis/example", "mock", nil)
	if err == nil || !strings.Contains(err.Error(), "99992356") || strings.Contains(err.Error(), "private-secret-fixture") {
		t.Fatal("unsafe or uninformative provider error", err)
	}
}

func TestIMOperationRecoveryKeepsActualConnectionState(t *testing.T) {
	for _, connection := range []string{"connected", "retrying"} {
		t.Run(connection, func(t *testing.T) {
			a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
			f := &reactionIM{events: make(chan string, 5), clearFails: 1}
			r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}, adapter: f}
			a.channels = map[string]*channelRuntime{"project": r}
			r.setState("connected")
			if err := r.clearThinking("message", "reaction"); err == nil {
				t.Fatal("fixture did not fail")
			}
			if connection == "retrying" {
				r.setState("retrying")
			}
			if err := r.clearThinking("message", "reaction"); err != nil {
				t.Fatal(err)
			}
			view := a.channelView("project").(map[string]any)
			if view["state"] != connection || view["error"] != "" {
				t.Fatal("successful operation left stale error or invented a connection", view["state"], view["error"])
			}
		})
	}
}
func (f *reactionIM) SendTo(context.Context, ChannelConfig, string, string) error {
	f.events <- "send"
	return f.sendErr
}

func TestIMDeliveryFailureClearsThinkingAndKeepsReplyForRetry(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex"}}}, tasks: make(chan queuedTask, 1), taskCancels: map[string]context.CancelFunc{}}
	f := &reactionIM{events: make(chan string, 10), sendErr: errors.New("delivery failed")}
	r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}, ctx: ctx, adapter: f, wake: make(chan struct{}, 1)}
	if err := r.receive(channelMessage{ID: "message", Text: "hello", ChatID: "oc_chat", ChatType: "user"}); err != nil {
		t.Fatal(err)
	}
	if <-f.events != "thinking" {
		t.Fatal("missing thinking")
	}
	task := <-a.tasks
	defer task.cancel()
	task.complete(nil)
	done := make(chan struct{})
	go func() { defer close(done); r.outbox() }()
	defer func() { cancel(); <-done }()
	for _, want := range []string{"send", "clear"} {
		select {
		case event := <-f.events:
			if event != want {
				t.Fatal(event, want)
			}
		case <-time.After(time.Second):
			t.Fatal("delivery failure left thinking", want)
		}
	}
	pending, err := a.historyStore().channelPendingItems(r.config, true)
	if err != nil || len(pending) != 1 {
		t.Fatal("delivery failure lost durable reply", err)
	}
}

func TestIMThinkingRemainsUntilReplyDeliveredAndClearsOnFailure(t *testing.T) {
	for _, outcome := range []string{"success", "error", "cancel"} {
		t.Run(outcome, func(t *testing.T) {
			root := t.TempDir()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex", Workspace: root}}}, tasks: make(chan queuedTask, 1), taskCancels: map[string]context.CancelFunc{}}
			f := &reactionIM{events: make(chan string, 10)}
			r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app"}, ctx: ctx, adapter: f, wake: make(chan struct{}, 1)}
			if err := r.receive(channelMessage{ID: "message", Text: "hello", ChatID: "oc_chat", ChatType: "user"}); err != nil {
				t.Fatal(err)
			}
			select {
			case event := <-f.events:
				if event != "thinking" {
					t.Fatal(event)
				}
			default:
				t.Fatal("no thinking reaction on incoming message")
			}
			task := <-a.tasks
			defer task.cancel()
			var err error
			if outcome == "error" {
				err = errors.New("executor failed")
			} else if outcome == "cancel" {
				err = context.Canceled
			}
			task.complete(err)
			if outcome == "success" {
				select {
				case event := <-f.events:
					t.Fatal("thinking cleared before delivery", event)
				default:
				}
				done := make(chan struct{})
				go func() { defer close(done); r.outbox() }()
				defer func() { cancel(); <-done }()
				for _, want := range []string{"send", "clear"} {
					select {
					case event := <-f.events:
						if event != want {
							t.Fatal("reaction lifecycle", event, want)
						}
					case <-time.After(time.Second):
						t.Fatal("missing event", want)
					}
				}
			} else {
				select {
				case event := <-f.events:
					if event != "clear" {
						t.Fatal(event)
					}
				case <-time.After(time.Second):
					t.Fatal("failed/cancelled task left reaction")
				}
			}
		})
	}
}

func TestFeishuThinkingUsesOfficialReactionAPI(t *testing.T) {
	var calls []string
	m := &imAdapter{c: ChannelConfig{Provider: "feishu"}, token: "mock", until: time.Now().Add(time.Hour), client: &http.Client{Transport: imRoundTripper(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method == "POST" {
			var body struct {
				ReactionType struct {
					EmojiType string `json:"emoji_type"`
				} `json:"reaction_type"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || body.ReactionType.EmojiType != "THINKING" {
				t.Error("wrong thinking reaction")
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"reaction_id":"reaction"}}`)), Header: make(http.Header)}, nil
	})}}
	id, err := m.Thinking(context.Background(), channelMessage{ID: "om_message"})
	if err != nil || id != "reaction" {
		t.Fatal(id, err)
	}
	if err = m.ClearThinking(context.Background(), "om_message", id); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "POST /open-apis/im/v1/messages/om_message/reactions" || calls[1] != "DELETE /open-apis/im/v1/messages/om_message/reactions/reaction" {
		t.Fatal(calls)
	}
}

func TestFeishuConversationNamesAreResolvedAndPersisted(t *testing.T) {
	requests := 0
	firstRequest := ""
	m := &imAdapter{c: ChannelConfig{Provider: "feishu", Name: "Helper"}, token: "mock", until: time.Now().Add(time.Hour), client: &http.Client{Transport: imRoundTripper(func(r *http.Request) (*http.Response, error) {
		requests++
		if firstRequest == "" {
			firstRequest = r.URL.Path
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"name":"Design team"}}`)), Header: make(http.Header)}, nil
	})}}
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), Services: []protocol.Service{{ID: "project", Provider: "codex"}}}, tasks: make(chan queuedTask, 2), taskCancels: map[string]context.CancelFunc{}}
	r := &channelRuntime{a: a, config: ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app", Name: "Helper"}, ctx: context.Background(), adapter: m, wake: make(chan struct{}, 1)}
	for _, id := range []string{"one", "two"} {
		if err := r.receive(channelMessage{ID: id, Text: "work", ChatID: "oc_team", ChatType: "group"}); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		task := <-a.tasks
		task.cancel()
		if task.result.Session.ChannelPeerName != "Design team" {
			t.Fatal("peer display name not persisted", task.result.Session.ChannelPeerName)
		}
	}
	// One metadata request plus two reaction requests; no repeated chat lookup.
	if requests != 3 {
		t.Fatal("conversation metadata was not cached", requests)
	}
	if !strings.HasSuffix(firstRequest, "/messages/one/reactions") {
		t.Fatal("optional metadata lookup delayed the thinking reaction", firstRequest)
	}
}
