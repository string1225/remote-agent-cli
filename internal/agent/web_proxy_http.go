package agent

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/string1225/remote-agent-cli/internal/webproxy"
)

func (p *webProxy) forwardHTTP(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() || r.URL.Host == "" || (r.URL.Scheme != "http" && r.URL.Scheme != "https") {
		http.Error(w, "absolute HTTP proxy URL required", http.StatusBadRequest)
		return
	}
	port := r.URL.Port()
	if port == "" {
		port = "80"
		if r.URL.Scheme == "https" {
			port = "443"
		}
	}
	target, err := webproxy.Address(net.JoinHostPort(r.URL.Hostname(), port))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if webproxy.RelayHost(r.URL.Hostname()) {
		if _, err := webproxy.Target(target); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
	ctx, release, ok := p.pool.Open(r.Context(), p.config.ID)
	if !ok {
		http.Error(w, "web proxy connection limit reached", http.StatusTooManyRequests)
		return
	}
	defer release()
	// Explicitly bypass environment proxies. Each transport owns this request's
	// connection so cancellation/off cannot leave a reusable tunnel behind.
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           func(context.Context, string, string) (net.Conn, error) { return p.dial(ctx, target) },
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
	}
	defer transport.CloseIdleConnections()
	defer http.NewResponseController(w).SetWriteDeadline(time.Time{})
	// The standard forwarding implementation preserves streaming, trailers and
	// WebSocket upgrades while stripping connection-specific proxy headers.
	proxy := &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			request.Out.Host = r.URL.Host
			request.Out.Header.Del("Proxy-Authorization")
		},
		Transport:     transport,
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			http.Error(w, "web proxy destination unavailable", http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(proxyResponseWriter{w}, r.WithContext(ctx))
}

type proxyResponseWriter struct{ http.ResponseWriter }

func (w proxyResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w proxyResponseWriter) Write(data []byte) (int, error) {
	controller := http.NewResponseController(w.ResponseWriter)
	_ = controller.SetWriteDeadline(time.Now().Add(60 * time.Second))
	return w.ResponseWriter.Write(data)
}
