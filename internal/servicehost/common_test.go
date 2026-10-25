package servicehost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateArgumentsDoNotNeedEnrollmentOrForwardCredentials(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config.json")
	if err := os.WriteFile(config, []byte(`{"server":"https://example.com/agents","credential":"fixture-private"}`), 0600); err != nil {
		t.Fatal(err)
	}
	r, check, err := ParseUpdate([]string{"--check", "--config", config})
	if err != nil || !check || r.Server != "https://example.com/agents" {
		t.Fatal(r, check, err)
	}
	r, check, err = ParseUpdate([]string{"--server", "https://example.com/agents", "--force"})
	if err != nil || check || !r.Force || r.Action != "apply" {
		t.Fatal(r, check, err)
	}
	for _, args := range [][]string{{"--check", "--rollback"}, {"--server", "https://secret@example.com"}, {"--server", "https://example.com/agents?token=private"}, {"--executable", "untrusted.exe"}, {"extra"}} {
		if _, _, err := ParseUpdate(args); err == nil {
			t.Fatal("unsafe or ambiguous args accepted", args)
		}
	}
}
