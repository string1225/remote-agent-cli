package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

type filePolicyTransport func(*http.Request) (*http.Response, error)

func (f filePolicyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTemporaryCLIAndMCPDoNotUseEnrollmentCredentials(t *testing.T) {
	var calls, uploads, deletes int
	transport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = transport })
	http.DefaultTransport = filePolicyTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") || r.Header.Get("Cookie") != "" {
			t.Error("platform credential sent to temporary provider")
		}
		value := any(map[string]any{"success": true})
		if r.URL.Host == "storage.to" {
			switch r.URL.Path {
			case "/api/upload/init":
				value = map[string]any{"success": true, "type": "single", "r2_key": "fixture", "upload_url": "https://0123456789abcdef0123456789abcdef.r2.cloudflarestorage.com/put"}
			case "/api/upload/confirm":
				uploads++
				var input struct {
					Name string `json:"filename"`
					Size int64  `json:"size"`
				}
				_ = json.NewDecoder(r.Body).Decode(&input)
				id := fmt.Sprintf("Code%05d", uploads)
				value = map[string]any{"success": true, "owner_token": "fixture-management-proof", "file": map[string]any{"id": id, "url": "https://storage.to/" + id, "filename": input.Name, "size": input.Size, "expires_at": time.Now().Add(time.Hour)}}
			default:
				if r.Method == "DELETE" {
					deletes++
					if r.Header.Get("Authorization") != "Owner fixture-management-proof" {
						t.Error("delete lost ownership")
					}
				} else {
					t.Error("unknown API path")
				}
			}
		} else {
			if r.URL.Host != "0123456789abcdef0123456789abcdef.r2.cloudflarestorage.com" || r.Header.Get("X-Visitor-Token") != "" || r.Header.Get("X-Owner-Token") != "" {
				t.Error("unexpected destination or R2 credential leak")
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != "generated fixture" {
				t.Error("upload bytes changed")
			}
		}
		raw, _ := json.Marshal(value)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(raw))}, nil
	})
	path := filepath.Join(t.TempDir(), "source.txt")
	_ = os.WriteFile(path, []byte("generated fixture"), 0600)
	cfg := filepath.Join(t.TempDir(), "config.json")
	_ = os.WriteFile(cfg, []byte("must not parse enrollment for temporary files"), 0600)
	var output bytes.Buffer
	if err := runFiles(context.Background(), []string{"upload", "--path", path, "--config", cfg}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"provider":"storage.to"`) || strings.Contains(output.String(), "fixture-management-proof") {
		t.Fatal("CLI default route or proof isolation failed")
	}
	before := calls
	if err := runFiles(context.Background(), []string{"upload", "--path", path, "--config", cfg, "--class", "sensitive"}, io.Discard); err == nil || calls != before {
		t.Fatal("CLI skipped user confirmation")
	}
	config := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "device", Credential: "fixture-device-secret", Server: "https://platform.test"}
	if err := config.Save(); err != nil {
		t.Fatal(err)
	}
	rpc := func(name string, arguments map[string]any) string {
		t.Helper()
		call, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": name, "arguments": arguments}})
		output.Reset()
		if err := runMCP(context.Background(), []string{"--config", config.Path}, bytes.NewReader(call), &output); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	if text := rpc("upload_file", map[string]any{"path": path}); strings.Contains(text, "isError") || !strings.Contains(text, "storage.to") {
		t.Fatal("MCP default route failed", text)
	}
	before = calls
	if text := rpc("upload_file", map[string]any{"path": path, "class": "sensitive"}); !strings.Contains(text, "isError") || calls != before {
		t.Fatal("MCP skipped user confirmation")
	}
	if text := rpc("delete_uploaded_file", map[string]any{"fileId": "Code00002"}); strings.Contains(text, "isError") || deletes != 1 {
		t.Fatal("MCP owned deletion failed", text)
	}
}
