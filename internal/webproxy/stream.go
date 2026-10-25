package webproxy

import (
	"context"
	"io"
	"net"
)

// Stream copies CONNECT bytes, including any bytes buffered with the headers.
// Cancellation closes both ends; TCP destinations retain write half-close.
func Stream(ctx context.Context, downstream net.Conn, reader io.Reader, upstream net.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	closeBoth := func() { downstream.Close(); upstream.Close() }
	stop := context.AfterFunc(ctx, closeBoth)
	defer func() { cancel(); closeBoth(); stop() }()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := io.Copy(upstream, reader)
		if half, ok := upstream.(interface{ CloseWrite() error }); ok && err == nil {
			_ = half.CloseWrite()
		} else {
			cancel()
		}
	}()
	_, _ = io.Copy(downstream, upstream)
	cancel()
	closeBoth()
	<-done
}
