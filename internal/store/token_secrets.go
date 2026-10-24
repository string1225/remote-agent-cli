package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"runtime"

	bolt "go.etcd.io/bbolt"
)

var ErrTokenNotSaved = errors.New("token was created before recoverable storage; import the original token to enable copying")

// Keep the key separate from database backups. Never replace a missing key if
// ciphertext exists: restoring the original key is required for recovery.
func tokenCipher(db *bolt.DB, path string) (cipher.AEAD, error) {
	if st, err := os.Lstat(path); err == nil {
		if !st.Mode().IsRegular() || runtime.GOOS != "windows" && st.Mode().Perm()&0077 != 0 {
			return nil, errors.New("token encryption key must be a private regular file (0600)")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	key, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		var populated bool
		if err = db.View(func(tx *bolt.Tx) error {
			k, _ := tx.Bucket([]byte("codex_secrets")).Cursor().First()
			populated = k != nil
			return nil
		}); err != nil {
			return nil, err
		}
		if populated {
			return nil, fmt.Errorf("restore token encryption key %s from backup", path)
		}
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return nil, e
		}
		_, err = f.Write(key)
		if err == nil {
			err = f.Sync()
		}
		err = errors.Join(err, f.Close())
	}
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("invalid token encryption key; restore it from backup")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *Store) sealToken(tx *bolt.Tx, id, secret string) error {
	nonce := make([]byte, s.tokenCipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	data := s.tokenCipher.Seal(nonce, nonce, []byte(secret), []byte(id))
	return tx.Bucket([]byte("codex_secrets")).Put([]byte(id), data)
}

func ownedToken(tx *bolt.Tx, userID, id string) bool {
	var info CodexToken
	return activeUser(tx, userID) && get(tx, "codex_tokens", id, &info) == nil && info.UserID == userID && !info.Revoked
}

func (s *Store) RevealCodexToken(userID, id string) (string, error) {
	var secret string
	err := s.db.View(func(tx *bolt.Tx) error {
		if !ownedToken(tx, userID, id) {
			return ErrDenied
		}
		data := tx.Bucket([]byte("codex_secrets")).Get([]byte(id))
		if data == nil {
			return ErrTokenNotSaved
		}
		n := s.tokenCipher.NonceSize()
		if len(data) < n {
			return errors.New("invalid saved token")
		}
		plain, err := s.tokenCipher.Open(nil, data[:n], data[n:], []byte(id))
		if err != nil {
			return err
		}
		secret = string(plain)
		return nil
	})
	return secret, err
}

// SaveCodexToken backfills an old hash-only record without issuing or rotating
// any credential. The user must already possess the exact original token.
func (s *Store) SaveCodexToken(userID, id, secret string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if !ownedToken(tx, userID, id) || string(tx.Bucket([]byte("codex_verifiers")).Get([]byte(Hash(secret)))) != id {
			return ErrDenied
		}
		return s.sealToken(tx, id, secret)
	})
}
