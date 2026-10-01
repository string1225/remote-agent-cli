// Build distributable agents and their SHA-256 manifests without external tooling.
package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
)

func main() {
	version := flag.String("version", buildinfo.Current(), "release revision (set explicitly when building from a source archive)")
	flag.Parse()
	if !regexp.MustCompile(`^[A-Za-z0-9._-]{1,80}$`).MatchString(*version) {
		log.Fatal("invalid release version")
	}
	if err := os.MkdirAll("dist", 0755); err != nil {
		log.Fatal(err)
	}
	for _, platform := range []string{"windows", "darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			name := "remote-agent-" + platform + "-" + arch
			if platform == "windows" {
				name += ".exe"
			}
			path := filepath.Join("dist", name)
			cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version="+*version, "-o", path, "./cmd/agent")
			cmd.Env = append(os.Environ(), "GOOS="+platform, "GOARCH="+arch, "CGO_ENABLED=0")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			digest := sha256.Sum256(data)
			if err := os.WriteFile(path+".sha256", []byte(fmt.Sprintf("%x  %s\n", digest, name)), 0644); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Built", name)
		}
	}
}
