package agent

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/sys/windows"
)

func applyFixtureSSHACL(t *testing.T, path, sid string) {
	t.Helper()
	sddl := strings.TrimPrefix(serviceSSHFileSDDL(sid), "O:BAG:BA")
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err = windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSSHConditionalACLDoesNotGrantOtherServiceTokens(t *testing.T) {
	if windows.GetCurrentProcessToken().IsElevated() {
		t.Skip("administrators intentionally retain key recovery access")
	}
	o := serviceSSHFixture(t)
	applyFixtureSSHACL(t, o.Identity, "S-1-5-80-111-222-333-444-555")
	if _, err := os.ReadFile(o.Identity); !os.IsPermission(err) {
		t.Fatal("unrelated token can read protected key", err)
	}
	applyFixtureSSHACL(t, o.Identity, "S-1-1-0") // Everyone is enabled in this isolated fixture.
	if _, err := os.ReadFile(o.Identity); err != nil {
		t.Fatal("matching token cannot read key", err)
	}
}

func TestNativeOpenSSHAcceptsConditionalReadACL(t *testing.T) {
	exe := filepath.Join(os.Getenv("SystemRoot"), "System32", "OpenSSH", "ssh.exe")
	if _, err := os.Stat(exe); err != nil {
		t.Skip("Windows OpenSSH client is not installed")
	}
	o := serviceSSHFixture(t)
	identity, _ := os.ReadFile(o.Identity)
	clientSigner, err := ssh.ParsePrivateKey(identity)
	if err != nil {
		t.Fatal(err)
	}
	_, hostPrivate, _ := ed25519.GenerateKey(rand.Reader)
	hostSigner, _ := ssh.NewSignerFromKey(hostPrivate)
	serverConfig := &ssh.ServerConfig{PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if !bytes.Equal(key.Marshal(), clientSigner.PublicKey().Marshal()) {
			return nil, fmt.Errorf("unexpected key")
		}
		return &ssh.Permissions{}, nil
	}}
	serverConfig.AddHostKey(hostSigner)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	o.Port, _ = strconv.Atoi(port)
	_ = os.WriteFile(o.KnownHosts, []byte(knownhosts.Line([]string{listener.Addr().String()}, hostSigner.PublicKey())+"\n"), 0600)
	_ = os.WriteFile(filepath.Join(filepath.Dir(o.Identity), "config"), []byte(serviceSSHConfig(o, filepath.Dir(o.Identity))), 0600)
	applyFixtureSSHACL(t, o.Identity, "S-1-1-0")
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		server, channels, requests, err := ssh.NewServerConn(conn, serverConfig)
		if err != nil {
			done <- err
			return
		}
		defer server.Close()
		go ssh.DiscardRequests(requests)
		for incoming := range channels {
			if incoming.ChannelType() != "session" {
				_ = incoming.Reject(ssh.UnknownChannelType, "session only")
				continue
			}
			channel, requests, err := incoming.Accept()
			if err != nil {
				done <- err
				return
			}
			for req := range requests {
				if req.Type == "exec" {
					var command struct{ Command string }
					_ = ssh.Unmarshal(req.Payload, &command)
					if command.Command != "probe" {
						done <- fmt.Errorf("unexpected command")
						return
					}
					_ = req.Reply(true, nil)
					_, _ = channel.Write([]byte("ssh-fixture-ok\n"))
					_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
					_ = channel.Close()
					done <- nil
					return
				}
				_ = req.Reply(false, nil)
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-F", filepath.Join(filepath.Dir(o.Identity), "config"), o.Alias, "probe")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	if err != nil || strings.TrimSpace(stdout.String()) != "ssh-fixture-ok" {
		t.Fatalf("native SSH rejected protected fixture key: %v %s %s", err, stdout.String(), stderr.String())
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestSSHSetupScriptPreservesSourcesAndPublishesConfigLast(t *testing.T) {
	o := serviceSSHFixture(t)
	script := serviceSSHSetupScript(o, `C:\ProgramData\RemoteAgent\ssh\fixture`, `C:\ProgramData\RemoteAgent\ssh\fixture\generations\new`, "S-1-5-80-111-222-333-444-555", strings.Repeat("a", 64), strings.Repeat("b", 64))
	for _, want := range []string{"Member_of", "ReparsePoint", "FileShare]::None", "File]::Replace", "previous-config"} {
		if !strings.Contains(script, want) {
			t.Error("missing setup protection", want)
		}
	}
	if strings.Contains(script, "Remove-Item") || strings.Contains(script, "StrictHostKeyChecking no") {
		t.Fatal("setup deletes source data or disables host identity verification")
	}
	// Parse with the real PowerShell parser, without executing or elevating.
	path := filepath.Join(t.TempDir(), "setup.ps1")
	_ = os.WriteFile(path, []byte(script), 0600)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	parse := "$e=$null;$null=[System.Management.Automation.Language.Parser]::ParseFile(" + quotePowerShell(path) + ",[ref]$null,[ref]$e);if($e.Count){$e|Out-String|Write-Error;exit 1}"
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(parse))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("invalid PowerShell: %v %s", err, out)
	}
}

func TestSSHSetupScriptCopiesAndRollsBackInTemporaryDirectories(t *testing.T) {
	o := serviceSSHFixture(t)
	key, _ := os.ReadFile(o.Identity)
	hosts, _ := os.ReadFile(o.KnownHosts)
	keyHash, hostsHash := sha256.Sum256(key), sha256.Sum256(hosts)
	base := filepath.Join(t.TempDir(), "SSH space's", "fixture")
	run := func(generation, checksum string) ([]byte, error) {
		script := serviceSSHSetupScript(o, base, generation, "S-1-1-0", checksum, hex.EncodeToString(hostsHash[:]))
		// Production runs elevated with administrator ownership/full control.
		// This fixture changes only its temporary ACLs so no elevation is used.
		script = strings.ReplaceAll(script, "O:BAG:BA", "")
		script = strings.ReplaceAll(script, ";;FRFX;;;", ";;FA;;;")
		script = strings.ReplaceAll(script, ";;FR;;;", ";;FA;;;")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script)).CombinedOutput()
	}
	first := filepath.Join(base, "generations", "first")
	if out, err := run(first, hex.EncodeToString(keyHash[:])); err != nil {
		t.Fatalf("first import: %v %s", err, out)
	}
	oldConfig, _ := os.ReadFile(filepath.Join(base, "config"))
	if string(oldConfig) != serviceSSHConfig(o, first) {
		t.Fatal("first config not published")
	}
	second := filepath.Join(base, "generations", "second")
	if out, err := run(second, hex.EncodeToString(keyHash[:])); err != nil {
		t.Fatalf("replacement import: %v %s", err, out)
	}
	backup, _ := os.ReadFile(filepath.Join(second, "previous-config"))
	if !bytes.Equal(backup, oldConfig) {
		t.Fatal("previous connection not retained")
	}
	before, _ := os.ReadFile(filepath.Join(base, "config"))
	third := filepath.Join(base, "generations", "third")
	if _, err := run(third, strings.Repeat("0", 64)); err == nil {
		t.Fatal("changed source accepted")
	}
	after, _ := os.ReadFile(filepath.Join(base, "config"))
	if !bytes.Equal(before, after) {
		t.Fatal("failed import changed active connection")
	}
	afterKey, _ := os.ReadFile(o.Identity)
	afterHosts, _ := os.ReadFile(o.KnownHosts)
	if !bytes.Equal(key, afterKey) || !bytes.Equal(hosts, afterHosts) {
		t.Fatal("original SSH files changed")
	}
}
