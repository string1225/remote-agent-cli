package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runConnect(args []string) error {
	flags := newCLIFlagSet("connect")
	server := flags.String("server", "https://codex.sunny-string.cn/agents", tr("platform URL"))
	config := flags.String("config", agent.DefaultPath(), tr("device configuration"))
	stdin := flags.Bool("token-stdin", false, tr("read the one-time device token from stdin"))
	prompt := flags.Bool("token-prompt", false, tr("prompt for a one-time device token without echo"))
	roots := flags.String("roots", "", tr("allowed project roots separated by the OS path separator"))
	write := flags.Bool("allow-write", true, tr("compatibility option; all sessions use full authorization"))
	noStart := flags.Bool("no-start", false, tr("bind the device without starting remote access"))
	rebind := flags.String("rebind", "", tr("recover this exact device ID using an owner-issued recovery token"))
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New(tr("unexpected connection arguments"))
	}
	if err := agent.ValidateServer(*server); err != nil {
		return err
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if _, err = agent.WindowsServiceStatus(); err != nil {
			return fmt.Errorf(tr("Install Remote Agent MSI first: %w"), err)
		}
	}
	if *rebind != "" {
		return runRebindConnect(*server, *rebind, path, *roots, *prompt, *stdin, *noStart)
	}
	if _, err = os.Stat(path); err == nil {
		existing, loadErr := agent.Load(path)
		if loadErr != nil {
			return loadErr
		}
		if strings.TrimSuffix(*server, "/") != existing.Server {
			return errors.New(tr("Existing device binding belongs to another server; nothing was changed"))
		}
		tokenErr := configureBoundCodex(existing)
		if tokenErr != nil {
			fmt.Fprintf(os.Stderr, tr("Device binding retained; automatic Codex setup failed: %v. Retry remote-agent codex setup --config %s.\n"), tokenErr, path)
		}
		fmt.Println(tr("Keeping existing device binding, projects and sessions; configuring the background service."))
		if runtime.GOOS == "windows" {
			fmt.Println(tr("Granting directory access may take several minutes. Keep this window open until setup finishes."))
		}
		if runtime.GOOS == "windows" {
			if err := agent.SetupWindowsService(path, *noStart); err != nil {
				return err
			}
			return tokenErr
		}
		if !*noStart {
			if err := agent.Autostart("install", path); err != nil {
				return err
			}
		}
		return tokenErr
	} else if !os.IsNotExist(err) {
		return err
	}
	token, err := readSetupToken(*prompt, *stdin)
	if err != nil {
		return err
	}
	var allowed []string
	if *roots != "" {
		allowed = filepath.SplitList(*roots)
	}
	c, err := agent.Enroll(strings.TrimSuffix(*server, "/"), token, path, allowed, *write)
	if err != nil {
		return err
	}
	fmt.Printf(tr("Bound client %s.\n"), c.Name)
	tokenErr := configureBoundCodex(c)
	if tokenErr != nil {
		fmt.Fprintf(os.Stderr, tr("Device binding retained; automatic Codex setup failed: %v. Retry remote-agent codex setup --config %s.\n"), tokenErr, path)
	}
	if runtime.GOOS == "windows" {
		fmt.Println(tr("Granting directory access may take several minutes. Keep this window open until setup finishes."))
		if err = agent.SetupWindowsService(path, *noStart); err != nil {
			return fmt.Errorf(tr("Device credentials saved, but service configuration failed: %w. Retry with remote-agent service configure --config <config-path>"), err)
		}
	} else if !*noStart {
		err = agent.Autostart("install", path)
	}
	if err != nil {
		return err
	}
	fmt.Println(tr("Client configured with full session authorization; use remote-agent status to check."))
	return tokenErr
}
