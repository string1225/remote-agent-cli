package agent

import (
	"context"
	"fmt"
	"net"
)

// SelectWebProxyAddress finds a usable loopback listener for this client.
func SelectWebProxyAddress(ctx context.Context, cfg Config, preferred string) (string, error) {
	if err := validateWebProxyListen(preferred); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	listener, bindErr := ListenWebProxy(preferred)
	if bindErr == nil {
		address := listener.Addr().String()
		return address, listener.Close()
	}
	// An occupied port is reusable only when its health endpoint identifies
	// this client. An unrelated proxy (including a wildcard listener) must
	// never become the user's system proxy through this command.
	if CheckWebProxy(ctx, cfg, preferred) == nil {
		return preferred, nil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	host, _, _ := net.SplitHostPort(preferred)
	listener, err := ListenWebProxy(net.JoinHostPort(host, "0"))
	if err != nil {
		return "", fmt.Errorf("web proxy cannot bind %s (%v) or select a free loopback port: %w", preferred, bindErr, err)
	}
	address := listener.Addr().String()
	return address, listener.Close()
}
