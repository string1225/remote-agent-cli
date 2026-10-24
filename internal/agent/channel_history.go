package agent

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type channelDelivery struct {
	ID, Destination, SessionID, Text string
	Sequence                         uint64
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
func (h historyStore) channelQueue(c ChannelConfig, id, text string, record bool) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("channel_outbox"))
		if err != nil {
			return err
		}
		if b.Stats().KeyN >= 1000 {
			return errors.New("通知队列已满，请检查渠道连接")
		}
		if b.Get([]byte(id)) != nil {
			return nil
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
		data, err := json.Marshal(channelDelivery{ID: id, Destination: c.destination(), SessionID: c.SessionID, Text: text, Sequence: sequence})
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}
func (h historyStore) channelPending(c ChannelConfig) (out []channelDelivery, err error) {
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
			if item.Destination == c.destination() {
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
