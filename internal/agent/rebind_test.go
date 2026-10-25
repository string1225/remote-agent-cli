package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestRebindPreservesRawConfigAndRejectsDifferentIdentityBeforeExchange(t *testing.T) {
	calls := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var in map[string]any
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatal(err)
		}
		if in["expectedId"] != "device" || in["rebind"] != true {
			t.Error("missing expected target")
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "device", "name": "Desktop", "credential": in["credential"]})
	}))
	defer site.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	original, _ := json.Marshal(map[string]any{"id": "device", "server": site.URL, "credential": "old", "name": "kept", "allowedRoots": []string{t.TempDir()}, "services": []protocol.Service{{ID: "project", Provider: "codex", Workspace: t.TempDir()}}, "channels": []ChannelConfig{{ServiceID: "project", Provider: "feishu", AppSecret: "mock-secret"}}, "futureSetting": map[string]string{"keep": "yes"}})
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Rebind(site.URL, "invite", path, "different", nil); err == nil || calls != 0 {
		t.Fatal("mismatch consumed recovery", err)
	}
	if _, _, err := Rebind(site.URL+"/other", "invite", path, "device", nil); err == nil || calls != 0 {
		t.Fatal("server mismatch consumed recovery", err)
	}
	restored, backup, err := Rebind(site.URL, "invite", path, "device", nil)
	if err != nil || restored.ID != "device" || restored.Credential == "old" || len(restored.Services) != 1 || restored.Channels[0].AppSecret != "mock-secret" {
		t.Fatal("recovery lost config", err)
	}
	data, _ := os.ReadFile(backup)
	if !bytes.Equal(data, original) {
		t.Fatal("original backup changed")
	}
	data, _ = os.ReadFile(path)
	var saved map[string]any
	json.Unmarshal(data, &saved)
	if saved["futureSetting"] == nil {
		t.Fatal("unknown original data lost")
	}
}

func TestRebindLostResponseCanConfirmAlreadyActivatedCredential(t *testing.T) {
	credential := ""
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/agent/catalog" {
			if r.Header.Get("Authorization") != "Bearer "+credential {
				w.WriteHeader(401)
				return
			}
			io.WriteString(w, `{"clients":[{"id":"device","name":"Recovered","projects":[]}]}`)
			return
		}
		var in map[string]any
		json.NewDecoder(r.Body).Decode(&in)
		credential = in["credential"].(string)
		connection, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		connection.Close()
	}))
	defer site.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	c, _, err := Rebind(site.URL, "invite", path, "device", []string{t.TempDir()})
	if err != nil || c.ID != "device" || c.Credential != credential || c.Name != "Recovered" {
		t.Fatal("lost response made recovery impossible", err)
	}
}

func TestRebindPendingCredentialSurvivesFailedLocalCommitAndCanRetry(t *testing.T) {
	proposed := ""
	exchanges := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/agent/catalog" {
			if proposed != "" && r.Header.Get("Authorization") == "Bearer "+proposed {
				io.WriteString(w, `{"clients":[{"id":"device","name":"Desktop","projects":[]}]}`)
			} else {
				w.WriteHeader(401)
			}
			return
		}
		exchanges++
		var in map[string]any
		json.NewDecoder(r.Body).Decode(&in)
		proposed = in["credential"].(string)
		json.NewEncoder(w).Encode(map[string]any{"id": "device", "name": "Desktop", "credential": proposed})
	}))
	defer site.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	before, _ := json.Marshal(Config{ID: "device", Server: site.URL, Credential: "old", AllowedRoots: []string{t.TempDir()}})
	os.WriteFile(path, before, 0600)
	_, _, err := rebindWithClient(site.Client(), site.URL, "invite", path, "device", nil, func(string, []byte) error { return errors.New("fixture disk error") })
	if err == nil {
		t.Fatal("failed config write accepted")
	}
	old, _ := os.ReadFile(path)
	if !bytes.Equal(old, before) {
		t.Fatal("failed write destroyed original config")
	}
	pending, _ := os.ReadFile(RebindPendingPath(path))
	if !bytes.Contains(pending, []byte(proposed)) {
		t.Fatal("new credential was not durably retained")
	}
	restored, _, err := Rebind(site.URL, "", path, "device", nil)
	if err != nil || restored.Credential != proposed || exchanges != 1 {
		t.Fatal("retry lost credential or redeemed twice", err, exchanges)
	}
}

func TestRebindMissingConfigRecoversSameProjectIDsFromHistory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	project := protocol.Service{ID: "project", Name: "Retained", Provider: "codex", Workspace: t.TempDir()}
	canonical, err := canonicalDir(project.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	project.Workspace = canonical
	history := historyStore{filepath.Join(root, "history.db")}
	session, err := history.create(project, "history")
	if err != nil {
		t.Fatal(err)
	}
	history.append(session.ID, protocol.ChatEntry{Role: "user", Text: "keep history"})
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		json.NewDecoder(r.Body).Decode(&in)
		json.NewEncoder(w).Encode(map[string]any{"id": "device", "name": "Desktop", "credential": in["credential"], "services": []protocol.Service{{ID: project.ID, Name: project.Name, Provider: project.Provider}}})
	}))
	defer site.Close()
	c, _, err := Rebind(site.URL, "invite", path, "device", nil)
	if err != nil || c.ID != "device" || len(c.Services) != 1 || c.Services[0].ID != project.ID || c.Services[0].Workspace != project.Workspace {
		t.Fatal("missing config recovery lost project", err)
	}
	page, err := history.read(session.ID, 0)
	if err != nil || len(page.Entries) != 1 {
		t.Fatal("original history changed", err)
	}
}
