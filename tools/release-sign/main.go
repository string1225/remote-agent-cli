// Sign release manifests offline; private keys never enter dist or server data.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
)

func main() {
	keygen := flag.String("keygen", "", "create a private signing seed at this new path; output only its public key")
	keyfile := flag.String("key", "", "offline private seed file")
	keydir := flag.String("key-dir", "", "offline directory containing the existing trusted signing seed")
	verify := flag.Bool("verify-only", false, "verify existing manifests and artifacts without reading a private key")
	id := flag.String("key-id", "release-2026", "embedded trust anchor ID")
	version := flag.String("version", "", "business version")
	sequence := flag.Uint64("sequence", 0, "monotonic release sequence")
	dir := flag.String("dist", "dist", "artifacts directory")
	platforms := flag.String("platforms", "windows,darwin,linux", "comma-separated release platforms")
	flag.Parse()
	if *keygen != "" {
		pub, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.OpenFile(*keygen, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.WriteString(hex.EncodeToString(key.Seed()))
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err != nil {
			log.Fatal(err)
		}
		if closeErr != nil {
			log.Fatal(closeErr)
		}
		fmt.Println(hex.EncodeToString(pub))
		return
	}
	var key ed25519.PrivateKey
	if !*verify {
		var err error
		key, err = trustedSigningKey(*keyfile, *keydir, *id, clientupdate.TrustKeys())
		if err != nil {
			log.Fatal(err)
		}
	}
	// Prepare and validate every platform before writing any new manifests.
	type manifest struct {
		path string
		data []byte
	}
	var manifests []manifest
	for _, platform := range strings.Split(*platforms, ",") {
		if platform != "windows" && platform != "darwin" && platform != "linux" {
			log.Fatal("unsupported release platform")
		}
		for _, arch := range []string{"amd64", "arm64"} {
			out := filepath.Join(*dir, "updates", "v1", platform, arch+".json")
			if *verify {
				if err := verifyRelease(*dir, out, platform, arch, *version, *sequence, clientupdate.TrustKeys()); err != nil {
					log.Fatal(err)
				}
				fmt.Println("Verified", platform, arch)
				continue
			}
			p := clientupdate.Payload{SchemaVersion: 1, Sequence: *sequence, Version: *version, OS: platform, Arch: arch, MinHostVersion: 1, WorkerProtocol: clientupdate.WorkerProtocol, DataVersion: clientupdate.DataVersion, MinRollbackDataVersion: clientupdate.DataVersion}
			if platform == "windows" {
				p.MinHostVersion = clientupdate.HostVersion
			}
			path := filepath.Join(*dir, p.Artifact())
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			sum := sha256.Sum256(data)
			p.Size, p.SHA256 = int64(len(data)), hex.EncodeToString(sum[:])
			release := clientupdate.Sign(p, *id, key)
			if err = release.Verify(clientupdate.TrustKeys(), platform, arch); err != nil {
				log.Fatal(err)
			}
			encoded, _ := json.MarshalIndent(release, "", "  ")
			manifests = append(manifests, manifest{out, append(encoded, '\n')})
		}
	}
	for _, m := range manifests {
		if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(m.path, m.data, 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Signed", filepath.Base(filepath.Dir(m.path)), filepath.Base(m.path), *version, *sequence)
	}
}

func trustedSigningKey(file, dir, id string, keys map[string]ed25519.PublicKey) (ed25519.PrivateKey, error) {
	expected, ok := keys[id]
	if !ok {
		return nil, fmt.Errorf("unknown signing key ID")
	}
	if file != "" && dir != "" || file == "" && dir == "" {
		return nil, fmt.Errorf("specify either -key or -key-dir")
	}
	paths := []string{file}
	if dir != "" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("signing directory unavailable; sign locally as its original desktop user")
		}
		paths = nil
		for _, entry := range entries {
			if entry.Type().IsRegular() {
				paths = append(paths, filepath.Join(dir, entry.Name()))
			}
		}
	}
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 128 {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		seed, err := hex.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil || len(seed) != ed25519.SeedSize {
			continue
		}
		key := ed25519.NewKeyFromSeed(seed)
		if key.Public().(ed25519.PublicKey).Equal(expected) {
			return key, nil
		}
	}
	return nil, fmt.Errorf("no readable private key matches the installed release trust anchor; existing files were retained")
}

func verifyRelease(dir, manifest, platform, arch, version string, sequence uint64, keys map[string]ed25519.PublicKey) error {
	f, err := os.Open(manifest)
	if err != nil {
		return fmt.Errorf("signed manifest unavailable for %s/%s", platform, arch)
	}
	defer f.Close()
	var release clientupdate.Release
	d := json.NewDecoder(io.LimitReader(f, 16385))
	d.DisallowUnknownFields()
	if d.Decode(&release) != nil || d.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("invalid signed manifest for %s/%s", platform, arch)
	}
	if err := release.Verify(keys, platform, arch); err != nil {
		return err
	}
	if err := clientupdate.Compatible(release); err != nil {
		return err
	}
	if version != "" && version != release.Version || sequence != 0 && sequence != release.Sequence {
		return fmt.Errorf("signed release does not match the requested version and sequence")
	}
	artifact, err := os.Open(filepath.Join(dir, release.Artifact()))
	if err != nil {
		return fmt.Errorf("signed artifact unavailable for %s/%s", platform, arch)
	}
	defer artifact.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(artifact, release.Size+1))
	if err != nil || n != release.Size || hex.EncodeToString(h.Sum(nil)) != release.SHA256 {
		return fmt.Errorf("signed artifact size or checksum mismatch for %s/%s", platform, arch)
	}
	return nil
}
