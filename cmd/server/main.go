package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/server"
	"github.com/string1225/remote-agent-cli/internal/store"
	"github.com/string1225/remote-agent-cli/web"
)

func env(name, defaultValue string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return defaultValue
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	if len(os.Args) > 1 {
		if os.Args[1] != "admin" {
			return errors.New("unknown command; use remote-server admin create|grant or run without arguments")
		}
		return runAdmin(os.Args[2:], os.Stdin, os.Stdout)
	}
	cfg, err := configFromEnv()
	if err != nil {
		return err
	}
	db, err := store.Open(env("DATABASE_PATH", "data/remote-agent.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	s, handler, err := server.New(db, cfg, web.Files)
	if err != nil {
		return err
	}
	httpServer := &http.Server{Addr: env("LISTEN_ADDR", "127.0.0.1:8080"), Handler: handler, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 16384}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			_ = db.Prune()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	go func() {
		<-ctx.Done()
		s.Close()
		shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		_ = httpServer.Shutdown(shutdown)
	}()
	log.Printf("Remote Agent console: %s", env("PUBLIC_URL", "http://localhost:8080"))
	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func configFromEnv() (server.Config, error) {
	cfg := server.Config{PublicURL: env("PUBLIC_URL", "http://localhost:8080"), Downloads: env("DOWNLOADS_DIR", "dist"), GatewayURL: strings.TrimSpace(os.Getenv("CODEX_GATEWAY_URL")), GatewayModel: env("CODEX_MODEL", "gpt-6.1-sol")}
	if cfg.GatewayURL == "" {
		return cfg, nil
	}
	path := os.Getenv("CODEX_GATEWAY_TOKEN_FILE")
	if path == "" {
		return cfg, errors.New("CODEX_GATEWAY_TOKEN_FILE is required when CODEX_GATEWAY_URL is set")
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("cannot open CODEX_GATEWAY_TOKEN_FILE: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return cfg, fmt.Errorf("cannot inspect CODEX_GATEWAY_TOKEN_FILE: %w", err)
	}
	if !info.Mode().IsRegular() {
		return cfg, errors.New("CODEX_GATEWAY_TOKEN_FILE must be a regular file")
	}
	if runtime.GOOS != "windows" && !privateGatewayToken(path, info.Mode()) {
		return cfg, errors.New("CODEX_GATEWAY_TOKEN_FILE must be private (chmod 600)")
	}
	raw, err := io.ReadAll(io.LimitReader(f, 16385))
	if err != nil {
		return cfg, errors.New("cannot read CODEX_GATEWAY_TOKEN_FILE")
	}
	cfg.GatewayToken = strings.TrimSpace(string(raw))
	if len(raw) > 16384 || cfg.GatewayToken == "" || strings.ContainsAny(cfg.GatewayToken, "\r\n") {
		return cfg, errors.New("CODEX_GATEWAY_TOKEN_FILE must contain one non-empty service token")
	}
	return cfg, nil
}

func systemdCredentialLocation(file, dir string) bool {
	return filepath.IsAbs(dir) && filepath.Clean(dir) == dir && filepath.Dir(dir) == "/run/credentials" &&
		filepath.IsAbs(file) && filepath.Clean(file) == file && filepath.Dir(file) == dir
}

func privateGatewayToken(path string, mode os.FileMode) bool {
	if mode.Perm()&0077 == 0 {
		return true
	}
	// LoadCredential may expose 0440 files inside a read-only, per-unit
	// namespace. Keep this exception confined to systemd's canonical directory;
	// an ordinary group-readable file must still be rejected.
	if runtime.GOOS != "linux" || mode.Perm() != 0440 || !systemdCredentialLocation(path, os.Getenv("CREDENTIALS_DIRECTORY")) {
		return false
	}
	resolved, err := filepath.EvalSymlinks(path)
	return err == nil && resolved == path
}
