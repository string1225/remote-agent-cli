package server_test

import (
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestDefaultHostRejectsLegacyClientWithoutChangingHost(t *testing.T) {
	f := setupAt(t, "/agents")
	owner := f.user("owner")
	host, _, _, _ := todoTestDevice(t, f, owner)
	f.request(owner, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 200)
	before, err := f.db.DefaultHost(host.UserID)
	if err != nil {
		t.Fatal(err)
	}
	legacy, _, device, browser := f.connectedDevice(owner)
	_ = browser.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = device.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			var msg protocol.Signal
			if device.ReadJSON(&msg) != nil {
				return
			}
			switch msg.Type {
			case "open":
				_ = device.WriteJSON(protocol.Signal{Type: "ready", PeerID: msg.PeerID, Transport: protocol.RelayTransport})
			case "request":
				if msg.Request != nil {
					_ = device.WriteJSON(protocol.Signal{Type: "event", PeerID: msg.PeerID, Event: &protocol.Event{ID: msg.Request.ID, Type: "error", Error: "unknown request type"}})
					return
				}
			}
		}
	}()
	raw := f.request(owner, "PUT", "/api/default-host", map[string]string{"agentId": legacy.ID}, 409)
	<-done
	if !strings.Contains(string(raw), "请先升级该客户端") || !strings.Contains(string(raw), "todos.call") || !strings.Contains(string(raw), "需要先发布") {
		t.Fatalf("missing actionable compatibility message: %s", raw)
	}
	after, err := f.db.DefaultHost(host.UserID)
	if err != nil || after != before {
		t.Fatal("unsupported target changed canonical host", before, after, err)
	}
	f.request(owner, "GET", "/api/todos", nil, 200)
}
