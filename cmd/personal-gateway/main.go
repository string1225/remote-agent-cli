package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/personalgateway"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("use serve, token-create, token-revoke or token-list")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	registry := flags.String("registry", "", "Path to the private device registry")
	name := flags.String("name", "", "Device name")
	output := flags.String("output", "", "Create this private token file; never prints the token")
	listen := flags.String("listen", "127.0.0.1:8088", "Loopback HTTP listener")
	codex := flags.String("codex", "/usr/local/bin/codex", "Native Codex executable")
	proxy := flags.String("proxy", "http://127.0.0.1:7890", "Trusted outbound proxy")
	concurrency := flags.Int("concurrency", 2, "Concurrent upstream requests")
	accountsRoot := flags.String("accounts-root", "", "Private account registry directory (enables account control)")
	controlToken := flags.String("control-token-file", "", "Private dedicated control credential file")
	legacyHome := flags.String("legacy-codex-home", "", "Existing Codex credential directory; retained on migration")
	adminConfig := flags.String("admin-config", "", "Private admin password verifier and session key")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *registry == "" && args[0] != "admin-create" {
		return fmt.Errorf("--registry is required")
	}
	switch args[0] {
	case "admin-create":
		if *adminConfig == "" || *output == "" {
			return fmt.Errorf("--admin-config and --output are required")
		}
		if err := os.MkdirAll(filepath.Dir(*output), 0700); err != nil {
			return err
		}
		file, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		defer file.Close()
		password, err := personalgateway.CreateAdmin(*adminConfig)
		if err != nil {
			os.Remove(*output)
			return err
		}
		if _, err := fmt.Fprintln(file, password); err != nil {
			return err
		}
		if err := file.Sync(); err != nil {
			return err
		}
		fmt.Printf("Admin password saved to %s\n", *output)
	case "token-create":
		if *output == "" {
			return fmt.Errorf("--output is required; tokens are written to a private file")
		}
		if err := os.MkdirAll(filepath.Dir(*output), 0700); err != nil {
			return err
		}
		file, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		ok := false
		defer func() {
			file.Close()
			if !ok {
				os.Remove(*output)
			}
		}()
		token, err := personalgateway.AddDevice(*registry, *name)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(file, token); err != nil {
			return err
		}
		if err := file.Sync(); err != nil {
			return err
		}
		ok = true
		fmt.Printf("Device %q created; secret saved to %s\n", *name, *output)
	case "token-revoke":
		if err := personalgateway.RevokeDevice(*registry, *name); err != nil {
			return err
		}
		fmt.Printf("Device %q revoked\n", *name)
	case "token-list":
		data, err := personalgateway.LoadRegistry(*registry)
		if err != nil {
			return err
		}
		for _, device := range data.Devices {
			fmt.Printf("%s\tenabled=%t\n", device.Name, device.Enabled)
		}
	case "serve":
		host, _, err := net.SplitHostPort(*listen)
		if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
			return fmt.Errorf("listener must bind to loopback")
		}
		handler, err := personalgateway.NewServer(*registry, &personalgateway.CodexSource{Binary: *codex, Proxy: *proxy}, *proxy, *concurrency)
		if err != nil {
			return err
		}
		if *accountsRoot != "" {
			info, err := os.Stat(*controlToken)
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
				return fmt.Errorf("control token must be a private regular file")
			}
			raw, err := os.ReadFile(*controlToken)
			if err != nil || len(raw) > 16384 {
				return fmt.Errorf("cannot read control credential")
			}
			handler.Accounts, err = personalgateway.NewAccountManager(*accountsRoot, *codex, *proxy, *legacyHome, strings.TrimSpace(string(raw)))
			if err != nil {
				return err
			}
			defer handler.Accounts.Close()
			handler.Credentials = handler.Accounts
		}
		if *adminConfig != "" {
			handler.Admin, err = personalgateway.LoadAdmin(*adminConfig, "https://codex.sunny-string.cn")
			if err != nil {
				return err
			}
		}
		server := &http.Server{Addr: *listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			server.Shutdown(shutdown)
		}()
		log.Printf("Personal model gateway listening on %s", *listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
	default:
		return fmt.Errorf("unknown command")
	}
	return nil
}
