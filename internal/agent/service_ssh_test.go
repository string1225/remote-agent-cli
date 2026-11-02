package agent

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func serviceSSHFixture(t *testing.T) ServiceSSHOptions {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(key, "test fixture")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	o := ServiceSSHOptions{Alias: "deploy-fixture", Host: "127.0.0.1", User: "fixture", Port: 22, Identity: filepath.Join(dir, "key"), KnownHosts: filepath.Join(dir, "known_hosts")}
	if err = os.WriteFile(o.Identity, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(o.KnownHosts, []byte(knownhosts.Line([]string{o.Host}, signer.PublicKey())+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return o
}

func TestServiceSSHValidatesSelectedInputsAndPreservesOriginals(t *testing.T) {
	o := serviceSSHFixture(t)
	before, _ := os.ReadFile(o.Identity)
	if err := o.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*ServiceSSHOptions){
		func(o *ServiceSSHOptions) { o.Alias = "../escape" },
		func(o *ServiceSSHOptions) { o.Alias = "bad." },
		func(o *ServiceSSHOptions) { o.Host = "host\nProxyCommand command" },
		func(o *ServiceSSHOptions) { o.User = "-oProxyCommand=x" },
		func(o *ServiceSSHOptions) { o.Port = 0 },
		func(o *ServiceSSHOptions) { o.Identity = filepath.Dir(o.Identity) },
		func(o *ServiceSSHOptions) { o.Identity = "relative-key" },
		func(o *ServiceSSHOptions) { o.KnownHosts = filepath.Join(t.TempDir(), "absent") },
	} {
		c := o
		change(&c)
		if err := c.Validate(); err == nil {
			t.Error("invalid SSH option accepted", c.Alias, c.Host)
		}
	}
	after, _ := os.ReadFile(o.Identity)
	if string(before) != string(after) {
		t.Fatal("source key changed")
	}
	config := serviceSSHConfig(o, filepath.Join(t.TempDir(), "generation"))
	for _, want := range []string{"BatchMode yes", "IdentityAgent none", "IdentitiesOnly yes", "StrictHostKeyChecking yes", "GlobalKnownHostsFile none"} {
		if !strings.Contains(config, want) {
			t.Error("missing fixed SSH policy", want)
		}
	}
	if strings.Contains(config, string(before)) {
		t.Fatal("key embedded in configuration")
	}
}

func TestServiceSSHRejectsEncryptedAndMalformedKeys(t *testing.T) {
	o := serviceSSHFixture(t)
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	block, err := ssh.MarshalPrivateKeyWithPassphrase(key, "fixture", []byte("mock-passphrase"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(o.Identity, pem.EncodeToMemory(block), 0600)
	if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "口令") {
		t.Fatal("encrypted key accepted", err)
	}
	_ = os.WriteFile(o.Identity, []byte("not a private key"), 0600)
	if err := o.Validate(); err == nil {
		t.Fatal("invalid key accepted")
	}
}
