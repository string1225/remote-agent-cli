package store

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	bolt "go.etcd.io/bbolt"
)

type ChannelBinding struct {
	AgentID     string `json:"agentId"`
	ServiceID   string `json:"serviceId"`
	Fingerprint string `json:"fingerprint"`
}

func channelKey(agentID, serviceID string) string { return agentID + "\x00" + serviceID }
func (s *Store) BindChannel(userID string, in ChannelBinding, finalize bool) error {
	return s.db.Update(func(tx *bolt.Tx) error { return bindChannel(tx, userID, in, finalize) })
}
func bindChannel(tx *bolt.Tx, userID string, in ChannelBinding, finalize bool) error {
	var a Agent
	if get(tx, "agents", in.AgentID, &a) != nil || a.Revoked || a.UserID != userID || !activeUser(tx, userID) {
		return ErrDenied
	}
	if in.ServiceID == "" || len(in.ServiceID) > 100 || strings.ContainsRune(in.ServiceID, 0) {
		return errors.New("invalid project")
	}
	key := channelKey(in.AgentID, in.ServiceID)
	projects, bots := tx.Bucket([]byte("channel_projects")), tx.Bucket([]byte("channel_bots"))
	if projects.Get([]byte(key)) == nil {
		count := 0
		_ = projects.ForEach(func(k, _ []byte) error {
			if strings.HasPrefix(string(k), in.AgentID+"\x00") {
				count++
			}
			return nil
		})
		if count >= 100 {
			return errors.New("100 project binding limit")
		}
	}
	if in.Fingerprint != "" {
		decoded, err := hex.DecodeString(in.Fingerprint)
		if err != nil || len(decoded) != 32 {
			return errors.New("invalid robot fingerprint")
		}
		var old ChannelBinding
		if err = get(tx, "channel_bots", in.Fingerprint, &old); err == nil && (old.AgentID != in.AgentID || old.ServiceID != in.ServiceID) {
			return errors.New("机器人已经绑定其他项目，请先解绑")
		}
	}
	// Reserve the new robot while retaining the previous reservation. Finalize
	// only after the client has atomically saved its configuration and stopped
	// its previous stream. A failed save can restore the old reservation safely.
	if finalize {
		var remove [][]byte
		err := bots.ForEach(func(k, v []byte) error {
			var old ChannelBinding
			if err := json.Unmarshal(v, &old); err != nil {
				return err
			}
			if old.AgentID == in.AgentID && old.ServiceID == in.ServiceID && string(k) != in.Fingerprint {
				remove = append(remove, append([]byte{}, k...))
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, k := range remove {
			if err := bots.Delete(k); err != nil {
				return err
			}
		}
	}
	if in.Fingerprint == "" {
		return projects.Delete([]byte(key))
	}
	if err := projects.Put([]byte(key), []byte(in.Fingerprint)); err != nil {
		return err
	}
	return put(tx, "channel_bots", in.Fingerprint, in)
}
func releaseAgentChannels(tx *bolt.Tx, id string) error {
	for _, name := range []string{"channel_projects", "channel_bots"} {
		b := tx.Bucket([]byte(name))
		var remove [][]byte
		err := b.ForEach(func(k, v []byte) error {
			match := strings.HasPrefix(string(k), id+"\x00")
			if name == "channel_bots" {
				var binding ChannelBinding
				if err := json.Unmarshal(v, &binding); err != nil {
					return err
				}
				match = binding.AgentID == id
			}
			if match {
				remove = append(remove, append([]byte{}, k...))
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, k := range remove {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Store) SyncChannels(userID, agentID string, bindings []ChannelBinding) error {
	if len(bindings) > 100 {
		return errors.New("100 channel limit")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		var a Agent
		if get(tx, "agents", agentID, &a) != nil || a.UserID != userID || a.Revoked || !activeUser(tx, userID) {
			return ErrDenied
		}
		if err := releaseAgentChannels(tx, agentID); err != nil {
			return err
		}
		for _, binding := range bindings {
			binding.AgentID = agentID
			if err := bindChannel(tx, userID, binding, true); err != nil {
				return err
			}
		}
		return nil
	})
}
