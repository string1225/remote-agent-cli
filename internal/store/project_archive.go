package store

import (
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

// SetProjectArchived changes only the account's working set. Local registration,
// task execution, channel bindings, native sessions and recordings are untouched.
// It also works offline and requires no client upgrade.
func (s *Store) SetProjectArchived(userID, agentID, projectID string, archived bool) (result protocol.Service, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		var a Agent
		if get(tx, "agents", agentID, &a) != nil || a.UserID != userID || a.Revoked || !activeUser(tx, userID) {
			return ErrDenied
		}
		for i, project := range a.Services {
			if project.ID != projectID || project.Scope != "" {
				continue
			}
			if archived && project.ArchivedAt == 0 {
				project.ArchivedAt = time.Now().UnixMilli()
			} else if !archived {
				project.ArchivedAt = 0
			}
			a.Services[i] = project
			result = project
			return put(tx, "agents", agentID, a)
		}
		return ErrDenied
	})
	return
}
