package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/gitauth"
	"github.com/string1225/remote-agent-cli/internal/i18n"
)

func gitHelp(output io.Writer) {
	fmt.Fprintln(output, tr("Windows background GitHub authorization (separate from desktop Git login):"))
	fmt.Fprintln(output, "  remote-agent git login [--config PATH] [--username ACCOUNT] [--force]")
	fmt.Fprintln(output, "  remote-agent git status [--config PATH]")
	fmt.Fprintln(output, "  remote-agent git logout --username ACCOUNT [--config PATH]")
}

func runGitCommand(args []string, output io.Writer, invoke func(context.Context, string, gitauth.Request, io.Writer) error) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		gitHelp(output)
		return nil
	}
	flags := newCLIFlagSet("git " + args[0])
	flags.SetOutput(output)
	config := flags.String("config", agent.DefaultPath(), tr("enrolled device configuration"))
	username := flags.String("username", "", tr("GitHub account (required for logout)"))
	force := flags.Bool("force", false, tr("renew GitHub device authorization (login only)"))
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			gitHelp(output)
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New(tr("unexpected Git authorization arguments"))
	}
	request := gitauth.Request{Version: gitauth.Version, Action: args[0], Username: *username, Force: *force}
	if err := request.Validate(); err != nil {
		return err
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	limit := 30 * time.Second
	if request.Action == "login" {
		limit = 15 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	return invoke(ctx, path, request, i18n.Writer(output))
}

func gitAuthReminder(output io.Writer, goos, config string, noStart bool) {
	if goos != "windows" {
		return
	}
	quoted := "'" + strings.ReplaceAll(config, "'", "''") + "'"
	fmt.Fprintln(output, tr("GitHub authorization for background tasks: desktop Git credentials are not shared with the Windows service."))
	fmt.Fprintln(output, tr("After configuring and starting the Windows service, authorize GitHub before using private repositories or pushing:"))
	if noStart {
		fmt.Fprintln(output, tr("If the service is disabled, start it from an administrator terminal first, then authorize from the device owner's terminal:"))
		fmt.Fprintf(output, "  remote-agent start --config %s\n", quoted)
	}
	fmt.Fprintf(output, "  remote-agent git login --config %s\n  remote-agent git status --config %s\n", quoted, quoted)
	fmt.Fprintln(output, tr("Login displays a device code to approve in your browser. Credentials stay on this client and remain available after sign-out or restart."))
}
