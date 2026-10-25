// Package githubproxy implements the restricted GitHub CONNECT relay.
package githubproxy

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"time"
)

// Target accepts only GitHub's HTTPS hosts, including API, source archives,
// release assets, raw files, and Git LFS objects. SSH and arbitrary sites are
// intentionally outside this service's scope.
func Target(authority string) (string, error) {
	host, port, err := net.SplitHostPort(authority)
	if err != nil || port != "443" {
		return "", errors.New("GitHub proxy requires a GitHub hostname on port 443")
	}
	host = strings.ToLower(host)
	if len(host) > 253 {
		return "", errors.New("invalid GitHub hostname")
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("invalid GitHub hostname")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", errors.New("invalid GitHub hostname")
			}
		}
	}
	for _, domain := range []string{"github.com", "githubusercontent.com", "githubassets.com"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return net.JoinHostPort(host, port), nil
		}
	}
	return "", errors.New("destination is outside the GitHub proxy allowlist")
}

// Dial resolves on the relay server and dials a checked IP directly, so DNS
// cannot switch the destination between validation and connection. Environment
// proxies are never consulted.
func Dial(ctx context.Context, authority string) (net.Conn, error) {
	target, err := Target(authority)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	host, port, _ := net.SplitHostPort(target)
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, errors.New("cannot resolve GitHub destination")
	}
	for _, ip := range ips {
		if !publicIP(ip) {
			return nil, errors.New("GitHub destination resolved to a non-public address")
		}
	}
	var dialer net.Dialer
	for _, ip := range ips {
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("cannot connect to GitHub destination")
}

func publicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}
