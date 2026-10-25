package agent

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type projectRegistration struct {
	protocol.Service
	Session protocol.Conversation `json:"session"`
}

// registerProject makes response-loss retries reuse the original project and
// its history. A failed history write leaves the saved project retryable.
func (a *Agent) registerProject(request protocol.Service) (result projectRegistration, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	workspace, err := a.Config.Workspace(request.Workspace)
	if err != nil {
		return result, err
	}
	for _, s := range a.Config.Services {
		existing, e := a.Config.Workspace(s.Workspace)
		if e == nil && existing == workspace && s.Provider == request.Provider {
			result.Service = s
			break
		}
	}
	if result.ID == "" {
		if len(a.Config.Services) >= 100 {
			return result, errors.New("100 service limit")
		}
		if !a.Providers[request.Provider].Available {
			return result, errors.New("install and sign in to a compatible provider CLI locally first")
		}
		if len(strings.TrimSpace(request.Name)) == 0 || len(request.Name) > 80 {
			return result, errors.New("service name must be 1–80 bytes")
		}
		result.Service = protocol.Service{ID: store.Token(), Name: strings.TrimSpace(request.Name), Provider: request.Provider, Workspace: workspace}
		a.Config.Services = append(a.Config.Services, result.Service)
		if err = a.Config.Save(); err != nil {
			a.Config.Services = a.Config.Services[:len(a.Config.Services)-1]
			return result, err
		}
	}
	result.Session, err = a.historyStore().ensureWebSession(result.Service)
	return result, err
}

func (h historyStore) ensureWebSession(service protocol.Service) (s protocol.Conversation, err error) {
	err = h.transact(true, func(tx *bolt.Tx) error {
		if err := tx.Bucket([]byte("sessions")).ForEach(func(_, v []byte) error {
			var candidate protocol.Conversation
			if err := json.Unmarshal(v, &candidate); err != nil {
				return err
			}
			if candidate.ServiceID == service.ID && (candidate.Kind == "" || candidate.Kind == "web") && (s.ID == "" || candidate.CreatedAt < s.CreatedAt) {
				s = candidate
			}
			return nil
		}); err != nil {
			return err
		}
		if s.ID != "" {
			return nil
		}
		now := time.Now().UnixMilli()
		s = protocol.Conversation{ID: store.Token(), ServiceID: service.ID, Title: "新会话", Provider: service.Provider, Workspace: service.Workspace, CreatedAt: now, UpdatedAt: now, Kind: "web", Source: "web"}
		if err := saveSession(tx, s); err != nil {
			return err
		}
		_, err := tx.Bucket([]byte("entries")).CreateBucket([]byte(s.ID))
		return err
	})
	s.NativeID = ""
	return
}
