package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestSensitiveSharingRequiresConfirmationBeforeReadingFileOrSending(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	for _, o := range []FileUploadOptions{{Class: "sensitive"}, {Class: "invalid", UserConfirmed: true}} {
		if _, err := UploadFileWithOptions(context.Background(), Config{Server: server.URL}, "does-not-exist", o); err == nil {
			t.Fatal("unconfirmed/invalid sharing accepted")
		}
	}
	if calls != 0 {
		t.Fatal("unconfirmed file sent")
	}
	for _, scope := range []string{"", "account", "client", "project"} {
		instructions := SessionInstructionsFor(protocol.Service{Scope: scope})
		if !strings.Contains(instructions, FileSharingInstructions) {
			t.Fatal("sharing rules missing")
		}
	}
}
