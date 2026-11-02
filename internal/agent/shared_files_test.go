package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFileRejectsUnsafeInputAndRedirects(t *testing.T) {
	redirected := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected = true }))
	defer target.Close()
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
	defer site.Close()
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	c := Config{Server: site.URL, ID: "device", Credential: "mock"}
	for _, candidate := range []string{"relative.txt", filepath.Dir(path), path} {
		if _, err := UploadFile(context.Background(), c, candidate); err == nil {
			t.Errorf("accepted invalid upload: %s", candidate)
		}
	}
	if redirected {
		t.Fatal("followed redirect with device credentials")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := UploadFile(ctx, c, path); err == nil {
		t.Fatal("ignored cancellation")
	}
}
