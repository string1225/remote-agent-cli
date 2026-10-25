package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestLegacyPermissionsMigrateInMemoryWithoutChangingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	before := []byte(`{"server":"https://example.com","id":"device","credential":"fixture","allowWrite":false,"services":[{"id":"kept","name":"old","workspace":"legacy","provider":"codex"}]}`)
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil || !c.AllowWrite || c.PermissionMode != "full" || c.Services[0].ID != "kept" {
		t.Fatalf("legacy migration: %+v %v", c, err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("loading rewrote original data")
	}
}

func TestProjectRegistrationRetriesReuseProjectAndWebSession(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), AllowedRoots: []string{root}}, Providers: map[string]Provider{"codex": {Available: true}}}
	request := protocol.Service{Name: "Project", Provider: "codex", Workspace: root}
	first, err := a.registerProject(request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Session.ID == "" || first.Session.ServiceID != first.ID || first.Session.Kind != "web" {
		t.Fatalf("missing initial session: %+v", first)
	}
	if _, err = a.historyStore().append(first.Session.ID, protocol.ChatEntry{Role: "user", Text: "retained history"}); err != nil {
		t.Fatal(err)
	}
	second, err := a.registerProject(request)
	if err != nil || second.ID != first.ID || second.Session.ID != first.Session.ID || len(a.Config.Services) != 1 {
		t.Fatalf("retry duplicated data: %+v %v", second, err)
	}
	page, err := a.historyStore().read(first.Session.ID, 0)
	if err != nil || len(page.Entries) != 1 {
		t.Fatal("history lost", err)
	}
}

func TestProjectRegistrationRecoversAfterHistoryStorageFailure(t *testing.T) {
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json"), AllowedRoots: []string{root}}, Providers: map[string]Provider{"codex": {Available: true}}}
	blocked := a.historyStore().path
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	request := protocol.Service{Name: "Project", Provider: "codex", Workspace: root}
	if _, err := a.registerProject(request); err == nil {
		t.Fatal("history storage failure was ignored")
	}
	if len(a.Config.Services) != 1 {
		t.Fatal("project identity was discarded")
	}
	id := a.Config.Services[0].ID
	if err := os.Remove(blocked); err != nil {
		t.Fatal(err)
	}
	result, err := a.registerProject(request)
	if err != nil || result.ID != id || result.Session.ID == "" || len(a.Config.Services) != 1 {
		t.Fatal("retry duplicated or lost project", err)
	}
}
