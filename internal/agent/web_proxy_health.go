package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const WebProxyHealthPath = "/.well-known/remote-agent/web-proxy"

func CheckWebProxy(ctx context.Context, cfg Config, address string) error {
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: time.Second}
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+WebProxyHealthPath, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.New("local web proxy is not listening")
	}
	defer res.Body.Close()
	var health struct {
		ClientID string `json:"clientId"`
	}
	if res.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&health) != nil || health.ClientID != cfg.ID {
		return errors.New("port is not served by this client's web proxy; update or restart the client")
	}
	return nil
}

func WaitWebProxy(ctx context.Context, cfg Config, address string) error {
	var err error
	for {
		if err = CheckWebProxy(ctx, cfg, address); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("proxy setting saved but listener is not ready: %w", err)
		case <-time.After(150 * time.Millisecond):
		}
	}
}

// ProbeWebProxyTunnel checks the same CONNECT path used by system browsers,
// before changing any Windows proxy setting. It transmits no HTTP auth tokens.
func ProbeWebProxyTunnel(ctx context.Context, address, target string) error {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))
	if _, err := fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target); err != nil {
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("139 relay CONNECT failed for %s: HTTP %d; system proxy was not changed", target, response.StatusCode)
	}
	return nil
}
