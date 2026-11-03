package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func runCoordination(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("use assistants | work dispatch/read/reply/cancel")
	}
	operation := "list"
	rest := args[1:]
	hostMode := args[0] == "host"
	if hostMode {
		if len(rest) == 0 || (rest[0] != "show" && rest[0] != "set") {
			return errors.New("use host show | host set --client CLIENT_ID")
		}
		operation, rest = rest[0], rest[1:]
	}
	if args[0] == "work" {
		if len(rest) == 0 {
			return errors.New("use work dispatch/read/reply/cancel --help")
		}
		operation, rest = rest[0], rest[1:]
		if operation != "dispatch" && operation != "read" && operation != "reply" && operation != "cancel" {
			return errors.New("unknown work operation")
		}
	}
	f := flag.NewFlagSet("coordination", flag.ContinueOnError)
	path := f.String("config", agent.DefaultPath(), "local config")
	scope := f.String("scope", "project", "account, client or project")
	client := f.String("client", "", "target client ID")
	project := f.String("project", "", "target original project ID")
	sourceProject := f.String("source-project", "", "real source project ID")
	session := f.String("session", "", "real source session ID")
	work := f.String("work", "", "work ID")
	request := f.String("request", "", "stable request ID (required for dispatch/reply)")
	prompt := f.String("prompt", "", "task or follow-up")
	title := f.String("title", "", "title")
	task := f.String("task", "", "expected current task ID for cancel")
	if e := f.Parse(rest); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return errors.New("unexpected arguments")
	}
	c, e := agent.Load(*path)
	if e != nil {
		return e
	}
	method, url := "GET", "/api/agent/assistants"
	var body any
	if hostMode {
		url = "/api/agent/default-host"
		if operation == "set" {
			if *client == "" {
				return errors.New("host set requires --client")
			}
			method = "PUT"
			body = map[string]string{"agentId": *client}
		}
	} else if operation != "list" {
		if *session == "" || *sourceProject == "" {
			return errors.New("real --source-project and --session are required")
		}
		if (operation == "dispatch" || operation == "reply") && *request == "" {
			return errors.New("--request is required; reuse the same ID when retrying")
		}
		method, url = "POST", "/api/agent/coordination"
		body = protocol.CoordinationCommand{Operation: operation, RequestID: *request, WorkID: *work, Prompt: *prompt, Title: *title, TaskID: *task, Source: protocol.AssistantRef{ClientID: c.ID, ProjectID: *sourceProject, SessionID: *session}, Target: protocol.AssistantRef{Scope: *scope, ClientID: *client, ProjectID: *project}}
	}
	data, e := agent.ClientAPI(ctx, c, method, url, body)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintln(out, string(data))
	return e
}
