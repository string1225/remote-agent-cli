package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runWebProxy(args []string) error {
	action := "status"
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		action, args = args[0], args[1:]
	}
	if action != "on" && action != "off" && action != "status" && action != "run" {
		return errors.New("使用 web-proxy on/off/status/run")
	}
	flags := newCLIFlagSet("web-proxy")
	path := flags.String("config", agent.DefaultPath(), "客户端配置文件")
	listen := flags.String("listen", agent.DefaultWebProxyListen, "Web 代理监听地址（仅限本机回环 IP）")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("web-proxy 不接受位置参数")
	}
	cfg, err := agent.Load(*path)
	if err != nil {
		return err
	}
	if action == "status" {
		address, err := agent.WebProxyAddress(*path)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"enabled": address != "", "listen": address, "server": cfg.Server})
	}
	if action == "on" || action == "off" {
		address := ""
		if action == "on" {
			address = *listen
		}
		backup, err := agent.SetWebProxy(*path, address)
		if err != nil {
			return err
		}
		if backup != "" {
			fmt.Printf("原配置备份：%s\n", backup)
		}
		fmt.Printf("Web 代理配置：%s；运行中的客户端约 1 秒内生效，客户端未运行时需先启动。\n", action)
		return nil
	}
	listener, err := agent.ListenWebProxy(*listen)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	fmt.Printf("Web 代理：http://%s；GitHub/ChatGPT 经 %s 中转，其他请求直连；Ctrl+C 停止。\n", listener.Addr(), cfg.Server)
	return agent.ServeWebProxy(ctx, cfg, listener)
}
