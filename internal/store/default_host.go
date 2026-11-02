package store

import (
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

var ErrDefaultHost = errors.New("请先迁移默认主机上的待办清单，再删除该客户端")

type DefaultHost struct {
	AgentID     string `json:"agentId"`
	Generation  string `json:"generation"`
	MigrationID string `json:"migrationId,omitempty"`
}

// Use metadata, rather than embedding this in User, so account edits cannot
// overwrite the host pointer and account credentials never leave the server.
func (s *Store) DefaultHost(userID string) (host DefaultHost, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) {
			return ErrDenied
		}
		if raw := tx.Bucket([]byte("meta")).Get([]byte("todo-host:" + userID)); raw != nil {
			return json.Unmarshal(raw, &host)
		}
		return nil
	})
	return
}

func (s *Store) SetDefaultHost(userID string, before, next DefaultHost) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) {
			return ErrDenied
		}
		var a Agent
		if get(tx, "agents", next.AgentID, &a) != nil || a.UserID != userID || a.Revoked || !a.Enrolled || next.Generation == "" {
			return ErrDenied
		}
		var current DefaultHost
		if raw := tx.Bucket([]byte("meta")).Get([]byte("todo-host:" + userID)); raw != nil {
			if err := json.Unmarshal(raw, &current); err != nil {
				return err
			}
		}
		if current != before {
			return errors.New("默认主机已改变，请刷新重试")
		}
		// Generation must be persisted, though it is not exposed in public JSON.
		return putDefaultHost(tx, userID, next)
	})
}

func putDefaultHost(tx *bolt.Tx, userID string, host DefaultHost) error {
	data, err := json.Marshal(host)
	if err != nil {
		return err
	}
	return tx.Bucket([]byte("meta")).Put([]byte("todo-host:"+userID), data)
}
