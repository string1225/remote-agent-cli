// Package servicehost is the stable Windows installation boundary. It never
// executes downloaded business code as the privileged updater identity.
package servicehost

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
)

const ServiceName = "RemoteAgent"
const UpdaterName = "RemoteAgentUpdater"
const PipeName = `\\.\pipe\RemoteAgent.Update.v1`

type Request struct {
	Action string `json:"action"`
	Server string `json:"server,omitempty"`
	Force  bool   `json:"force,omitempty"`
}
type Event struct {
	Message string `json:"message,omitempty"`
	Done    bool   `json:"done,omitempty"`
	Error   string `json:"error,omitempty"`
}
type Profile struct {
	OwnerSID     string `json:"ownerSid"`
	UpdateServer string `json:"updateServer"`
}

func ParseUpdate(args []string) (Request, bool, error) {
	f := flag.NewFlagSet("update", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	check := f.Bool("check", false, "")
	rollback := f.Bool("rollback", false, "")
	force := f.Bool("force", false, "")
	server := f.String("server", "", "")
	config := f.String("config", "", "")
	if err := f.Parse(args); err != nil {
		return Request{}, false, err
	}
	if f.NArg() != 0 || *check && *rollback {
		return Request{}, false, errors.New("use update [--check | --rollback] [--force] [--server URL] [--config PATH]")
	}
	if *server == "" && *config != "" {
		data, err := os.ReadFile(*config)
		if err != nil {
			return Request{}, false, err
		}
		var v struct {
			Server string `json:"server"`
		}
		if err = json.Unmarshal(data, &v); err != nil {
			return Request{}, false, err
		}
		*server = v.Server
	}
	if *server != "" {
		normalized, err := clientupdate.NormalizeServer(*server)
		if err != nil {
			return Request{}, false, err
		}
		*server = normalized
	}
	action := "apply"
	if *rollback {
		action = "rollback"
	}
	return Request{Action: action, Server: *server, Force: *force}, *check, nil
}
