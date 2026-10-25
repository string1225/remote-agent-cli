package agent

import (
	"bytes"
	"encoding/json"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestClaimRepairRetainsDataAndRecoversLostResponse(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	active := ""
	claims := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/repair/inspect" {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "target", "name": "Target", "os": runtime.GOOS, "enrolled": true, "bindingValid": active != "" && r.Header.Get("Authorization") == "Bearer "+active, "machineMatch": true, "repairSourceId": "source"})
			return
		}
		if r.URL.Path != "/api/repair/claim" {
			t.Error("unexpected path", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		claims++
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in["id"] != "target" || in["force"] != true {
			t.Error("missing target/confirmation")
		}
		active = in["credential"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "target", "name": "Target", "credential": "mismatched-response"})
	}))
	defer site.Close()
	original, _ := json.Marshal(map[string]any{"id": "source", "name": "Source", "server": site.URL, "credential": "fixture-old", "allowedRoots": []string{root}, "services": []protocol.Service{{ID: "project", Workspace: root, Provider: "codex"}}, "channels": []ChannelConfig{{ServiceID: "project", Provider: "feishu", AppSecret: "fixture-only"}}, "custom": "retained"})
	_ = os.WriteFile(path, original, 0600)
	_, backup, e := ClaimRepair(site.Client(), site.URL, "target", "capability", path, strings.Repeat("a", 64), "actual", true)
	if e == nil {
		t.Fatal("incorrect response accepted")
	}
	current, _ := os.ReadFile(path)
	if !bytes.Equal(current, original) {
		t.Fatal("failed repair destroyed original")
	}
	saved, _ := os.ReadFile(backup)
	if !bytes.Equal(saved, original) {
		t.Fatal("backup missing")
	}
	c, _, e := ClaimRepair(site.Client(), site.URL, "target", "used", path, strings.Repeat("a", 64), "actual", true)
	if e != nil || claims != 1 || c.ID != "target" || len(c.Services) != 1 || c.Services[0].ID != "project" || len(c.Channels) != 1 {
		t.Fatal(c.ID, claims, e)
	}
	current, _ = os.ReadFile(path)
	if !bytes.Contains(current, []byte(`"custom": "retained"`)) {
		t.Fatal("unknown data lost")
	}
	if _, e = os.Stat(path + ".claim-pending.json"); !os.IsNotExist(e) {
		t.Fatal("journal not retired")
	}
	stale := Config{Path: path, Server: site.URL, ID: "source", Credential: "fixture-old"}
	if stale.Save() == nil {
		t.Fatal("old daemon overwrote new binding")
	}
}

func TestClaimRepairWithoutLocalConfigRecoversProjectsAfterResponseLoss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	active := ""
	claims := 0
	services := []protocol.Service{{ID: "retained-project", Name: "Original", Provider: "codex", Workspace: t.TempDir()}}
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/repair/inspect" {
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "target", "os": runtime.GOOS, "bindingValid": active != "" && r.Header.Get("Authorization") == "Bearer "+active, "machineMatch": true, "services": services})
			return
		}
		claims++
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		active = in["credential"].(string)
		_, _ = w.Write([]byte("lost-response"))
	}))
	defer site.Close()
	if _, _, err := ClaimRepair(site.Client(), site.URL, "target", "fixture", path, strings.Repeat("a", 64), "new-machine", true); err == nil {
		t.Fatal("lost response accepted")
	}
	c, _, err := ClaimRepair(site.Client(), site.URL, "target", "used", path, strings.Repeat("a", 64), "new-machine", true)
	if err != nil || claims != 1 || len(c.Services) != 1 || c.Services[0].ID != services[0].ID {
		t.Fatal("recovery lost original projects", err, claims, c.Services)
	}
}
