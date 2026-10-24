package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestClientCapabilitiesStayWithinAccountAndUseServerAssignedRouting(t *testing.T) {
	f := setupAt(t, "/agents")
	alice, bob := f.user("alice"), f.user("bob")
	source, credential, _, _ := f.connectedDevice(alice)
	target, _, device, _ := f.connectedDevice(alice)
	foreign, _, _, _ := f.connectedDevice(bob)
	call := func(method, path string, body any, status int) []byte {
		var data bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&data).Encode(body)
		}
		r, _ := http.NewRequest(method, f.url+path+"?id="+source.ID, &data)
		r.Header.Set("Authorization", "Bearer "+credential)
		r.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Error(err)
			return nil
		}
		defer res.Body.Close()
		got, _ := io.ReadAll(res.Body)
		if res.StatusCode != status {
			t.Errorf("client capability got %d want %d", res.StatusCode, status)
		}
		return got
	}
	catalog := call("GET", "/api/agent/catalog", nil, 200)
	if !bytes.Contains(catalog, []byte(target.ID)) || bytes.Contains(catalog, []byte(foreign.ID)) || bytes.Contains(catalog, []byte(credential)) {
		t.Fatal("catalog crossed account scope or exposed credentials")
	}
	call("POST", "/api/agent/tools", map[string]any{"agentId": foreign.ID, "type": "tasks.start", "body": map[string]string{"prompt": "denied"}}, 404)
	call("POST", "/api/agent/tools", map[string]any{"agentId": target.ID, "type": "run", "body": map[string]string{}}, 400)
	result := make(chan []byte, 1)
	go func() {
		result <- call("POST", "/api/agent/tools", map[string]any{"agentId": target.ID, "type": "tasks.start", "body": map[string]string{"serviceId": "project-b", "prompt": "work", "parentAgentId": foreign.ID, "parentSessionId": "origin-session"}}, 200)
	}()
	device.SetReadDeadline(time.Now().Add(5 * time.Second))
	var open protocol.Signal
	if err := device.ReadJSON(&open); err != nil || open.Type != "open" {
		t.Fatal("no relay session", err)
	}
	if err := device.WriteJSON(protocol.Signal{Type: "ready", PeerID: open.PeerID, Transport: protocol.RelayTransport}); err != nil {
		t.Fatal(err)
	}
	var request protocol.Signal
	if err := device.ReadJSON(&request); err != nil || request.Type != "request" || request.PeerID != open.PeerID {
		t.Fatal("request not routed to reserved session", err)
	}
	var task protocol.TaskStart
	if request.Request == nil || json.Unmarshal(request.Request.Body, &task) != nil || task.ParentAgentID != source.ID || task.ParentSessionID != "origin-session" || task.ServiceID != "project-b" {
		t.Fatal("caller spoofed source or project context")
	}
	event := protocol.Event{ID: request.Request.ID, Type: "result", Data: map[string]any{"session": map[string]string{"id": "new-web", "kind": "web"}, "taskId": "new-task"}}
	if err := device.WriteJSON(protocol.Signal{Type: "event", PeerID: open.PeerID, Event: &event}); err != nil {
		t.Fatal(err)
	}
	select {
	case response := <-result:
		if !bytes.Contains(response, []byte("new-web")) {
			t.Fatal("did not return created web session")
		}
	case <-time.After(time.Second):
		t.Fatal("operation response not forwarded")
	}
}
