package serviceaccess

import (
	"strings"
	"testing"
)

func TestOptionsRejectAmbiguousOrBroadPaths(t *testing.T) {
	for _, path := range []string{"", ".", `C:\`, `C:/`, `C:relative`, `\\server\share\repo`, `\\?\C:\repo`, `C:\repo\*`, `C:\repo\..\other`, `C:\repo\.`, `C:\repo\file:secret`, "C:\\repo\n[safe]", `C:\repo\NUL.txt`, `C:\repo\COM1`, `C:\repo\a.`, `C:\repo\a `, `C:\repo\\child`} {
		if err := (Options{Action: "grant", Path: path, Permission: "read"}).Validate(); err == nil {
			t.Errorf("accepted unsafe path %q", path)
		}
	}
	for _, path := range []string{`C:\Work\O'Brien # repo`, `D:/data/a=b`, `C:\Users\张三\资料.txt`} {
		if err := (Options{Action: "grant", Path: path, Permission: "modify"}).Validate(); err != nil {
			t.Errorf("rejected explicit path %q: %v", path, err)
		}
	}
	for _, permission := range []string{"", "full", "write", "Read"} {
		if err := (Options{Action: "grant", Path: `C:\repo`, Permission: permission}).Validate(); err == nil {
			t.Errorf("accepted permission %q", permission)
		}
	}
	for _, id := range []string{"", "../backup", `C:\backup`, strings.Repeat("f", 33)} {
		if err := (Options{Action: "restore", Backup: id}).Validate(); err == nil {
			t.Errorf("accepted backup ID %q", id)
		}
	}
}

func TestGitConfigOnlyContainsExactSelectedPaths(t *testing.T) {
	config, err := gitConfig([]string{`C:\Work\O'Brien # repo`, `D:/private/repo`})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, `directory = "C:/Work/O'Brien # repo"`) || strings.Contains(config, "directory = *") || strings.Contains(config, "credential.") {
		t.Fatal(config)
	}
	if _, err := gitConfig([]string{`C:/repo/*`}); err == nil {
		t.Fatal("wildcard trust accepted")
	}
}
