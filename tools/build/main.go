// Build distributable agents and their SHA-256 manifests without external tooling.
package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
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
			cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", path, "./cmd/agent")
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
