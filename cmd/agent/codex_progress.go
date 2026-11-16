package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

type codexInstallProgress struct {
	mu      sync.Mutex
	out     io.Writer
	current codexsetup.Progress
	started time.Time
}

func installCodex(ctx context.Context, releaseURL, bin, goos, arch string, client *http.Client, output io.Writer) (string, error) {
	reporter := &codexInstallProgress{out: output}
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				reporter.tick()
			case <-stop:
				return
			}
		}
	}()
	defer func() { close(stop); <-done }()
	return codexsetup.InstallWithProgress(ctx, releaseURL, bin, goos, arch, client, reporter.update)
}

func codexInstallStep(stage string) string {
	switch stage {
	case "release":
		return tr("[1/6] Fetching official Codex release information...")
	case "download":
		return tr("[2/6] Downloading official Codex CLI...")
	case "verify":
		return tr("[3/6] Verifying SHA-256 checksum...")
	case "extract":
		return tr("[4/6] Extracting and checking the complete Codex package...")
	case "publish":
		return tr("[5/6] Installing Codex and companion files (preserving the previous version on failure)...")
	case "complete":
		return tr("[6/6] Codex CLI installation complete.")
	}
	return ""
}

func (p *codexInstallProgress) update(progress codexsetup.Progress) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if progress.Stage != p.current.Stage {
		if p.current.Stage == "download" {
			p.showDownload()
		}
		p.started = time.Now()
		fmt.Fprintln(p.out, codexInstallStep(progress.Stage))
	}
	p.current = progress
}

func (p *codexInstallProgress) tick() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current.Stage == "" || p.current.Stage == "complete" {
		return
	}
	if p.current.Stage == "download" {
		p.showDownload()
		return
	}
	fmt.Fprintf(p.out, tr("  %s (elapsed %s)\n"), codexInstallStep(p.current.Stage), time.Since(p.started).Round(time.Second))
}

func (p *codexInstallProgress) showDownload() {
	const mib = 1024 * 1024
	elapsed := time.Since(p.started).Round(time.Second)
	if p.current.Total > 0 {
		fmt.Fprintf(p.out, tr("  Downloaded %.1f / %.1f MiB (%.0f%%, elapsed %s)\n"), float64(p.current.Downloaded)/mib, float64(p.current.Total)/mib, 100*float64(p.current.Downloaded)/float64(p.current.Total), elapsed)
	} else {
		fmt.Fprintf(p.out, tr("  Downloaded %.1f MiB (total size unknown, elapsed %s)\n"), float64(p.current.Downloaded)/mib, elapsed)
	}
}
