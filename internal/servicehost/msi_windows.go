package servicehost

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/string1225/remote-agent-cli/internal/clientupdate"
	"golang.org/x/sys/windows"
	"io"
	"os"
)

func adoptMSISeed(store clientupdate.Store, path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return e
	}
	return store.AdoptMSISeed(hex.EncodeToString(h.Sum(nil)))
}
func finishMSISeed(rollback bool) error {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("MSI runtime maintenance requires administrator")
	}
	store, e := runtimeStore()
	if e != nil {
		return e
	}
	return store.FinishMSISeed(rollback)
}
