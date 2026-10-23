package main

import (
	"slices"
	"strings"
	"testing"
)

func TestWindowsAppActionsOpenStore(t *testing.T) {
	for _, action := range []string{"install", "upgrade"} {
		t.Run(action, func(t *testing.T) {
			name, args, err := appCommand("windows", action)
			want := []string{"-NoProfile", "-NonInteractive", "-Command", "Start-Process -FilePath 'ms-windows-store://pdp/?ProductId=9PLM9XGG6VKS'"}
			if err != nil || name != "powershell.exe" || !slices.Equal(args, want) {
				t.Fatal("Windows app action must open its Store page", name, args, err)
			}
		})
	}
	if _, _, err := appCommand("windows", "upgrade; whoami"); err == nil {
		t.Fatal("unexpected app action accepted")
	}
}

func TestAppUpgradeTargetsOnlyCodex(t *testing.T) {
	name, args, err := appCommand("darwin", "upgrade")
	if err != nil || name != "brew" || strings.Join(args, " ") != "upgrade --cask codex-app" {
		t.Fatal("wrong macOS update", name, args, err)
	}
	if _, _, err := appCommand("linux", "upgrade"); err == nil {
		t.Fatal("unsupported app update accepted")
	}
}
