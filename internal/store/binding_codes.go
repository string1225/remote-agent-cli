package store

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

type savedBindingCode struct {
	Code      string `json:"code"`
	ExpiresAt int64  `json:"expiresAt"`
}

// Reuse the existing encrypted-secret key and its backup/recovery rules. The
// namespace keeps device codes separate from personal model tokens.
func (s *Store) saveBindingCode(tx *bolt.Tx, agentID, code string, expires int64) error {
	id := "binding:" + agentID
	data, err := json.Marshal(savedBindingCode{code, expires})
	if err != nil {
		return err
	}
	nonce := make([]byte, s.tokenCipher.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	return tx.Bucket([]byte("codex_secrets")).Put([]byte(id), s.tokenCipher.Seal(nonce, nonce, data, []byte(id)))
}

// BindingCode reveals an unconsumed code only to its owner. Old hash-only or
// expired invitations are replaced atomically without creating another device.
func (s *Store) BindingCode(userID, id string, renew bool) (a Agent, code string, expires int64, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) || get(tx, "agents", id, &a) != nil || a.UserID != userID || a.Revoked || a.Enrolled {
			return ErrDenied
		}
		key := "binding:" + id
		if data := tx.Bucket([]byte("codex_secrets")).Get([]byte(key)); !renew && len(data) > s.tokenCipher.NonceSize() {
			n := s.tokenCipher.NonceSize()
			plain, e := s.tokenCipher.Open(nil, data[:n], data[n:], []byte(key))
			if e != nil {
				return e
			}
			var saved savedBindingCode
			if e = json.Unmarshal(plain, &saved); e != nil {
				return e
			}
			var invite enrollment
			if saved.ExpiresAt > time.Now().Unix() && get(tx, "enrollments", Hash(saved.Code), &invite) == nil && !invite.Recovery && invite.AgentID == id {
				code, expires = saved.Code, saved.ExpiresAt
				return nil
			}
		}
		cursor := tx.Bucket([]byte("enrollments")).Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var invite enrollment
			if json.Unmarshal(v, &invite) == nil && invite.AgentID == id && !invite.Recovery {
				if e := cursor.Delete(); e != nil {
					return e
				}
			}
		}
		code, expires = Token(), time.Now().Add(15*time.Minute).Unix()
		if e := put(tx, "enrollments", Hash(code), enrollment{AgentID: id, ExpiresAt: expires}); e != nil {
			return e
		}
		return s.saveBindingCode(tx, id, code, expires)
	})
	a.TokenHash = ""
	if err != nil {
		code = ""
	}
	return
}

// A durable client-generated credential makes a lost activation response safe
// to retry. Other callers still cannot reuse an already consumed code.
func (s *Store) RedeemActivation(code, osName, arch, credential string) (a Agent, err error) {
	raw, e := base64.RawURLEncoding.DecodeString(credential)
	if e != nil || len(raw) != 32 {
		return a, ErrDenied
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		var invite enrollment
		if get(tx, "enrollments", Hash(code), &invite) != nil || invite.Recovery || invite.ExpiresAt <= time.Now().Unix() {
			return ErrDenied
		}
		if get(tx, "agents", invite.AgentID, &a) != nil || a.Revoked || !activeUser(tx, a.UserID) || a.OS != "" && a.OS != osName {
			return ErrDenied
		}
		hash := Hash(credential)
		if a.Enrolled {
			if invite.ActivationHash != hash || a.TokenHash != hash || a.Arch != arch {
				return ErrDenied
			}
			return nil
		}
		a.Enrolled, a.OS, a.Arch, a.TokenHash = true, osName, arch, hash
		if e := put(tx, "agents", a.ID, a); e != nil {
			return e
		}
		invite.ActivationHash = hash
		if e := put(tx, "enrollments", Hash(code), invite); e != nil {
			return e
		}
		return tx.Bucket([]byte("codex_secrets")).Delete([]byte("binding:" + a.ID))
	})
	return
}
