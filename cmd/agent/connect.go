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
	write := flags.Bool("allow-write", false, tr("allow project edits (default: read-only)"))
	noStart := flags.Bool("no-start", false, tr("bind the device without starting remote access"))
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
	if _, err = os.Stat(path); err == nil {
		existing, loadErr := agent.Load(path)
		if loadErr != nil {
			return loadErr
		}
		if strings.TrimSuffix(*server, "/") != existing.Server {
			return errors.New(tr("Existing device binding belongs to another server; nothing was changed"))
		}
		fmt.Println(tr("Keeping existing device binding, projects and sessions; configuring the background service."))
		if runtime.GOOS == "windows" {
			fmt.Println(tr("Granting directory access may take several minutes. Keep this window open until setup finishes."))
		}
		if runtime.GOOS == "windows" {
			return agent.SetupWindowsService(path, *noStart)
		}
		if !*noStart {
			return agent.Autostart("install", path)
		}
		return nil
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
	fmt.Println(tr("Client configured; use remote-agent status to check. Configure the Codex proxy token separately with remote-agent codex setup."))
	return nil
}
