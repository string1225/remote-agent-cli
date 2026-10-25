package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runRebindConnect(server, id, path, rootList string, prompt, stdin, noStart bool) error {
	token := ""
	if prompt || stdin {
		value, err := readSetupToken(prompt, stdin)
		if err != nil {
			return err
		}
		token = value
	}
	var roots []string
	if rootList != "" {
		roots = filepath.SplitList(rootList)
	}
	config, backup, err := agent.Rebind(server, token, path, id, roots)
	if err != nil {
		return err
	}
	if backup != "" {
		fmt.Printf(tr("Original device configuration backed up to %s\n"), backup)
	}
	fmt.Printf(tr("Recovered device %s (%s); projects and local history are retained.\n"), config.Name, config.ID)
	tokenErr := configureBoundCodex(config)
	if tokenErr != nil {
		fmt.Fprintf(os.Stderr, tr("Device binding retained; automatic Codex setup failed: %v. Retry remote-agent codex setup --config %s.\n"), tokenErr, path)
	}
	if runtime.GOOS == "windows" {
		fmt.Println(tr("Granting directory access may take several minutes. Keep this window open until setup finishes."))
		if err := agent.SetupWindowsService(path, noStart); err != nil {
			return err
		}
	} else if !noStart {
		if err := agent.Autostart("install", path); err != nil {
			return err
		}
	}
	return tokenErr
}
