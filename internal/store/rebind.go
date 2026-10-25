package store

import (
	"encoding/base64"
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

// RebindEnrollment grants a short-lived recovery capability without changing
// the active device credential or any associated project/model credential.
func (s *Store) RebindEnrollment(userID, id string) (a Agent, token string, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) || get(tx, "agents", id, &a) != nil || a.UserID != userID || a.Revoked || !a.Enrolled {
			return ErrDenied
		}
		token = Token()
		return put(tx, "enrollments", Hash(token), enrollment{AgentID: a.ID, ExpiresAt: time.Now().Add(15 * time.Minute).Unix(), Recovery: true})
	})
	a.TokenHash = ""
	if err != nil {
		token = ""
	}
	return
}

// The client persists its proposed credential before this transaction. A lost
// response or failed local config replacement can therefore be recovered.
func (s *Store) RedeemRebind(token, expectedID, osName, arch, credential string) (a Agent, err error) {
	raw, e := base64.RawURLEncoding.DecodeString(credential)
	if expectedID == "" || e != nil || len(raw) != 32 {
		return a, ErrDenied
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		var invite enrollment
		if get(tx, "enrollments", Hash(token), &invite) != nil || !invite.Recovery || invite.AgentID != expectedID || invite.ExpiresAt <= time.Now().Unix() {
			return ErrDenied
		}
		if get(tx, "agents", expectedID, &a) != nil || a.Revoked || !a.Enrolled || !activeUser(tx, a.UserID) {
			return ErrDenied
		}
		a.TokenHash = Hash(credential)
		a.OS, a.Arch = osName, arch
		if err := put(tx, "agents", a.ID, a); err != nil {
			return err
		}
		// Every pending recovery for this identity predates this credential change.
		// Retire them together so a stale second link cannot rotate it later.
		b := tx.Bucket([]byte("enrollments"))
		cursor := b.Cursor()
		for key, data := cursor.First(); key != nil; key, data = cursor.Next() {
			var other enrollment
			if json.Unmarshal(data, &other) == nil && other.AgentID == a.ID && other.Recovery {
				if err := cursor.Delete(); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return
}
