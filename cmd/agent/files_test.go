package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

func TestFileUploadCLIAndMCP(t *testing.T) {
	content := []byte("mock apk")
	path := filepath.Join(t.TempDir(), "test.apk")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	var site *httptest.Server
	calls := 0
	site = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		data, _ := io.ReadAll(r.Body)
		if r.URL.Path != "/agents/api/agent/files" || r.URL.Query().Get("id") != "device" || r.Header.Get("Authorization") != "Bearer mock-secret" || !bytes.Equal(data, content) {
			t.Error("incorrect upload")
		}
		hash := sha256.Sum256(data)
		id := store.Token()
		_ = json.NewEncoder(w).Encode(protocol.SharedFile{ID: id, URL: site.URL + "/agents/files/" + id, Name: "test.apk", Size: int64(len(data)), SHA256: hex.EncodeToString(hash[:]), ExpiresAt: time.Now().Add(time.Hour)})
	}))
	defer site.Close()
	c := agent.Config{Path: filepath.Join(t.TempDir(), "config.json"), ID: "device", Credential: "mock-secret", Server: site.URL + "/agents"}
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runFiles(context.Background(), []string{"upload", "--config", c.Path, "--path", path}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "/agents/files/") || strings.Contains(out.String(), "mock-secret") {
		t.Fatal("invalid CLI output", out.String())
	}
	call, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": "upload_file", "arguments": map[string]string{"path": path}}})
	out.Reset()
	if err := runMCP(context.Background(), []string{"--config", c.Path}, bytes.NewReader(call), &out); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || strings.Contains(out.String(), "isError") || !strings.Contains(out.String(), "/agents/files/") {
		t.Fatal("MCP did not upload", out.String())
	}
	if err := runFiles(context.Background(), []string{"upload", "--config", c.Path}, io.Discard); err == nil {
		t.Fatal("missing file accepted")
	}
}
