package agent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type rebindJournal struct {
	Server     string          `json:"server"`
	ID         string          `json:"id"`
	Credential string          `json:"credential"`
	Backup     string          `json:"backup,omitempty"`
	Config     json.RawMessage `json:"config"`
}

func RebindPendingPath(path string) string { return path + ".rebind-pending.json" }

// Rebind replaces only the credential of an explicitly selected identity.
// The proposed credential is durable before redemption; a failed local commit
// or lost HTTP response can be recovered by rerunning with the same device ID.
func Rebind(server, invite, path, id string, roots []string) (Config, string, error) {
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return rebindWithClient(client, server, invite, path, id, roots, writeRecoveryFile)
}

func readRebindConfig(path, server, id string) (map[string]json.RawMessage, []byte, error) {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, errors.New("rebind configuration must not be a symbolic link")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var document map[string]json.RawMessage
	var identity struct {
		ID     string `json:"id"`
		Server string `json:"server"`
	}
	if json.Unmarshal(data, &document) != nil || document == nil || json.Unmarshal(data, &identity) != nil {
		return nil, nil, errors.New("existing device configuration is invalid; restore its backup before rebinding")
	}
	if identity.ID != id || strings.TrimSuffix(identity.Server, "/") != server {
		return nil, nil, errors.New("recovery target differs from the existing device ID or server; nothing was changed")
	}
	return document, data, nil
}

func rebindWithClient(client *http.Client, server, invite, path, id string, roots []string, commit func(string, []byte) error) (c Config, backup string, err error) {
	server = strings.TrimSuffix(server, "/")
	if err = ValidateServer(server); err != nil {
		return
	}
	if id == "" || len(id) > 200 || strings.ContainsAny(id, "\r\n\t ") {
		return c, "", errors.New("rebind requires the exact device ID")
	}
	document, original, err := readRebindConfig(path, server, id)
	if err != nil {
		return c, "", err
	}
	var journal rebindJournal
	pending := RebindPendingPath(path)
	data, readErr := os.ReadFile(pending)
	resumed := readErr == nil
	if resumed {
		if json.Unmarshal(data, &journal) != nil || !validRebindJournal(journal, server, id) {
			return c, "", errors.New("pending recovery belongs to another device or is invalid; preserve it and inspect the original backup")
		}
		backup = journal.Backup
	} else {
		if !errors.Is(readErr, os.ErrNotExist) {
			return c, "", readErr
		}
		if invite == "" {
			return c, "", errors.New("supply a recovery token with --token-prompt or --token-stdin")
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return c, "", err
		}
		if document == nil {
			if len(roots) == 0 {
				home, e := os.UserHomeDir()
				if e != nil {
					return c, "", e
				}
				roots = []string{home}
			}
			for i, root := range roots {
				roots[i], err = canonicalDir(root)
				if err != nil {
					return
				}
			}
			initial := Config{Server: server, ID: id, AllowedRoots: roots, AllowWrite: true, PermissionMode: "full", Services: []protocol.Service{}}
			raw, _ := json.Marshal(initial)
			json.Unmarshal(raw, &document)
		}
		if len(original) > 0 {
			backup = path + ".rebind-backup-" + time.Now().UTC().Format("20060102T150405.000000000")
			if err = writeRecoveryFile(backup, original); err != nil {
				return c, "", err
			}
		}
		journal = rebindJournal{Server: server, ID: id, Credential: store.Token(), Backup: backup}
		document["credential"], _ = json.Marshal(journal.Credential)
		journal.Config, _ = json.MarshalIndent(document, "", "  ")
		raw, _ := json.MarshalIndent(journal, "", "  ")
		if err = writeRecoveryFile(pending, raw); err != nil {
			return
		}
	}
	fail := func(cause error) (Config, string, error) {
		return Config{}, backup, fmt.Errorf("rebind was not saved: %w; original configuration retained; recovery file: %s; rerun connect --rebind %s with the same server", cause, pending, id)
	}
	var confirmed struct {
		ID         string             `json:"id"`
		Name       string             `json:"name"`
		Credential string             `json:"credential"`
		Services   []protocol.Service `json:"services"`
	}
	verify := func() bool {
		req, e := http.NewRequest("GET", server+"/api/agent/catalog?id="+url.QueryEscape(id), nil)
		if e != nil {
			return false
		}
		req.Header.Set("Authorization", "Bearer "+journal.Credential)
		response, e := client.Do(req)
		if e != nil {
			return false
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return false
		}
		var catalog struct {
			Clients []struct {
				ID       string             `json:"id"`
				Name     string             `json:"name"`
				Projects []protocol.Service `json:"projects"`
			} `json:"clients"`
		}
		if json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&catalog) != nil {
			return false
		}
		for _, item := range catalog.Clients {
			if item.ID == id {
				confirmed.ID, confirmed.Name, confirmed.Credential, confirmed.Services = id, item.Name, journal.Credential, item.Projects
				return true
			}
		}
		return false
	}
	if !resumed || !verify() {
		if invite == "" {
			return fail(errors.New("pending credential is not active; provide a new owner recovery token"))
		}
		body, _ := json.Marshal(map[string]any{"token": invite, "os": runtime.GOOS, "arch": runtime.GOARCH, "rebind": true, "expectedId": id, "credential": journal.Credential})
		response, e := client.Post(server+"/api/enroll", "application/json", bytes.NewReader(body))
		if e != nil {
			if !verify() {
				return fail(errors.New("cannot confirm recovery with the enrollment server"))
			}
		} else {
			decodeErr := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&confirmed)
			response.Body.Close()
			if response.StatusCode != 200 || decodeErr != nil {
				if !verify() {
					return fail(fmt.Errorf("recovery exchange returned HTTP %d", response.StatusCode))
				}
			}
		}
	}
	if confirmed.ID != id || confirmed.Credential != journal.Credential {
		return fail(errors.New("recovery response does not match the expected device"))
	}
	// Merge the newest local data, since a background client may have saved a
	// project/channel while the recovery exchange was in flight.
	document, _, err = readRebindConfig(path, server, id)
	if err != nil {
		return fail(err)
	}
	if document == nil {
		if json.Unmarshal(journal.Config, &document) != nil {
			return fail(errors.New("pending configuration is invalid"))
		}
		var saved Config
		json.Unmarshal(journal.Config, &saved)
		if len(saved.Services) == 0 {
			saved.Services = confirmed.Services
			recoverRebindProjects(path, &saved)
			document["services"], _ = json.Marshal(saved.Services)
			document["allowedRoots"], _ = json.Marshal(saved.AllowedRoots)
		}
		document["name"], _ = json.Marshal(confirmed.Name)
	}
	document["credential"], _ = json.Marshal(journal.Credential)
	journal.Config, _ = json.MarshalIndent(document, "", "  ")
	raw, _ := json.MarshalIndent(journal, "", "  ")
	if err = writeRecoveryFile(pending, raw); err != nil {
		return fail(err)
	}
	if err = commit(path, journal.Config); err != nil {
		return fail(err)
	}
	if err = json.Unmarshal(journal.Config, &c); err != nil {
		return fail(err)
	}
	c.Path = path
	c.PermissionMode, c.AllowWrite = "full", true
	if err = os.Remove(pending); err != nil {
		return c, backup, fmt.Errorf("device recovered; remove the private recovery file %s after checking the configuration: %w", pending, err)
	}
	return c, backup, nil
}

func validRebindJournal(j rebindJournal, server, id string) bool {
	credential, err := base64.RawURLEncoding.DecodeString(j.Credential)
	if err != nil || len(credential) != 32 || j.ID != id || j.Server != server {
		return false
	}
	var candidate Config
	return json.Unmarshal(j.Config, &candidate) == nil && candidate.ID == id && strings.TrimSuffix(candidate.Server, "/") == server && candidate.Credential == j.Credential
}

func recoverRebindProjects(path string, c *Config) {
	history := historyStore{filepath.Join(filepath.Dir(path), "history.db")}
	if _, err := os.Stat(history.path); err != nil {
		return
	}
	_ = history.transact(false, func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("sessions")).ForEach(func(_, data []byte) error {
			var session protocol.Conversation
			if json.Unmarshal(data, &session) != nil {
				return nil
			}
			for i := range c.Services {
				if c.Services[i].ID == session.ServiceID && c.Services[i].Workspace == "" && session.Provider == c.Services[i].Provider {
					if root, err := canonicalDir(session.Workspace); err == nil {
						c.Services[i].Workspace = root
						c.AllowedRoots = append(c.AllowedRoots, root)
					}
				}
			}
			return nil
		})
	})
}

func writeRecoveryFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".rebind-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = protectRebindFile(f.Name()); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
