package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func configureBoundCodex(config agent.Config) error {
	home, err := codexsetup.Home()
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return applyBoundCodex(client, config, home, os.Stdout)
}

func applyBoundCodex(client *http.Client, config agent.Config, home string, output io.Writer) error {
	connection, err := codexsetup.DeviceConnection(client, config.Server, config.ID, config.Credential)
	if err != nil {
		return err
	}
	if connection.AccountID == "" || connection.AccountUsername == "" {
		return errors.New(tr("server did not return the token account; update the server and retry"))
	}
	backup, err := codexsetup.Apply(home, connection)
	if err != nil {
		return err
	}
	if backup != "" {
		fmt.Fprintf(output, tr("Previous configuration backed up to %s\n"), backup)
	}
	if _, err := codexsetup.InstallRemoteAgentSkill(home); err != nil {
		return fmt.Errorf("安装 remote-agent skill 失败：%w", err)
	}
	fmt.Fprintf(output, tr("Codex Token belongs to platform account %s (%s) at %s; source: device binding.\n"), connection.AccountUsername, connection.AccountID, config.Server)
	return nil
}
