package agent

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestHistoryPersistenceAndPagination(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	service := protocol.Service{ID: "project-a", Provider: "codex", Workspace: t.TempDir()}
	s, err := h.create(service, "")
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Repeat("你好🙂", 6000)
	if _, err = h.append(s.ID, protocol.ChatEntry{Role: "user", TurnID: "turn-1", Kind: "message", Text: input}); err != nil {
		t.Fatal(err)
	}
	if err = h.nativeID(s.ID, "thread-123"); err != nil {
		t.Fatal(err)
	}
	// Reopen the on-disk store as after an Agent restart.
	reopened := historyStore{h.path}
	var got strings.Builder
	var cursor uint64
	var pages int
	for {
		page, err := reopened.read(s.ID, cursor)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(page)
		if len(raw) > 32000 {
			t.Fatalf("page exceeds transport budget: %d", len(raw))
		}
		for _, e := range page.Entries {
			got.WriteString(e.Text)
		}
		pages++
		cursor = page.NextCursor
		if cursor == 0 {
			break
		}
	}
	if got.String() != input || pages < 2 {
		t.Fatal("paginated history lost Unicode content")
	}
	saved, err := reopened.get(s.ID)
	if err != nil || saved.NativeID != "thread-123" {
		t.Fatal("native session ID not persisted")
	}
	page, err := reopened.list("project-a", "你好", 0)
	if err != nil || len(page.Sessions) != 1 || page.Sessions[0].NativeID != "" {
		t.Fatal("search or metadata filtering failed")
	}
	page, err = reopened.list("project-b", "", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatal("history crossed project scope")
	}
	if _, err = reopened.read("../config.json", 0); err == nil {
		t.Fatal("unknown session accepted")
	}
}

func TestTranscriptParsingAndNativeResume(t *testing.T) {
	var text strings.Builder
	var native string
	d := &transcript{text: func(_, s string) error { text.WriteString(s); return nil }, native: func(id string) error { native = id; return nil }}
	chunks := []string{`{"type":"thread.started","thread_id":"thread-123"}` + "\n", `{"type":"item.completed","item":{"id":"x","type":"agent_message","text":"你`, "好" + `"}}` + "\n"}
	for _, c := range chunks {
		if err := d.write("stdout", c); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.flush(); err != nil {
		t.Fatal(err)
	}
	if native != "thread-123" || text.String() != "你好\n" {
		t.Fatalf("incorrect transcript: %q, %q", native, text.String())
	}
	args := strings.Join(providerSessionArgs("codex", false, native), " ")
	if !strings.Contains(args, "exec resume") || !strings.Contains(args, `sandbox_mode="read-only"`) || !strings.Contains(args, `approval_policy="never"`) || strings.Contains(args, "--last") {
		t.Fatalf("unsafe or wrong resume arguments: %s", args)
	}
	args = strings.Join(providerSessionArgs("qoder", false, native), " ")
	if !strings.Contains(args, "--resume thread-123") || !strings.Contains(args, "--permission-mode plan") {
		t.Fatal(args)
	}
}
