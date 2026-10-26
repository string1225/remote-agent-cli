package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/systemproxy"
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
	backend, systemSupported := systemproxy.Native()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	system := systemproxy.Manager{Backend: backend, Path: filepath.Join(home, ".remote-agent", "web-proxy-system.json")}
	if action == "status" {
		address, err := agent.WebProxyAddress(*path)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		status := map[string]any{"enabled": address != "", "listen": address, "server": cfg.Server, "listening": false, "systemProxySupported": systemSupported}
		if address != "" {
			status["listening"] = agent.CheckWebProxy(ctx, cfg, address) == nil
		}
		if systemSupported {
			state, err := backend.Read()
			if err != nil {
				return err
			}
			status["systemProxyEnabled"] = state.Flags&systemproxy.Proxy != 0
			status["systemProxyManaged"] = state.Flags&systemproxy.Proxy != 0 && state.Server == address
		}
		return json.NewEncoder(os.Stdout).Encode(status)
	}
	if action == "on" || action == "off" {
		address := ""
		if action == "on" {
			address = *listen
		}
		if action == "off" && systemSupported {
			previous, err := agent.WebProxyAddress(*path)
			if err != nil {
				return err
			}
			if previous == "" {
				previous = agent.DefaultWebProxyListen
			}
			if err := system.Disable(previous); err != nil {
				return err
			}
		}
		backup, err := agent.SetWebProxy(*path, address)
		if err != nil {
			return err
		}
		if backup != "" {
			fmt.Printf("原配置备份：%s\n", backup)
		}
		if action == "on" {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			err := agent.WaitWebProxy(ctx, cfg, address)
			cancel()
			if err != nil {
				return fmt.Errorf("%w；请升级并启动客户端服务，系统代理保持原设置", err)
			}
			probeCtx, stop := context.WithTimeout(context.Background(), 22*time.Second)
			err = agent.ProbeWebProxyTunnel(probeCtx, address, "github.com:443")
			stop()
			if err != nil {
				return err
			}
			if systemSupported {
				if err := system.Enable(address); err != nil {
					return err
				}
				fmt.Printf("Web 代理已就绪并启用 Windows 系统代理：http://%s。关闭用 remote-agent web-proxy off，原系统设置已备份。\n", address)
			} else {
				fmt.Printf("Web 代理已就绪：http://%s；请在应用中选择该代理。\n", address)
			}
		} else {
			if systemSupported {
				fmt.Println("Web 代理已关闭；Windows 系统代理原设置已恢复（外部修改的设置会保留）。")
			} else {
				fmt.Println("Web 代理已关闭。")
			}
		}
		return nil
	}
	listener, err := agent.ListenWebProxy(*listen)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	fmt.Printf("Web 代理：http://%s；GitHub/ChatGPT/OpenAI/Google 经 %s 中转，其他请求直连；Ctrl+C 停止。\n", listener.Addr(), cfg.Server)
	return agent.ServeWebProxy(ctx, cfg, listener)
}
