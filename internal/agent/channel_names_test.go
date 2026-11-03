package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestPrivateChatNameUsesMembersWhenDirectoryIsUnavailable(t *testing.T) {
	var paths []string
	m := &imAdapter{c: ChannelConfig{Provider: "feishu", Name: "研发助理"}, botID: "ou_bot", token: "mock", until: time.Now().Add(time.Hour)}
	m.client = &http.Client{Transport: imRoundTripper(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		body := `{"code":0,"data":{"name":"机器人应用名"}}`
		if strings.Contains(r.URL.Path, "/contact/") {
			body = `{"code":99991672,"msg":"fixture directory permission denied"}`
		} else if strings.HasSuffix(r.URL.Path, "/members") {
			body = `{"code":0,"data":{"items":[{"member_id":"ou_bot","name":"机器人"},{"member_id":"ou_alice","name":"小林"}]}}`
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	in := channelMessage{ChatID: "oc_chat", ChatType: "user", PeerID: "ou_alice"}
	got := m.ResolveConversation(context.Background(), in)
	if got.Name != "研发助理" || got.PeerName != "小林" || len(paths) != 3 {
		t.Fatalf("unexpected private chat identity: %#v, calls=%d", got, len(paths))
	}
	m.ResolveConversation(context.Background(), in)
	if len(paths) != 3 {
		t.Fatal("successful identity was not cached")
	}
}

func TestMissingChatNamesKeepHistoryAndNeverTurnIDsIntoNames(t *testing.T) {
	project := protocol.Service{ID: "project", Provider: "codex"}
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{project}}}
	c := ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "fixture", Name: "研发助理"}
	in := channelMessage{ChatID: "oc_private", ChatType: "user", PeerName: "小林"}
	session, _, err := a.historyStore().channelSession(c, project, in)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().nativeID(session.ID, "native-context"); err != nil {
		t.Fatal(err)
	}
	// Old versions wrote the opaque ID as the display name.
	if err = a.historyStore().update(session.ID, func(s *protocol.Conversation) { s.ChannelPeerName = in.ChatID }); err != nil {
		t.Fatal(err)
	}
	c.Name = ""
	in.PeerName = ""
	updated, _, err := a.historyStore().channelSession(c, project, in)
	if err != nil || updated.ID != session.ID || updated.NativeID != "native-context" || updated.ChannelName != "研发助理" || updated.ChannelPeerName != "" || updated.Title != "研发助理 · 私聊" {
		t.Fatalf("identity refresh changed context or exposed a raw ID: %#v, %v", updated, err)
	}
	in.PeerName = "小林"
	updated, _, err = a.historyStore().channelSession(c, project, in)
	if err != nil || updated.Title != "研发助理 · 小林" || updated.ID != session.ID {
		t.Fatalf("resolved name did not enrich the same session: %#v, %v", updated, err)
	}
}

func TestChannelNameRoundTripsInPublicViewsWithoutCredentials(t *testing.T) {
	c := ChannelConfig{ServiceID: "project", Provider: "feishu", AppID: "private-app", AppSecret: "secret-fixture", Name: "研发助理", SessionID: "legacy"}
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{{ID: c.ServiceID}}, Channels: []ChannelConfig{c}}}
	runtime := &channelRuntime{a: a, config: c, adapter: &imAdapter{c: c, botName: "机器人默认名称"}, state: "connected"}
	a.channels = map[string]*channelRuntime{c.ServiceID: runtime}
	services := a.catalogServices()
	if services[0].Channel == nil || services[0].Channel.Name != c.Name {
		t.Fatal("catalog lost configured channel name")
	}
	data, _ := json.Marshal(services)
	if strings.Contains(string(data), c.AppSecret) || strings.Contains(string(data), c.AppID) {
		t.Fatal("catalog included credentials")
	}
	view := a.channelView(c.ServiceID).(map[string]any)
	if view["name"] != c.Name {
		t.Fatal("settings lost configured name", view["name"])
	}
	runtime.config.Name = ""
	runtime.adapter = &imAdapter{botName: "机器人默认名称"}
	view = a.channelView(c.ServiceID).(map[string]any)
	if view["name"] != "机器人默认名称" {
		t.Fatal("settings lost resolved bot name", view["name"])
	}
}

func TestEmptyLegacyChannelPlaceholderIsHiddenWithoutDeletingHistory(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	s, err := h.create(protocol.Service{ID: "project"}, "IM · feishu")
	if err != nil {
		t.Fatal(err)
	}
	if err = h.update(s.ID, func(s *protocol.Conversation) { s.Kind = "im" }); err != nil {
		t.Fatal(err)
	}
	page, err := h.list("project", "", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatal("empty placeholder still shown", err)
	}
	if _, err = h.get(s.ID); err != nil {
		t.Fatal("hidden legacy session was deleted", err)
	}
	if _, err = h.append(s.ID, protocol.ChatEntry{Role: "user", Text: "preserved old message"}); err != nil {
		t.Fatal(err)
	}
	page, err = h.list("project", "", 0)
	if err != nil || len(page.Sessions) != 1 || page.Sessions[0].ID != s.ID {
		t.Fatal("legacy messages became inaccessible", err)
	}
}
