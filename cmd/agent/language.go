package main

import (
	"flag"
	"os"

	"github.com/string1225/remote-agent-cli/internal/i18n"
)

var tr = i18n.Text

func newCLIFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(i18n.Writer(os.Stderr))
	return flags
}
