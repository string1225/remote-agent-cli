package access

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshFixture struct {
	host          string
	port          int
	home          string
	passwordCalls atomic.Int32
}

func fixtureSSH(t *testing.T) *sshFixture {
	t.Helper()
	shell, err := exec.LookPath("sh")
	if err != nil {
		shell = `C:\Program Files\Git\bin\sh.exe`
		if _, err = os.Stat(shell); err != nil {
			t.Skip("POSIX shell fixture unavailable")
		}
	}
	home := t.TempDir()
	if err = os.Mkdir(filepath.Join(home, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	original := []byte("# original authorization retained\n")
	if err = os.WriteFile(filepath.Join(home, ".ssh", "authorized_keys"), original, 0600); err != nil {
		t.Fatal(err)
	}
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &sshFixture{home: home}
	config := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			fixture.passwordCalls.Add(1)
			if string(p) != "mock-password-never-persist" {
				return nil, errors.New("invalid fixture password")
			}
			return nil, nil
		},
		PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			raw, _ := os.ReadFile(filepath.Join(home, ".ssh", "authorized_keys"))
			if !bytes.Contains(raw, bytes.TrimSpace(ssh.MarshalAuthorizedKey(key))) {
				return nil, errors.New("fixture key not installed")
			}
			return nil, nil
		},
	}
	config.AddHostKey(signer)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	fixture.host, _, _ = net.SplitHostPort(listener.Addr().String())
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	fixture.port, _ = strconv.Atoi(port)
	go func() {
		for {
			conn, e := listener.Accept()
			if e != nil {
				return
			}
			go func() {
				defer conn.Close()
				_, channels, requests, e := ssh.NewServerConn(conn, config)
				if e != nil {
					return
				}
				go ssh.DiscardRequests(requests)
				for channel := range channels {
					if channel.ChannelType() != "session" {
						_ = channel.Reject(ssh.UnknownChannelType, "session only")
						continue
					}
					ch, reqs, e := channel.Accept()
					if e != nil {
						continue
					}
					go func() {
						defer ch.Close()
						for req := range reqs {
							if req.Type != "exec" {
								_ = req.Reply(false, nil)
								continue
							}
							var command struct{ Command string }
							if ssh.Unmarshal(req.Payload, &command) != nil {
								_ = req.Reply(false, nil)
								return
							}
							_ = req.Reply(true, nil)
							status := uint32(0)
							if command.Command == "sh -s" {
								cmd := exec.Command(shell, "-s")
								cmd.Env = append(os.Environ(), "HOME="+filepath.ToSlash(home))
								cmd.Stdin = ch
								cmd.Stdout = io.Discard
								cmd.Stderr = io.Discard
								if cmd.Run() != nil {
									status = 1
								}
							} else if command.Command != "true" {
								status = 1
							}
							_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
							return
						}
					}()
				}
			}()
		}
	}()
	return fixture
}
func TestSSHBootstrapPreservesAuthorizationAndWorksWithNativeOpenSSH(t *testing.T) {
	fixture := fixtureSSH(t)
	m, _ := testManager(t)
	request := Request{Name: "fixture", Host: fixture.host, Port: fixture.port, User: "fixture", Password: "mock-password-never-persist"}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	server, err := m.AddServer(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if server.Fingerprint == "" || server.HostKey == "" {
		t.Fatal("host key was not pinned")
	}
	original := filepath.Join(fixture.home, ".ssh", "authorized_keys")
	raw, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("# original authorization retained\n")) {
		t.Fatal("original authorization was lost")
	}
	backups, _ := filepath.Glob(original + ".remote-agent-backup-*")
	if len(backups) != 1 {
		t.Fatal("original authorization was not backed up")
	}
	calls := fixture.passwordCalls.Load()
	request.Password = "not-the-password"
	if _, err = m.AddServer(ctx, request); err != nil {
		t.Fatal("idempotent retry failed", err)
	}
	if fixture.passwordCalls.Load() != calls {
		t.Fatal("existing connection unexpectedly used password")
	}
	repeated, _ := os.ReadFile(original)
	if !bytes.Equal(raw, repeated) {
		t.Fatal("retry changed authorization")
	}
	err = filepath.WalkDir(m.root, func(path string, item os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !item.IsDir() {
			data, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			if bytes.Contains(data, []byte("mock-password-never-persist")) {
				return errors.New("password persisted")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := m.ListServers()
	if err != nil || len(listed) != 1 {
		t.Fatal("missing server", err)
	}
	serialized, _ := json.Marshal(listed)
	if bytes.Contains(serialized, []byte("PRIVATE KEY")) {
		t.Fatal("private key returned")
	}
	native, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("native SSH not installed")
	}
	cmd := exec.CommandContext(ctx, native, "-F", server.ConfigPath, "-o", "ConnectTimeout=5", server.Name, "true")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native SSH rejected managed key/config: %v: %s", err, out)
	}
	// Simulate a lost local publication after the remote accepted the key.
	if err = os.Remove(filepath.Join(filepath.Dir(server.ConfigPath), "server.json")); err != nil {
		t.Fatal(err)
	}
	request.Password = ""
	if _, err = m.AddServer(ctx, request); err != nil {
		t.Fatal("pending key did not recover", err)
	}
	recovered, _ := os.ReadFile(original)
	if !bytes.Equal(raw, recovered) {
		t.Fatal("recovery duplicated or changed remote authorization")
	}
	if ok, err := m.RemoveServer("fixture"); err != nil || !ok {
		t.Fatal(err)
	}
	listed, err = m.ListServers()
	if err != nil || len(listed) != 0 {
		t.Fatal("removed server still listed")
	}
	after, _ := os.ReadFile(original)
	if !bytes.Equal(raw, after) {
		t.Fatal("local removal changed remote keys")
	}
}
func TestSSHInvalidInputAndFailedAuthenticationDoNotPublishServer(t *testing.T) {
	m, _ := testManager(t)
	for _, r := range []Request{{Name: "../escape", Host: "127.0.0.1", User: "root"}, {Name: "ok", Host: "bad;host", User: "root"}, {Name: "ok", Host: "127.0.0.1", User: "root;command"}} {
		if _, err := m.AddServer(context.Background(), r); err == nil {
			t.Fatal("invalid endpoint accepted")
		}
	}
	fixture := fixtureSSH(t)
	_, err := m.AddServer(context.Background(), Request{Name: "failed", Host: fixture.host, Port: fixture.port, User: "fixture", Password: "wrong-password"})
	if err == nil || strings.Contains(err.Error(), "wrong-password") {
		t.Fatal("failed auth accepted or password echoed")
	}
	servers, err := m.ListServers()
	if err != nil || len(servers) != 0 {
		t.Fatal("failed connection was published")
	}
}
