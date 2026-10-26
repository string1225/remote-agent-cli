package webproxy

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

func TestTargetRestrictsRelayHostsAndWebPorts(t *testing.T) {
	for _, target := range []string{"github.com:443", "github.com:80", "API.GITHUB.COM:443", "codeload.github.com:443", "raw.githubusercontent.com:443", "release-assets.githubusercontent.com:443", "github.githubassets.com:443", "chatgpt.com:443", "ws.chatgpt.com:443", "AUTH.CHATGPT.COM:80", "chatgpt.com.:443", "github.com:0443", "openai.com:443", "api.openai.com:443", "google.com:443", "www.google.com:80"} {
		if _, err := Target(target); err != nil {
			t.Errorf("rejected %q: %v", target, err)
		}
	}
	for _, target := range []string{"github.com", "github.com:22", "github.com.evil.test:443", "evilgithub.com:443", "chatgpt.com.evil.test:443", "evilchatgpt.com:443", "chatgpt.com:8443", "github.io:443", "127.0.0.1:443", "[::1]:443", "169.254.169.254:443", "user@github.com:443", "github.com/path:443", "foo..github.com:443", "-foo.github.com:443", "foo_.github.com:443"} {
		if _, err := Target(target); err == nil {
			t.Errorf("accepted forbidden target %q", target)
		}
	}
}

func TestRelayHostLeavesOtherDestinationsDirect(t *testing.T) {
	for _, host := range []string{"github.com", "api.github.com", "raw.githubusercontent.com", "chatgpt.com", "ws.chatgpt.com", "FOO.CHATGPT.COM.", "openai.com", "api.openai.com", "google.com", "www.google.com"} {
		if !RelayHost(host) {
			t.Errorf("missed relay host %s", host)
		}
	}
	for _, host := range []string{"example.com", "github.com.example.com", "chatgpt.com.example.com", "notchatgpt.com", "127.0.0.1", "::1", "github.io", "openai.com.evil.test", "evilgoogle.com"} {
		if RelayHost(host) {
			t.Errorf("unexpected relay host %s", host)
		}
	}
	for _, address := range []string{"example.com:1234", "127.0.0.1:8080", "[::1]:8080"} {
		if _, err := Address(address); err != nil {
			t.Error("direct address rejected", address, err)
		}
	}
}

func TestPublicIPRejectsLocalAndSpecialRanges(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.1.2.3", "172.16.1.1", "192.168.1.1", "169.254.169.254", "100.100.100.200", "0.1.2.3", "192.0.2.5", "198.18.0.1", "255.255.255.255", "::", "::1", "::ffff:127.0.0.1", "fd00::1", "fe80::1", "64:ff9b::7f00:1", "2001:db8::1", "2002:7f00:1::"} {
		if publicIP(netip.MustParseAddr(ip)) {
			t.Errorf("accepted non-public IP %s", ip)
		}
	}
	for _, ip := range []string{"140.82.112.3", "185.199.108.133", "2606:50c0:8000::154", "::ffff:140.82.112.3"} {
		if !publicIP(netip.MustParseAddr(ip)) {
			t.Errorf("rejected public IP %s", ip)
		}
	}
}

func TestPoolLimitsReleasesAndCancels(t *testing.T) {
	var pool Pool
	var releases []func()
	var contexts []context.Context
	for device := 0; device < MaxConnections/MaxPerDevice; device++ {
		for range MaxPerDevice {
			ctx, release, ok := pool.Open(context.Background(), string(rune('a'+device)))
			if !ok {
				t.Fatal("connection refused before limit")
			}
			if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > MaxLifetime {
				t.Fatal("missing finite lease")
			}
			contexts = append(contexts, ctx)
			releases = append(releases, release)
		}
		if _, _, ok := pool.Open(context.Background(), string(rune('a'+device))); ok {
			t.Fatal("per-device limit exceeded")
		}
	}
	if _, _, ok := pool.Open(context.Background(), "another-device"); ok {
		t.Fatal("global limit exceeded")
	}
	releases[0]()
	_, release, ok := pool.Open(context.Background(), "a")
	if !ok {
		t.Fatal("slot was not released")
	}
	defer release()
	pool.Close()
	for _, ctx := range contexts {
		if ctx.Err() == nil {
			t.Fatal("shutdown did not cancel connection")
		}
	}
	for _, release := range releases {
		release()
	}
	if _, _, ok := pool.Open(context.Background(), "a"); ok {
		t.Fatal("closed pool admitted a connection")
	}
}
