package access

import (
	"context"
	"encoding/json"
	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testManager(t *testing.T) (*Manager, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(ctx, filepath.Join(base, "access"), gitauth.Manager{})
	if err != nil {
		t.Fatal(err)
	}
	return m, cancel
}
func waitGitHub(t *testing.T, m *Manager, status string) GitHubState {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		state := m.github
		m.mu.Unlock()
		if state.Status == status {
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("did not reach %s", status)
	return GitHubState{}
}
func TestGitHubSurvivesRequestDisconnectAndSavesOnlyPublicState(t *testing.T) {
	m, _ := testManager(t)
	approved := make(chan struct{})
	logins := 0
	m.runGit = func(ctx context.Context, r gitauth.Request, w io.Writer) error {
		if r.Action == "status" {
			_, err := io.WriteString(w, "fixture-user\n")
			return err
		}
		logins++
		_, _ = io.WriteString(w, "Visit https://github.com/login/device A1B2-")
		_, _ = io.WriteString(w, "C3D4\nmock-sensitive-output-never-persisted\n")
		select {
		case <-approved:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	request, cancel := context.WithCancel(context.Background())
	raw := json.RawMessage(`{"action":"github.start"}`)
	result, err := m.Call(request, raw)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	state := result.(GitHubState)
	if state.Code != "A1B2-C3D4" || state.Status != "waiting" {
		t.Fatalf("missing device grant: %+v", state)
	}
	second, err := m.startGitHub(Request{})
	if err != nil || second.ID != state.ID || logins != 1 {
		t.Fatal("duplicate grant started")
	}
	close(approved)
	saved := waitGitHub(t, m, "authorized")
	if len(saved.Accounts) != 1 || saved.Code != "" {
		t.Fatal("authorization was not verified")
	}
	data, err := os.ReadFile(filepath.Join(m.root, "github.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "mock-sensitive") || strings.Contains(string(data), "A1B2-C3D4") {
		t.Fatal("completed grant leaked output or code")
	}
}
func TestGitHubCancellationAndRestartAreExplicit(t *testing.T) {
	m, _ := testManager(t)
	m.runGit = func(ctx context.Context, _ gitauth.Request, w io.Writer) error {
		_, _ = io.WriteString(w, "A1B2-C3D4")
		<-ctx.Done()
		return ctx.Err()
	}
	if _, err := m.startGitHub(Request{}); err != nil {
		t.Fatal(err)
	}
	restarted, err := New(context.Background(), m.root, gitauth.Manager{})
	if err != nil {
		t.Fatal(err)
	}
	if restarted.github.Status != "interrupted" || restarted.github.Code != "" {
		t.Fatal("stale grant presented as active")
	}
	if _, err = m.Call(context.Background(), json.RawMessage(`{"action":"github.cancel"}`)); err != nil {
		t.Fatal(err)
	}
	waitGitHub(t, m, "cancelled")
}
func TestCredentialRequestsRejectUnknownFieldsWithoutEchoingSecrets(t *testing.T) {
	m, _ := testManager(t)
	for _, raw := range []string{`{"action":"servers.add","password":"mock-secret","command":"anything"}`, `{"action":"arbitrary","password":"mock-secret"}`} {
		_, err := m.Call(context.Background(), json.RawMessage(raw))
		if err == nil || strings.Contains(err.Error(), "mock-secret") {
			t.Fatal("invalid request accepted or echoed")
		}
	}
}
