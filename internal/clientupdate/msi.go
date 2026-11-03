package clientupdate

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

// MSISequence is supplied by the installer build; it does not bypass release
// signature validation or replace the signed updater's anti-rollback floor.
var MSISequence = "0"

// AdoptMSISeed is called only by the elevated MSI custom action after installing
// the protected seed. Signed background releases keep their normal verification
// and sequence floor. MSI rollback restores the entire former runtime pointer.
func (s Store) AdoptMSISeed(digest string) error {
	state, e := s.Load()
	if e != nil {
		return e
	}
	if state.MSISeedDigest == digest {
		return nil
	}
	root, e := os.OpenRoot(s.Path)
	if e != nil {
		return e
	}
	defer root.Close()
	if _, e = root.Stat("msi-rollback.json"); e == nil {
		return errors.New("unfinished MSI runtime transaction; restore it before retrying")
	}
	data, e := json.Marshal(state)
	if e != nil {
		return e
	}
	if e = atomicWrite(root, "msi-rollback.json", data); e != nil {
		return e
	}
	if state.Current != nil {
		state.Previous = state.Current
		state.HasPrevious = true
	}
	state.Current = nil
	state.Pending = false
	state.MSISeedDigest = digest
	sequence, e := strconv.ParseUint(MSISequence, 10, 64)
	if e != nil {
		return e
	}
	state.MSISequence = sequence
	if state.MSISequence < state.HighestSequence {
		state.MSISequence = state.HighestSequence
	}
	return s.Save(state)
}

func (s Store) FinishMSISeed(rollback bool) error {
	root, e := os.OpenRoot(s.Path)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	defer root.Close()
	data, e := root.ReadFile("msi-rollback.json")
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	if rollback {
		// Store exactly the old schema, so an older rolled-back host can read it.
		if e = atomicWrite(root, "state.json", data); e != nil {
			return e
		}
	}
	return root.Remove("msi-rollback.json")
}
