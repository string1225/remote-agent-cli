package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type RepairTarget struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	OS             string             `json:"os"`
	Enrolled       bool               `json:"enrolled"`
	MachineID      string             `json:"machineId"`
	Hostname       string             `json:"hostname"`
	BindingValid   bool               `json:"bindingValid"`
	MachineMatch   bool               `json:"machineMatch"`
	RepairSourceID string             `json:"repairSourceId"`
	Services       []protocol.Service `json:"services"`
}

func InspectRepair(client *http.Client, server, id, token, machine string, local Config) (RepairTarget, error) {
	var target RepairTarget
	if err := ValidateServer(server); err != nil {
		return target, err
	}
	body, _ := json.Marshal(map[string]string{"id": id, "token": token, "machineId": machine})
	req, err := http.NewRequest("POST", strings.TrimSuffix(server, "/")+"/api/repair/inspect", bytes.NewReader(body))
	if err != nil {
		return target, err
	}
	req.Header.Set("Content-Type", "application/json")
	if local.ID == id && strings.TrimSuffix(local.Server, "/") == strings.TrimSuffix(server, "/") {
		req.Header.Set("Authorization", "Bearer "+local.Credential)
	}
	res, err := client.Do(req)
	if err != nil {
		return target, errors.New("Cannot inspect target client; check the network and retry")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return target, errors.New("Repair authorization is invalid or expired; copy a new repair command from the target client")
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 16384)).Decode(&target); err != nil || target.ID != id {
		return target, errors.New("Repair target does not match the client ID in the command")
	}
	if target.OS != "" && target.OS != runtime.GOOS {
		return target, fmt.Errorf("Target client OS is %s; this computer runs %s. OS mismatch cannot be overridden", target.OS, runtime.GOOS)
	}
	return target, nil
}

// ClaimRepair durably retains the proposed credential before the capability is
// consumed. A failed save/response can be retried using the private journal.
func ClaimRepair(client *http.Client, server, id, token, path, machine, hostname string, force bool) (c Config, backup string, err error) {
	server = strings.TrimSuffix(server, "/")
	if err = ValidateServer(server); err != nil {
		return
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	unlock, e := lockFile(path + ".write.lock")
	if e != nil {
		return c, "", errors.New("Another configuration or repair operation is in progress")
	}
	defer unlock()
	if info, e := os.Lstat(path); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return c, "", errors.New("Binding configuration must not be a symbolic link")
	}
	var original []byte
	var document map[string]json.RawMessage
	original, err = os.ReadFile(path)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return
	}
	var source Config
	if len(original) > 0 {
		if err = json.Unmarshal(original, &document); err != nil {
			return c, "", errors.New("Local configuration is damaged; restore a backup first. Binding was not changed")
		}
		if err = json.Unmarshal(original, &source); err != nil {
			return
		}
	}
	pending := path + ".claim-pending.json"
	var journal rebindJournal
	if raw, e := os.ReadFile(pending); e == nil {
		if json.Unmarshal(raw, &journal) != nil || !validRebindJournal(journal, server, id) {
			return c, "", errors.New("A recovery transaction for another target is pending; retain and inspect the pending file")
		}
		backup = journal.Backup
	} else if !os.IsNotExist(e) {
		return c, "", e
	}
	if journal.Credential == "" {
		if token == "" {
			return c, "", errors.New("Rebinding requires a target client repair authorization issued by the website")
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return
		}
		if len(original) > 0 {
			backup = path + ".claim-backup-" + time.Now().UTC().Format("20060102T150405.000000000")
			if err = writeRecoveryFile(backup, original); err != nil {
				return
			}
		}
		if document == nil {
			root, e := os.UserHomeDir()
			if e != nil {
				return c, backup, e
			}
			initial := Config{AllowedRoots: []string{root}, Services: []protocol.Service{}, AllowWrite: true, PermissionMode: "full"}
			raw, _ := json.Marshal(initial)
			json.Unmarshal(raw, &document)
		}
		document["id"], _ = json.Marshal(id)
		document["server"], _ = json.Marshal(server)
		journal = rebindJournal{ID: id, Server: server, Credential: store.Token(), Backup: backup}
		document["credential"], _ = json.Marshal(journal.Credential)
		journal.Config, _ = json.MarshalIndent(document, "", "  ")
		raw, _ := json.Marshal(journal)
		if err = writeRecoveryFile(pending, raw); err != nil {
			return
		}
	}
	fail := func(e error) (Config, string, error) {
		return Config{}, backup, fmt.Errorf("Rebinding is incomplete: %w; original configuration and recovery file retained: %s", e, pending)
	}
	var result struct {
		ID                  string             `json:"id"`
		Name                string             `json:"name"`
		Credential          string             `json:"credential"`
		Services            []protocol.Service `json:"services"`
		ChannelsTransferred bool               `json:"channelsTransferred"`
	}
	// The inspect endpoint also authenticates an already committed credential,
	// allowing recovery after response loss without reusing a one-time token.
	probe := Config{ID: id, Server: server, Credential: journal.Credential}
	target, e := InspectRepair(client, server, id, token, machine, probe)
	if e == nil && target.BindingValid && target.MachineMatch {
		result.ID = id
		result.Name = target.Name
		result.Credential = journal.Credential
		result.Services = target.Services
		result.ChannelsTransferred = source.ID == id || source.ID == "" || target.RepairSourceID == source.ID
	} else {
		in := map[string]any{"id": id, "token": token, "credential": journal.Credential, "machineId": machine, "hostname": hostname, "os": runtime.GOOS, "arch": runtime.GOARCH, "force": force}
		if source.Server == server {
			in["sourceId"] = source.ID
			in["sourceCredential"] = source.Credential
		}
		raw, _ := json.Marshal(in)
		res, e := client.Post(server+"/api/repair/claim", "application/json", bytes.NewReader(raw))
		if e != nil {
			return fail(errors.New("Cannot confirm the server binding result; retry the same repair command"))
		}
		e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&result)
		res.Body.Close()
		if e != nil || res.StatusCode != 200 {
			return fail(errors.New("Repair authorization is invalid or the target OS does not match"))
		}
	}
	if result.ID != id || result.Credential != journal.Credential {
		return fail(errors.New("Server returned a different binding target"))
	}
	if err = json.Unmarshal(journal.Config, &document); err != nil {
		return fail(err)
	}
	document["name"], _ = json.Marshal(result.Name)
	if source.ID != "" && source.ID != id && !result.ChannelsTransferred {
		// Secrets stay in the private original backup; another account's robot
		// reservations cannot be silently moved by a target-only capability.
		document["channels"] = json.RawMessage("[]")
		var services []protocol.Service
		json.Unmarshal(document["services"], &services)
		for i := range services {
			services[i].Channel = nil
		}
		document["services"], _ = json.Marshal(services)
	}
	if source.ID == "" {
		var recovered Config
		json.Unmarshal(journal.Config, &recovered)
		if len(recovered.Services) == 0 {
			recovered.Services = result.Services
			recoverRebindProjects(path, &recovered)
			document["services"], _ = json.Marshal(recovered.Services)
			document["allowedRoots"], _ = json.Marshal(recovered.AllowedRoots)
		}
	}
	journal.Config, _ = json.MarshalIndent(document, "", "  ")
	raw, _ := json.Marshal(journal)
	if err = writeRecoveryFile(pending, raw); err != nil {
		return fail(err)
	}
	if err = writeRecoveryFile(path, journal.Config); err != nil {
		return fail(err)
	}
	if err = json.Unmarshal(journal.Config, &c); err != nil {
		return fail(err)
	}
	c.Path = path
	c.AllowWrite = true
	c.PermissionMode = "full"
	if err = os.Remove(pending); err != nil {
		return c, backup, errors.New("Binding completed; keep the backup and remove the claim-pending file")
	}
	return
}
