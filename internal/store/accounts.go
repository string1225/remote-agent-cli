package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	bolt "go.etcd.io/bbolt"
)

var ErrInvalid = errors.New("invalid account data")
var ErrLastAdmin = errors.New("cannot disable or demote the last active administrator")

// NormalizeUsername preserves the display name and uses Unicode simple case folding
// for uniqueness, including non-ASCII equivalents such as sigma and Kelvin sign.
func NormalizeUsername(value string) (name, key string, err error) {
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "", "", ErrInvalid
	}
	name = strings.TrimSpace(value)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 64 {
		return "", "", ErrInvalid
	}
	return name, foldUsername(name), nil
}

func foldUsername(value string) string {
	var b strings.Builder
	for _, r := range value {
		folded := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < folded {
				folded = next
			}
		}
		b.WriteRune(folded)
	}
	return b.String()
}

func migrateAccounts(tx *bolt.Tx) error {
	meta := tx.Bucket([]byte("meta"))
	if meta.Get([]byte("accounts_v2")) != nil {
		return nil
	}
	err := tx.Bucket([]byte("users")).ForEach(func(_, v []byte) error {
		var u User
		if err := json.Unmarshal(v, &u); err != nil {
			return err
		}
		name, key, err := NormalizeUsername(u.Username)
		if err != nil {
			return fmt.Errorf("legacy username %q: %w", u.Username, err)
		}
		if u.ID == "" {
			return errors.New("legacy user has no immutable ID")
		}
		names := tx.Bucket([]byte("usernames"))
		if existing := names.Get([]byte(key)); existing != nil && !bytes.Equal(existing, []byte(u.ID)) {
			return fmt.Errorf("legacy usernames conflict after normalization: %q", u.Username)
		}
		if tx.Bucket([]byte("accounts")).Get([]byte(u.ID)) != nil {
			return errors.New("duplicate legacy user ID")
		}
		u.Username = name
		if err := put(tx, "accounts", u.ID, u); err != nil {
			return err
		}
		return names.Put([]byte(key), []byte(u.ID))
	})
	if err != nil {
		return err
	}
	// All records migrate in one transaction; retaining no stale password copies.
	if err := tx.DeleteBucket([]byte("users")); err != nil {
		return err
	}
	if _, err := tx.CreateBucket([]byte("users")); err != nil {
		return err
	}
	return meta.Put([]byte("accounts_v2"), []byte{1})
}

func activeUser(tx *bolt.Tx, id string) bool {
	var u User
	return get(tx, "accounts", id, &u) == nil && !u.Disabled
}

func (s *Store) UserByID(id string) (u User, err error) {
	err = s.db.View(func(tx *bolt.Tx) error { return get(tx, "accounts", id, &u) })
	return
}

func (s *Store) Users(search string) (list []User, err error) {
	list = []User{}
	search = foldUsername(strings.TrimSpace(search))
	err = s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("accounts")).ForEach(func(_, v []byte) error {
			var u User
			if err := json.Unmarshal(v, &u); err != nil {
				return err
			}
			if search == "" || strings.Contains(foldUsername(u.Username), search) {
				u.Password = nil
				list = append(list, u)
			}
			return nil
		})
	})
	return
}

func invalidateSessions(tx *bolt.Tx, userID string) error {
	c := tx.Bucket([]byte("sessions")).Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var session Session
		if err := json.Unmarshal(v, &session); err != nil {
			return err
		}
		if session.UserID == userID {
			if err := c.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
}

// ChangeAccount compares the password snapshot inside the same transaction as the
// update, preventing a concurrent password reset from authorizing stale changes.
func (s *Store) ChangeAccount(id, username string, password, expectedPassword []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		var u User
		if err := get(tx, "accounts", id, &u); err != nil || u.Disabled {
			return ErrDenied
		}
		if !bytes.Equal(u.Password, expectedPassword) {
			return ErrDenied
		}
		if username != "" {
			name, key, err := NormalizeUsername(username)
			if err != nil {
				return err
			}
			names := tx.Bucket([]byte("usernames"))
			if existing := names.Get([]byte(key)); existing != nil && !bytes.Equal(existing, []byte(id)) {
				return ErrExists
			}
			_, oldKey, _ := NormalizeUsername(u.Username)
			if err := names.Delete([]byte(oldKey)); err != nil {
				return err
			}
			if err := names.Put([]byte(key), []byte(id)); err != nil {
				return err
			}
			u.Username = name
		}
		if password != nil {
			u.Password = password
		}
		if err := put(tx, "accounts", id, u); err != nil {
			return err
		}
		return invalidateSessions(tx, id)
	})
}

// UpdateUser serializes the last-active-admin invariant with every privilege
// mutation. actorID may be empty only for the offline administrator command.
func (s *Store) UpdateUser(actorID, id string, admin, disabled *bool, password []byte) (out User, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		if actorID != "" {
			var actor User
			if get(tx, "accounts", actorID, &actor) != nil || !actor.Admin || actor.Disabled {
				return ErrDenied
			}
		}
		var u User
		if err := get(tx, "accounts", id, &u); err != nil {
			return err
		}
		out = u
		if admin != nil {
			out.Admin = *admin
		}
		if disabled != nil {
			out.Disabled = *disabled
		}
		if password != nil {
			out.Password = password
		}
		if u.Admin && !u.Disabled && (!out.Admin || out.Disabled) {
			count := 0
			if err := tx.Bucket([]byte("accounts")).ForEach(func(_, v []byte) error {
				var other User
				if err := json.Unmarshal(v, &other); err != nil {
					return err
				}
				if other.Admin && !other.Disabled {
					count++
				}
				return nil
			}); err != nil {
				return err
			}
			if count <= 1 {
				return ErrLastAdmin
			}
		}
		if err := put(tx, "accounts", id, out); err != nil {
			return err
		}
		if password != nil || u.Admin != out.Admin || u.Disabled != out.Disabled {
			return invalidateSessions(tx, id)
		}
		return nil
	})
	out.Password = nil
	return
}

func (s *Store) CreateUserAsAdmin(actorID string, u User) error {
	// Authorization is checked again within the write transaction to serialize
	// demotion/disable against administrative account creation.
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
		var actor User
		if get(tx, "accounts", actorID, &actor) != nil || actor.Disabled || !actor.Admin {
			return ErrDenied
		}
		if tx.Bucket([]byte("usernames")).Get([]byte(key)) != nil || tx.Bucket([]byte("accounts")).Get([]byte(u.ID)) != nil {
			return ErrExists
		}
		if err := put(tx, "accounts", u.ID, u); err != nil {
			return err
		}
		return tx.Bucket([]byte("usernames")).Put([]byte(key), []byte(u.ID))
	})
}
