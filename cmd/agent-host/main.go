package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/string1225/remote-agent-cli/internal/i18n"
	"github.com/string1225/remote-agent-cli/internal/servicehost"
)

func main() {
	if err := servicehost.Run(os.Args[1:]); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, i18n.Error(err))
		os.Exit(1)
	}
}
