package server_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestSharedFilePublicDownloadAndDailyCleanup(t *testing.T) {
	f := setupAt(t, "/agents")
	f.s.Config.SharedFiles = t.TempDir()
	device, credential, _, _ := f.connectedDevice(f.user("uploader"))
	content := []byte("apk-test-content\x00\xff")
	path := filepath.Join(t.TempDir(), "安装包.apk")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	c := agent.Config{ID: device.ID, Credential: credential, Server: f.url}
	file, err := agent.UploadFile(context.Background(), c, path)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(content)
	if len(file.ID) != 43 || file.Name != "安装包.apk" || file.SHA256 != hex.EncodeToString(hash[:]) || file.URL != f.url+"/files/"+file.ID {
		t.Fatal("incorrect upload capability", file)
	}
	local := file.ExpiresAt.In(time.FixedZone("CN", 8*3600))
	if local.Hour() != 4 || local.Minute() != 0 || time.Until(file.ExpiresAt) > 24*time.Hour {
		t.Fatal("incorrect expiry", file.ExpiresAt)
	}
	req, _ := http.NewRequest("GET", file.URL, nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Range", "bytes=0-2")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 206 || !bytes.Equal(data, content[:3]) || !strings.HasPrefix(res.Header.Get("Content-Disposition"), "attachment;") || res.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("public range download failed", res.StatusCode, res.Header, data)
	}
	if err := f.s.PruneSharedFiles(file.ExpiresAt.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	res, err = http.Get(file.URL)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if !bytes.Equal(data, content) {
		t.Fatal("live file removed before cutoff")
	}
	if err := f.s.PruneSharedFiles(file.ExpiresAt); err != nil {
		t.Fatal(err)
	}
	res, err = http.Get(file.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatal("expired file still public")
	}
	entries, _ := os.ReadDir(f.s.Config.SharedFiles)
	if len(entries) != 0 {
		t.Fatal("cleanup retained files", entries)
	}
}

func TestSharedFileRejectsInvalidUploadsAndLeaksNoPartialFile(t *testing.T) {
	f := setup(t)
	f.s.Config.SharedFiles = t.TempDir()
	device, credential, _, _ := f.connectedDevice(f.user("owner"))
	for _, tc := range []struct {
		name, credential, filename, contentType string
		size                                    int64
		status                                  int
	}{
		{"anonymous", "", "a.txt", "application/octet-stream", 3, 401},
		{"wrong credential", "wrong", "a.txt", "application/octet-stream", 3, 401},
		{"traversal", credential, "../a.txt", "application/octet-stream", 3, 400},
		{"control", credential, "a\nb.txt", "application/octet-stream", 3, 400},
		{"content type", credential, "a.txt", "application/json", 3, 415},
		{"size", credential, "a.txt", "application/octet-stream", protocol.MaxSharedFileBytes + 1, 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", f.url+"/api/agent/files?id="+device.ID, strings.NewReader("abc"))
			req.Header.Set("Authorization", "Bearer "+tc.credential)
			req.Header.Set("X-File-Name", url.QueryEscape(tc.filename))
			req.Header.Set("Content-Type", tc.contentType)
			if tc.size > 3 {
				// A header-only request proves the server rejects before reading.
				req.Body = io.NopCloser(io.LimitReader(zeroReader{}, tc.size))
				req.ContentLength = tc.size
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
			if res.StatusCode != tc.status {
				t.Fatal(res.StatusCode, tc.status)
			}
		})
	}
	// Chunked clients cannot bypass the limit by omitting Content-Length.
	req, _ := http.NewRequest("POST", f.url+"/api/agent/files?id="+device.ID, io.LimitReader(zeroReader{}, protocol.MaxSharedFileBytes+1))
	req.Header.Set("Authorization", "Bearer "+credential)
	req.Header.Set("X-File-Name", "large.bin")
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 413 {
		t.Fatal("chunked limit bypass", res.StatusCode)
	}
	_ = filepath.WalkDir(f.s.Config.SharedFiles, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			t.Error("partial upload retained", path)
		}
		return err
	})
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestSharedFilesExpireWithoutCleanupAndRetainOtherData(t *testing.T) {
	f := setup(t)
	f.s.Config.SharedFiles = t.TempDir()
	device, credential, _, _ := f.connectedDevice(f.user("owner"))
	path := filepath.Join(t.TempDir(), "empty.txt")
	_ = os.WriteFile(path, nil, 0600)
	file, err := agent.UploadFile(context.Background(), agent.Config{ID: device.ID, Credential: credential, Server: f.url}, path)
	if err != nil {
		t.Fatal(err)
	}
	metadataPaths, _ := filepath.Glob(filepath.Join(f.s.Config.SharedFiles, "*", file.ID, "metadata.json"))
	if len(metadataPaths) != 1 {
		t.Fatal("metadata missing")
	}
	file.ExpiresAt = time.Now().Add(-time.Second)
	raw, _ := json.Marshal(file)
	_ = os.WriteFile(metadataPaths[0], raw, 0600)
	res, err := http.Get(file.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Fatal("expired capability still works")
	}
	// Restart catch-up removes old buckets, and never touches unrelated data.
	_ = os.MkdirAll(filepath.Join(f.s.Config.SharedFiles, "1000000000", "partial"), 0700)
	_ = os.WriteFile(filepath.Join(f.s.Config.SharedFiles, "keep.txt"), []byte("keep"), 0600)
	if err := f.s.PruneSharedFiles(time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(f.s.Config.SharedFiles, "1000000000")); !os.IsNotExist(err) {
		t.Fatal("old bucket retained")
	}
	if _, err := os.Stat(filepath.Join(f.s.Config.SharedFiles, "keep.txt")); err != nil {
		t.Fatal("unrelated file removed")
	}
}
