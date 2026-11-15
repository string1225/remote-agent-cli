package store

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

// EnsureEndpointGuideToken issues a separate personal token for distribution.
// Administrators can only recover tokens issued for this purpose, never an
// arbitrary existing personal/device token. Retries and concurrent copies reuse
// this token until its owner or an administrator revokes it.
func (s *Store) EnsureEndpointGuideToken(actorID, userID string) (info CodexToken, secret string, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		var actor User
		if get(tx, "accounts", actorID, &actor) != nil || actor.Disabled || !actor.Admin || !activeUser(tx, userID) {
			return ErrDenied
		}
		err := tx.Bucket([]byte("codex_tokens")).ForEach(func(_, v []byte) error {
			var token CodexToken
			if e := json.Unmarshal(v, &token); e != nil {
				return e
			}
			if token.UserID == userID && token.Purpose == "endpoint-guide" && !token.Revoked {
				info = token
			}
			return nil
		})
		if err != nil {
			return err
		}
		if info.ID != "" {
			secret, err = s.openToken(tx, info.ID)
			return err
		}
		info, secret, err = s.newCodexToken(tx, userID, "Endpoint / Codex 安装指引", "")
		if err != nil {
			return err
		}
		info.Purpose = "endpoint-guide"
		return put(tx, "codex_tokens", info.ID, info)
	})
	if err != nil {
		secret = ""
	}
	return
}
