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
	"github.com/string1225/remote-agent-cli/internal/i18n"
)

func main() {
	if err := run(); err != nil {
		log.Print(i18n.Error(err))
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) == 3 && os.Args[1] == "internal-worker" {
		input, output, closePipes, err := agent.WorkerControlPipes()
		if err != nil {
			return err
		}
		defer closePipes()
		return agent.RunWorker(context.Background(), os.Args[2], input, output)
	}
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if handled, err := frameworkUpdate(context.Background(), os.Args[2:], i18n.Writer(os.Stdout)); handled {
			return err
		}
	}
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println("github-proxy on/off/status/run [--config PATH] [--listen 127.0.0.1:17891]：本机 GitHub HTTPS 代理（默认关闭）。")
		fmt.Println(tr("remote-agent: connect --server URL --token-prompt | enroll | run | start | stop | status | version | mcp | update [--check] | service profile/configure/status | autostart install/remove | codex install/update/setup [--server URL --token-prompt] | app install/upgrade\nUse <command> --help for options.\nWindows: install the MSI first, then run a token/setup command in a new PowerShell. The MSI installs PATH and an automatic Windows Service."))
		return nil
	}
	if os.Args[1] == "codex" {
		return runCodex(os.Args[2:])
	}
	if os.Args[1] == "github-proxy" {
		return runGitHubProxy(os.Args[2:])
	}
	if os.Args[1] == "connect" {
		return runConnect(os.Args[2:])
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
			return errors.New(tr("use autostart install/remove"))
		}
		action = args[0]
		args = args[1:]
	}
	if os.Args[1] == "mcp" {
		return runMCP(context.Background(), os.Args[2:], os.Stdin, os.Stdout)
	}
	flags := newCLIFlagSet(os.Args[1])
	config := flags.String("config", agent.DefaultPath(), tr("configuration file"))
	server := flags.String("server", "", tr("HTTPS control server URL, optionally including a base path"))
	stdin := flags.Bool("token-stdin", false, tr("read one-time enrollment token from stdin"))
	roots := flags.String("roots", "", tr("allowed workspace roots separated by the OS path-list separator (default: home)"))
	write := flags.Bool("allow-write", true, tr("compatibility option; all sessions use full authorization"))
	logPath := flags.String("log", "", tr("write operational logs to this file"))
	check := flags.Bool("check", false, tr("check the published build without downloading or restarting (update only)"))
	installer := flags.Bool("installer", false, tr("upgrade the standard installation using this downloaded installer binary (update only)"))
	preserveStart := flags.Bool("preserve-start", false, tr("do not enable a stopped service during installer updates"))
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 0 {
		return errors.New(tr("unexpected positional arguments"))
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf(tr("remote-agent %s (%s/%s)\n"), buildinfo.Current(), runtime.GOOS, runtime.GOARCH)
		return nil
	case "update":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return agent.Update(ctx, path, agent.UpdateOptions{CheckOnly: *check, Installer: *installer, PreserveStart: *preserveStart, Server: *server}, i18n.Writer(os.Stdout))
	case "enroll":
		if !*stdin {
			return errors.New(tr("supply --token-stdin to read the one-time token from stdin"))
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
		fmt.Printf(tr("Bound device %s (%s)\n"), c.Name, c.ID)
		if err := configureBoundCodex(c); err != nil {
			return fmt.Errorf(tr("Device binding retained; automatic Codex setup failed: %v. Retry remote-agent codex setup --config %s.\n"), err, path)
		}
		return nil
	case "run":
		c, err := agent.Load(path)
		if err != nil {
			return err
		}
		unlock, err := agent.LockRun(path)
		if errors.Is(err, agent.ErrAgentRunning) {
			fmt.Fprintln(os.Stdout, tr("Agent already running; keeping the existing process."))
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
		return errors.New(tr("unknown command; use --help to list commands"))
	}
}
