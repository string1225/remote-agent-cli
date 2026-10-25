package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func runCodex(args []string) error {
	if len(args) == 0 || (args[0] != "setup" && args[0] != "install" && args[0] != "update") {
		return errors.New(tr("use codex install/update [--server URL] or codex setup [--config PATH] [--server URL --token-stdin]"))
	}
	flags := newCLIFlagSet("codex " + args[0])
	configPath := flags.String("config", agent.DefaultPath(), tr("enrolled device configuration"))
	serverURL := flags.String("server", "", tr("platform URL for personal-token setup"))
	tokenStdin := flags.Bool("token-stdin", false, tr("read a personal model token from stdin, without device enrollment"))
	tokenPrompt := flags.Bool("token-prompt", false, tr("prompt for a personal model token without echo or shell history"))
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New(tr("unexpected setup arguments"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 25 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var connection codexsetup.Connection
	server := strings.TrimSuffix(*serverURL, "/")
	if args[0] != "setup" {
		if server == "" {
			if config, err := agent.Load(*configPath); err == nil {
				server = config.Server
			} else {
				server = "https://codex.sunny-string.cn/agents"
			}
		}
		if err := agent.ValidateServer(server); err != nil {
			return err
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path, err := codexsetup.Install(ctx, server+"/install/codex-release", filepath.Join(home, ".remote-agent", "bin"), runtime.GOOS, runtime.GOARCH, client)
		if err != nil {
			return err
		}
		fmt.Printf(tr("Official Codex CLI installed at %s\n"), path)
		if err := installRemoteAgentSkill(); err != nil {
			return err
		}
		return nil
	}
	if *tokenStdin || *tokenPrompt {
		if err := agent.ValidateServer(server); err != nil {
			return err
		}
		token, err := readSetupToken(*tokenPrompt, *tokenStdin)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(token, "ra_codex_") || len(token) < 40 || strings.ContainsAny(token, "\r\n") {
			return errors.New(tr("supply a personal token created in /agents"))
		}
		connection, err = codexsetup.PersonalConnection(client, server, token)
		if err != nil {
			return err
		}
		fmt.Printf(tr("Codex Token belongs to platform account %s (%s) at %s; source: supplied personal Token.\n"), connection.AccountUsername, connection.AccountID, server)
	} else {
		config, err := agent.Load(*configPath)
		if err != nil {
			return err
		}
		if server != "" && server != config.Server {
			return errors.New(tr("setup server differs from enrolled server"))
		}
		server = config.Server
		connection, err = codexsetup.DeviceConnection(client, server, config.ID, config.Credential)
		if err != nil {
			return err
		}
		if connection.AccountID == "" || connection.AccountUsername == "" {
			return errors.New(tr("server did not return the token account; update the server and retry"))
		}
		fmt.Printf(tr("Codex Token belongs to platform account %s (%s) at %s; source: device binding.\n"), connection.AccountUsername, connection.AccountID, server)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	bin := filepath.Join(home, ".remote-agent", "bin")
	if !agent.Detect()["codex"].Available {
		fmt.Println(tr("Installing official Codex CLI (SHA-256 verified)..."))
		if _, err := codexsetup.Install(ctx, server+"/install/codex-release", bin, runtime.GOOS, runtime.GOARCH, client); err != nil {
			return err
		}
	}
	codexHome, err := codexsetup.Home()
	if err != nil {
		return err
	}
	if err := installRemoteAgentSkill(); err != nil {
		return err
	}
	backup, err := codexsetup.Apply(codexHome, connection)
	if err != nil {
		return err
	}
	if backup != "" {
		fmt.Printf(tr("Previous configuration backed up to %s\n"), backup)
	}
	fmt.Printf(tr("Codex configured for %s. Restart Codex App and create a new local chat.\n"), connection.BaseURL)
	return nil
}
