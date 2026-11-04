// Prepare runtime manifests for the configured HTTPS publication site.
package main

import (
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
	verify := flag.Bool("verify-only", false, "verify existing manifests and artifacts")
	version := flag.String("version", "", "business version")
	sequence := flag.Uint64("sequence", 0, "monotonic release sequence")
	dir := flag.String("dist", "dist", "artifacts directory")
	platforms := flag.String("platforms", "windows,darwin,linux", "comma-separated release platforms")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("unexpected release arguments")
	}
	if err := prepare(*dir, strings.Split(*platforms, ","), *version, *sequence, *verify); err != nil {
		log.Fatal(err)
	}
}

func prepare(dir string, platforms []string, version string, sequence uint64, verify bool) error {
	type manifest struct {
		path string
		data []byte
	}
	var manifests []manifest
	for _, platform := range platforms {
		if platform != "windows" && platform != "darwin" && platform != "linux" {
			return fmt.Errorf("unsupported release platform")
		}
		for _, arch := range []string{"amd64", "arm64"} {
			path := filepath.Join(dir, "updates", "v1", platform, arch+".json")
			if verify {
				if err := verifyRelease(dir, path, platform, arch, version, sequence); err != nil {
					return err
				}
				continue
			}
			p := clientupdate.Payload{SchemaVersion: 1, Sequence: sequence, Version: version, OS: platform, Arch: arch, MinHostVersion: 1, WorkerProtocol: clientupdate.WorkerProtocol, DataVersion: clientupdate.DataVersion, MinRollbackDataVersion: clientupdate.DataVersion}
			if platform == "windows" {
				p.MinHostVersion = clientupdate.HostVersion
			}
			raw, err := os.ReadFile(filepath.Join(dir, p.Artifact()))
			if err != nil {
				return err
			}
			sum := sha256.Sum256(raw)
			p.Size, p.SHA256 = int64(len(raw)), hex.EncodeToString(sum[:])
			release := clientupdate.Release{Payload: p}
			if err := release.Verify(platform, arch); err != nil {
				return err
			}
			encoded, _ := json.MarshalIndent(release, "", "  ")
			manifests = append(manifests, manifest{path, append(encoded, '\n')})
		}
	}
	// Validate every platform before changing any manifest.
	for _, m := range manifests {
		if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(m.path, m.data, 0644); err != nil {
			return err
		}
	}
	fmt.Println("Runtime manifests and artifact checksums verified; publication trusts the configured HTTPS site.")
	return nil
}

func verifyRelease(dir, manifest, platform, arch, version string, sequence uint64) error {
	f, err := os.Open(manifest)
	if err != nil {
		return fmt.Errorf("runtime manifest unavailable for %s/%s", platform, arch)
	}
	defer f.Close()
	var release clientupdate.Release
	d := json.NewDecoder(io.LimitReader(f, 16385))
	d.DisallowUnknownFields()
	if d.Decode(&release) != nil || d.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("invalid runtime manifest for %s/%s", platform, arch)
	}
	if err := release.Verify(platform, arch); err != nil {
		return err
	}
	if err := clientupdate.Compatible(release); err != nil {
		return err
	}
	if release.KeyID != "" || release.Signature != "" {
		return fmt.Errorf("new runtime manifests must omit legacy signature fields")
	}
	if version != "" && version != release.Version || sequence != 0 && sequence != release.Sequence {
		return fmt.Errorf("runtime release does not match the requested version and sequence")
	}
	artifact, err := os.Open(filepath.Join(dir, release.Artifact()))
	if err != nil {
		return err
	}
	defer artifact.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(artifact, release.Size+1))
	if err != nil || n != release.Size || hex.EncodeToString(h.Sum(nil)) != release.SHA256 {
		return fmt.Errorf("runtime size or checksum mismatch for %s/%s", platform, arch)
	}
	return nil
}
