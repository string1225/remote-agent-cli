package clientupdate

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignedReleasesAcrossPlatforms(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]ed25519.PublicKey{"fixture": pub}
	for _, platform := range []string{"windows", "darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run(platform+"-"+arch, func(t *testing.T) {
				body := []byte("native fixture " + platform + " " + arch)
				sum := sha256.Sum256(body)
				release := Sign(Payload{SchemaVersion: 1, Sequence: 1, Version: "fixture", OS: platform, Arch: arch, MinHostVersion: 1, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:])}, "fixture", key)
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
				r, err := Fetch(context.Background(), HTTPClient(), server.URL, platform, arch, keys)
				if err != nil {
					t.Fatal(err)
				}
				if err = Download(context.Background(), HTTPClient(), server.URL, r, io.Discard); err != nil {
					t.Fatal(err)
				}
				if err = r.Verify(keys, platform, "386"); err == nil {
					t.Fatal("wrong architecture accepted")
				}
				r.OS = "freebsd"
				if err = r.Verify(keys, "freebsd", arch); err == nil {
					t.Fatal("unsupported OS accepted")
				}
				r = release
				r.Version = "tampered"
				if err = r.Verify(keys, platform, arch); err == nil {
					t.Fatal("tampered release accepted")
				}
			})
		}
	}
}
