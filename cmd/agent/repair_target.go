package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func confirmRepair(input io.Reader, output io.Writer, local agent.Config, target agent.RepairTarget, hostname string) bool {
	current := tr("Not bound")
	if local.ID != "" {
		current = local.Name + "（" + local.ID + "）"
	}
	fmt.Fprintf(output, tr("\nExecuting computer: %s\nOriginal local binding: %s\nTarget client: %s (%s)\nOriginal target computer: %s\nBinding or machine identity mismatch. Rebinding will disconnect the previous target computer and retire this computer's original binding. Local configuration will be backed up; projects, sessions and recordings are retained. Files on the previous target computer will not be transferred. Cross-account/site robot configuration stays in the backup and needs reconfiguration.\nTo bind this computer to the target client, type BIND and press Enter; any other input cancels: "), hostname, current, target.Name, target.ID, target.Hostname)
	line, err := bufio.NewReader(input).ReadString('\n')
	return err == nil && (strings.TrimSpace(line) == "BIND" || strings.TrimSpace(line) == "绑定")
}

func runTargetRepair(server, id, path string, prompt, stdin, force, noStart bool) error {
	if len(id) > 200 || strings.ContainsAny(id, "\r\n\t ") {
		return errors.New(tr("Invalid target client ID"))
	}
	token := ""
	if prompt || stdin {
		var err error
		token, err = readSetupToken(prompt, stdin)
		if err != nil {
			return err
		}
	}
	machine, err := agent.MachineIdentity(path)
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	client := &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	local, err := agent.Load(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var config agent.Config
	backup := ""
	if _, err = os.Stat(path + ".claim-pending.json"); err == nil {
		config, backup, err = agent.ClaimRepair(client, server, id, token, path, machine, hostname, true)
	} else if !os.IsNotExist(err) {
		return err
	} else {
		var target agent.RepairTarget
		target, err = agent.InspectRepair(client, server, id, token, machine, local)
		if err != nil {
			return err
		}
		matched := local.ID == id && strings.TrimSuffix(local.Server, "/") == strings.TrimSuffix(server, "/") && target.BindingValid && target.MachineMatch
		if matched {
			config = local
			fmt.Printf(tr("Matched target client %s (%s); keeping local device credentials.\n"), target.Name, id)
		} else {
			needsConfirmation := target.Enrolled || local.ID != ""
			if needsConfirmation && !force {
				console, e := repairConsole()
				if e != nil {
					return errors.New(tr("Binding mismatch; configuration was not changed. Confirm in an interactive terminal or explicitly pass --force-rebind"))
				}
				ok := confirmRepair(console, os.Stderr, local, target, hostname)
				console.Close()
				if !ok {
					return errors.New(tr("Rebinding cancelled; existing binding and configuration were preserved"))
				}
			}
			config, backup, err = agent.ClaimRepair(client, server, id, token, path, machine, hostname, needsConfirmation)
		}
	}
	if err != nil {
		return err
	}
	if backup != "" {
		fmt.Printf(tr("Original binding configuration backed up to %s\n"), backup)
	}
	fmt.Printf(tr("Repairing service for target client %s (%s).\n"), config.Name, config.ID)
	tokenErr := configureBoundCodex(config)
	if runtime.GOOS == "windows" {
		if err = agent.SetupWindowsService(path, noStart); err != nil {
			return err
		}
	} else if !noStart {
		if err = agent.Autostart("install", path); err != nil {
			return err
		}
	}
	return tokenErr
}
