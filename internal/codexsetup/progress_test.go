package codexsetup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestInstallProgressTracksStagesAndDownloadSizes(t *testing.T) {
	for _, knownSize := range []bool{true, false} {
		t.Run(strconv.FormatBool(knownSize), func(t *testing.T) {
			events, err := installProgressFixture(t, knownSize, false)
			if err != nil {
				t.Fatal(err)
			}
			var stages []string
			var downloaded int64
			for _, event := range events {
				if len(stages) == 0 || stages[len(stages)-1] != event.Stage {
					stages = append(stages, event.Stage)
				}
				if event.Downloaded < downloaded {
					t.Fatal("download progress moved backwards")
				}
				downloaded = event.Downloaded
			}
			if !reflect.DeepEqual(stages, []string{"release", "download", "verify", "extract", "publish", "complete"}) {
				t.Fatalf("stages = %v", stages)
			}
			last := events[len(events)-1]
			if downloaded <= 0 || knownSize && last.Total != downloaded || !knownSize && last.Total != -1 {
				t.Fatalf("incorrect byte counts: %+v", last)
			}
		})
	}
}

func TestInstallProgressNeverCompletesAfterChecksumFailure(t *testing.T) {
	events, err := installProgressFixture(t, true, true)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("checksum failure = %v", err)
	}
	for _, event := range events {
		if event.Stage == "extract" || event.Stage == "publish" || event.Stage == "complete" {
			t.Fatalf("reported success after failure: %+v", event)
		}
	}
	if events[len(events)-1].Stage != "verify" {
		t.Fatalf("wrong failure stage: %v", events)
	}
}

func installProgressFixture(t *testing.T, knownSize, corrupt bool) ([]Progress, error) {
	t.Helper()
	archive := windowsPackage(t, "progress", "", nil)
	sum := sha256.Sum256(archive)
	if corrupt {
		sum[0] ^= 1
	}
	name, err := AssetName("windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	var fixture *httptest.Server
	fixture = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/release" {
			_ = json.NewEncoder(w).Encode(Release{Assets: []Asset{{Name: name, Digest: "sha256:" + hex.EncodeToString(sum[:]), URL: fixture.URL + "/archive"}}})
			return
		}
		if knownSize {
			w.Header().Set("Content-Length", strconv.Itoa(len(archive)))
		} else {
			w.(http.Flusher).Flush()
		}
		_, _ = w.Write(archive)
	}))
	defer fixture.Close()
	var events []Progress
	_, err = InstallWithProgress(context.Background(), fixture.URL+"/release", t.TempDir(), "windows", "amd64", fixture.Client(), func(progress Progress) { events = append(events, progress) })
	return events, err
}
