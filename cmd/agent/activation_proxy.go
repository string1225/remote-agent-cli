package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/systemproxy"
)

func applyActivationProxy(c agent.Config, on bool) error {
	backend, supported := systemproxy.Native()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	manager := systemproxy.Manager{Backend: backend, Path: filepath.Join(home, ".remote-agent", "web-proxy-system.json")}
	if !on {
		address, err := agent.WebProxyAddress(c.Path)
		if err != nil {
			return err
		}
		if address == "" {
			return nil
		}
		if supported {
			if err = manager.Disable(address); err != nil {
				return err
			}
		}
		_, err = agent.SetWebProxy(c.Path, "")
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	address, err := configureWebProxyListener(ctx, c, "", io.Discard)
	if err != nil {
		return err
	}
	if err = agent.ProbeWebProxyTunnel(ctx, address, "github.com:443"); err != nil {
		return err
	}
	if supported {
		return manager.Enable(address)
	}
	return nil
}
