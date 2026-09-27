package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
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
func split(v string) []string {
	out := []string{}
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	db, err := store.Open(env("DATABASE_PATH", "data/remote-agent.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	s, handler, err := server.New(db, server.Config{PublicURL: env("PUBLIC_URL", "http://localhost:8080"), Downloads: env("DOWNLOADS_DIR", "dist"), STUN: split(env("STUN_URLS", "stun:stun.l.google.com:19302")), TURN: split(os.Getenv("TURN_URLS")), TURNSecret: os.Getenv("TURN_SECRET"), AllowSignup: env("ALLOW_SIGNUP", "true") == "true"}, web.Files)
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
