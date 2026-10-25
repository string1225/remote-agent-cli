package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestSaysoProjectMigrationAndActionCreatesIndependentWebSession(t *testing.T) {
	for _, provider := range []string{"codex", "qoder"} {
		t.Run(provider, func(t *testing.T) { verifySaysoProjectAction(t, provider) })
	}
}
func verifySaysoProjectAction(t *testing.T, provider string) {
	s, p := soFixture(t)
	s.a.Config.Services[0].Provider = provider
	daemon, stop := context.WithCancel(context.Background())
	s.a.initTasks(daemon)
	t.Cleanup(func() { stop(); <-s.a.taskDone; <-s.a.notifyDone })
	s.a.Providers[provider] = Provider{Name: provider, Available: true}
	old := s.state.Projects[0]
	old.ServiceID = "" // pre-unified project with existing discussion and recordings
	v := soNewSession(t, s, p)
	discussion := s.session(v.ID)
	discussion.Recordings = []soRecording{{ID: "retained-recording", Filename: "retained.webm"}}
	project, err := s.bindProject("demo")
	if err != nil || project.ID != old.ID || project.PinnedSessionID != v.ID || len(discussion.Recordings) != 1 {
		t.Fatal("project migration lost discussion/recording identity", err)
	}
	again, err := s.bindProject("demo")
	if err != nil || again.PinnedSessionID != v.ID || len(s.state.Projects) != 1 {
		t.Fatal("pinned SaySo discussion not stable")
	}
	started, release := make(chan struct{}), make(chan struct{})
	s.a.Execute = func(ctx context.Context, _ Provider, service protocol.Service, _ string, write bool, send func(protocol.Event) error, id string) error {
		if service.ID != "demo" || !write {
			t.Error("action changed project/permissions")
		}
		if provider == "qoder" && service.Model != "" {
			t.Error("Codex execution model leaked into Qoder project")
		}
		close(started)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
		}
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": "done"}})
	}
	raw := soCall(t, s, p, "POST", "/api/sessions/"+v.ID+"/actions", map[string]string{"title": "work", "detail": "do project work"})
	var action soAction
	if json.Unmarshal(raw, &action) != nil || action.ID == "" {
		t.Fatal("action not created")
	}
	soCall(t, s, p, "PATCH", "/api/actions/"+action.ID, map[string]string{"status": "ready"})
	soCall(t, s, p, "POST", "/api/actions/"+action.ID+"/dispatch", nil)
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("action not started")
	}
	p.cancel() // Closing SaySo/browser must not cancel the delegated Web task.
	close(release)
	soWait(t, s, func() bool { return s.state.Actions[0].Status == "done" })
	run := s.state.Runs[0]
	web, err := s.a.historyStore().get(run.WebSessionID)
	if err != nil || web.Kind != "web" || web.Source != "sayso" || web.ActionID != action.ID || web.ParentSessionID != v.ID || web.ServiceID != "demo" || web.TaskStatus != "done" {
		t.Fatal("action not linked to completed new Web session", web, err)
	}
}
