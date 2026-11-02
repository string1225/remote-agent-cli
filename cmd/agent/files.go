package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func runFiles(ctx context.Context, args []string, output io.Writer) error {
	if len(args) == 0 || args[0] != "upload" {
		return errors.New("用法：remote-agent files upload --path FILE [--config PATH]；返回公开下载链接，下次北京时间凌晨 4 点失效")
	}
	flags := flag.NewFlagSet("files upload", flag.ContinueOnError)
	flags.SetOutput(output)
	config := flags.String("config", agent.DefaultPath(), "客户端配置路径")
	path := flags.String("path", "", "需要上传的本地文件路径（最多 100 MiB）")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *path == "" || flags.NArg() != 0 {
		return errors.New("请使用 --path 指定上传文件")
	}
	c, err := agent.Load(*config)
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(*path)
	if err != nil {
		return err
	}
	file, err := agent.UploadFile(ctx, c, absolute)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(file)
}
