// Package webproxy defines web proxy routing and the restricted server relay.
package webproxy

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// Address validates a CONNECT/dial authority without restricting direct
// destinations. Relay-only policy is enforced separately by Target.
func Address(authority string) (string, error) {
	host, port, err := net.SplitHostPort(authority)
	number, portErr := strconv.Atoi(port)
	if err != nil || host == "" || portErr != nil || number < 1 || number > 65535 || strings.ContainsAny(host, " /\\@?#\r\n\t") {
		return "", errors.New("invalid proxy destination")
	}
	return net.JoinHostPort(strings.ToLower(strings.TrimSuffix(host, ".")), strconv.Itoa(number)), nil
}

// RelayHost matches names on label boundaries, including bare chatgpt.com.
// Other hosts are connected directly by the client, never by the server.
func RelayHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, domain := range []string{"github.com", "githubusercontent.com", "githubassets.com", "chatgpt.com", "openai.com", "google.com"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// Target is the server's allowlist: GitHub, its assets, and ChatGPT over web
// ports. A relay-selected host on another port must fail rather than go direct.
func Target(authority string) (string, error) {
	target, err := Address(authority)
	if err != nil {
		return "", err
	}
	host, port, _ := net.SplitHostPort(target)
	if port != "443" && port != "80" {
		return "", errors.New("web relay requires port 80 or 443")
	}
	host = strings.ToLower(host)
	if len(host) > 253 {
		return "", errors.New("invalid relay hostname")
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("invalid relay hostname")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", errors.New("invalid relay hostname")
			}
		}
	}
	if RelayHost(host) {
		return target, nil
	}
	return "", errors.New("destination is outside the web relay allowlist")
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
	ips, err := resolveRelayIPs(ctx, host, net.DefaultResolver.LookupNetIP, lookupPublicDNS)
	if err != nil {
		return nil, err
	}
	var dialer net.Dialer
	for _, ip := range ips {
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, errors.New("cannot connect to relay destination")
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
