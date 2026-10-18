package main

import (
	"strings"
	"testing"
)

func TestAppUpgradeTargetsOnlyCodex(t *testing.T) {
	name, args, err := appCommand("windows", "upgrade")
	if err != nil || name != "winget" || !strings.Contains(strings.Join(args, " "), "--id 9PLM9XGG6VKS --source msstore") || strings.Contains(strings.Join(args, " "), "--all") {
		t.Fatal("unsafe Windows app update", name, args, err)
	}
	name, args, err = appCommand("darwin", "upgrade")
	if err != nil || name != "brew" || strings.Join(args, " ") != "upgrade --cask codex-app" {
		t.Fatal("wrong macOS update", name, args, err)
	}
	if _, _, err := appCommand("linux", "upgrade"); err == nil {
		t.Fatal("unsupported app update accepted")
	}
}
