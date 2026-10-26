package webproxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"
)

// A transparent proxy on the relay host may answer ordinary DNS with fake IPs
// (198.18.0.0/15). Never permit those IPs as destinations. Resolve them over
// authenticated HTTPS pinned to public resolver IPs instead.
func resolveRelayIPs(ctx context.Context, host string, local func(context.Context, string, string) ([]netip.Addr, error), public func(context.Context, string) ([]netip.Addr, error)) ([]netip.Addr, error) {
	ips, err := local(ctx, "ip", host)
	if err == nil && validPublicIPs(ips) {
		return ips, nil
	}
	ips, err = public(ctx, host)
	if err != nil || !validPublicIPs(ips) {
		return nil, errors.New("cannot resolve relay destination to public addresses")
	}
	return ips, nil
}

func validPublicIPs(ips []netip.Addr) bool {
	if len(ips) == 0 {
		return false
	}
	for _, ip := range ips {
		if !publicIP(ip) {
			return false
		}
	}
	return true
}

type dnsCacheEntry struct {
	ips   []netip.Addr
	until time.Time
}

var publicDNSCache = struct {
	sync.Mutex
	entries map[string]dnsCacheEntry
}{entries: map[string]dnsCacheEntry{}}

type publicResolver struct {
	endpoint string
	client   *http.Client
}

func pinnedDNSResolver(host, ip, path string) publicResolver {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return publicResolver{endpoint: "https://" + host + path, client: &http.Client{
		Timeout:       6 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, "443"))
		}, TLSClientConfig: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}, TLSHandshakeTimeout: 5 * time.Second, MaxIdleConnsPerHost: 8},
	}}
}

var publicDNSResolvers = []publicResolver{
	pinnedDNSResolver("cloudflare-dns.com", "1.1.1.1", "/dns-query"),
	pinnedDNSResolver("dns.google", "8.8.8.8", "/resolve"),
}

func lookupPublicDNS(ctx context.Context, host string) ([]netip.Addr, error) {
	publicDNSCache.Lock()
	cached, ok := publicDNSCache.entries[host]
	publicDNSCache.Unlock()
	if ok && time.Now().Before(cached.until) {
		return append([]netip.Addr(nil), cached.ips...), nil
	}
	for _, resolver := range publicDNSResolvers {
		ips, err := queryPublicDNS(ctx, resolver.client, resolver.endpoint, host)
		if err != nil || !validPublicIPs(ips) {
			continue
		}
		publicDNSCache.Lock()
		if len(publicDNSCache.entries) >= 256 {
			for key := range publicDNSCache.entries {
				delete(publicDNSCache.entries, key)
				break
			}
		}
		publicDNSCache.entries[host] = dnsCacheEntry{ips: append([]netip.Addr(nil), ips...), until: time.Now().Add(30 * time.Second)}
		publicDNSCache.Unlock()
		return ips, nil
	}
	return nil, errors.New("public DNS-over-HTTPS lookup failed")
}

func queryPublicDNS(ctx context.Context, client *http.Client, endpoint, host string) ([]netip.Addr, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("name", host)
	q.Set("type", "A")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("public DNS returned an unsuccessful response")
	}
	var reply struct {
		Status int
		Answer []struct {
			Type int
			Data string
		}
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 32*1024)).Decode(&reply); err != nil {
		return nil, err
	}
	if reply.Status != 0 {
		return nil, errors.New("public DNS did not resolve the hostname")
	}
	var ips []netip.Addr
	for _, answer := range reply.Answer {
		if answer.Type != 1 {
			continue
		}
		ip, err := netip.ParseAddr(answer.Data)
		if err != nil || !publicIP(ip) {
			return nil, errors.New("public DNS returned a non-public address")
		}
		ips = append(ips, ip)
	}
	if len(ips) == 0 {
		return nil, errors.New("public DNS returned no IPv4 addresses")
	}
	return ips, nil
}
