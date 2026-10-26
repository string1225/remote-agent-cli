package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
	bolt "go.etcd.io/bbolt"
)

func TestMCPToolsUseEnrolledSiteAndProjectSessionContext(t *testing.T) {
	var operations []string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-device-secret" || r.URL.Query().Get("id") != "source-client" {
			t.Error("wrong device identity")
		}
		operations = append(operations, r.URL.Path)
		if r.URL.Path == "/api/agent/catalog" {
			io.WriteString(w, `{"clients":[{"id":"target","projects":[{"id":"project-b"}]}]}`)
			return
		}
		var in struct {
			AgentID string         `json:"agentId"`
			Type    string         `json:"type"`
			Body    map[string]any `json:"body"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			t.Error("invalid tool request")
		}
		if in.Type == "tasks.start" {
			if in.AgentID != "target" || in.Body["serviceId"] != "project-b" || in.Body["parentSessionId"] != "origin-im" || in.Body["source"] != "im" || in.Body["notify"] != true || in.Body["notifySessionId"] != "origin-im" {
				t.Error("lost caller or target context")
			}
			io.WriteString(w, `{"session":{"id":"new-web","kind":"web"},"taskId":"task"}`)
		} else if in.Type == "channels.notify" {
			if in.AgentID != "source-client" || in.Body["serviceId"] != "project-a" || in.Body["sessionId"] != "origin-im" || in.Body["message"] != "done" {
				t.Error("notification not scoped to current project")
			}
			io.WriteString(w, `{"queued":true}`)
		} else {
			t.Error("unexpected operation")
		}
	}))
	defer up.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), Server: up.URL, ID: "source-client", Credential: "local-device-secret"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_clients_projects","arguments":{}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"start_web_session","arguments":{"clientId":"target","projectId":"project-b","prompt":"work"}}}
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"done"}}}
`)
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "project-a", "--session", "origin-im", "--source", "im"}, input, &out); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 3 || !strings.Contains(out.String(), "new-web") || strings.Contains(out.String(), "local-device-secret") {
		t.Fatal("tool calls missing or leaked credentials")
	}
	decoder := json.NewDecoder(&out)
	count := 0
	for decoder.More() {
		var v map[string]any
		if err := decoder.Decode(&v); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 5 {
		t.Fatal("notifications produced replies or lost responses", count)
	}
}

func TestMCPDelegationRetainsOriginIMConversation(t *testing.T) {
	operations := 0
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			AgentID string         `json:"agentId"`
			Type    string         `json:"type"`
			Body    map[string]any `json:"body"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			t.Error("invalid request")
		}
		operations++
		if in.Type == "tasks.start" {
			if in.Body["parentSessionId"] != "web-child" || in.Body["notifySessionId"] != "original-im" || in.Body["notify"] != true || in.Body["notifyAgentId"] != "origin-client" || in.Body["notifyServiceId"] != "origin-project" {
				t.Error("nested task lost original IM route")
			}
		} else if in.AgentID != "origin-client" || in.Body["serviceId"] != "origin-project" || in.Body["sessionId"] != "original-im" {
			t.Error("progress notification used worker Web session")
		}
		io.WriteString(w, `{"ok":true}`)
	}))
	defer site.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), Server: site.URL, ID: "worker-client", Credential: "mock"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"start_web_session","arguments":{"prompt":"nested work"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"notify_im_session","arguments":{"message":"progress"}}}`)
	var out bytes.Buffer
	if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "worker-project", "--session", "web-child", "--source", "web", "--notify-client", "origin-client", "--notify-project", "origin-project", "--notify-session", "original-im"}, input, &out); err != nil {
		t.Fatal(err)
	}
	if operations != 2 {
		t.Fatal("missing nested operations", out.String())
	}
}

func TestMCPFinalNotificationOnlySuppressesAfterOriginAccepts(t *testing.T) {
	for _, tc := range []struct {
		name               string
		final, fail, other bool
		want               bool
	}{
		{name: "progress"}, {name: "final", final: true, want: true}, {name: "failed", final: true, fail: true}, {name: "other-session", final: true, other: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.fail {
					w.WriteHeader(503)
					io.WriteString(w, `{"error":"offline"}`)
					return
				}
				io.WriteString(w, `{"queued":true}`)
			}))
			defer site.Close()
			root := t.TempDir()
			c := agent.Config{Path: filepath.Join(root, "config.json"), Server: site.URL, ID: "worker", Credential: "mock"}
			if err := c.Save(); err != nil {
				t.Fatal(err)
			}
			history := filepath.Join(root, "history.db")
			db, err := bolt.Open(history, 0600, nil)
			if err != nil {
				t.Fatal(err)
			}
			err = db.Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists([]byte("sessions"))
				if err != nil {
					return err
				}
				return b.Put([]byte("web"), []byte(`{"id":"web","taskId":"task"}`))
			})
			db.Close()
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]any{"message": "结论、验证和提交", "final": tc.final}
			if tc.other {
				args["sessionId"] = "other-im"
			}
			request, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": "notify_im_session", "arguments": args}})
			var output bytes.Buffer
			if err := runMCP(context.Background(), []string{"--config", c.Path, "--project", "work", "--session", "web", "--source", "web", "--notify-client", "origin", "--notify-project", "original-project", "--notify-session", "original-im"}, bytes.NewReader(append(request, '\n')), &output); err != nil {
				t.Fatal(err)
			}
			db, err = bolt.Open(history, 0600, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			notified := false
			err = db.View(func(tx *bolt.Tx) error {
				if b := tx.Bucket([]byte("task_results")); b != nil {
					var result struct{ Notified bool }
					if data := b.Get([]byte("task")); data != nil {
						if err := json.Unmarshal(data, &result); err != nil {
							return err
						}
						notified = result.Notified
					}
				}
				return nil
			})
			if err != nil || notified != tc.want {
				t.Fatal("incorrect automatic completion suppression", notified, err, output.String())
			}
		})
	}
}
