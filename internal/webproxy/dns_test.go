package webproxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestRelayDNSReplacesFakeIPWithoutPermittingPrivateDestinations(t *testing.T) {
	for _, localIP := range []string{"198.18.0.12", "127.0.0.1", "169.254.169.254"} {
		called := false
		ips, err := resolveRelayIPs(context.Background(), "github.com", func(context.Context, string, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr(localIP)}, nil
		}, func(context.Context, string) ([]netip.Addr, error) {
			called = true
			return []netip.Addr{netip.MustParseAddr("140.82.112.3")}, nil
		})
		if err != nil || !called || len(ips) != 1 || ips[0].String() != "140.82.112.3" {
			t.Fatal("fake/private DNS did not use validated fallback", localIP, ips, err)
		}
	}
	for _, publicIP := range []string{"198.18.0.12", "10.0.0.1", "::1"} {
		_, err := resolveRelayIPs(context.Background(), "github.com", func(context.Context, string, string) ([]netip.Addr, error) {
			return nil, errors.New("fixture DNS unavailable")
		}, func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr(publicIP)}, nil
		})
		if err == nil {
			t.Fatal("unsafe fallback was allowed", publicIP)
		}
	}
}

func TestRelayDNSKeepsValidLocalResolution(t *testing.T) {
	ips, err := resolveRelayIPs(context.Background(), "github.com", func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("140.82.112.3")}, nil
	}, func(context.Context, string) ([]netip.Addr, error) {
		t.Fatal("unnecessary public DNS lookup")
		return nil, nil
	})
	if err != nil || len(ips) != 1 {
		t.Fatal(ips, err)
	}
}

func TestPublicDNSParsesOnlyValidatedAddressRecords(t *testing.T) {
	for _, body := range []string{`{"Status":0,"Answer":[{"type":1,"data":"140.82.112.3"}]}`, `{"Status":0,"Answer":[{"type":1,"data":"198.18.0.12"}]}`, `{"Status":3}`, `{"Status":0,"Answer":[{"type":5,"data":"alias.example.test"}]}`} {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("name") != "github.com" || r.URL.Query().Get("type") != "A" || r.Header.Get("Authorization") != "" {
				t.Error("wrong DNS request or leaked auth")
			}
			w.Write([]byte(body))
		}))
		ips, err := queryPublicDNS(context.Background(), server.Client(), server.URL, "github.com")
		server.Close()
		valid := body == `{"Status":0,"Answer":[{"type":1,"data":"140.82.112.3"}]}`
		if valid != (err == nil) || (valid && len(ips) != 1) {
			t.Fatal("wrong public DNS validation", body, ips, err)
		}
	}
}
