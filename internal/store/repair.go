package store

import (
	"time"

	bolt "go.etcd.io/bbolt"
)

// RepairEnrollment leaves enrolled credentials untouched. A pending device can
// receive another short-lived invite for its original identity.
func (s *Store) RepairEnrollment(userID, id string) (a Agent, token string, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) || get(tx, "agents", id, &a) != nil || a.UserID != userID || a.Revoked {
			return ErrDenied
		}
		if a.Enrolled {
			a.TokenHash = ""
			return nil
		}
		token = Token()
		return put(tx, "enrollments", Hash(token), enrollment{AgentID: a.ID, ExpiresAt: time.Now().Add(15 * time.Minute).Unix()})
	})
	return
}
