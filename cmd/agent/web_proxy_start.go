package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func configureWebProxyListener(ctx context.Context, cfg agent.Config, preferred string, out io.Writer) (string, error) {
	if preferred == "" {
		var err error
		preferred, err = agent.WebProxyAddress(cfg.Path)
		if err != nil {
			return "", err
		}
		if preferred == "" {
			preferred = agent.DefaultWebProxyListen
		}
	}
	address, err := agent.SelectWebProxyAddress(ctx, cfg, preferred)
	if err != nil {
		return "", err
	}
	for attempt := 0; ; attempt++ {
		if address != preferred {
			if _, port, _ := net.SplitHostPort(preferred); port == "0" {
				fmt.Fprintf(out, "Web 代理自动切换到系统分配的空闲端口 %s。\n", address)
			} else {
				fmt.Fprintf(out, "Web 代理监听地址 %s 被占用或不可绑定，自动切换到空闲端口 %s。\n", preferred, address)
			}
		}
		backup, err := agent.SetWebProxy(cfg.Path, address)
		if err != nil {
			return "", err
		}
		if backup != "" {
			fmt.Fprintf(out, "原配置备份：%s\n", backup)
		}
		waitCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		err = agent.WaitWebProxy(waitCtx, cfg, address)
		cancel()
		if err == nil {
			return address, nil
		}
		if ctx.Err() != nil || attempt == 2 {
			return "", err
		}
		// Selection cannot reserve a port across the CLI/service boundary.
		// Retry only if a competing bind made it unavailable, rather than
		// rotating ports when the service is stopped or outdated.
		next, selectErr := agent.SelectWebProxyAddress(ctx, cfg, address)
		if selectErr != nil || next == address {
			return "", err
		}
		preferred, address = address, next
	}
}
