package store

import (
	"strings"
	"unicode"
	"unicode/utf8"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) RenameAgent(userID, id, name string) error {
	return s.UpdateAgentSettings(userID, id, name, "")
}

func (s *Store) UpdateAgentSettings(userID, id, name, selectedOS string) error {
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || len([]rune(name)) < 1 || len([]rune(name)) > 80 || strings.ContainsFunc(name, unicode.IsControl) {
		return ErrInvalid
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		var a Agent
		if !activeUser(tx, userID) || get(tx, "agents", id, &a) != nil || a.UserID != userID || a.Revoked {
			return ErrDenied
		}
		a.Name = name
		if selectedOS != "" {
			if !ValidDeviceOS(selectedOS) || a.OS != "" && a.OS != selectedOS {
				return ErrInvalid
			}
			a.OS = selectedOS
		}
		return put(tx, "agents", id, a)
	})
}

func ValidDeviceOS(value string) bool {
	return value == "windows" || value == "darwin" || value == "linux"
}
func (s *Store) EnrollForOS(userID, name, osName string) (Agent, string, error) {
	if !ValidDeviceOS(osName) {
		return Agent{}, "", ErrInvalid
	}
	return s.Enroll(userID, name, osName)
}
