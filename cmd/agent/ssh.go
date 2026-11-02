package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runSSH(ctx context.Context, args []string, output io.Writer) error {
	if len(args) == 0 || args[0] != "configure" {
		return errors.New("用法：remote-agent ssh configure --alias NAME --host HOST --user USER --identity PATH --known-hosts PATH [--port 22]；Windows 会请求一次 UAC 授权")
	}
	flags := flag.NewFlagSet("ssh configure", flag.ContinueOnError)
	flags.SetOutput(output)
	var o agent.ServiceSSHOptions
	flags.StringVar(&o.Alias, "alias", "", "后台 SSH 连接别名")
	flags.StringVar(&o.Host, "host", "", "目标 IP 或域名")
	flags.StringVar(&o.User, "user", "", "远程 SSH 用户名")
	flags.IntVar(&o.Port, "port", 22, "远程 SSH 端口")
	flags.StringVar(&o.Identity, "identity", "", "要导入的私钥；保留原文件")
	flags.StringVar(&o.KnownHosts, "known-hosts", "", "已经核对过目标主机身份的 known_hosts 文件")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || o.Identity == "" || o.KnownHosts == "" {
		return errors.New("请指定 SSH 私钥、known_hosts 及目标连接参数")
	}
	var err error
	o.Identity, err = filepath.Abs(o.Identity)
	if err != nil {
		return err
	}
	o.KnownHosts, err = filepath.Abs(o.KnownHosts)
	if err != nil {
		return err
	}
	if err = o.Validate(); err != nil {
		return err
	}
	return agent.ConfigureServiceSSH(ctx, o, output)
}
