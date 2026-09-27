package agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func soFixture(t *testing.T) (*saysoStore, *peer) {
	t.Helper()
	root := t.TempDir()
	a := &Agent{Config: Config{Path: filepath.Join(root, "config.json"), AllowedRoots: []string{root}, Services: []protocol.Service{{ID: "demo", Name: "Demo", Workspace: root, Provider: "codex"}}}, Providers: map[string]Provider{}}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	p := &peer{ctx: ctx, cancel: cancel}
	s, err := a.saysoStore()
	if err != nil {
		t.Fatal(err)
	}
	return s, p
}
func soCall(t *testing.T, s *saysoStore, p *peer, method, path string, body any) json.RawMessage {
	t.Helper()
	raw, _ := json.Marshal(body)
	result, err := s.call(p, method, path, raw)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return result
}
func soNewSession(t *testing.T, s *saysoStore, p *peer) *soSession {
	t.Helper()
	var v soSession
	if err := json.Unmarshal(soCall(t, s, p, "POST", "/api/sessions", map[string]string{"projectId": s.state.Projects[0].ID}), &v); err != nil {
		t.Fatal(err)
	}
	return &v
}
func soWait(t *testing.T, s *saysoStore, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		ok := done()
		s.mu.Unlock()
		if ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("SaySo operation did not finish")
}
func TestSaysoLocalWorkflowAndPersistence(t *testing.T) {
	s, p := soFixture(t)
	v := soNewSession(t, s, p)
	if v.Status != "paused" || s.state.Settings.AnalysisProvider != "local" {
		t.Fatal("Codex must remain optional and capture opt-in")
	}
	soCall(t, s, p, "POST", "/api/sessions/"+v.ID+"/utterances", map[string]any{"text": "请今天完成移动端的登录页面", "atMs": 123})
	soWait(t, s, func() bool { return s.session(v.ID).AnalysisStatus == "done" })
	var state soState
	_ = json.Unmarshal(soCall(t, s, p, "GET", "/api/state", nil), &state)
	if len(state.Actions) != 1 || state.Actions[0].Status != "draft" || state.Actions[0].Priority != "high" {
		t.Fatalf("unexpected actions: %+v", state.Actions)
	}
	a := state.Actions[0]
	if _, err := s.call(p, "POST", "/api/actions/"+a.ID+"/dispatch", nil); err == nil {
		t.Fatal("draft dispatch accepted")
	}
	soCall(t, s, p, "PATCH", "/api/actions/"+a.ID, map[string]string{"status": "ready"})
	soCall(t, s, p, "PATCH", "/api/actions/"+a.ID, map[string]string{"detail": "增加验证码验收要求"})
	if s.action(a.ID).Status != "draft" {
		t.Fatal("edited action stayed ready")
	}
	s2 := &saysoStore{a: s.a, dir: s.dir, uploads: map[string]*soUpload{}}
	if err := s2.load(); err != nil {
		t.Fatal(err)
	}
	if len(s2.state.Sessions) != 1 || len(s2.state.Sessions[0].Transcript) != 1 || s2.state.Actions[0].Detail != "增加验证码验收要求" {
		t.Fatal("local history did not survive reopen")
	}
	bad, _ := json.Marshal(map[string]string{"name": "outside", "rootPath": t.TempDir()})
	if _, err := s.call(p, "POST", "/api/projects", bad); err == nil {
		t.Fatal("allowed roots bypassed")
	}
}
func TestSaysoAudioChunkOwnershipAndRoundtrip(t *testing.T) {
	s, p := soFixture(t)
	v := soNewSession(t, s, p)
	payload := bytes.Repeat([]byte{0, 1, 255, 7}, 10001)
	var begin struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(soCall(t, s, p, "POST", "/api/audio/begin", map[string]any{"sessionId": v.ID, "sizeBytes": len(payload), "mimeType": "audio/webm;codecs=opus", "startedAt": soNow(), "endedAt": soNow(), "durationMs": 1000}), &begin)
	path := "/api/audio/" + begin.ID
	otherCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	other := &peer{ctx: otherCtx}
	if _, err := s.call(other, "POST", path+"/finish", nil); err == nil {
		t.Fatal("another peer finalized upload")
	}
	if _, err := s.call(p, "POST", path+"/finish", nil); err == nil {
		t.Fatal("incomplete upload finalized")
	}
	bad, _ := json.Marshal(map[string]any{"offset": 1, "data": "AA=="})
	if _, err := s.call(p, "POST", path+"/chunk", bad); err == nil {
		t.Fatal("out-of-order chunk accepted")
	}
	for offset := 0; offset < len(payload); offset += 16000 {
		soCall(t, s, p, "POST", path+"/chunk", map[string]any{"offset": offset, "data": base64.StdEncoding.EncodeToString(payload[offset:min(offset+16000, len(payload))])})
	}
	soCall(t, s, p, "POST", path+"/finish", nil)
	var got []byte
	for len(got) < len(payload) {
		var chunk struct {
			Data []byte `json:"data"`
		}
		_ = json.Unmarshal(soCall(t, s, p, "POST", path+"/read", map[string]int{"offset": len(got)}), &chunk)
		got = append(got, chunk.Data...)
	}
	if !bytes.Equal(payload, got) {
		t.Fatal("audio corrupted")
	}
	if _, err := s.call(p, "POST", "/api/audio/../../config.json/read", []byte(`{"offset":0}`)); err == nil {
		t.Fatal("path traversal accepted")
	}
	if _, err := s.call(p, "POST", "/api/audio/unknown/read", []byte(`{"offset":0}`)); err == nil {
		t.Fatal("unindexed file accepted")
	}
}
func TestSaysoAnalysisMergePermissionsAndCancellation(t *testing.T) {
	if !json.Valid([]byte(soSchema)) {
		t.Fatal("analysis output schema is invalid")
	}
	s, p := soFixture(t)
	s.a.Providers["codex"] = Provider{Name: "codex", Available: true}
	s.state.Settings.AnalysisProvider = "codex"
	v := soNewSession(t, s, p)
	soCall(t, s, p, "POST", "/api/sessions/"+v.ID+"/actions", map[string]string{"title": "登录页面", "detail": "初始需求"})
	a := s.state.Actions[0]
	s.a.Execute = func(ctx context.Context, provider Provider, service protocol.Service, prompt string, write bool, send func(protocol.Event) error, id string) error {
		if write {
			t.Error("analysis bypassed read-only policy")
		}
		if service.OutputSchema == "" {
			t.Error("missing output schema")
		}
		output, _ := json.Marshal(map[string]any{"actions": []soCandidate{{ActionID: a.ID, Title: "登录页面", Detail: "补充手机布局及验收", Source: "请支持手机上的登录页面", Priority: "medium"}}})
		if err := send(protocol.Event{Type: "output", Data: map[string]string{"stream": "stdout", "text": "{\"type\":\"item.completed\",\"item\":{\"type\":\"reasoning\",\"text\":\"分析用户请求\"}}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"正在整理\"}}\n"}}); err != nil {
			return err
		}
		event, _ := json.Marshal(map[string]any{"type": "item.completed", "item": map[string]string{"type": "agent_message", "text": string(output)}})
		return send(protocol.Event{Type: "output", Data: map[string]string{"stream": "stdout", "text": string(event) + "\n"}})
	}
	soCall(t, s, p, "POST", "/api/sessions/"+v.ID+"/utterances", map[string]string{"text": "请支持手机上的登录页面"})
	soWait(t, s, func() bool { return s.session(v.ID).AnalysisStatus == "done" })
	if len(s.state.Actions) != 1 || !strings.Contains(a.Detail, "手机布局") || len(s.session(v.ID).AnalysisTraces[0].UpdatedActionIDs) != 1 {
		t.Fatal("analysis failed to merge existing action")
	}
	started := make(chan struct{})
	s.a.Execute = func(ctx context.Context, _ Provider, _ protocol.Service, _ string, write bool, _ func(protocol.Event) error, _ string) error {
		if write {
			t.Error("dispatch escalated local write permission")
		}
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}
	soCall(t, s, p, "PATCH", "/api/actions/"+a.ID, map[string]string{"status": "ready"})
	soCall(t, s, p, "POST", "/api/actions/"+a.ID+"/dispatch", nil)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("dispatch did not start")
	}
	s.a.mu.Lock()
	busy := s.a.busy
	s.a.mu.Unlock()
	if !busy {
		t.Fatal("shared run limit not held")
	}
	p.cancel()
	soWait(t, s, func() bool { return a.Status == "failed" })
	if !strings.Contains(s.state.Runs[0].Error, "canceled") {
		t.Fatal("disconnect did not stop action")
	}
}
func TestSaysoCorruptDataPreserved(t *testing.T) {
	s, _ := soFixture(t)
	path := filepath.Join(s.dir, "state.json")
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	s2 := &saysoStore{a: s.a, dir: s.dir}
	if err := s2.load(); err == nil {
		t.Fatal("corrupt data silently replaced")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "broken" {
		t.Fatal("original data lost")
	}
}
