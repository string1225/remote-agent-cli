package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	gitPath, _ := exec.LookPath("git")
	gitProxy := systemproxy.GitManager{Git: gitPath, Path: filepath.Join(home, ".remote-agent", "web-proxy-git.json")}
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
		status["gitProxySupported"] = gitPath != ""
		if gitPath != "" {
			gitState, err := gitProxy.Status(ctx, address)
			if err != nil {
				return err
			}
			status["gitProxyEnabled"], status["gitProxyManaged"] = gitState.Enabled, gitState.Managed
		}
		return json.NewEncoder(os.Stdout).Encode(status)
	}
	if action == "on" || action == "off" {
		address := ""
		if action == "on" {
			preferred := ""
			flags.Visit(func(f *flag.Flag) {
				if f.Name == "listen" {
					preferred = *listen
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			address, err = configureWebProxyListener(ctx, cfg, preferred, os.Stdout)
			cancel()
			if err != nil {
				return fmt.Errorf("%w；请检查客户端服务，系统代理保持原设置", err)
			}
		}
		if action == "off" {
			previous, err := agent.WebProxyAddress(*path)
			if err != nil {
				return err
			}
			if previous == "" {
				previous = agent.DefaultWebProxyListen
			}
			if gitPath != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := gitProxy.Disable(ctx, previous)
				cancel()
				if err != nil {
					return err
				}
			}
			if systemSupported {
				if err := system.Disable(previous); err != nil {
					return err
				}
			}
		}
		if action == "on" {
			probeCtx, stop := context.WithTimeout(context.Background(), 22*time.Second)
			err := agent.ProbeWebProxyTunnel(probeCtx, address, "github.com:443")
			stop()
			if err != nil {
				return err
			}
			if systemSupported {
				if err := system.Enable(address); err != nil {
					return err
				}
			}
			if gitPath != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := gitProxy.Enable(ctx, address)
				cancel()
				if err != nil {
					if systemSupported {
						return errors.Join(err, system.Disable(address))
					}
					return err
				}
				fmt.Printf("Git HTTP 代理已同步：http://%s；关闭时恢复原设置。\n", address)
			}
			if systemSupported {
				fmt.Printf("Web 代理已就绪并启用 Windows 系统代理：http://%s。关闭用 remote-agent web-proxy off，原系统设置已备份。\n", address)
			} else {
				fmt.Printf("Web 代理已就绪：http://%s；请在应用中选择该代理。\n", address)
			}
		} else {
			backup, err := agent.SetWebProxy(*path, "")
			if err != nil {
				return err
			}
			if backup != "" {
				fmt.Printf("原配置备份：%s\n", backup)
			}
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
