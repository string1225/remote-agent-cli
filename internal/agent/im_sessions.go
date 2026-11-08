package agent

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

// Read recipient routes without initializing or modifying history. This path
// must also work before the first incoming IM message has created a database.
func (h historyStore) readIMRoutes(channels []ChannelConfig) (out []protocol.Conversation, err error) {
	db, err := bolt.Open(h.path, 0600, &bolt.Options{ReadOnly: true, Timeout: 5 * time.Second})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bindings := map[string]ChannelConfig{}
	for _, channel := range channels {
		bindings[channel.ServiceID] = channel
	}
	err = db.View(func(tx *bolt.Tx) error {
		routes := tx.Bucket([]byte("channel_routes"))
		if routes == nil || tx.Bucket([]byte("sessions")) == nil {
			return nil
		}
		return routes.ForEach(func(_, id []byte) error {
			s, err := sessionFrom(tx, string(id))
			if err != nil {
				return err
			}
			channel, exists := bindings[s.ServiceID]
			if !exists || s.ChannelProvider != channel.Provider {
				return nil
			}
			channel.ChatID, channel.ChatType = s.ChannelChatID, s.ChannelChatType
			if string(routes.Get([]byte(channelRouteKey(channel)))) == s.ID {
				out = append(out, s)
			}
			return nil
		})
	})
	return out, err
}

func (a *Agent) listIMSessions(in protocol.IMSessionQuery) (protocol.IMSessionPage, error) {
	a.mu.Lock()
	page := protocol.IMSessionPage{ClientID: a.Config.ID, Sessions: []protocol.IMSessionRecipient{}}
	projects := append([]protocol.Service{}, a.Config.Services...)
	channels := append([]ChannelConfig{}, a.Config.Channels...)
	a.mu.Unlock()
	if len(in.Search) > 512 || len(in.ServiceID) > 160 {
		return page, errors.New("IM 会话查询参数过长")
	}
	search := strings.ToLower(strings.TrimSpace(in.Search))
	projectNames := map[string]string{}
	for _, project := range projects {
		projectNames[project.ID] = project.Name
	}
	all := []protocol.IMSessionRecipient{}
	seen := map[string]bool{}
	conversations, err := a.historyStore().readIMRoutes(channels)
	if err != nil {
		return page, err
	}
	for _, channel := range channels {
		name, exists := projectNames[channel.ServiceID]
		if !exists || in.ServiceID != "" && channel.ServiceID != in.ServiceID {
			continue
		}
		for _, session := range conversations {
			if session.ServiceID != channel.ServiceID || seen[session.ID] || session.DeletedAt != 0 || session.ArchivedAt != 0 || session.Kind != "im" || session.ChannelChatID == "" || session.ChannelChatType != "user" && session.ChannelChatType != "group" {
				continue
			}
			seen[session.ID] = true
			peer := soBound(strings.TrimSpace(session.ChannelPeerName), 160)
			if peer == session.ChannelChatID {
				peer = ""
			}
			channelName := channel.Name
			if channelName == "" {
				channelName = session.ChannelName
			}
			channelName = soBound(channelName, 160)
			name = soBound(name, 80)
			title := soBound(session.Title, 80)
			if !strings.Contains(strings.ToLower(strings.Join([]string{name, channelName, peer, title, channel.Provider}, "\n")), search) {
				continue
			}
			all = append(all, protocol.IMSessionRecipient{SessionID: session.ID, ProjectID: channel.ServiceID, ProjectName: name, Provider: channel.Provider, ChannelName: channelName, ChatType: session.ChannelChatType, PeerName: peer, NameKnown: peer != "", Title: title})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].SessionID < all[j].SessionID })
	if in.Cursor >= uint64(len(all)) {
		return page, nil
	}
	size := 0
	for i := in.Cursor; i < uint64(len(all)); i++ {
		data, _ := json.Marshal(all[i])
		if len(page.Sessions) > 0 && (size+len(data) > 24000 || len(page.Sessions) >= 50) {
			page.NextCursor = i
			break
		}
		page.Sessions = append(page.Sessions, all[i])
		size += len(data)
	}
	return page, nil
}

func (a *Agent) handleIMSessionList(req protocol.Request, send func(protocol.Event) error) {
	var in protocol.IMSessionQuery
	if err := json.Unmarshal(req.Body, &in); err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: "invalid IM session query"})
		return
	}
	page, err := a.listIMSessions(in)
	if err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()})
		return
	}
	_ = send(protocol.Event{ID: req.ID, Type: "result", Data: page})
}
