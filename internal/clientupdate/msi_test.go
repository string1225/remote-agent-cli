package clientupdate

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestMSIAdoptsSeedPreservesFloorAndRollsBack(t *testing.T) {
	s, makeRelease := fixture(t)
	release, body := makeRelease(10, "old-runtime")
	if e := s.Stage(release, func(w io.Writer) error { _, e := w.Write(body); return e }); e != nil {
		t.Fatal(e)
	}
	original := State{SchemaVersion: 1, Current: &release, HighestSequence: 10, HighestSHA256: release.SHA256}
	if e := s.Save(original); e != nil {
		t.Fatal(e)
	}
	if e := s.AdoptMSISeed("new-protected-seed"); e != nil {
		t.Fatal(e)
	}
	state, e := s.Load()
	if e != nil || state.Current != nil || state.Previous == nil || state.HighestSequence != 10 {
		t.Fatal(state, e)
	}
	if path, e := s.Binary(state.Current, "protected-seed.exe"); e != nil || path != "protected-seed.exe" {
		t.Fatal(path, e)
	}
	// A stale online channel must not reactivate the former core.
	engine := Engine{Store: s, Service: &fakeService{store: s}}
	if e = engine.Apply(context.Background(), release, func(io.Writer) error { t.Fatal("stale update downloaded"); return nil }, false); e == nil {
		t.Fatal("old online release replaced MSI")
	}
	if e = s.FinishMSISeed(true); e != nil {
		t.Fatal(e)
	}
	state, e = s.Load()
	if e != nil || state.Current == nil || state.Current.SHA256 != release.SHA256 || state.MSISeedDigest != "" {
		t.Fatal("rollback lost former pointer", e)
	}
	if e = s.AdoptMSISeed("new-protected-seed"); e != nil {
		t.Fatal(e)
	}
	if e = s.FinishMSISeed(false); e != nil {
		t.Fatal(e)
	}
	if e = s.AdoptMSISeed("new-protected-seed"); e != nil {
		t.Fatal("repair changed selection", e)
	}
	if _, e = os.Stat(filepath.Join(s.Path, "msi-rollback.json")); !os.IsNotExist(e) {
		t.Fatal("committed backup remains", e)
	}
}
