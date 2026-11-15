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
	if len(args) == 0 || (args[0] != "upload" && args[0] != "delete") {
		return errors.New("用法：remote-agent files upload --path FILE [--class temporary|sensitive --confirm-sensitive] [--config PATH]；files delete --id ID 删除本机有回执的storage.to副本")
	}
	flags := flag.NewFlagSet("files "+args[0], flag.ContinueOnError)
	flags.SetOutput(output)
	config := flags.String("config", agent.DefaultPath(), "客户端配置路径")
	path := flags.String("path", "", "需要上传的本地文件路径（最多 100 MiB）")
	class := flags.String("class", "temporary", "temporary=storage.to临时文件；sensitive=经用户确认后的139分享")
	confirmed := flags.Bool("confirm-sensitive", false, "已取得用户针对该敏感文件及公开分享范围的确认；此参数不替代询问")
	id := flags.String("id", "", "需要删除的storage.to文件ID（本机须保留管理回执）")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	options := agent.FileUploadOptions{Class: *class, UserConfirmed: *confirmed}
	if err := options.Validate(); err != nil {
		return err
	}
	c := agent.Config{Path: *config}
	if args[0] == "delete" {
		if *id == "" || *path != "" || flags.NArg() != 0 {
			return errors.New("请使用 files delete --id ID")
		}
		if err := agent.DeleteUploadedFile(ctx, c, *id); err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(map[string]any{"id": *id, "deleted": true, "provider": "storage.to"})
	}
	if *path == "" || *id != "" || flags.NArg() != 0 {
		return errors.New("请使用 --path 指定上传文件")
	}
	if *class == "sensitive" {
		var err error
		c, err = agent.Load(*config)
		if err != nil {
			return err
		}
	}
	absolute, err := filepath.Abs(*path)
	if err != nil {
		return err
	}
	file, err := agent.UploadFileWithOptions(ctx, c, absolute, options)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(file)
}
