package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestTodoPersistenceConflictReceiptsAndMigration(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	call := func(command protocol.TodoCommand) (any, error) {
		return h.todos(protocol.TodoStorageCall{AccountID: "alice", Generation: "original", TodoCommand: command})
	}
	value, err := call(protocol.TodoCommand{Operation: "create", RequestID: "request", Title: "选择部署窗口", Body: "周五还是周六？", ClientID: "client", ProjectID: "project", SessionID: "session"})
	if err != nil {
		t.Fatal(err)
	}
	row := value.(protocol.Todo)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, title := range []string{"周五", "周六"} {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			_, e := call(protocol.TodoCommand{Operation: "update", ID: row.ID, Version: row.Version, Title: title, Body: "test"})
			results <- e
		}(title)
	}
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for e := range results {
		if e == nil {
			success++
		} else if errors.Is(e, errTodoConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatal("concurrent edits overwrote one another", success, conflicts)
	}
	value, err = call(protocol.TodoCommand{Operation: "reply", ID: row.ID, Version: 2, RequestID: "reply", Message: "周五发布"})
	if err != nil {
		t.Fatal(err)
	}
	row = value.(protocol.Todo)
	if _, err = call(protocol.TodoCommand{Operation: "complete", ID: row.ID, Version: row.Version}); err == nil {
		t.Fatal("pending delivery was lost")
	}
	value, err = call(protocol.TodoCommand{Operation: "reply", ID: row.ID, Version: 1, RequestID: "reply", Message: "周五发布"})
	if err != nil || value.(protocol.Todo).Version != row.Version {
		t.Fatal("lost-ack retry failed", err)
	}
	if _, err = call(protocol.TodoCommand{Operation: "reply", ID: row.ID, RequestID: "reply", Message: "different"}); err == nil {
		t.Fatal("same reply ID changed its payload")
	}
	value, err = call(protocol.TodoCommand{Operation: "ack", ID: row.ID, RequestID: "reply"})
	if err != nil {
		t.Fatal(err)
	}
	row = value.(protocol.Todo)
	if row.Reply.State != "delivered" || row.Reply.TaskID != todoReplyTaskID(row.ID, "reply") {
		t.Fatal(row)
	}
	large, err := call(protocol.TodoCommand{Operation: "create", RequestID: "escaped", Title: "escaped", Body: strings.Repeat("\n", 7999), ClientID: "client", ProjectID: "project", SessionID: "session"})
	if err != nil {
		t.Fatal(err)
	}
	escaped := large.(protocol.Todo)
	if _, err = call(protocol.TodoCommand{Operation: "reply", ID: escaped.ID, Version: 1, RequestID: "oversized", Message: strings.Repeat("\n", 7999)}); err == nil {
		t.Fatal("JSON escaping exceeded migration budget")
	}
	check, err := call(protocol.TodoCommand{Operation: "get", ID: escaped.ID})
	if err != nil || check.(protocol.Todo).Reply != nil {
		t.Fatal("failed reply left a receipt or partial record", err)
	}
	for i := 0; i < 70; i++ {
		if _, err = call(protocol.TodoCommand{Operation: "create", RequestID: fmt.Sprint(i), Title: fmt.Sprint(i), Body: strings.Repeat("🙂", 1500)}); err != nil {
			t.Fatal(err)
		}
	}
	// Copy a complete generation with transport-sized pages. Former generations
	// remain intact and an unrelated account never sees their contents.
	destination := historyStore{filepath.Join(t.TempDir(), "history.db")}
	cursor := ""
	pages := 0
	for {
		value, err = call(protocol.TodoCommand{Operation: "export", Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		page := value.(protocol.TodoExport)
		pages++
		in := protocol.TodoStorageCall{AccountID: "alice", Generation: "new", TodoCommand: protocol.TodoCommand{Operation: "import"}, Records: page.Records}
		encoded, _ := json.Marshal(in)
		if len(encoded) > protocol.MaxRequestBytes-1024 {
			t.Fatal("migration exceeds request budget", len(encoded))
		}
		if _, err = destination.todos(in); err != nil {
			t.Fatal(err)
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if pages < 3 {
		t.Fatal("fixture did not exercise pagination")
	}
	count := 0
	cursor = ""
	for {
		v, e := destination.todos(protocol.TodoStorageCall{AccountID: "alice", Generation: "new", TodoCommand: protocol.TodoCommand{Operation: "list", Cursor: cursor}})
		if e != nil {
			t.Fatal(e)
		}
		page := v.(protocol.TodoPage)
		count += len(page.Todos)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if count != 72 {
		t.Fatal("migration dropped or duplicated todos", count)
	}
	v, err := destination.todos(protocol.TodoStorageCall{AccountID: "bob", Generation: "new", TodoCommand: protocol.TodoCommand{Operation: "list"}})
	if err != nil || len(v.(protocol.TodoPage).Todos) != 0 {
		t.Fatal("account isolation failed", err)
	}
	v, err = destination.todos(protocol.TodoStorageCall{AccountID: "alice", Generation: "new", TodoCommand: protocol.TodoCommand{Operation: "reply", ID: row.ID, RequestID: "reply", Message: "周五发布"}})
	if err != nil || v.(protocol.Todo).Reply.State != "delivered" {
		t.Fatal("migration lost idempotency receipts", err)
	}
	if todoReplyTaskID("first", "reply") == todoReplyTaskID("second", "reply") {
		t.Fatal("reply IDs collided across todos")
	}
}

func TestTodoRepliesResumeOriginalIMSessionWithoutBrowserPrivileges(t *testing.T) {
	a, project := concurrencyFixture(t)
	a.Config.ID = "client"
	ctx, cancel := context.WithCancel(context.Background())
	a.initTasks(ctx)
	t.Cleanup(func() { cancel(); <-a.taskDone; <-a.notifyDone })
	session, err := a.historyStore().create(project, "IM decision")
	if err != nil {
		t.Fatal(err)
	}
	if err = a.historyStore().update(session.ID, func(s *protocol.Conversation) { s.Kind = "im"; s.NativeID = "native-context" }); err != nil {
		t.Fatal(err)
	}
	started := make(chan string, 2)
	a.Execute = func(ctx context.Context, _ Provider, p protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		if p.ResumeID != "native-context" || p.ToolSource != "im" {
			t.Error("lost original IM context", p.ResumeID, p.ToolSource)
		}
		started <- prompt
		return nil
	}
	row := protocol.Todo{ID: "todo", ClientID: "client", ProjectID: project.ID, SessionID: session.ID, Title: "which?", Reply: &protocol.TodoReply{ID: "reply", Text: "choose A"}}
	body, _ := json.Marshal(row)
	var denied protocol.Event
	a.handle(&peer{ctx: ctx}, protocol.Request{ID: "browser", Type: "todos.deliver", Body: body}, func(e protocol.Event) error { denied = e; return nil })
	if denied.Type != "error" {
		t.Fatal("untrusted browser invoked an account tool")
	}
	for i := 0; i < 2; i++ {
		if _, err = a.handleTodos(protocol.Request{Type: "todos.deliver", Body: body}); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case text := <-started:
		if !strings.Contains(text, "choose A") {
			t.Fatal(text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reply not executed")
	}
	waitSessionStatus(t, a, session.ID, "done")
	select {
	case <-started:
		t.Fatal("retry executed twice")
	default:
	}
}

func TestTodoFreezeFencesDelayedWritesAndSurvivesRollback(t *testing.T) {
	h := historyStore{filepath.Join(t.TempDir(), "history.db")}
	in := protocol.TodoStorageCall{AccountID: "alice", Generation: "canonical", MigrationID: "migration"}
	call := func(op string) (any, error) { in.Operation = op; return h.todos(in) }
	if _, err := call("freeze"); err != nil {
		t.Fatal(err)
	}
	in.RequestID, in.Title = "late-create", "must wait"
	if _, err := call("create"); err == nil {
		t.Fatal("a delayed write entered the frozen snapshot")
	}
	if _, err := call("unfreeze"); err != nil {
		t.Fatal(err)
	}
	if _, err := call("freeze"); err == nil {
		t.Fatal("a delayed freeze undid rollback")
	}
	if _, err := call("create"); err != nil {
		t.Fatal("canonical host did not recover", err)
	}
	in.MigrationID = "second-migration"
	if _, err := call("freeze"); err != nil {
		t.Fatal(err)
	}
	value, err := call("export")
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range value.(protocol.TodoExport).Records {
		if !strings.HasPrefix(record.Key, "t/") && !strings.HasPrefix(record.Key, "r/") {
			t.Fatal("migration state leaked into new canonical generation", record.Key)
		}
	}
}
