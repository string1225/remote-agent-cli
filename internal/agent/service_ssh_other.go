//go:build !windows

package agent

import (
	"context"
	"errors"
	"io"
)

func ConfigureServiceSSH(context.Context, ServiceSSHOptions, io.Writer) error {
	return errors.New("SSH 服务授权仅用于 Windows；其他系统请使用运行客户端的用户配置 SSH")
}
