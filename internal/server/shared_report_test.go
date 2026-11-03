package server_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

func TestSharedHTMLReportPreviewIsSandboxedAndDownloadStaysAvailable(t *testing.T) {
	f := setupAt(t, "/agents")
	f.s.Config.SharedFiles = t.TempDir()
	device, credential, _, _ := f.connectedDevice(f.user("reporter"))
	cfg := agent.Config{ID: device.ID, Credential: credential, Server: f.url}
	for _, name := range []string{"方案.HTML", "report.htm", "payload.svg", "note.txt"} {
		t.Run(name, func(t *testing.T) {
			content := "<!doctype html><style>body{color:navy}</style><h1>方案</h1><script>fetch('/api/me')</script>"
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			file, err := agent.UploadFile(context.Background(), cfg, path)
			if err != nil {
				t.Fatal(err)
			}
			html := strings.HasSuffix(name, ".HTML") || strings.HasSuffix(name, ".htm")
			if html && file.PreviewURL != file.URL+"?view=1" || !html && file.PreviewURL != "" {
				t.Fatal("incorrect preview capability", file)
			}
			for _, method := range []string{"GET", "HEAD"} {
				req, _ := http.NewRequest(method, file.URL+"?view=1", nil)
				req.Header.Set("Sec-Fetch-Site", "cross-site")
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				data, _ := io.ReadAll(res.Body)
				res.Body.Close()
				if !html {
					if res.StatusCode != http.StatusUnsupportedMediaType {
						t.Fatal("non-HTML preview accepted", res.StatusCode)
					}
					continue
				}
				if res.StatusCode != 200 || res.Header.Get("Content-Type") != "text/html; charset=utf-8" || !strings.HasPrefix(res.Header.Get("Content-Disposition"), "inline;") {
					t.Fatal("HTML preview unavailable", res.StatusCode, res.Header)
				}
				policy := res.Header.Get("Content-Security-Policy")
				for _, required := range []string{"default-src 'none'", "sandbox", "style-src 'unsafe-inline'", "form-action 'none'", "base-uri 'none'"} {
					if !strings.Contains(policy, required) {
						t.Fatal("missing report isolation", required, policy)
					}
				}
				if strings.Contains(policy, "allow-") || strings.Contains(policy, "script-src") || res.Header.Get("X-Content-Type-Options") != "nosniff" || res.Header.Get("Cache-Control") != "no-store" {
					t.Fatal("report can escape static preview", res.Header)
				}
				if method == "GET" && string(data) != content || method == "HEAD" && len(data) != 0 {
					t.Fatal("unexpected preview body", method, string(data))
				}
			}
			res, err := http.Get(file.URL)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
			if res.Header.Get("Content-Type") != "application/octet-stream" || !strings.HasPrefix(res.Header.Get("Content-Disposition"), "attachment;") {
				t.Fatal("default download changed", res.Header)
			}
		})
	}
}
