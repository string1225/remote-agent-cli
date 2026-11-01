package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestProjectSettingsPreserveIdentityNativeContextAndConfigBackup(t *testing.T) {
	a, project := concurrencyFixture(t)
	if err := a.Config.Save(); err != nil {
		t.Fatal(err)
	}
	session, _ := a.historyStore().create(project, "retained history")
	_, _ = a.historyStore().append(session.ID, protocol.ChatEntry{Role: "user", Text: "retained"})
	_ = a.historyStore().nativeID(session.ID, "native-kept")
	before, _ := os.ReadFile(a.Config.Path)
	nextRoot := filepath.Join(project.Workspace, "new-root")
	if err := os.Mkdir(nextRoot, 0700); err != nil {
		t.Fatal(err)
	}
	updated, err := a.updateProject(protocol.Service{ID: project.ID, Name: "Renamed project", Workspace: nextRoot})
	expectedRoot, _ := a.Config.Workspace(nextRoot)
	if err != nil || updated.ID != project.ID || updated.Provider != project.Provider || updated.Workspace != expectedRoot {
		t.Fatal(updated, err)
	}
	retained, _ := a.historyStore().get(session.ID)
	if retained.NativeID != "native-kept" || retained.Workspace != project.Workspace {
		t.Fatal("historical data was rewritten", retained)
	}
	backups, _ := filepath.Glob(a.Config.Path + ".project-settings-backup-*")
	if len(backups) != 1 {
		t.Fatal(backups)
	}
	backup, _ := os.ReadFile(backups[0])
	if string(backup) != string(before) {
		t.Fatal("configuration recovery copy differs")
	}
	ctx, stop := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { stop(); <-a.taskDone; <-a.notifyDone })
	seen := make(chan protocol.Service, 1)
	a.Execute = func(_ context.Context, _ Provider, s protocol.Service, _ string, _ bool, _ func(protocol.Event) error, _ string) error {
		seen <- s
		return nil
	}
	if _, err := a.startSessionTask(protocol.Request{ID: "after-move", ServiceID: project.ID, SessionID: session.ID, Prompt: "continue"}); err != nil {
		t.Fatal(err)
	}
	service := <-seen
	if service.Workspace != expectedRoot || service.ResumeID != "native-kept" {
		t.Fatal(service)
	}
	if _, err := a.updateProject(protocol.Service{ID: project.ID, Name: "invalid", Workspace: t.TempDir()}); err == nil {
		t.Fatal("unauthorized root accepted")
	}
}

func TestSessionRenameArchiveAndSoftDeleteRetainOriginalData(t *testing.T) {
	a, project := concurrencyFixture(t)
	session, _ := a.historyStore().create(project, "original")
	_, _ = a.historyStore().append(session.ID, protocol.ChatEntry{Role: "user", Text: "keep my history"})
	_, err := a.updateSession(protocol.Request{Type: "sessions.rename", ServiceID: project.ID, SessionID: session.ID, Title: "Renamed"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.updateSession(protocol.Request{Type: "sessions.archive", ServiceID: project.ID, SessionID: session.ID, Archived: true})
	if err != nil {
		t.Fatal(err)
	}
	active, _ := a.historyStore().list(project.ID, "", 0)
	archived, _ := a.historyStore().listArchived(project.ID, "", 0, true)
	if len(active.Sessions) != 0 || len(archived.Sessions) != 1 || archived.Sessions[0].Title != "Renamed" {
		t.Fatal(active, archived)
	}
	_, err = a.updateSession(protocol.Request{Type: "sessions.archive", ServiceID: project.ID, SessionID: session.ID})
	if err != nil {
		t.Fatal(err)
	}
	active, _ = a.historyStore().list(project.ID, "", 0)
	if len(active.Sessions) != 1 {
		t.Fatal("restore failed")
	}
	_, err = a.updateSession(protocol.Request{Type: "sessions.delete", ServiceID: project.ID, SessionID: session.ID})
	if err != nil {
		t.Fatal(err)
	}
	active, _ = a.historyStore().list(project.ID, "", 0)
	archived, _ = a.historyStore().listArchived(project.ID, "", 0, true)
	if len(active.Sessions) != 0 || len(archived.Sessions) != 0 {
		t.Fatal("deleted row still listed")
	}
	page, err := a.historyStore().read(session.ID, 0)
	if err != nil || len(page.Entries) != 1 || page.Entries[0].Text != "keep my history" || page.Session.DeletedAt == 0 {
		t.Fatal("recovery data lost", page, err)
	}
	if _, err := a.updateSession(protocol.Request{Type: "sessions.rename", ServiceID: "other-project", SessionID: session.ID, Title: "cross-project"}); err == nil {
		t.Fatal("cross-project mutation allowed")
	}
}
