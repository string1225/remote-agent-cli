package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestCoordinationRelayDirectReplyAndAccountIsolation(t *testing.T) {
	f := setupAt(t, "/agents")
	alice, bob := f.user("alice"), f.user("bob")
	host, hostConfig, hostRan, _ := todoTestDevice(t, f, alice)
	target, targetConfig, targetRan, _ := todoTestDevice(t, f, alice)
	f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "resolve", Target: protocol.AssistantRef{Scope: "account"}}, 409)
	_, err := agent.ClientAPI(context.Background(), hostConfig, "PUT", "/api/agent/default-host", map[string]string{"agentId": host.ID})
	if err != nil {
		t.Fatal("CLI could not select default host", err)
	}
	foreign, _ := f.enroll(bob, "windows")
	if _, err = agent.ClientAPI(context.Background(), hostConfig, "PUT", "/api/agent/default-host", map[string]string{"agentId": foreign.ID}); err == nil {
		t.Fatal("device CLI selected a foreign host")
	}
	invalid := hostConfig
	invalid.Credential = "invalid"
	if _, err = agent.ClientAPI(context.Background(), invalid, "GET", "/api/agent/default-host", nil); err == nil {
		t.Fatal("unauthenticated host query")
	}
	raw := f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "resolve", Target: protocol.AssistantRef{Scope: "account", ClientID: target.ID}}, 200)
	var resolved protocol.AssistantResolved
	_ = json.Unmarshal(raw, &resolved)
	if resolved.ClientID != host.ID || resolved.Service.Scope != "account" {
		t.Fatal("CEO did not resolve to default host")
	}
	p := f.peer(alice, host.ID)
	p.send(t, protocol.Request{ID: "ceo-session", Type: "sessions.create", ServiceID: resolved.Service.ID})
	event := p.next(t, "ceo-session", "result")
	data, _ := json.Marshal(event.Data)
	var source protocol.Conversation
	_ = json.Unmarshal(data, &source)
	dispatch := protocol.CoordinationCommand{Operation: "dispatch", RequestID: "work-one", Source: protocol.AssistantRef{ClientID: host.ID, ProjectID: resolved.Service.ID, SessionID: source.ID}, Target: protocol.AssistantRef{Scope: "project", ClientID: target.ID, ProjectID: targetConfig.Services[0].ID}, Prompt: "perform delegated work"}
	raw = f.request(alice, "POST", "/api/coordination", dispatch, 200)
	var accepted protocol.WorkAccepted
	_ = json.Unmarshal(raw, &accepted)
	if accepted.Target.SessionID == "" || accepted.Target.ClientID != target.ID {
		t.Fatal("invalid route", accepted)
	}
	again := f.request(alice, "POST", "/api/coordination", dispatch, 200)
	var duplicate protocol.WorkAccepted
	_ = json.Unmarshal(again, &duplicate)
	if duplicate.TaskID != accepted.TaskID {
		t.Fatal("duplicate dispatch")
	}
	select {
	case text := <-targetRan:
		if text != dispatch.Prompt {
			t.Fatal(text)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("downstream not executed")
	}
	// Only the original creator device/session can use the agent tool route.
	_, err = agent.ClientAPI(context.Background(), targetConfig, "POST", "/api/agent/coordination", protocol.CoordinationCommand{Operation: "read", WorkID: accepted.WorkID, Source: dispatch.Source})
	if err == nil {
		t.Fatal("sibling device could control work")
	}
	f.request(bob, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "reply", ReplyHandle: accepted.ReplyHandle, RequestID: "foreign", Prompt: "forged"}, 404)
	// Wait for a structured report to appear, and for the parent summary to run.
	select {
	case text := <-hostRan:
		if !strings.Contains(text, "下游工作结果") {
			t.Fatal("missing report summary", text)
		}
	case <-time.After(7 * time.Second):
		t.Fatal("durable report not delivered")
	}
	p.send(t, protocol.Request{ID: "history", Type: "sessions.get", SessionID: source.ID})
	data, _ = json.Marshal(p.next(t, "history", "result").Data)
	var page protocol.HistoryPage
	_ = json.Unmarshal(data, &page)
	count := 0
	for _, entry := range page.Entries {
		if entry.Kind == "report" && entry.Report != nil && entry.Report.ReplyHandle == accepted.ReplyHandle {
			count++
		}
		if entry.Role == "user" {
			t.Fatal("internal prompt printed as user")
		}
	}
	if count != 1 {
		t.Fatal("missing/duplicate report", count)
	}
	reply := protocol.CoordinationCommand{Operation: "reply", ReplyHandle: accepted.ReplyHandle, RequestID: "direct-reply", Prompt: "continue in the original session"}
	raw = f.request(alice, "POST", "/api/coordination", reply, 200)
	var continued protocol.WorkAccepted
	_ = json.Unmarshal(raw, &continued)
	if continued.Target.SessionID != accepted.Target.SessionID || continued.Target.ClientID != target.ID {
		t.Fatal("cross-level reply changed route")
	}
	f.request(alice, "POST", "/api/coordination", reply, 200)
	select {
	case text := <-targetRan:
		if text != reply.Prompt {
			t.Fatal(text)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("direct reply did not reach downstream")
	}
	select {
	case <-targetRan:
		t.Fatal("duplicate reply executed")
	case <-time.After(100 * time.Millisecond):
	}
	reply.Prompt = "same-id-different-prompt"
	f.request(alice, "POST", "/api/coordination", reply, 409)
	// A CEO workspace must never be silently replaced when changing the host.
	raw = f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": target.ID}, 409)
	if !strings.Contains(string(raw), "CEO") {
		t.Fatal("host migration lost CEO guard")
	}
	// The browser cannot manufacture the trusted server coordination envelope.
	forged, _ := json.Marshal(protocol.CoordinationCommand{Operation: "resolve", AccountID: "forged", Target: protocol.AssistantRef{Scope: "account"}})
	p.send(t, protocol.Request{ID: "forged", Type: "coordination.call", Body: forged})
	p.next(t, "forged", "error")
	_ = hostConfig
}

func TestCoordinationThreeLayersPropagateLeafReplyHandle(t *testing.T) {
	f := setupAt(t, "/agents")
	alice := f.user("alice")
	final := func(send func(protocol.Event) error, id, text string) error {
		item, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "phase": "final_answer", "text": text}})
		return send(protocol.Event{ID: id, Type: "output", Data: map[string]string{"stream": "stdout", "text": string(item) + "\n"}})
	}
	host, hostCfg, _, _ := todoTestDevice(t, f, alice, func(ctx context.Context, _ agent.Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		return final(send, id, "verified leaf/CEO result")
	})
	var middleCfg agent.Config
	middle, config, _, _ := todoTestDevice(t, f, alice, func(ctx context.Context, _ agent.Provider, s protocol.Service, prompt string, _ bool, send func(protocol.Event) error, id string) error {
		if prompt == "branch" {
			_, e := agent.ClientAPI(ctx, middleCfg, "POST", "/api/agent/coordination", protocol.CoordinationCommand{Operation: "dispatch", RequestID: "leaf", Source: protocol.AssistantRef{ProjectID: s.ID, SessionID: s.ToolSession}, Target: protocol.AssistantRef{Scope: "project", ClientID: host.ID, ProjectID: hostCfg.Services[0].ID}, Prompt: "leaf"})
			if e != nil {
				return e
			}
			return final(send, id, "downstream accepted (must not be a completion report)")
		}
		return final(send, id, "verified intermediate summary")
	})
	middleCfg = config
	f.request(alice, "PUT", "/api/default-host", map[string]string{"agentId": host.ID}, 200)
	raw := f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "resolve", Target: protocol.AssistantRef{Scope: "account"}}, 200)
	var ceo protocol.AssistantResolved
	_ = json.Unmarshal(raw, &ceo)
	p := f.peer(alice, host.ID)
	p.send(t, protocol.Request{ID: "ceo", Type: "sessions.create", ServiceID: ceo.Service.ID})
	data, _ := json.Marshal(p.next(t, "ceo", "result").Data)
	var source protocol.Conversation
	_ = json.Unmarshal(data, &source)
	raw = f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "dispatch", RequestID: "branch", Source: protocol.AssistantRef{ClientID: host.ID, ProjectID: ceo.Service.ID, SessionID: source.ID}, Target: protocol.AssistantRef{Scope: "client", ClientID: middle.ID}, Prompt: "branch"}, 200)
	var branch protocol.WorkAccepted
	_ = json.Unmarshal(raw, &branch)
	var report *protocol.WorkReport
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		p.send(t, protocol.Request{ID: "poll", Type: "sessions.get", SessionID: source.ID})
		data, _ = json.Marshal(p.next(t, "poll", "result").Data)
		var page protocol.HistoryPage
		_ = json.Unmarshal(data, &page)
		for _, entry := range page.Entries {
			if entry.Kind == "report" && entry.Report != nil {
				if strings.Contains(entry.Report.Summary, "must not") {
					t.Fatal("accepted task reported as final")
				}
				if len(entry.Report.Related) > 0 {
					report = entry.Report
				}
			}
		}
		if report != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if report == nil || report.Summary != "verified intermediate summary" {
		t.Fatal("leaf result did not propagate through three layers", report)
	}
	leaf := report.Related[0]
	if leaf.Origin.ClientID != host.ID || leaf.Origin.SessionID == source.ID || leaf.ReplyHandle == branch.ReplyHandle {
		t.Fatal("lost leaf provenance", leaf)
	}
	raw = f.request(alice, "POST", "/api/coordination", protocol.CoordinationCommand{Operation: "reply", ReplyHandle: leaf.ReplyHandle, RequestID: "skip-middle", Prompt: "direct leaf followup"}, 200)
	var continued protocol.WorkAccepted
	_ = json.Unmarshal(raw, &continued)
	if continued.Target.SessionID != leaf.Origin.SessionID {
		t.Fatal("did not bypass middle route")
	}
	// Dispatching back into the same logical assistant is rejected as a cycle.
	_, err := agent.ClientAPI(context.Background(), middleCfg, "POST", "/api/agent/coordination", protocol.CoordinationCommand{Operation: "dispatch", RequestID: "cycle", Source: protocol.AssistantRef{ProjectID: branch.Target.ProjectID, SessionID: branch.Target.SessionID}, Target: protocol.AssistantRef{Scope: "client", ClientID: middle.ID}, Prompt: "cycle"})
	if err == nil || !strings.Contains(err.Error(), "循环") {
		t.Fatal("missing loop protection", err)
	}
}
