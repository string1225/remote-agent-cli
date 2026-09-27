package agent

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type historyStore struct{ path string }

func (a *Agent) historyStore() historyStore {
	return historyStore{filepath.Join(filepath.Dir(a.Config.Path), "history.db")}
}

func (h historyStore) transact(write bool, fn func(*bolt.Tx) error) error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0700); err != nil {
		return err
	}
	db, err := bolt.Open(h.path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"sessions", "entries"} {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if write {
		return db.Update(fn)
	}
	return db.View(fn)
}

func sessionFrom(tx *bolt.Tx, id string) (s protocol.Conversation, err error) {
	b := tx.Bucket([]byte("sessions")).Get([]byte(id))
	if b == nil {
		return s, errors.New("session not found")
	}
	err = json.Unmarshal(b, &s)
	return
}
func saveSession(tx *bolt.Tx, s protocol.Conversation) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return tx.Bucket([]byte("sessions")).Put([]byte(s.ID), b)
}
func (h historyStore) create(service protocol.Service, title string) (s protocol.Conversation, err error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "新会话"
	}
	if len([]rune(title)) > 80 {
		title = string([]rune(title)[:80])
	}
	now := time.Now().UnixMilli()
	s = protocol.Conversation{ID: store.Token(), ServiceID: service.ID, Title: title, Provider: service.Provider, Workspace: service.Workspace, CreatedAt: now, UpdatedAt: now}
	err = h.transact(true, func(tx *bolt.Tx) error {
		if err := saveSession(tx, s); err != nil {
			return err
		}
		_, err := tx.Bucket([]byte("entries")).CreateBucket([]byte(s.ID))
		return err
	})
	return
}
func (h historyStore) get(id string) (s protocol.Conversation, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error { var e error; s, e = sessionFrom(tx, id); return e })
	return
}
func (h historyStore) list(serviceID, search string, offset uint64) (page protocol.ConversationPage, err error) {
	page.Sessions = []protocol.Conversation{}
	err = h.transact(false, func(tx *bolt.Tx) error {
		all := []protocol.Conversation{}
		if err := tx.Bucket([]byte("sessions")).ForEach(func(_, v []byte) error {
			var s protocol.Conversation
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}
			if s.ServiceID == serviceID && strings.Contains(strings.ToLower(s.Title), strings.ToLower(search)) {
				s.NativeID = ""
				all = append(all, s)
			}
			return nil
		}); err != nil {
			return err
		}
		sort.Slice(all, func(i, j int) bool {
			if all[i].UpdatedAt == all[j].UpdatedAt {
				return all[i].ID < all[j].ID
			}
			return all[i].UpdatedAt > all[j].UpdatedAt
		})
		size := 0
		for i := offset; i < uint64(len(all)); i++ {
			b, _ := json.Marshal(all[i])
			if len(page.Sessions) > 0 && (size+len(b) > 24000 || len(page.Sessions) >= 50) {
				page.NextCursor = i
				break
			}
			page.Sessions = append(page.Sessions, all[i])
			size += len(b)
		}
		return nil
	})
	return
}
func (h historyStore) read(id string, after uint64) (page protocol.HistoryPage, err error) {
	page.Entries = []protocol.ChatEntry{}
	err = h.transact(false, func(tx *bolt.Tx) error {
		var err error
		page.Session, err = sessionFrom(tx, id)
		if err != nil {
			return err
		}
		page.Session.NativeID = ""
		bucket := tx.Bucket([]byte("entries")).Bucket([]byte(id))
		if bucket == nil {
			return errors.New("session history missing")
		}
		if after == ^uint64(0) {
			return nil
		}
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, after+1)
		cursor := bucket.Cursor()
		size := 0
		for k, v := cursor.Seek(key); k != nil; k, v = cursor.Next() {
			if len(page.Entries) > 0 && size+len(v) > 24000 {
				page.NextCursor = page.Entries[len(page.Entries)-1].Seq
				break
			}
			var entry protocol.ChatEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return err
			}
			page.Entries = append(page.Entries, entry)
			size += len(v)
		}
		return nil
	})
	return
}
func (h historyStore) append(id string, entry protocol.ChatEntry) (entries []protocol.ChatEntry, err error) {
	err = h.transact(true, func(tx *bolt.Tx) error {
		s, err := sessionFrom(tx, id)
		if err != nil {
			return err
		}
		bucket := tx.Bucket([]byte("entries")).Bucket([]byte(id))
		if bucket == nil {
			return errors.New("session history missing")
		}
		if bucket.Sequence() > 100000 {
			return errors.New("session history is full; start a new session")
		}
		if entry.Role == "user" && s.Title == "新会话" {
			r := []rune(strings.TrimSpace(entry.Text))
			if len(r) > 60 {
				r = r[:60]
			}
			s.Title = string(r)
		}
		runes := []rune(entry.Text)
		for {
			part := entry
			n := min(len(runes), 2000)
			part.Text = string(runes[:n])
			runes = runes[n:]
			part.Seq, err = bucket.NextSequence()
			if err != nil {
				return err
			}
			part.CreatedAt = time.Now().UnixMilli()
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, part.Seq)
			b, err := json.Marshal(part)
			if err != nil {
				return err
			}
			if err = bucket.Put(key, b); err != nil {
				return err
			}
			entries = append(entries, part)
			if len(runes) == 0 {
				break
			}
		}
		s.UpdatedAt = time.Now().UnixMilli()
		return saveSession(tx, s)
	})
	return
}
func (h historyStore) nativeID(id, nativeID string) error {
	if len(nativeID) == 0 || len(nativeID) > 128 || strings.ContainsAny(nativeID, " \r\n\t/\\") || strings.HasPrefix(nativeID, "-") {
		return errors.New("invalid provider session ID")
	}
	return h.transact(true, func(tx *bolt.Tx) error {
		s, err := sessionFrom(tx, id)
		if err != nil {
			return err
		}
		s.NativeID = nativeID
		return saveSession(tx, s)
	})
}
