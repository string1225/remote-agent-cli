package clientupdate

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (Store, func(uint64, string) (Release, []byte)) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s := Store{Path: t.TempDir(), Keys: map[string]ed25519.PublicKey{"fixture": pub}, Arch: "amd64"}
	return s, func(seq uint64, version string) (Release, []byte) {
		b := []byte("fixture executable " + version)
		sum := sha256.Sum256(b)
		return Sign(Payload{SchemaVersion: 1, Sequence: seq, Version: version, OS: "windows", Arch: "amd64", MinHostVersion: 1, WorkerProtocol: 1, DataVersion: 1, MinRollbackDataVersion: 1, Size: int64(len(b)), SHA256: hex.EncodeToString(sum[:])}, "fixture", key), b
	}
}

func TestSignatureAndCredentialFreeDownload(t *testing.T) {
	s, makeRelease := fixture(t)
	release, body := makeRelease(1, "fixture-one")
	var corrupt, redirect bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("credentials leaked")
		}
		switch r.URL.Path {
		case "/agents/updates/v1/windows/amd64.json":
			if redirect {
				http.Redirect(w, r, "/elsewhere", 302)
				return
			}
			_ = json.NewEncoder(w).Encode(release)
		case "/agents/downloads/" + release.Artifact():
			if corrupt {
				_, _ = w.Write([]byte("damaged"))
			} else {
				_, _ = w.Write(body)
			}
		default:
			t.Error("unexpected download path", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	ctx := context.Background()
	base := server.URL + "/agents"
	r, err := Fetch(ctx, HTTPClient(), base, "windows", "amd64", s.Keys)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Stage(r, func(w io.Writer) error { return Download(ctx, HTTPClient(), base, r, w) }); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Binary(&r, ""); err != nil {
		t.Fatal(err)
	}
	release.Version = "tampered"
	if _, err = Fetch(ctx, HTTPClient(), base, "windows", "amd64", s.Keys); err == nil {
		t.Fatal("tampered manifest accepted")
	}
	release, _ = makeRelease(2, "fixture-two")
	corrupt = true
	if err = s.Stage(release, func(w io.Writer) error { return Download(ctx, HTTPClient(), base, release, w) }); err == nil {
		t.Fatal("corrupt download accepted")
	}
	redirect = true
	if _, err = Fetch(ctx, HTTPClient(), base, "windows", "amd64", s.Keys); err == nil {
		t.Fatal("redirect followed")
	}
	state, err := s.Load()
	if err != nil || state.Current != nil {
		t.Fatal("failed downloads changed current", err)
	}
}

type fakeService struct {
	store                                      Store
	running, busy, paused, failStart, fallback bool
	stops, starts, pauses                      int
	health                                     Health
}

func (f *fakeService) Running() (bool, error) { return f.running, nil }
func (f *fakeService) Pause(force bool) error {
	f.pauses++
	if f.busy && !force {
		return errors.New("busy")
	}
	f.paused = true
	f.busy = false
	return nil
}
func (f *fakeService) Resume() error { f.paused = false; return nil }
func (f *fakeService) Stop() error   { f.stops++; f.running = false; return nil }
func (f *fakeService) Start() error {
	f.starts++
	if f.failStart {
		f.failStart = false
		return errors.New("candidate crashed")
	}
	f.running = true
	s, err := f.store.Load()
	if err != nil {
		return err
	}
	current := s.Current
	if f.fallback {
		current = s.Previous
		f.fallback = false
	}
	f.health = Health{Digest: "seed", Ready: true}
	if current != nil {
		f.health.Digest = current.SHA256
	}
	return nil
}
func (f *fakeService) Health() (Health, error) { return f.health, nil }
func apply(e Engine, r Release, b []byte, force bool) error {
	return e.Apply(context.Background(), r, func(w io.Writer) error { _, err := w.Write(b); return err }, force)
}

func TestTransactionPreservesRegistrationAndRollsBack(t *testing.T) {
	s, m := fixture(t)
	f := &fakeService{store: s, running: true}
	var output bytes.Buffer
	e := Engine{Store: s, Service: f, Output: &output}
	r1, b1 := m(1, "one")
	r2, b2 := m(2, "two")
	r3, b3 := m(3, "three")
	if err := apply(e, r1, b1, false); err != nil {
		t.Fatal(err)
	}
	if err := apply(e, r2, b2, false); err != nil {
		t.Fatal(err)
	}
	f.failStart = true
	if err := apply(e, r3, b3, false); err == nil {
		t.Fatal("failed startup accepted")
	}
	state, _ := s.Load()
	if state.Current.Version != "two" || state.Previous.Version != "one" || state.HighestSequence != 3 || state.Pending || !f.running {
		t.Fatalf("unsafe rollback: %+v", state)
	}
	if err := apply(e, r1, b1, false); err == nil {
		t.Fatal("stale release replayed")
	}
	if err := e.Rollback(false); err != nil {
		t.Fatal(err)
	}
	state, _ = s.Load()
	if state.Current.Version != "one" || state.Previous.Version != "two" || state.HighestSequence != 3 {
		t.Fatal("explicit rollback lost sequence floor")
	}
	if !strings.Contains(output.String(), "MSI registration") {
		t.Fatal("missing result")
	}
}

func TestStoppedServiceStaysStoppedAndInterruptedCommitRecovers(t *testing.T) {
	s, m := fixture(t)
	f := &fakeService{store: s}
	e := Engine{Store: s, Service: f}
	r, b := m(1, "one")
	if err := apply(e, r, b, false); err != nil {
		t.Fatal(err)
	}
	state, _ := s.Load()
	if !state.Pending || f.starts != 0 || f.stops != 0 || f.pauses != 0 {
		t.Fatal("stopped service changed")
	}
	if err := apply(e, r, b, false); err == nil {
		t.Fatal("overwrote a pending stopped-service update")
	}
	if err := f.Start(); err != nil {
		t.Fatal(err)
	}
	if err := e.Reconcile(); err != nil {
		t.Fatal(err)
	}
	state, _ = s.Load()
	if state.Pending || state.Current.Version != "one" {
		t.Fatal("pending activation not committed")
	}
	// Simulate power loss after state commit but before successful activation.
	r2, b2 := m(2, "two")
	if err := s.Stage(r2, func(w io.Writer) error { _, err := w.Write(b2); return err }); err != nil {
		t.Fatal(err)
	}
	state.Previous = state.Current
	state.RollbackPrevious = nil
	state.RollbackHasPrevious = true
	state.Current = &r2
	state.HasPrevious = true
	state.Pending = true
	state.HighestSequence = 2
	state.HighestSHA256 = r2.SHA256
	if err := s.Save(state); err != nil {
		t.Fatal(err)
	}
	f.fallback = true
	if err := f.Start(); err != nil {
		t.Fatal(err)
	}
	if err := e.Reconcile(); err != nil {
		t.Fatal(err)
	}
	state, _ = s.Load()
	if state.Pending || state.Current.Version != "one" || state.HighestSequence != 2 {
		t.Fatal("interrupted candidate did not safely reconcile")
	}
}

func TestBusyAndCorruptUpdatesNeverStopActiveService(t *testing.T) {
	s, m := fixture(t)
	f := &fakeService{store: s, running: true, busy: true}
	e := Engine{Store: s, Service: f}
	r, b := m(1, "one")
	if err := apply(e, r, b, false); err == nil || f.stops != 0 {
		t.Fatal("busy task interrupted")
	}
	r2, _ := m(2, "two")
	if err := apply(e, r2, []byte("corrupt"), true); err == nil || f.stops != 0 {
		t.Fatal("corrupt update stopped service")
	}
	if err := apply(e, r, b, true); err != nil {
		t.Fatal(err)
	}
	if !f.running || f.stops != 1 {
		t.Fatal("forced idle transition failed")
	}
}

func TestRuntimeConfinementAndHostCompatibility(t *testing.T) {
	s, m := fixture(t)
	r, b := m(1, "one")
	if err := s.Stage(r, func(w io.Writer) error { _, err := w.Write(b); return err }); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Path, "versions", r.SHA256, "remote-agent-core.exe"), []byte("replaced"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Binary(&r, ""); err == nil {
		t.Fatal("changed installed runtime accepted")
	}
	if Compatible(Release{Payload: Payload{MinHostVersion: HostVersion + 1, WorkerProtocol: 1}}) == nil {
		t.Fatal("new host requirement ignored")
	}
	r.SHA256 = "../../outside"
	if _, err := s.Binary(&r, ""); err == nil {
		t.Fatal("runtime traversal accepted")
	}
	out := t.TempDir()
	keep := filepath.Join(out, "keep")
	_ = os.WriteFile(keep, []byte("retained"), 0600)
	name := fmt.Sprintf("%064x", 9)
	if err := os.Symlink(out, filepath.Join(s.Path, "versions", name)); err != nil {
		t.Log("symlink unavailable:", err)
		return
	}
	if err := s.Clean(State{Current: &r}); err == nil {
		t.Fatal("cleanup accepted a symlink")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatal("cleanup escaped runtime root")
	}
}
