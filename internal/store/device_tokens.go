package store

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	bolt "go.etcd.io/bbolt"
)

// EnsureDeviceCodexToken reuses recoverable active credentials atomically. Old
// hash-only tokens stay valid; they cannot be recovered or silently rotated.
func (s *Store) EnsureDeviceCodexToken(userID, name, agentID string) (info CodexToken, secret string, err error) {
	name = strings.TrimSpace(name)
	if agentID == "" || name == "" || len([]rune(name)) > 80 || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return info, "", ErrInvalid
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		var agent Agent
		if !activeUser(tx, userID) || get(tx, "agents", agentID, &agent) != nil || agent.UserID != userID || agent.Revoked || !agent.Enrolled {
			return ErrDenied
		}
		if err := tx.Bucket([]byte("codex_tokens")).ForEach(func(_, v []byte) error {
			var candidate CodexToken
			if err := json.Unmarshal(v, &candidate); err != nil {
				return err
			}
			if candidate.UserID != userID || candidate.AgentID != agentID || candidate.Revoked || info.ID != "" {
				return nil
			}
			data := tx.Bucket([]byte("codex_secrets")).Get([]byte(candidate.ID))
			if data == nil {
				return nil
			}
			n := s.tokenCipher.NonceSize()
			if len(data) < n {
				return errors.New("invalid saved device token")
			}
			plain, err := s.tokenCipher.Open(nil, data[:n], data[n:], []byte(candidate.ID))
			if err != nil {
				return err
			}
			info, secret = candidate, string(plain)
			return nil
		}); err != nil {
			return err
		}
		if info.ID != "" {
			return nil
		}
		var err error
		info, secret, err = s.newCodexToken(tx, userID, name, agentID)
		return err
	})
	if err != nil {
		secret = ""
	}
	return
}
