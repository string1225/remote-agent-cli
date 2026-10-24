package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runConnect(args []string) error {
	flags := flag.NewFlagSet("connect", flag.ContinueOnError)
	server := flags.String("server", "https://codex.sunny-string.cn/agents", "platform URL")
	config := flags.String("config", agent.DefaultPath(), "device configuration")
	stdin := flags.Bool("token-stdin", false, "read the one-time device token from stdin")
	prompt := flags.Bool("token-prompt", false, "prompt for a one-time device token without echo")
	roots := flags.String("roots", "", "allowed project roots separated by the OS path separator")
	write := flags.Bool("allow-write", false, "allow project edits (default: read-only)")
	noStart := flags.Bool("no-start", false, "bind the device without starting remote access")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected connection arguments")
	}
	if err := agent.ValidateServer(*server); err != nil {
		return err
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if _, err = agent.WindowsServiceStatus(); err != nil {
			return fmt.Errorf("请先下载安装 Remote Agent MSI：%w", err)
		}
	}
	if _, err = os.Stat(path); err == nil {
		existing, loadErr := agent.Load(path)
		if loadErr != nil {
			return loadErr
		}
		if strings.TrimSuffix(*server, "/") != existing.Server {
			return errors.New("已有设备绑定属于其他服务器，未改动")
		}
		fmt.Println("保留现有设备绑定、项目和会话；正在配置后台服务。")
		if runtime.GOOS == "windows" {
			return agent.SetupWindowsService(path, *noStart)
		}
		if !*noStart {
			return agent.Autostart("install", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	token, err := readSetupToken(*prompt, *stdin)
	if err != nil {
		return err
	}
	var allowed []string
	if *roots != "" {
		allowed = filepath.SplitList(*roots)
	}
	c, err := agent.Enroll(strings.TrimSuffix(*server, "/"), token, path, allowed, *write)
	if err != nil {
		return err
	}
	fmt.Printf("已绑定客户端 %s。\n", c.Name)
	if runtime.GOOS == "windows" {
		if err = agent.SetupWindowsService(path, *noStart); err != nil {
			return fmt.Errorf("设备凭证已保存，服务尚未完成配置：%w。可运行 remote-agent service configure --config <配置路径> 重试", err)
		}
	} else if !*noStart {
		err = agent.Autostart("install", path)
	}
	if err != nil {
		return err
	}
	fmt.Println("客户端配置完成；可使用 remote-agent status 检查。Codex 代理 Token 使用 remote-agent codex setup 单独配置。")
	return nil
}
