package store

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

var ErrDenied = errors.New("invalid or expired credentials")
var ErrExists = errors.New("account already exists")

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  []byte `json:"password,omitempty"`
	Admin     bool   `json:"admin"`
	Disabled  bool   `json:"disabled"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}
type Session struct {
	UserID    string `json:"userId"`
	ExpiresAt int64  `json:"expiresAt"`
}
type Agent struct {
	MachineID      string             `json:"machineId,omitempty"`
	Hostname       string             `json:"hostname,omitempty"`
	RepairSourceID string             `json:"repairSourceId,omitempty"`
	ID             string             `json:"id"`
	UserID         string             `json:"userId"`
	Name           string             `json:"name"`
	OS             string             `json:"os"`
	Arch           string             `json:"arch"`
	TokenHash      string             `json:"tokenHash,omitempty"`
	Revoked        bool               `json:"revoked"`
	Enrolled       bool               `json:"enrolled"`
	LastSeen       int64              `json:"lastSeen"`
	Services       []protocol.Service `json:"services"`
}
type enrollment struct {
	AgentID   string `json:"agentId"`
	ExpiresAt int64  `json:"expiresAt"`
	Recovery  bool   `json:"recovery,omitempty"`
}
type Store struct {
	db          *bolt.DB
	tokenCipher cipher.AEAD
}

func Token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func Hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"users", "accounts", "usernames", "sessions", "agents", "enrollments", "codex_tokens", "codex_verifiers", "codex_secrets", "meta", "channel_projects", "channel_bots"} {
			if _, e := tx.CreateBucketIfNotExists([]byte(name)); e != nil {
				return e
			}
		}
		return migrateAccounts(tx)
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	aead, err := tokenCipher(db, path+".token-key")
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db, tokenCipher: aead}, nil
}
func (s *Store) Close() error { return s.db.Close() }
func put(tx *bolt.Tx, bucket, key string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return tx.Bucket([]byte(bucket)).Put([]byte(key), b)
}
func get(tx *bolt.Tx, bucket, key string, v any) error {
	b := tx.Bucket([]byte(bucket)).Get([]byte(key))
	if b == nil {
		return ErrDenied
	}
	return json.Unmarshal(b, v)
}
func (s *Store) CreateUser(u User) error {
	name, key, err := NormalizeUsername(u.Username)
	if err != nil {
		return err
	}
	u.Username = name
	if u.ID == "" {
		u.ID = Token()
	}
	if u.CreatedAt == 0 {
		u.CreatedAt = time.Now().Unix()
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("usernames")).Get([]byte(key)) != nil || tx.Bucket([]byte("accounts")).Get([]byte(u.ID)) != nil {
			return ErrExists
		}
		if err := put(tx, "accounts", u.ID, u); err != nil {
			return err
		}
		return tx.Bucket([]byte("usernames")).Put([]byte(key), []byte(u.ID))
	})
}
func (s *Store) User(name string) (u User, err error) {
	_, key, e := NormalizeUsername(name)
	if e != nil {
		return u, ErrDenied
	}
	err = s.db.View(func(tx *bolt.Tx) error {
		id := tx.Bucket([]byte("usernames")).Get([]byte(key))
		if id == nil {
			return ErrDenied
		}
		return get(tx, "accounts", string(id), &u)
	})
	return
}
func (s *Store) NewSession(userID string) (string, error) {
	return s.newSession(userID, nil)
}
func (s *Store) NewSessionForUser(u User) (string, error) {
	return s.newSession(u.ID, u.Password)
}
func (s *Store) newSession(userID string, expectedPassword []byte) (string, error) {
	token := Token()
	err := s.db.Update(func(tx *bolt.Tx) error {
		var u User
		if get(tx, "accounts", userID, &u) != nil || u.Disabled || (expectedPassword != nil && !bytes.Equal(expectedPassword, u.Password)) {
			return ErrDenied
		}
		return put(tx, "sessions", Hash(token), Session{userID, time.Now().Add(30 * 24 * time.Hour).Unix()})
	})
	return token, err
}
func (s *Store) Session(token string) (v Session, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if err := get(tx, "sessions", Hash(token), &v); err != nil {
			return err
		}
		if !activeUser(tx, v.UserID) {
			return ErrDenied
		}
		return nil
	})
	if v.ExpiresAt <= time.Now().Unix() {
		err = ErrDenied
	}
	return
}
func (s *Store) Logout(token string) error {
	return s.db.Update(func(tx *bolt.Tx) error { return tx.Bucket([]byte("sessions")).Delete([]byte(Hash(token))) })
}
func (s *Store) Enroll(userID, name string, selectedOS ...string) (a Agent, token string, err error) {
	a = Agent{ID: Token(), UserID: userID, Name: name, Services: []protocol.Service{}}
	if len(selectedOS) > 0 {
		a.OS = selectedOS[0]
	}
	token = Token()
	err = s.db.Update(func(tx *bolt.Tx) error {
		if !activeUser(tx, userID) {
			return ErrDenied
		}
		if e := put(tx, "agents", a.ID, a); e != nil {
			return e
		}
		return put(tx, "enrollments", Hash(token), enrollment{AgentID: a.ID, ExpiresAt: time.Now().Add(15 * time.Minute).Unix()})
	})
	return
}
func (s *Store) Redeem(token, osName, arch string) (a Agent, credential string, err error) {
	credential = Token()
	err = s.db.Update(func(tx *bolt.Tx) error {
		var e enrollment
		if err := get(tx, "enrollments", Hash(token), &e); err != nil {
			return ErrDenied
		}
		if e.Recovery || e.ExpiresAt <= time.Now().Unix() {
			return ErrDenied
		}
		if err := get(tx, "agents", e.AgentID, &a); err != nil || a.Revoked || a.Enrolled {
			return ErrDenied
		}
		if !activeUser(tx, a.UserID) {
			return ErrDenied
		}
		if a.OS != "" && a.OS != osName {
			return ErrDenied
		}
		a.Enrolled = true
		a.OS = osName
		a.Arch = arch
		a.TokenHash = Hash(credential)
		if err := put(tx, "agents", a.ID, a); err != nil {
			return err
		}
		return tx.Bucket([]byte("enrollments")).Delete([]byte(Hash(token)))
	})
	return
}
func (s *Store) Agent(id string) (a Agent, err error) {
	err = s.db.View(func(tx *bolt.Tx) error { return get(tx, "agents", id, &a) })
	return
}
func (s *Store) AuthenticateAgent(id, credential string) (a Agent, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if credential == "" || get(tx, "agents", id, &a) != nil || a.Revoked || !a.Enrolled || !activeUser(tx, a.UserID) || subtle.ConstantTimeCompare([]byte(a.TokenHash), []byte(Hash(credential))) != 1 {
			return ErrDenied
		}
		return nil
	})
	return
}
func (s *Store) Agents(userID string) (list []Agent, err error) {
	list = []Agent{}
	err = s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("agents")).ForEach(func(_, v []byte) error {
			var a Agent
			if e := json.Unmarshal(v, &a); e != nil {
				return e
			}
			if a.UserID == userID && !a.Revoked {
				a.TokenHash = ""
				list = append(list, a)
			}
			return nil
		})
	})
	return
}
func (s *Store) Heartbeat(id string, msg protocol.Signal) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var a Agent
		if err := get(tx, "agents", id, &a); err != nil || a.Revoked || !activeUser(tx, a.UserID) {
			return ErrDenied
		}
		a.LastSeen = time.Now().Unix()
		if msg.MachineID != "" {
			if !validMachineID(msg.MachineID) || a.MachineID != "" && a.MachineID != msg.MachineID {
				return ErrDenied
			}
			a.MachineID = msg.MachineID
			a.Hostname = cleanHostname(msg.Hostname)
		}
		a.Services = msg.Services
		return put(tx, "agents", id, a)
	})
}
func (s *Store) Revoke(userID, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var a Agent
		if err := get(tx, "agents", id, &a); err != nil || a.UserID != userID {
			return ErrDenied
		}
		a.Revoked = true
		a.TokenHash = ""
		if err := put(tx, "agents", id, a); err != nil {
			return err
		}
		if err := releaseAgentChannels(tx, id); err != nil {
			return err
		}
		return revokeAgentTokens(tx, id)
	})
}

// Prune keeps expired authentication records from accumulating indefinitely.
func (s *Store) Prune() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range []string{"sessions", "enrollments"} {
			c := tx.Bucket([]byte(name)).Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				var record struct {
					ExpiresAt int64 `json:"expiresAt"`
				}
				if json.Unmarshal(v, &record) == nil && record.ExpiresAt < time.Now().Unix() {
					if err := c.Delete(); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}
