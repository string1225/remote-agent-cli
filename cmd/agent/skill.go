package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func currentSkillHomes() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(home), `\system32\config\`) {
		return "", "", errors.New("skill installation needs the device owner's user profile; use skill install --home PATH")
	}
	codex, err := codexsetup.Home()
	return home, codex, err
}

func syncCurrentUserSkill() (codexsetup.UserSkillInstall, error) {
	home, codex, err := currentSkillHomes()
	if err != nil {
		return codexsetup.UserSkillInstall{}, err
	}
	result, err := codexsetup.SyncUserRemoteAgentSkill(home, codex, buildinfo.Current())
	if err == nil {
		err = agent.PrepareUserSkillServiceAccess(home)
	}
	return result, err
}

func runSkill(args []string, out io.Writer) error {
	action := "status"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action, args = args[0], args[1:]
	}
	if action != "install" && action != "update" && action != "status" {
		return errors.New("use skill install/update/status [--home PATH] [--codex-home PATH]")
	}
	flags := newCLIFlagSet("skill " + action)
	home := flags.String("home", "", "owning user's home, default current user")
	codex := flags.String("codex-home", "", "legacy Codex home to migrate, default current profile")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("skill does not accept positional arguments")
	}
	currentUser := *home == ""
	if *home == "" {
		resolved, legacy, err := currentSkillHomes()
		if err != nil {
			return err
		}
		*home = resolved
		if *codex == "" {
			*codex = legacy
		}
	} else {
		resolved, err := filepath.Abs(*home)
		if err != nil {
			return err
		}
		*home = resolved
		if *codex == "" {
			*codex = filepath.Join(*home, ".codex")
		}
	}
	legacy, err := filepath.Abs(*codex)
	if err != nil {
		return err
	}
	if action == "status" {
		status, err := codexsetup.UserRemoteAgentSkillStatus(*home, buildinfo.Current())
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(status)
	}
	result, err := codexsetup.SyncUserRemoteAgentSkill(*home, legacy, buildinfo.Current())
	if err != nil {
		return err
	}
	if currentUser {
		if err = agent.PrepareUserSkillServiceAccess(*home); err != nil {
			return err
		}
	}
	return json.NewEncoder(out).Encode(result)
}
