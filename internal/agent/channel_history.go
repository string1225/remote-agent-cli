package agent

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type channelDelivery struct {
	ID, Destination, SessionID, Text         string
	Sequence                                 uint64
	Fingerprint, ServiceID, ChatID, ChatType string
	MessageID, ReactionID                    string
	Delivered                                bool
}
type channelReaction struct {
	Fingerprint, ServiceID, MessageID, ReactionID string
	Pending                                       bool
}

func channelReactionKey(c ChannelConfig, messageID string) []byte {
	return []byte(store.Hash(c.fingerprint() + "\x00" + c.ServiceID + "\x00" + messageID))
}
func (h historyStore) channelTrackReaction(c ChannelConfig, messageID, reactionID string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("channel_reactions"))
		if err != nil {
			return err
		}
		if b.Stats().KeyN >= 1000 {
			return errors.New("IM 表情清理队列已满")
		}
		data, err := json.Marshal(channelReaction{Fingerprint: c.fingerprint(), ServiceID: c.ServiceID, MessageID: messageID, ReactionID: reactionID})
		if err != nil {
			return err
		}
		return b.Put(channelReactionKey(c, messageID), data)
	})
}
func (h historyStore) channelRetireReaction(c ChannelConfig, messageID string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_reactions"))
		if b == nil {
			return nil
		}
		key := channelReactionKey(c, messageID)
		data := b.Get(key)
		if data == nil {
			return nil
		}
		var item channelReaction
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		item.Pending = true
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return b.Put(key, data)
	})
}
func (h historyStore) channelPendingReactions(c ChannelConfig, all bool) (items []channelReaction, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_reactions"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, data []byte) error {
			var item channelReaction
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}
			if item.Fingerprint == c.fingerprint() && item.ServiceID == c.ServiceID && (all || item.Pending) {
				items = append(items, item)
			}
			return nil
		})
	})
	return
}
func (h historyStore) channelReactionCleared(c ChannelConfig, messageID, reactionID string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte("channel_reactions")); b != nil {
			if err := b.Delete(channelReactionKey(c, messageID)); err != nil {
				return err
			}
		}
		b := tx.Bucket([]byte("channel_outbox"))
		if b == nil {
			return nil
		}
		var changed []channelDelivery
		if err := b.ForEach(func(_, data []byte) error {
			var item channelDelivery
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}
			if item.Fingerprint == c.fingerprint() && item.ServiceID == c.ServiceID && item.MessageID == messageID && item.ReactionID == reactionID {
				item.MessageID, item.ReactionID = "", ""
				changed = append(changed, item)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, item := range changed {
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if err = b.Put([]byte(item.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func channelRouteKey(c ChannelConfig) string {
	return store.Hash(c.fingerprint() + "\x00" + c.ServiceID + "\x00" + c.ChatType + "\x00" + c.ChatID)
}
func (c ChannelConfig) validLegacyTarget() bool {
	return c.ChatID != "" && (c.ChatType == "user" || c.ChatType == "group") && (c.Provider != "feishu" || strings.HasPrefix(c.ChatID, "oc_"))
}

// Capture the old fixed destination before replacing the client configuration.
// Legacy outbox rows contained only a destination hash, so retaining their
// actual target now prevents a later default-target edit from hiding them.
func (h historyStore) channelMigrateLegacy(c ChannelConfig, project protocol.Service) error {
	if !c.validLegacyTarget() || c.SessionID == "" {
		return nil
	}
	if _, _, err := h.channelSession(c, project, channelMessage{ChatID: c.ChatID, ChatType: c.ChatType}); err != nil {
		return err
	}
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_outbox"))
		if b == nil {
			return nil
		}
		var changed []channelDelivery
		if err := b.ForEach(func(_, data []byte) error {
			var item channelDelivery
			if err := json.Unmarshal(data, &item); err != nil {
				return err
			}
			if item.Fingerprint == "" && item.Destination == c.destination() {
				item.Fingerprint, item.ServiceID, item.ChatID, item.ChatType = c.fingerprint(), c.ServiceID, c.ChatID, c.ChatType
				changed = append(changed, item)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, item := range changed {
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if err = b.Put([]byte(item.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

// Route identities are separate from the old fixed session. An upgrade never
// moves existing entries or guesses which new peer owns an old conversation.
func (h historyStore) channelSession(c ChannelConfig, project protocol.Service, in channelMessage) (session protocol.Conversation, route ChannelConfig, err error) {
	route = c
	if in.ChatID != "" {
		route.ChatID, route.ChatType = in.ChatID, in.ChatType
	}
	if route.ChatID == "" || route.ChatType != "user" && route.ChatType != "group" {
		return session, route, errors.New("IM 会话身份无效")
	}
	err = h.transact(true, func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte("channel_routes"))
		if e != nil {
			return e
		}
		key := []byte(channelRouteKey(route))
		if id := b.Get(key); id != nil {
			session, e = sessionFrom(tx, string(id))
			if e != nil {
				return e
			}
		} else {
			// Only an exact old destination can retain the old native context.
			reuseLegacy := false
			if c.SessionID != "" && c.ChatID == route.ChatID && c.ChatType == route.ChatType {
				session, e = sessionFrom(tx, c.SessionID)
				if e != nil {
					return e
				}
				reuseLegacy = session.ChannelChatID == "" && session.ServiceID == project.ID
			}
			if !reuseLegacy {
				now := time.Now().UnixMilli()
				session = protocol.Conversation{ID: store.Token(), ServiceID: project.ID, Provider: project.Provider, Workspace: project.Workspace, CreatedAt: now, UpdatedAt: now}
				if _, e = tx.Bucket([]byte("entries")).CreateBucket([]byte(session.ID)); e != nil {
					return e
				}
			}
			if e = b.Put(key, []byte(session.ID)); e != nil {
				return e
			}
		}
		session.Kind, session.Source = "im", "im"
		session.ChannelProvider, session.ChannelChatID, session.ChannelChatType = c.Provider, route.ChatID, route.ChatType
		session.ChannelName = c.Name
		if session.ChannelName == "" {
			session.ChannelName = in.Name
		}
		if session.ChannelName == "" {
			session.ChannelName = c.Provider
		}
		if in.PeerName != "" {
			session.ChannelPeerName = soBound(strings.TrimSpace(in.PeerName), 160)
		}
		if session.ChannelPeerName == "" {
			session.ChannelPeerName = route.ChatID
		}
		session.Title = soBound("IM · "+session.ChannelName+" · "+session.ChannelPeerName, 80)
		route.SessionID = session.ID
		return saveSession(tx, session)
	})
	return
}

func (h historyStore) channelConversations(c ChannelConfig) (out []protocol.Conversation, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_routes"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, id []byte) error {
			s, e := sessionFrom(tx, string(id))
			if e != nil {
				return e
			}
			candidate := c
			candidate.ChatID, candidate.ChatType = s.ChannelChatID, s.ChannelChatType
			if s.ServiceID == c.ServiceID && s.ChannelProvider == c.Provider && string(b.Get([]byte(channelRouteKey(candidate)))) == s.ID {
				s.NativeID = ""
				out = append(out, s)
			}
			return nil
		})
	})
	return
}

func (h historyStore) channelTarget(c ChannelConfig, sessionID string) (ChannelConfig, error) {
	all, err := h.channelConversations(c)
	if err != nil {
		return c, err
	}
	for _, s := range all {
		if s.ID == sessionID {
			c.ChatID, c.ChatType, c.SessionID = s.ChannelChatID, s.ChannelChatType, s.ID
			return c, nil
		}
	}
	if sessionID != "" {
		if sessionID == c.SessionID && c.validLegacyTarget() {
			return c, nil
		}
		return c, errors.New("请选择此渠道已收到消息的 IM 会话")
	}
	if c.validLegacyTarget() {
		return c, nil
	}
	if len(all) == 1 {
		s := all[0]
		c.ChatID, c.ChatType, c.SessionID = s.ChannelChatID, s.ChannelChatType, s.ID
		return c, nil
	}
	if len(all) > 1 {
		return c, errors.New("项目有多个 IM 会话，请指定通知会话")
	}
	return c, errors.New("等待第一条 IM 来信后可发送通知")
}

func (c ChannelConfig) destination() string {
	return store.Hash(c.fingerprint() + "\x00" + c.ChatType + "\x00" + c.ChatID + "\x00" + c.SessionID)
}
func (h historyStore) channelIncoming(sessionID, id, text string) (fresh bool, err error) {
	err = h.transact(true, func(tx *bolt.Tx) error {
		seen, e := tx.CreateBucketIfNotExists([]byte("channel_seen"))
		if e != nil {
			return e
		}
		if seen.Get([]byte(id)) != nil {
			return nil
		}
		order, e := tx.CreateBucketIfNotExists([]byte("channel_seen_order"))
		if e != nil {
			return e
		}
		if _, e = appendHistory(tx, sessionID, protocol.ChatEntry{TurnID: id, Role: "user", Text: text, Kind: "message"}); e != nil {
			return e
		}
		if e = seen.Put([]byte(id), []byte{1}); e != nil {
			return e
		}
		seq, e := order.NextSequence()
		if e != nil {
			return e
		}
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, seq)
		if e = order.Put(key, []byte(id)); e != nil {
			return e
		}
		if order.Stats().KeyN > 20000 {
			c := order.Cursor()
			k, v := c.First()
			if k != nil {
				if e = seen.Delete(v); e != nil {
					return e
				}
				if e = c.Delete(); e != nil {
					return e
				}
			}
		}
		fresh = true
		return nil
	})
	return
}
func (h historyStore) channelNotification(c ChannelConfig, id, text string) error {
	return h.channelQueue(c, id, text, true)
}

// Accept all parts of the session's answer atomically. A full queue must not
// leave a partial answer active, and thinking ends after the last delivery.
func (h historyStore) channelReplyParts(c ChannelConfig, id string, parts []string, messageID, reactionID string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("channel_outbox"))
		if err != nil {
			return err
		}
		needed := 0
		for i := range parts {
			if b.Get([]byte(fmt.Sprintf("%s-%04d", id, i))) == nil {
				needed++
			}
		}
		if b.Stats().KeyN+needed > 1000 {
			return errors.New("通知队列已满，请检查渠道连接")
		}
		for i, text := range parts {
			partID := fmt.Sprintf("%s-%04d", id, i)
			if b.Get([]byte(partID)) != nil {
				continue
			}
			sequence, err := b.NextSequence()
			if err != nil {
				return err
			}
			item := channelDelivery{ID: partID, Destination: c.destination(), SessionID: c.SessionID, Text: text, Sequence: sequence, Fingerprint: c.fingerprint(), ServiceID: c.ServiceID, ChatID: c.ChatID, ChatType: c.ChatType}
			if i == len(parts)-1 {
				item.MessageID, item.ReactionID = messageID, reactionID
			}
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if err = b.Put([]byte(partID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h historyStore) channelQueue(c ChannelConfig, id, text string, record bool) error {
	return h.channelQueueReply(c, id, text, record, "", "")
}
func (h historyStore) channelQueueReply(c ChannelConfig, id, text string, record bool, messageID, reactionID string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("channel_outbox"))
		if err != nil {
			return err
		}
		if b.Get([]byte(id)) != nil {
			return nil
		}
		// Keep accepted IDs after delivery, including across relay retries.
		if record {
			seen, err := tx.CreateBucketIfNotExists([]byte("channel_notifications"))
			if err != nil {
				return err
			}
			key := []byte(c.destination() + "\x00" + id)
			if seen.Get(key) != nil {
				return nil
			}
			if err = seen.Put(key, []byte{1}); err != nil {
				return err
			}
		}
		if b.Stats().KeyN >= 1000 {
			return errors.New("通知队列已满，请检查渠道连接")
		}
		if record {
			if _, err = appendHistory(tx, c.SessionID, protocol.ChatEntry{TurnID: id, Role: "assistant", Text: text, Kind: "notification"}); err != nil {
				return err
			}
		}
		sequence, err := b.NextSequence()
		if err != nil {
			return err
		}
		data, err := json.Marshal(channelDelivery{ID: id, Destination: c.destination(), SessionID: c.SessionID, Text: text, Sequence: sequence, Fingerprint: c.fingerprint(), ServiceID: c.ServiceID, ChatID: c.ChatID, ChatType: c.ChatType, MessageID: messageID, ReactionID: reactionID})
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}
func (h historyStore) channelPending(c ChannelConfig) (out []channelDelivery, err error) {
	return h.channelPendingItems(c, false)
}
func (h historyStore) channelPendingItems(c ChannelConfig, all bool) (out []channelDelivery, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_outbox"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, v []byte) error {
			var item channelDelivery
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
			// Keep legacy automatic deliveries dormant even if archival failed.
			if automaticCompletionID(item.ID) {
				return nil
			}
			sameRoute := item.Destination == c.destination() && (item.Fingerprint != "" || c.validLegacyTarget())
			sameBinding := item.Fingerprint == c.fingerprint() && item.ServiceID == c.ServiceID
			if sameRoute || all && sameBinding {
				out = append(out, item)
			}
			return nil
		})
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	return
}
func (h historyStore) channelDelivered(id string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_outbox"))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(id))
	})
}
func (h historyStore) channelSent(id string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("channel_outbox"))
		if b == nil {
			return nil
		}
		data := b.Get([]byte(id))
		if data == nil {
			return nil
		}
		var item channelDelivery
		if err := json.Unmarshal(data, &item); err != nil {
			return err
		}
		item.Delivered = true
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}
func (h historyStore) turnText(sessionID, turnID string, limit int) string {
	var text strings.Builder
	for cursor := uint64(0); ; {
		page, err := h.read(sessionID, cursor)
		if err != nil {
			break
		}
		for _, entry := range page.Entries {
			if entry.TurnID == turnID && entry.Role == "assistant" && text.Len() < limit {
				text.WriteString(entry.Text)
			}
		}
		if page.NextCursor == 0 || text.Len() >= limit {
			break
		}
		cursor = page.NextCursor
	}
	return soBound(text.String(), limit)
}
