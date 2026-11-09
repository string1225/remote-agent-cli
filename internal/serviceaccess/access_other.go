//go:build !windows

package serviceaccess

import (
	"errors"
	"io"
)

func trustPath() string { return "" }

func Execute(Options, io.Writer) error {
	return errors.New("access 仅用于 Windows RemoteAgent 服务；其他平台请使用该平台的文件权限管理工具")
}
