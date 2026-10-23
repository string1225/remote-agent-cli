package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runService(args []string) error {
	if len(args) == 0 {
		return errors.New("use service run/profile/configure/status")
	}
	flags := flag.NewFlagSet("service "+args[0], flag.ContinueOnError)
	config := flags.String("config", agent.DefaultPath(), "existing device configuration (data is kept in place)")
	profile := flags.String("profile", "", "profile captured in the original user's environment")
	output := flags.String("output", "", "output file for service profile")
	noStart := flags.Bool("no-start", false, "configure but disable and stop the service")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected service arguments")
	}
	switch args[0] {
	case "run":
		return agent.RunWindowsService()
	case "prepare":
		return agent.PrepareWindowsService()
	case "status":
		status, err := agent.WindowsServiceStatus()
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(status)
	case "profile", "configure":
		var p agent.ServiceProfile
		var err error
		if *profile != "" {
			p, err = agent.ReadServiceProfile(*profile)
		} else {
			p, err = agent.CaptureServiceProfile(*config)
		}
		if err != nil {
			return err
		}
		if args[0] == "profile" {
			if *output == "" {
				return errors.New("service profile requires --output")
			}
			return p.Save(*output)
		}
		return agent.ConfigureWindowsService(p, *noStart)
	default:
		return errors.New("use service run/profile/configure/status")
	}
}
