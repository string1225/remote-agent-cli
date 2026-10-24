package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

type remoteNotification struct{ ID, AgentID, ServiceID, Text string }

func (a *Agent) initNotifications(ctx context.Context) {
	a.notifyWake = make(chan struct{}, 1)
	a.notifyDone = make(chan struct{})
	go func() {
		defer close(a.notifyDone)
		if _, err := os.Stat(filepath.Join(filepath.Dir(a.Config.Path), "notifications.enabled")); err == nil {
			// A task may already have published a wake while this worker was
			// starting. Never block startup (and cancellation) on a full signal.
			select {
			case a.notifyWake <- struct{}{}:
			default:
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.notifyWake:
			}
			for ctx.Err() == nil {
				items, err := a.historyStore().remotePending()
				if err == nil && len(items) == 0 {
					break
				}
				if err == nil {
					for _, item := range items {
						a.mu.Lock()
						config := a.Config
						a.mu.Unlock()
						_, err = ClientTool(ctx, config, item.AgentID, "channels.notify", map[string]string{"serviceId": item.ServiceID, "message": item.Text})
						if err != nil {
							break
						}
						if err = a.historyStore().remoteDelivered(item.ID); err != nil {
							break
						}
					}
				}
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(30 * time.Second):
					case <-a.notifyWake:
					}
				}
			}
		}
	}()
}
func (a *Agent) queueRemoteNotification(v remoteNotification) error {
	// Publish the recovery marker before the durable queue entry. A crash can
	// leave an empty enabled queue, never an undiscoverable pending delivery.
	marker := filepath.Join(filepath.Dir(a.Config.Path), "notifications.enabled")
	f, err := os.OpenFile(marker, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	f.Close()
	err = a.historyStore().transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("remote_notifications"))
		if err != nil {
			return err
		}
		if b.Stats().KeyN >= 1000 {
			return errors.New("跨设备通知队列已满")
		}
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return b.Put([]byte(v.ID), data)
	})
	if err == nil {
		select {
		case a.notifyWake <- struct{}{}:
		default:
		}
	}
	return err
}
func (h historyStore) remotePending() (items []remoteNotification, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("remote_notifications"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, v []byte) error {
			var item remoteNotification
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	return
}
func (h historyStore) remoteDelivered(id string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("remote_notifications"))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(id))
	})
}
