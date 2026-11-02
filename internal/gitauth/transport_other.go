//go:build !windows

package gitauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func localPath(config string) string { return filepath.Join(filepath.Dir(config), ".access.sock") }
func ListenLocal(ctx context.Context, config string, manager Manager, handler func(context.Context, json.RawMessage) (any, error)) (func(), error) {
	path := localPath(config)
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, errors.New("local credential endpoint is not a socket")
		}
		if c, e := net.DialTimeout("unix", path, time.Second); e == nil {
			c.Close()
			return nil, errors.New("local credential endpoint already active")
		}
		if err = os.Remove(path); err != nil {
			return nil, err
		}
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err = os.Chmod(path, 0600); err != nil {
		listener.Close()
		return nil, err
	}
	server := &Server{Manager: manager, Manage: handler, Authorize: func(net.Conn) error { return nil }}
	child, cancel := context.WithCancel(ctx)
	var once sync.Once
	stop := func() { once.Do(func() { cancel(); listener.Close() }) }
	go func() { <-child.Done(); listener.Close() }()
	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				return
			}
			go server.Serve(child, c)
		}
	}()
	return stop, nil
}
func CallLocal(ctx context.Context, config string, r Request, output io.Writer) error {
	c, err := (&net.Dialer{}).DialContext(ctx, "unix", localPath(config))
	if err != nil {
		return errors.New("请先启动或升级本机客户端")
	}
	return Exchange(ctx, c, r, output)
}
