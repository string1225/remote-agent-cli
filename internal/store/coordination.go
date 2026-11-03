package store

import (
	"encoding/json"
	"errors"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

// Routing metadata only: prompts, reports and recordings stay on clients.
type WorkRoute struct {
	ID          string                `json:"id"`
	AccountID   string                `json:"accountId"`
	ReplyHandle string                `json:"replyHandle"`
	RequestHash string                `json:"requestHash"`
	Source      protocol.AssistantRef `json:"source"`
	Target      protocol.AssistantRef `json:"target"`
	Path        []string              `json:"path,omitempty"`
	CreatedAt   int64                 `json:"createdAt"`
}

func (s *Store) CoordinationRoute(id string, create *WorkRoute) (out WorkRoute, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte("coordination_routes"))
		if e != nil {
			return e
		}
		if data := b.Get([]byte(id)); data != nil {
			return json.Unmarshal(data, &out)
		}
		if create == nil {
			return errors.New("任务或回复引用不存在")
		}
		out = *create
		data, e := json.Marshal(out)
		if e != nil {
			return e
		}
		if e = b.Put([]byte(out.ReplyHandle), data); e != nil {
			return e
		}
		return b.Put([]byte(id), data)
	})
	return
}

func (s *Store) SaveCoordinationTarget(id string, target protocol.AssistantRef) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("coordination_routes"))
		if b == nil {
			return errors.New("任务不存在")
		}
		var route WorkRoute
		if json.Unmarshal(b.Get([]byte(id)), &route) != nil {
			return errors.New("任务不存在")
		}
		route.Target = target
		parents, e := tx.CreateBucketIfNotExists([]byte("coordination_parents"))
		if e != nil {
			return e
		}
		if e = parents.Put([]byte(Hash(route.AccountID+"\x00"+target.ClientID+"\x00"+target.SessionID)), []byte(id)); e != nil {
			return e
		}
		data, e := json.Marshal(route)
		if e != nil {
			return e
		}
		if e = b.Put([]byte(route.ReplyHandle), data); e != nil {
			return e
		}
		return b.Put([]byte(id), data)
	})
}

func (s *Store) CoordinationParent(account, client, session string) (id string, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("coordination_parents"))
		if b == nil {
			return errors.New("not found")
		}
		id = string(b.Get([]byte(Hash(account + "\x00" + client + "\x00" + session))))
		return nil
	})
	if err == nil && id == "" {
		err = errors.New("not found")
	}
	return
}
