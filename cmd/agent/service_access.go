package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/string1225/remote-agent-cli/internal/serviceaccess"
)

func accessHelp(w io.Writer) {
	fmt.Fprintln(w, `Windows 后台访问授权（需要时请求 UAC，不重启服务）：
  remote-agent access grant --path ABSOLUTE_PATH --permission read|modify [--recursive] [--git-trust]
  remote-agent access list
  remote-agent access restore --backup ID [--force]
目录默认只授权目录本身；--recursive 包括现有内容及以后继承权限的内容。
--git-trust 只信任指定仓库根目录，作用于升级后的后台任务，不修改用户 Git 配置。
restore 撤销本次授权并恢复备份；--force 允许覆盖之后的 ACL 修改，应先核对备份。
ACL、Git 仓库信任与 GitHub 登录互相独立；凭据仍用 git login / github login。
Windows 授权不能解密其他身份的 DPAPI 凭据，也不能提供 Session 0 桌面交互。`)
}

func runAccess(ctx context.Context, args []string, w io.Writer, invoke func(context.Context, serviceaccess.Options, io.Writer) error) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		accessHelp(w)
		return nil
	}
	o := serviceaccess.Options{Action: args[0]}
	f := flag.NewFlagSet("access "+o.Action, flag.ContinueOnError)
	f.SetOutput(w)
	switch o.Action {
	case "grant":
		f.StringVar(&o.Path, "path", "", "明确选择本机绝对文件/目录路径（拒绝链接、通配符和驱动器根目录）")
		f.StringVar(&o.Permission, "permission", "", "read（读取/执行）或 modify（读取/写入/删除，不含更改 ACL/属主）")
		f.BoolVar(&o.Recursive, "recursive", false, "授权整个目录树，包括受保护的子目录；先备份所有现有 DACL")
		f.BoolVar(&o.GitTrust, "git-trust", false, "明确允许此仓库配置/钩子被后台 Git 使用；路径必须是仓库根目录")
	case "restore":
		f.StringVar(&o.Backup, "backup", "", "access grant/list 输出的备份 ID（备份始终保留）")
		f.BoolVar(&o.Force, "force", false, "覆盖授权之后的 ACL 修改；中断/失败的授权也可用此选项恢复")
	case "list":
	default:
		return errors.New("请使用 remote-agent access grant/list/restore；用 access --help 查看帮助")
	}
	if err := f.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			accessHelp(w)
			return nil
		}
		return err
	}
	if f.NArg() != 0 {
		return errors.New("access 不接受位置参数，请使用 --path 或 --backup")
	}
	if err := o.Validate(); err != nil {
		return err
	}
	return invoke(ctx, o, w)
}
