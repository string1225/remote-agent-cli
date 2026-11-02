package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/access"
	"github.com/string1225/remote-agent-cli/internal/agent"
	"golang.org/x/term"
)

type credentialCaller func(context.Context, string, access.Request) (json.RawMessage, error)

func runServers(ctx context.Context, args []string, input io.Reader, out io.Writer, call credentialCaller) error {
	if len(args) == 0 {
		return errors.New("用法：remote-agent servers add [--name NAME] --host IP --user USER [--port 22] [--password-stdin] | list | test --name NAME | remove --name NAME")
	}
	flags := newCLIFlagSet("servers " + args[0])
	flags.SetOutput(out)
	config := flags.String("config", agent.DefaultPath(), "客户端配置")
	r := access.Request{Action: "servers." + args[0]}
	flags.StringVar(&r.Name, "name", "", "服务器别名")
	flags.StringVar(&r.Host, "host", "", "服务器 IP")
	flags.StringVar(&r.User, "user", "", "SSH 用户名")
	flags.IntVar(&r.Port, "port", 22, "SSH 端口")
	stdin := flags.Bool("password-stdin", false, "从标准输入读取一行密码；默认在终端隐藏输入")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("不接受位置参数")
	}
	switch args[0] {
	case "add", "list", "test", "remove":
	default:
		return errors.New("使用 servers add/list/test/remove")
	}
	if args[0] == "add" {
		if r.Host == "" || r.User == "" {
			return errors.New("请提供 --host 和 --user；连接别名可留空自动生成")
		}
		if *stdin {
			value, err := bufio.NewReader(io.LimitReader(input, 4098)).ReadString('\n')
			if err != nil && err != io.EOF {
				return errors.New("无法读取 SSH 密码")
			}
			r.Password = strings.TrimRight(value, "\r\n")
		} else {
			file, ok := input.(*os.File)
			if !ok || !term.IsTerminal(int(file.Fd())) {
				return errors.New("请在交互终端输入密码，或显式使用 --password-stdin；不支持命令行密码参数")
			}
			fmt.Fprint(out, "SSH 密码（输入隐藏，仅用于初始化）：")
			value, err := term.ReadPassword(int(file.Fd()))
			fmt.Fprintln(out)
			if err != nil {
				return errors.New("无法读取 SSH 密码")
			}
			r.Password = string(value)
			clear(value)
		}
	}
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	result, err := call(ctx, path, r)
	r.Password = ""
	if err != nil {
		return err
	}
	var formatted any
	if json.Unmarshal(result, &formatted) != nil {
		return errors.New("无效的客户端响应")
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(formatted)
}

func runGitHub(ctx context.Context, args []string, out io.Writer, call credentialCaller) error {
	if len(args) == 0 {
		return errors.New("用法：remote-agent github login [--username ACCOUNT] [--force] [--wait] | status | accounts | cancel | logout --username ACCOUNT")
	}
	flags := newCLIFlagSet("github " + args[0])
	flags.SetOutput(out)
	config := flags.String("config", agent.DefaultPath(), "客户端配置")
	r := access.Request{}
	flags.StringVar(&r.Username, "username", "", "GitHub 用户名，可留空；邮箱在网页登录时输入")
	flags.BoolVar(&r.Force, "force", false, "重新授权")
	wait := flags.Bool("wait", false, "显示设备码后继续等待；退出终端不取消后台授权")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("不接受位置参数")
	}
	action := args[0]
	if action == "login" {
		action = "start"
	}
	switch action {
	case "start", "status", "accounts", "cancel", "logout":
	default:
		return errors.New("使用 github login/status/accounts/cancel/logout")
	}
	r.Action = "github." + action
	path, err := filepath.Abs(*config)
	if err != nil {
		return err
	}
	for {
		raw, err := call(ctx, path, r)
		if err != nil {
			return err
		}
		if action != "start" {
			_, err = fmt.Fprintln(out, string(raw))
			return err
		}
		var state access.GitHubState
		if json.Unmarshal(raw, &state) != nil {
			return errors.New("无效的授权状态")
		}
		if state.Code != "" {
			fmt.Fprintf(out, "授权网址：%s\n设备码：%s\n", state.URL, state.Code)
		}
		fmt.Fprintln(out, state.Message)
		if !*wait {
			return nil
		}
		for state.Status == "starting" || state.Status == "waiting" {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			raw, err = call(ctx, path, access.Request{Action: "github.status"})
			if err != nil {
				return err
			}
			previous := state.Code
			if json.Unmarshal(raw, &state) != nil {
				return errors.New("无效的授权状态")
			}
			if state.Code != "" && state.Code != previous {
				fmt.Fprintf(out, "授权网址：%s\n设备码：%s\n", state.URL, state.Code)
			}
		}
		fmt.Fprintln(out, state.Message)
		if state.Status != "authorized" {
			return errors.New("GitHub 授权尚未完成")
		}
		return nil
	}
}
