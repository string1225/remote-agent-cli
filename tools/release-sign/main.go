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
	"log"
	"os"
	"path/filepath"

	"github.com/string1225/remote-agent-cli/internal/clientupdate"
)

func main() {
	keygen := flag.String("keygen", "", "create a private signing seed at this new path; output only its public key")
	keyfile := flag.String("key", "", "offline private seed file")
	id := flag.String("key-id", "release-2026", "embedded trust anchor ID")
	version := flag.String("version", "", "business version")
	sequence := flag.Uint64("sequence", 0, "monotonic release sequence")
	dir := flag.String("dist", "dist", "artifacts directory")
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
	seedHex, err := os.ReadFile(*keyfile)
	if err != nil {
		log.Fatal(err)
	}
	seed, err := hex.DecodeString(string(seedHex))
	if err != nil || len(seed) != ed25519.SeedSize {
		log.Fatal("invalid private signing seed")
	}
	key := ed25519.NewKeyFromSeed(seed)
	for _, arch := range []string{"amd64", "arm64"} {
		path := filepath.Join(*dir, "remote-agent-core-windows-"+arch+".exe")
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		sum := sha256.Sum256(data)
		p := clientupdate.Payload{SchemaVersion: 1, Sequence: *sequence, Version: *version, OS: "windows", Arch: arch, MinHostVersion: clientupdate.HostVersion, WorkerProtocol: clientupdate.WorkerProtocol, DataVersion: clientupdate.DataVersion, MinRollbackDataVersion: clientupdate.DataVersion, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:])}
		release := clientupdate.Sign(p, *id, key)
		if err = release.Verify(map[string]ed25519.PublicKey{*id: key.Public().(ed25519.PublicKey)}, "windows", arch); err != nil {
			log.Fatal(err)
		}
		encoded, _ := json.MarshalIndent(release, "", "  ")
		out := filepath.Join(*dir, "updates", "v1", "windows")
		if err = os.MkdirAll(out, 0755); err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, arch+".json"), append(encoded, '\n'), 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Signed", arch, *version, *sequence)
	}
}
