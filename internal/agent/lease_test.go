package agent

import (
	"context"
	"testing"
	"time"

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

}

func TestPeerRelativeLeaseExpiresDespiteClockSkew(t *testing.T) {
	a := &Agent{}
	p, err := a.newPeer(context.Background(), protocol.Signal{
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
