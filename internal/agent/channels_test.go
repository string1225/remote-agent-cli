package agent

import (
	"context"
	"encoding/json"
	"errors"
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
	a.Execute = func(ctx context.Context, _ Provider, s protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
		executions.Add(1)
		if write {
			t.Error("channel escalated local write permission")
		}
		text := "web-complete"
		if s.ToolSource == "im" {
			text = "request-accepted"
			if !strings.Contains(prompt, "固定的 IM 渠道会话") {
				t.Error("missing channel context")
			}
			if _, err := a.startTask(ctx, protocol.TaskStart{ServiceID: s.ID, Prompt: "delegated work", Title: "web-from-im", Source: "im", ParentAgentID: "client", ParentSessionID: s.ToolSession, Notify: true}, nil); err != nil {
				return err
			}
		}
		data := `{"type":"thread.started","thread_id":"native-thread"}` + "\n" + `{"type":"item.completed","item":{"type":"agent_message","text":"` + text + `"}}` + "\n"
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": data}})
	}
	config := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "app-public", AppSecret: "secret-fixture", ChatID: "chat", ChatType: "user"}
	if err := a.configureChannel(ctx, config); err != nil {
		t.Fatal(err)
	}
	a.channelMu.Lock()
	r := a.channels["project"]
	a.channelMu.Unlock()
	if err := r.receive(channelMessage{"message-1", "start work"}); err != nil {
		t.Fatal(err)
	}
	if err := r.receive(channelMessage{"message-1", "start work"}); err != nil {
		t.Fatal(err)
	}
	gotReply, gotCompletion := false, false
	for deadline := time.After(10 * time.Second); !gotReply || !gotCompletion; {
		select {
		case text := <-sent:
			gotReply = gotReply || strings.Contains(text, "request-accepted")
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
		if s.Kind == "im" {
			channel = s
		} else if s.Kind == "web" {
			web = s
		}
	}
	if channel.ID == "" || web.ID == "" || channel.ID == web.ID || web.ParentSessionID != channel.ID || web.Source != "im" || web.TaskStatus != "done" {
		t.Fatal("incorrect project/session relationships")
	}
	private, err := a.historyStore().get(channel.ID)
	if err != nil || private.NativeID != "native-thread" {
		t.Fatal("fixed native channel session not retained", err)
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
	for _, e := range []*larkim.P2MessageReceiveV1{decode("app", "approved", "text", mention), decode("user", "other", "text", mention), decode("user", "approved", "image", mention), decode("user", "approved", "text", `[]`), decode("user", "approved", "text", `[{"key":"@bot","id":{"open_id":"another-bot"}}]`)} {
		if _, ok := feishu.feishuMessage(e); ok {
			t.Fatal("unsafe Feishu message accepted")
		}
	}
	ding := &imAdapter{c: ChannelConfig{Provider: "dingtalk", ChatID: "approved", ChatType: "group"}}
	if _, ok := ding.dingMessage(`{"msgId":"id","msgtype":"text","conversationId":"approved","conversationType":"2","isInAtList":true,"text":{"content":"work"}}`); !ok {
		t.Fatal("valid DingTalk text rejected")
	}
	if _, ok := ding.dingMessage(`{"msgId":"id","msgtype":"text","conversationId":"other","conversationType":"2","isInAtList":true,"text":{"content":"work"}}`); ok {
		t.Fatal("wrong DingTalk destination accepted")
	}
}
