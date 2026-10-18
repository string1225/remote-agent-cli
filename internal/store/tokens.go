package store

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"

	bolt "go.etcd.io/bbolt"
)

type CodexToken struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	CreatedAt  int64  `json:"createdAt"`
	LastUsedAt int64  `json:"lastUsedAt"`
	Revoked    bool   `json:"revoked"`
	AgentID    string `json:"agentId,omitempty"`
}

func revokeAgentTokens(tx *bolt.Tx, agentID string) error {
	var revoked []CodexToken
	err := tx.Bucket([]byte("codex_tokens")).ForEach(func(_, v []byte) error {
		var token CodexToken
		if err := json.Unmarshal(v, &token); err != nil {
			return err
		}
		if token.AgentID == agentID && !token.Revoked {
			token.Revoked = true
			revoked = append(revoked, token)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, token := range revoked {
		if err := put(tx, "codex_tokens", token.ID, token); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) NewCodexToken(userID, name, agentID string) (info CodexToken, secret string, err error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 80 || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return info, "", ErrInvalid
	}
	secret = "ra_codex_" + Token()
	info = CodexToken{ID: Token(), UserID: userID, Name: name, Prefix: secret[:17], CreatedAt: time.Now().Unix(), AgentID: agentID}
	err = s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) {
			return ErrDenied
		}
		if agentID != "" {
			var a Agent
			if get(tx, "agents", agentID, &a) != nil || a.UserID != userID || a.Revoked || !a.Enrolled {
				return ErrDenied
			}
			// Setup may fail before the new credential is written locally. Keep
			// previous tokens usable until explicit token or device revocation.
		}
		if err := put(tx, "codex_tokens", info.ID, info); err != nil {
			return err
		}
		return tx.Bucket([]byte("codex_verifiers")).Put([]byte(Hash(secret)), []byte(info.ID))
	})
	if err != nil {
		secret = ""
	}
	return
}

func (s *Store) CodexTokens(userID string) (list []CodexToken, err error) {
	list = []CodexToken{}
	err = s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("codex_tokens")).ForEach(func(_, v []byte) error {
			var info CodexToken
			if err := json.Unmarshal(v, &info); err != nil {
				return err
			}
			if userID == "" || info.UserID == userID {
				list = append(list, info)
			}
			return nil
		})
	})
	return
}

func (s *Store) RevokeCodexToken(actorID, id string, admin bool) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var actor User
		if get(tx, "accounts", actorID, &actor) != nil || actor.Disabled {
			return ErrDenied
		}
		var info CodexToken
		if get(tx, "codex_tokens", id, &info) != nil {
			return ErrDenied
		}
		if admin {
			if !actor.Admin {
				return ErrDenied
			}
		} else if info.UserID != actorID {
			return ErrDenied
		}
		info.Revoked = true
		return put(tx, "codex_tokens", id, info)
	})
}

func (s *Store) AuthenticateCodexToken(secret string) (info CodexToken, err error) {
	if len(secret) != 52 || !strings.HasPrefix(secret, "ra_codex_") {
		return info, ErrDenied
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		id := tx.Bucket([]byte("codex_verifiers")).Get([]byte(Hash(secret)))
		if id == nil || get(tx, "codex_tokens", string(id), &info) != nil || info.Revoked || !activeUser(tx, info.UserID) {
			return ErrDenied
		}
		if info.AgentID != "" {
			var a Agent
			if get(tx, "agents", info.AgentID, &a) != nil || a.Revoked || !a.Enrolled || a.UserID != info.UserID {
				return ErrDenied
			}
		}
		now := time.Now().Unix()
		if info.LastUsedAt != now {
			info.LastUsedAt = now
			return put(tx, "codex_tokens", info.ID, info)
		}
		return nil
	})
	return
}
