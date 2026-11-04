package clientupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnsignedReleasesAcrossPlatforms(t *testing.T) {
	for _, platform := range []string{"windows", "darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run(platform+"-"+arch, func(t *testing.T) {
				body := []byte("native fixture " + platform + " " + arch)
				sum := sha256.Sum256(body)
				release := Release{Payload: Payload{SchemaVersion: 1, Sequence: 1, Version: "fixture", OS: platform, Arch: arch, MinHostVersion: 1, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:])}}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
						t.Error("download leaked credentials")
					}
					switch r.URL.Path {
					case "/updates/v1/" + platform + "/" + arch + ".json":
						_ = json.NewEncoder(w).Encode(release)
					case "/downloads/" + release.Artifact():
						_, _ = w.Write(body)
					default:
						t.Error("unexpected artifact path", r.URL.Path)
						http.NotFound(w, r)
					}
				}))
				defer server.Close()
				r, err := Fetch(context.Background(), HTTPClient(), server.URL, platform, arch)
				if err != nil {
					t.Fatal(err)
				}
				if err = Download(context.Background(), HTTPClient(), server.URL, r, io.Discard); err != nil {
					t.Fatal(err)
				}
				if err = r.Verify(platform, "386"); err == nil {
					t.Fatal("wrong architecture accepted")
				}
				r.OS = "freebsd"
				if err = r.Verify("freebsd", arch); err == nil {
					t.Fatal("unsupported OS accepted")
				}
				r = release
				r.Size = MaxSize + 1
				if err = r.Verify(platform, arch); err == nil {
					t.Fatal("oversized release accepted")
				}
			})
		}
	}
}

func TestLegacyRuntimeStateRetainsRollbackWithoutSigningKeys(t *testing.T) {
	s, makeRelease := fixture(t)
	release, body := makeRelease(7, "legacy")
	release.KeyID = "old-release-key"
	release.Signature = "legacy-field-no-longer-verified"
	if err := s.Stage(release, func(w io.Writer) error { _, err := w.Write(body); return err }); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(State{SchemaVersion: 1, Current: &release, Previous: &release, HasPrevious: true, HighestSequence: 7, HighestSHA256: release.SHA256}); err != nil {
		t.Fatal(err)
	}
	state, err := s.Load()
	if err != nil || !state.HasPrevious || state.Previous.SHA256 != release.SHA256 {
		t.Fatal("legacy rollback point was lost", err)
	}
	if _, err := s.Binary(state.Previous, ""); err != nil {
		t.Fatal("legacy binary no longer loadable", err)
	}
}

func TestUnsignedUpdatesStillRequireTrustedTLS(t *testing.T) {
	_, makeRelease := fixture(t)
	release, _ := makeRelease(1, "https-fixture")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()
	if _, err := Fetch(context.Background(), HTTPClient(), server.URL, "windows", "amd64"); err == nil {
		t.Fatal("untrusted TLS certificate accepted")
	}
	if _, err := Fetch(context.Background(), server.Client(), server.URL, "windows", "amd64"); err != nil {
		t.Fatal("trusted HTTPS manifest rejected", err)
	}
	if _, err := NormalizeServer("http://updates.example.com/agents"); err == nil {
		t.Fatal("non-loopback HTTP update authority accepted")
	}
}
