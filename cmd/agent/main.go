package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/buildinfo"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println("remote-agent: enroll | run | start | stop | status | version | update [--check] | service profile/configure/status | autostart install/remove | codex install/update/setup | app install/upgrade\nUse <command> --help for options.\nWindows MSI uses a machine service; service configuration and start/stop require an administrator.")
		return nil
	}
	if os.Args[1] == "codex" {
		return runCodex(os.Args[2:])
	}
	if os.Args[1] == "app" {
		return runApp(os.Args[2:])
	}
	if os.Args[1] == "service" {
		return runService(os.Args[2:])
	}
	args := os.Args[2:]
	action := ""
	if os.Args[1] == "autostart" {
		if len(args) == 0 {
			return errors.New("use autostart install/remove")
		}
		action = args[0]
		args = args[1:]
	}
	flags := flag.NewFlagSet(os.Args[1], flag.ContinueOnError)
	config := flags.String("config", agent.DefaultPath(), "configuration file")
	server := flags.String("server", "", "HTTPS control server URL, optionally including a base path")
	stdin := flags.Bool("token-stdin", false, "read one-time enrollment token from stdin")
	roots := flags.String("roots", "", "allowed workspace roots separated by the OS path-list separator (default: home)")
	write := flags.Bool("allow-write", false, "allow Codex workspace writes / Qoder file edits (enroll only)")
	logPath := flags.String("log", "", "write operational logs to this file")
	check := flags.Bool("check", false, "check the published build without downloading or restarting (update only)")
	installer := flags.Bool("installer", false, "upgrade the standard installation using this downloaded installer binary (update only)")
	preserveStart := flags.Bool("preserve-start", false, "do not enable a stopped service during installer updates")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return errors.New("unexpected positional arguments")
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("remote-agent %s (%s/%s)\n", buildinfo.Current(), runtime.GOOS, runtime.GOARCH)
		return nil
	case "update":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return agent.Update(ctx, path, agent.UpdateOptions{CheckOnly: *check, Installer: *installer, PreserveStart: *preserveStart, Server: *server}, os.Stdout)
	case "enroll":
		if !*stdin {
			return errors.New("supply --token-stdin to read the one-time token from stdin")
		}
		token, err := io.ReadAll(io.LimitReader(os.Stdin, 4096))
		if err != nil {
			return err
		}
		var allowed []string
		if *roots != "" {
			allowed = filepath.SplitList(*roots)
		}
		c, err := agent.Enroll(*server, strings.TrimSpace(string(token)), path, allowed, *write)
		if err != nil {
			return err
		}
		fmt.Printf("Bound device %s (%s)\n", c.Name, c.ID)
		return nil
	case "run":
		c, err := agent.Load(path)
		if err != nil {
			return err
		}
		unlock, err := agent.LockRun(path)
		if errors.Is(err, agent.ErrAgentRunning) {
			fmt.Fprintln(os.Stdout, "Agent already running; keeping the existing process.")
			return nil
		}
		if err != nil {
			return err
		}
		defer unlock()
		if *logPath != "" {
			if st, err := os.Stat(*logPath); err == nil && st.Size() > 10*1024*1024 {
				_ = os.Rename(*logPath, *logPath+".old")
			}
			f, err := os.OpenFile(*logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return err
			}
			previous := log.Writer()
			defer func() { log.SetOutput(previous); f.Close() }()
			log.SetOutput(f)
		}
		log.Printf("agent starting (version %s, pid %d)", buildinfo.Current(), os.Getpid())
		if err := agent.RepairAutostart(path); err != nil {
			log.Printf("cannot repair autostart recovery policy: %v", err)
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		err = agent.New(c).Run(ctx)
		if err != nil {
			log.Printf("agent stopped: %v", err)
		} else {
			log.Print("agent stopped after shutdown signal")
		}
		return err
	case "status":
		c, err := agent.Load(path)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"id": c.ID, "name": c.Name, "server": c.Server, "allowedRoots": c.AllowedRoots, "allowWrite": c.AllowWrite, "services": c.Services, "providers": agent.Detect()})
	case "autostart":
		return agent.Autostart(action, path)
	case "start":
		return agent.Autostart("install", path)
	case "stop":
		return agent.Autostart("remove", path)
	default:
		return errors.New("unknown command; use --help to list commands")
	}
}
