package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

const codexStoreURI = "ms-windows-store://pdp/?ProductId=9PLM9XGG6VKS"

func appCommand(goos, action string) (string, []string, error) {
	if action != "install" && action != "upgrade" {
		return "", nil, errors.New("use app install or app upgrade")
	}
	switch goos {
	case "windows":
		return "powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", "Start-Process -FilePath '" + codexStoreURI + "'"}, nil
	case "darwin":
		return "brew", []string{action, "--cask", "codex-app"}, nil
	default:
		return "", nil, errors.New("Codex App installation is supported here on Windows and macOS; use codex install on Linux")
	}
}

func runApp(args []string) error {
	if len(args) != 1 {
		return errors.New("use app install or app upgrade")
	}
	name, arguments, err := appCommand(runtime.GOOS, args[0])
	if err != nil {
		return err
	}
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required; use the official Codex App installer or its update menu", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arguments...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		fmt.Println("Opened the Codex App page. Complete installation or updates in the Microsoft Store.")
	}
	return nil
}
