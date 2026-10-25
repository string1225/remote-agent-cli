package agent

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentUploadRetryScopeAndImageArguments(t *testing.T) {
	a := &Agent{Config: Config{Path: filepath.Join(t.TempDir(), "config.json")}}
	data := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...)
	first := attachmentUpload{Name: "image.png", Size: int64(len(data)), Data: base64.StdEncoding.EncodeToString(data[:12])}
	r, next, err := a.uploadAttachment("session-a", first)
	if err != nil || next != 12 || r.Complete {
		t.Fatalf("first: %d %v", next, err)
	}
	first.ID = r.ID
	if _, n, err := a.uploadAttachment("session-a", first); err != nil || n != 12 {
		t.Fatalf("retry: %d %v", n, err)
	}
	if _, _, _, err := a.resolveAttachments("session-a", []string{r.ID}); err == nil {
		t.Fatal("unfinished attachment accepted")
	}
	r, next, err = a.uploadAttachment("session-a", attachmentUpload{ID: r.ID, Offset: 12, Data: base64.StdEncoding.EncodeToString(data[12:]), Complete: true})
	if err != nil || next != int64(len(data)) || !r.Complete {
		t.Fatalf("finish: %d %v", next, err)
	}
	refs, paths, images, err := a.resolveAttachments("session-a", []string{r.ID})
	if err != nil || len(refs) != 1 || len(images) != 1 || !strings.HasSuffix(images[0], ".png") {
		t.Fatalf("resolve image: %v %v", images, err)
	}
	if raw, err := os.ReadFile(paths[0]); err != nil || string(raw) != string(data) {
		t.Fatal("upload changed bytes", err)
	}
	if _, _, _, err = a.resolveAttachments("session-b", []string{r.ID}); err == nil {
		t.Fatal("cross-session attachment accepted")
	}
	if _, _, err = a.uploadAttachment("../outside", first); err == nil {
		t.Fatal("traversal accepted")
	}
	if _, _, err = a.uploadAttachment("session-a", attachmentUpload{Name: "../bad", Size: 1, Data: "YQ==", Complete: true}); err == nil {
		t.Fatal("unsafe name accepted")
	}
	if _, _, err = a.uploadAttachment("session-a", attachmentUpload{Name: "huge", Size: maxAttachmentSize + 1, Data: "YQ=="}); err == nil {
		t.Fatal("oversized file accepted")
	}
}
