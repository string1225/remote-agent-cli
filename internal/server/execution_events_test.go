package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestExecutionSubscriptionClosedBrowserCheckpointsAndIsolation(t *testing.T) {
	f := setupAt(t, "/agents")
	alice, bob := f.user("alice"), f.user("bob")
	host, hostCfg, hostRan, _ := todoTestDevice(t, f, alice)
	release := make(chan struct{})
	var releaseOnce sync.Once
	finish := func() { releaseOnce.Do(func() { close(release) }) }
	defer finish()
	target, targetCfg, _, _ := todoTestDevice(t, f, alice, func(ctx context.Context, _ agent.Provider, _ protocol.Service, _ string, _ bool, send func(protocol.Event) error, id string) error {
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
		item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": "verified final artifact"}})
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": string(item) + "\n"}})
	})
	raw := f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "resolve", Target: protocol.AssistantRef{Scope: "project", ClientID: host.ID, ProjectID: hostCfg.Services[0].ID}}, 200)
	var manager protocol.AssistantResolved
	json.Unmarshal(raw, &manager)
	p := f.peer(alice, host.ID)
	p.send(t, protocol.Request{ID: "source", Type: "sessions.create", ServiceID: manager.Service.ID})
	data, _ := json.Marshal(p.next(t, "source", "result").Data)
	var source protocol.Conversation
	json.Unmarshal(data, &source)
	p.conn.Close() // no page and no manager model remains open
	hostCfg, err := agent.Load(hostCfg.Path)
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := agent.StartExecutionTopic(context.Background(), hostCfg, manager.Service.ID, source.ID, "first", "event topic", target.ID, protocol.TaskStart{ServiceID: targetCfg.Services[0].ID, Prompt: "work", Checkpoints: []string{"plan"}})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case v := <-hostRan:
		t.Fatal("idle waiting ran a model", v)
	case <-time.After(100 * time.Millisecond):
	}
	event := protocol.CoordinationCommand{Operation: "checkpoint", Source: protocol.AssistantRef{ProjectID: targetCfg.Services[0].ID, SessionID: accepted.Session.ID}, TaskID: accepted.TaskID, Stage: "plan", Status: "checkpoint", Summary: "plan selected", Evidence: []string{"plan.md"}}
	for i := 0; i < 3; i++ {
		if _, err = agent.ClientAPI(context.Background(), targetCfg, "POST", "/api/agent/coordination", event); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case v := <-hostRan:
		if !strings.Contains(v, `"event":"checkpoint"`) || !strings.Contains(v, `"topic":"event topic"`) {
			t.Fatal("lost structured provenance", v)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("checkpoint did not wake manager")
	}
	select {
	case <-hostRan:
		t.Fatal("duplicate checkpoint model wake")
	case <-time.After(100 * time.Millisecond):
	}
	event.TaskID = "old-turn"
	if _, err = agent.ClientAPI(context.Background(), targetCfg, "POST", "/api/agent/coordination", event); err == nil {
		t.Fatal("stale event accepted")
	}
	event.TaskID = accepted.TaskID
	if _, err = agent.ClientAPI(context.Background(), hostCfg, "POST", "/api/agent/coordination", event); err == nil {
		t.Fatal("sibling source forged execution event")
	}
	// A different manager session cannot subscribe to this one's direct child.
	p = f.peer(alice, host.ID)
	p.send(t, protocol.Request{ID: "sibling", Type: "sessions.create", ServiceID: manager.Service.ID})
	data, _ = json.Marshal(p.next(t, "sibling", "result").Data)
	var sibling protocol.Conversation
	json.Unmarshal(data, &sibling)
	sub := protocol.CoordinationCommand{Operation: "subscribe", Topic: "event topic", Source: protocol.AssistantRef{ProjectID: manager.Service.ID, SessionID: sibling.ID}, Target: protocol.AssistantRef{ClientID: target.ID, ProjectID: targetCfg.Services[0].ID, SessionID: accepted.Session.ID}}
	if _, err = agent.ClientAPI(context.Background(), hostCfg, "POST", "/api/agent/coordination", sub); err == nil {
		t.Fatal("sibling subscribed to another parent's child")
	}
	f.request(bob, "POST", "/api/coordination", sub, 400)
	event.Status, event.Stage, event.Summary = "blocked", "validation", "requires missing target input"
	if _, err = agent.ClientAPI(context.Background(), targetCfg, "POST", "/api/agent/coordination", event); err != nil {
		t.Fatal(err)
	}
	select {
	case <-hostRan:
		t.Fatal("blocked event delivered while execution is still running")
	case <-time.After(100 * time.Millisecond):
	}
	// Complete through normal task shutdown; exactly one terminal review event.
	finish()
	select {
	case v := <-hostRan:
		if !strings.Contains(v, `"event":"blocked"`) {
			t.Fatal(v)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("terminal review not delivered")
	}
	select {
	case <-hostRan:
		t.Fatal("ready and blocked both woke manager")
	case <-time.After(100 * time.Millisecond):
	}
	// Real report cards retain event IDs/source/task and the evidence, not logs.
	p.send(t, protocol.Request{ID: "history", Type: "sessions.get", SessionID: source.ID})
	data, _ = json.Marshal(p.next(t, "history", "result").Data)
	var page protocol.HistoryPage
	json.Unmarshal(data, &page)
	count := 0
	for _, entry := range page.Entries {
		if entry.Report != nil {
			count++
			if entry.Report.TaskID != accepted.TaskID || entry.Report.ID == "" || entry.Report.Origin.SessionID != accepted.Session.ID {
				t.Fatal("lost routing evidence", entry.Report)
			}
		}
	}
	if count != 2 {
		t.Fatal("report card dedup failed", count)
	}
}
