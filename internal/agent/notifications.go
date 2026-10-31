package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

// Old releases persisted automatic completion deliveries. Keep their exact
// bytes in a recovery bucket, but never send them after a session-managed upgrade.
func (h historyStore) retireAutomaticNotifications() error {
	if _, err := os.Stat(h.path); os.IsNotExist(err) {
		return nil
	}
	return h.transact(true, func(tx *bolt.Tx) error {
		for _, name := range []string{"remote_notifications", "channel_outbox"} {
			b := tx.Bucket([]byte(name))
			if b == nil {
				continue
			}
			archive, err := tx.CreateBucketIfNotExists([]byte("retired_" + name))
			if err != nil {
				return err
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if name == "channel_outbox" {
					var item channelDelivery
					if json.Unmarshal(v, &item) != nil || !automaticCompletionID(item.ID) {
						continue
					}
				}
				archiveKey := k
				if previous := archive.Get(k); previous != nil && !bytes.Equal(previous, v) {
					archiveKey = []byte(string(k) + "\x00" + store.Hash(string(v)))
				}
				if err = archive.Put(archiveKey, v); err != nil {
					return err
				}
				if err = c.Delete(); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func automaticCompletionID(id string) bool {
	if strings.HasSuffix(id, "-completion") {
		return true
	}
	i := strings.LastIndex(id, "-completion-")
	if i < 0 {
		return false
	}
	part := id[i+len("-completion-"):]
	if len(part) != 4 {
		return false
	}
	for _, r := range part {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (a *Agent) initNotifications(context.Context) {
	// Retain the shutdown channel for existing daemon lifecycle callers.
	a.notifyDone = make(chan struct{})
	if err := a.historyStore().retireAutomaticNotifications(); err != nil {
		log.Printf("retaining legacy task notifications without delivery: %v", err)
	}
	close(a.notifyDone)
}
