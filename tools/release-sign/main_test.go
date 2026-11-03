package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
)

func TestSigningDirectorySelectsOnlyTrustedKey(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	_, wrong, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	write := func(name string, key ed25519.PrivateKey) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(hex.EncodeToString(key.Seed())+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("a-unrelated.seed", wrong)
	keys := map[string]ed25519.PublicKey{"release-2026": public}
	if _, err := trustedSigningKey("", dir, "release-2026", keys); err == nil {
		t.Fatal("accepted an unrelated signing key")
	}
	write("z-original.seed", private)
	got, err := trustedSigningKey("", dir, "release-2026", keys)
	if err != nil || !got.Equal(private) {
		t.Fatal("did not select the original trust anchor", err)
	}
}

func TestVerifyReleaseRejectsTamperingAndWrongRelease(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	keys := map[string]ed25519.PublicKey{"release-2026": public}
	dir := t.TempDir()
	raw := []byte("isolated runtime fixture")
	sum := sha256.Sum256(raw)
	p := clientupdate.Payload{SchemaVersion: 1, Sequence: 10, Version: "test-10", OS: "windows", Arch: "amd64", MinHostVersion: 2, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(raw)), SHA256: hex.EncodeToString(sum[:])}
	manifest := filepath.Join(dir, "amd64.json")
	encoded, _ := json.Marshal(clientupdate.Sign(p, "release-2026", private))
	if err := os.WriteFile(manifest, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(dir, p.Artifact())
	if err := os.WriteFile(artifact, raw, 0600); err != nil {
		t.Fatal(err)
	}
	verify := func(version string, sequence uint64) error {
		return verifyRelease(dir, manifest, "windows", "amd64", version, sequence, keys)
	}
	if err := verify("test-10", 10); err != nil {
		t.Fatal(err)
	}
	if verify("test-11", 10) == nil || verify("test-10", 11) == nil {
		t.Fatal("accepted a different release than requested")
	}
	if err := os.WriteFile(artifact, []byte("changed runtime fixture!"), 0600); err != nil {
		t.Fatal(err)
	}
	if verify("test-10", 10) == nil {
		t.Fatal("accepted changed artifact bytes")
	}
	if err := os.WriteFile(manifest, append(encoded, []byte(" {}")...), 0600); err != nil {
		t.Fatal(err)
	}
	if verify("test-10", 10) == nil {
		t.Fatal("accepted trailing manifest data")
	}
}
