package store

import (
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrProjectSessionToken = errors.New("token must contain 32–256 ASCII letters, digits, hyphens or underscores")

type ProjectSessionToken struct {
	UserID     string `json:"-"`
	Configured bool   `json:"configured"`
	UpdatedAt  int64  `json:"updatedAt,omitempty"`
}

type projectSessionCredential struct {
	Hash      string `json:"hash"`
	UpdatedAt int64  `json:"updatedAt"`
}

func projectSessionKey(agentID, projectID string) string { return agentID + "\x00" + projectID }
func projectSessionSecretKey(agentID, projectID string) string {
	return "project-session:" + projectSessionKey(agentID, projectID)
}

func ownedSessionProject(tx *bolt.Tx, userID, agentID, projectID string) (Agent, error) {
	var a Agent
	if agentID == "" || projectID == "" || strings.ContainsRune(agentID+projectID, 0) || get(tx, "agents", agentID, &a) != nil || a.Revoked || !a.Enrolled || a.UserID != userID || !activeUser(tx, userID) {
		return a, ErrDenied
	}
	for _, project := range a.Services {
		if project.ID == projectID && project.Scope == "" {
			return a, nil
		}
	}
	return a, ErrDenied
}

func (s *Store) ProjectSessionToken(userID, agentID, projectID string) (info ProjectSessionToken, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if _, err := ownedSessionProject(tx, userID, agentID, projectID); err != nil {
			return err
		}
		raw := tx.Bucket([]byte("project_session_tokens")).Get([]byte(projectSessionKey(agentID, projectID)))
		if raw == nil {
			return nil
		}
		var record projectSessionCredential
		if err := get(tx, "project_session_tokens", projectSessionKey(agentID, projectID), &record); err != nil {
			return err
		}
		info = ProjectSessionToken{UserID: userID, Configured: true, UpdatedAt: record.UpdatedAt}
		return nil
	})
	return
}

func (s *Store) SetProjectSessionToken(userID, agentID, projectID, secret string) (info ProjectSessionToken, token string, err error) {
	if secret == "" {
		secret = "ra_project_" + Token()
	}
	if len(secret) < 32 || len(secret) > 256 || strings.IndexFunc(secret, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_')
	}) >= 0 {
		return info, "", ErrProjectSessionToken
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		if _, err := ownedSessionProject(tx, userID, agentID, projectID); err != nil {
			return err
		}
		now := time.Now().Unix()
		if err := s.sealToken(tx, projectSessionSecretKey(agentID, projectID), secret); err != nil {
			return err
		}
		if err := put(tx, "project_session_tokens", projectSessionKey(agentID, projectID), projectSessionCredential{Hash: Hash(secret), UpdatedAt: now}); err != nil {
			return err
		}
		info = ProjectSessionToken{UserID: userID, Configured: true, UpdatedAt: now}
		return nil
	})
	if err == nil {
		token = secret
	}
	return
}

func (s *Store) RevealProjectSessionToken(userID, agentID, projectID string) (secret string, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if _, err := ownedSessionProject(tx, userID, agentID, projectID); err != nil {
			return err
		}
		var record projectSessionCredential
		if err := get(tx, "project_session_tokens", projectSessionKey(agentID, projectID), &record); err != nil {
			return err
		}
		key := projectSessionSecretKey(agentID, projectID)
		data := tx.Bucket([]byte("codex_secrets")).Get([]byte(key))
		n := s.tokenCipher.NonceSize()
		if len(data) < n {
			return ErrTokenNotSaved
		}
		plain, err := s.tokenCipher.Open(nil, data[:n], data[n:], []byte(key))
		if err != nil {
			return err
		}
		secret = string(plain)
		return nil
	})
	return
}

func (s *Store) DeleteProjectSessionToken(userID, agentID, projectID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := ownedSessionProject(tx, userID, agentID, projectID); err != nil {
			return err
		}
		if err := tx.Bucket([]byte("project_session_tokens")).Delete([]byte(projectSessionKey(agentID, projectID))); err != nil {
			return err
		}
		return tx.Bucket([]byte("codex_secrets")).Delete([]byte(projectSessionSecretKey(agentID, projectID)))
	})
}

func (s *Store) AuthenticateProjectSessionToken(agentID, projectID, secret string) (info ProjectSessionToken, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		var a Agent
		if secret == "" || len(secret) > 256 || get(tx, "agents", agentID, &a) != nil {
			return ErrDenied
		}
		if _, err := ownedSessionProject(tx, a.UserID, agentID, projectID); err != nil {
			return ErrDenied
		}
		var record projectSessionCredential
		if get(tx, "project_session_tokens", projectSessionKey(agentID, projectID), &record) != nil || subtle.ConstantTimeCompare([]byte(record.Hash), []byte(Hash(secret))) != 1 {
			return ErrDenied
		}
		info = ProjectSessionToken{UserID: a.UserID, Configured: true, UpdatedAt: record.UpdatedAt}
		return nil
	})
	return
}
