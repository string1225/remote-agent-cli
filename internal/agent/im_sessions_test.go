package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestIMSessionDirectoryUsesCurrentBindingsAndOnlyReturnsMetadata(t *testing.T) {
	project := protocol.Service{ID: "project", Name: "研发项目", Workspace: "private-workspace"}
	other := protocol.Service{ID: "other", Name: "另一个项目"}
	channel := ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "private-app", AppSecret: "private-secret", Name: "研发助理"}
	second := ChannelConfig{ServiceID: other.ID, Provider: "dingtalk", AppID: "other-app", AppSecret: "other-secret", Name: "项目助理"}
	a := &Agent{Config: Config{ID: "client", Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{project, other}, Channels: []ChannelConfig{channel, second}}}
	create := func(c ChannelConfig, p protocol.Service, chat, kind, name string) protocol.Conversation {
		t.Helper()
		s, _, err := a.historyStore().channelSession(c, p, channelMessage{ChatID: chat, ChatType: kind, PeerName: name})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	alice := create(channel, project, "private-chat-id", "user", "小林")
	create(channel, project, "another-chat-id", "group", "研发群")
	create(second, other, "other-chat-id", "user", "小林")
	create(channel, project, "unnamed-chat-id", "user", "")
	old := channel
	old.AppID = "replaced-app"
	create(old, project, "stale-chat-id", "user", "旧机器人对象")
	deleted := create(channel, project, "deleted-chat-id", "user", "已删除对象")
	if err := a.historyStore().update(deleted.ID, func(s *protocol.Conversation) { s.DeletedAt = 1 }); err != nil {
		t.Fatal(err)
	}
	if _, err := a.historyStore().append(alice.ID, protocol.ChatEntry{Role: "user", Text: "private-message-body"}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(a.historyStore().path)
	if err != nil {
		t.Fatal(err)
	}
	query := func(body string) map[string]any {
		t.Helper()
		var event protocol.Event
		a.handle(&peer{ctx: context.Background()}, protocol.Request{ID: "query", Type: "channels.sessions", Body: json.RawMessage(body)}, func(e protocol.Event) error { event = e; return nil })
		if event.Type != "result" {
			t.Fatalf("cannot discover IM recipients: %s %s", event.Type, event.Error)
		}
		raw, _ := json.Marshal(event.Data)
		for _, hidden := range []string{"private-secret", "private-app", "private-chat-id", "private-message-body", "private-workspace", "旧机器人对象", "已删除对象"} {
			if strings.Contains(string(raw), hidden) {
				t.Fatalf("recipient directory included private/stale data: %s", hidden)
			}
		}
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	all := query(`{}`)
	if all["clientId"] != "client" || len(all["sessions"].([]any)) != 4 {
		t.Fatal("directory lost valid recipients", all)
	}
	unknownFound := false
	for _, item := range all["sessions"].([]any) {
		recipient := item.(map[string]any)
		if recipient["peerName"] == "" {
			unknownFound = true
			if recipient["nameKnown"] != false || recipient["projectName"] != project.Name {
				t.Fatal("unknown recipient was given a fabricated name", recipient)
			}
		}
	}
	if !unknownFound {
		t.Fatal("unknown-name session was silently lost")
	}
	matches := query(`{"search":"小林"}`)["sessions"].([]any)
	if len(matches) != 2 {
		t.Fatal("same-name recipients were silently merged", matches)
	}
	filtered := query(`{"serviceId":"project","search":"小林"}`)["sessions"].([]any)
	if len(filtered) != 1 || filtered[0].(map[string]any)["sessionId"] != alice.ID || filtered[0].(map[string]any)["peerName"] != "小林" {
		t.Fatal("name-to-session mapping is incorrect", filtered)
	}
	unknown := query(`{"search":"unnamed-chat-id"}`)["sessions"].([]any)
	if len(unknown) != 0 {
		t.Fatal("raw platform IDs were treated as recipient names")
	}
	if a.channels != nil || a.tasks != nil {
		t.Fatal("read-only discovery started a bot or task subsystem")
	}
	after, err := os.ReadFile(a.historyStore().path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("read-only discovery changed the history database", err)
	}
}

func TestIMSessionDirectoryDoesNotCreateHistory(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{{ID: "project"}}, Channels: []ChannelConfig{{ServiceID: "project", Provider: "feishu", AppID: "fixture"}}}}
	page, err := a.listIMSessions(protocol.IMSessionQuery{})
	if err != nil || len(page.Sessions) != 0 {
		t.Fatal("new client must return no recipients", page, err)
	}
	if _, err := os.Stat(a.historyStore().path); !os.IsNotExist(err) {
		t.Fatal("recipient discovery created a history database", err)
	}
}

func TestIMSessionDirectoryPaginationAndInvalidQueries(t *testing.T) {
	project := protocol.Service{ID: "project"}
	channel := ChannelConfig{ServiceID: project.ID, Provider: "feishu", AppID: "fixture"}
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), Services: []protocol.Service{project}, Channels: []ChannelConfig{channel}}}
	for i := 0; i < 52; i++ {
		if _, _, err := a.historyStore().channelSession(channel, project, channelMessage{ChatID: string(rune('A' + i)), ChatType: "group", PeerName: "Team"}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := a.listIMSessions(protocol.IMSessionQuery{})
	if err != nil || len(first.Sessions) != 50 || first.NextCursor != 50 {
		t.Fatal("first recipient page incorrect", first, err)
	}
	second, err := a.listIMSessions(protocol.IMSessionQuery{Cursor: first.NextCursor})
	if err != nil || len(second.Sessions) != 2 || second.NextCursor != 0 || second.Sessions[0].SessionID <= first.Sessions[49].SessionID {
		t.Fatal("second recipient page lost/duplicated targets", second, err)
	}
	empty, err := a.listIMSessions(protocol.IMSessionQuery{Cursor: ^uint64(0)})
	if err != nil || len(empty.Sessions) != 0 {
		t.Fatal("large cursor must produce an empty page", err)
	}
	if _, err = a.listIMSessions(protocol.IMSessionQuery{Search: strings.Repeat("x", 513)}); err == nil {
		t.Fatal("unbounded search accepted")
	}
	var event protocol.Event
	a.handle(&peer{ctx: context.Background()}, protocol.Request{Type: "channels.sessions", Body: json.RawMessage(`{`)}, func(e protocol.Event) error { event = e; return nil })
	if event.Type != "error" {
		t.Fatal("invalid query accepted")
	}
	a.Config.Services[0].Name = strings.Repeat("项", 26)
	a.Config.Channels[0].Name = strings.Repeat("助", 80)
	items, err := a.historyStore().channelConversations(channel)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if err = a.historyStore().update(item.ID, func(s *protocol.Conversation) {
			s.ChannelPeerName = strings.Repeat("名", 160)
			s.Title = strings.Repeat("题", 80)
		}); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	for cursor := uint64(0); ; {
		page, err := a.listIMSessions(protocol.IMSessionQuery{Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(page)
		if len(raw) > 25000 {
			t.Fatal("recipient page exceeds relay-safe byte budget", len(raw))
		}
		for _, recipient := range page.Sessions {
			if seen[recipient.SessionID] {
				t.Fatal("byte pagination duplicated a recipient")
			}
			seen[recipient.SessionID] = true
		}
		if page.NextCursor == 0 {
			break
		}
		cursor = page.NextCursor
	}
	if len(seen) != 52 {
		t.Fatal("byte pagination lost recipients", len(seen))
	}
}
