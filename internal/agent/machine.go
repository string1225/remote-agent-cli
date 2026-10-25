package agent

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// MachineIdentity exposes only a domain-separated digest, never the OS identity.
// A private installation marker is used on systems without a readable OS ID.
func MachineIdentity(config string) (string, error) {
	identity, err := systemMachineIdentity()
	if err != nil || strings.TrimSpace(identity) == "" {
		path := filepath.Join(filepath.Dir(config), "machine-id")
		data, e := os.ReadFile(path)
		if os.IsNotExist(e) {
			if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
				return "", e
			}
			var b [32]byte
			if _, e = rand.Read(b[:]); e != nil {
				return "", e
			}
			f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if os.IsExist(e) {
				data, e = os.ReadFile(path)
			} else if e == nil {
				data = []byte(hex.EncodeToString(b[:]))
				_, e = f.Write(data)
				if e == nil {
					e = f.Sync()
				}
				e = errors.Join(e, f.Close())
			}
		}
		if e != nil || len(data) < 32 {
			return "", errors.New("cannot read a stable local machine identity")
		}
		identity = "installation:" + string(data)
	}
	sum := sha256.Sum256([]byte("remote-agent-machine/v1\x00" + strings.TrimSpace(identity)))
	return hex.EncodeToString(sum[:]), nil
}
