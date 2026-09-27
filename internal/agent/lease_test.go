package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestPeerLeaseClockSkewAndBounds(t *testing.T) {
	serverNow := time.Unix(1790522789, 0)
	for _, skew := range []time.Duration{-2 * time.Hour, -108 * time.Second, 108 * time.Second, 2 * time.Hour} {
		msg := protocol.Signal{ExpiresAt: serverNow.Add(30 * time.Minute).Unix(), LeaseSeconds: 1800}
		lifetime, err := peerLifetime(msg, serverNow.Add(skew))
		if err != nil || lifetime != 30*time.Minute {
			t.Fatalf("clock skew %v: lifetime %v, error %v", skew, lifetime, err)
		}
	}
	for _, seconds := range []int64{-1, 1801, 1<<63 - 1} {
		if _, err := peerLifetime(protocol.Signal{LeaseSeconds: seconds}, serverNow); err == nil {
			t.Fatalf("accepted invalid lifetime %d", seconds)
		}
	}
	if lifetime, err := peerLifetime(protocol.Signal{ExpiresAt: serverNow.Add(10 * time.Minute).Unix()}, serverNow); err != nil || lifetime != 10*time.Minute {
		t.Fatalf("legacy server compatibility: %v, %v", lifetime, err)
	}
	for _, expiry := range []int64{serverNow.Unix() - 1, serverNow.Add(32 * time.Minute).Unix()} {
		if _, err := peerLifetime(protocol.Signal{ExpiresAt: expiry}, serverNow); err == nil {
			t.Fatal("accepted invalid legacy expiry")
		}
	}
}

func TestPeerRelativeLeaseExpiresDespiteClockSkew(t *testing.T) {
	remote, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	if _, err := remote.CreateDataChannel("remote-agent", nil); err != nil {
		t.Fatal(err)
	}
	offer, err := remote.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	sdp, err := json.Marshal(offer)
	if err != nil {
		t.Fatal(err)
	}
	a := &Agent{}
	p, err := a.newPeer(context.Background(), protocol.Signal{
		Secret: strings.Repeat("x", 32), SDP: sdp,
		ExpiresAt: time.Now().Add(2 * time.Hour).Unix(), LeaseSeconds: 1,
	}, func(protocol.Signal) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
	select {
	case <-p.ctx.Done():
		if p.ctx.Err() != context.DeadlineExceeded {
			t.Fatalf("unexpected termination: %v", p.ctx.Err())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("relative lease did not expire")
	}
}
