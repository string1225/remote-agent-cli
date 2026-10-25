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
			if platform == "windows" {
				core := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version="+*version, "-o", filepath.Join("dist", "remote-agent-core-windows-"+arch+".exe"), "./cmd/agent")
				core.Env = append(os.Environ(), "GOOS=windows", "GOARCH="+arch, "CGO_ENABLED=0")
				core.Stdout = os.Stdout
				core.Stderr = os.Stderr
				if err := core.Run(); err != nil {
					log.Fatal(err)
				}
				if err := checksum(filepath.Join("dist", "remote-agent-core-windows-"+arch+".exe")); err != nil {
					log.Fatal(err)
				}
			}
			packagePath := "./cmd/agent"
			if platform == "windows" {
				packagePath = "./cmd/agent-host"
			}
			cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version="+*version, "-o", path, packagePath)
			cmd.Env = append(os.Environ(), "GOOS="+platform, "GOARCH="+arch, "CGO_ENABLED=0")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
			if err := checksum(path); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Built", name)
		}
	}
}

func checksum(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	return os.WriteFile(path+".sha256", []byte(fmt.Sprintf("%x  %s\n", digest, filepath.Base(path))), 0644)
}
