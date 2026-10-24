package personalgateway

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Device struct {
	Name      string `json:"name"`
	SHA256    string `json:"sha256"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

type Registry struct {
	Devices []Device `json:"devices"`
}

func LoadRegistry(path string) (Registry, error) {
	var registry Registry
	data, err := os.ReadFile(path)
	if err != nil {
		return registry, err
	}
	if err = json.Unmarshal(data, &registry); err != nil {
		return registry, errors.New("invalid device registry")
	}
	for _, device := range registry.Devices {
		digest, err := hex.DecodeString(device.SHA256)
		if err != nil || len(digest) != sha256.Size || device.Name == "" {
			return registry, errors.New("invalid device registry entry")
		}
	}
	return registry, nil
}

func (registry Registry) Authorize(token string) bool {
	if !strings.HasPrefix(token, "sunny_") || len(token) < 45 || len(token) > 100 {
		return false
	}
	digest := sha256.Sum256([]byte(token))
	valid := 0
	for _, device := range registry.Devices {
		expected, _ := hex.DecodeString(device.SHA256)
		match := subtle.ConstantTimeCompare(digest[:], expected)
		if device.Enabled {
			valid |= match
		}
	}
	return valid == 1
}

// AddDevice returns a secret once. Only its SHA-256 verifier is stored server-side.
func AddDevice(path, name string) (string, error) {
	if strings.TrimSpace(name) == "" || len(name) > 80 {
		return "", errors.New("device name must contain 1-80 characters")
	}
	for _, character := range name {
		if character < 32 || character == 127 {
			return "", errors.New("device name must not contain control characters")
		}
	}
	unlock, err := lockRegistry(path)
	if err != nil {
		return "", err
	}
	defer unlock()
	registry, err := LoadRegistry(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	for _, device := range registry.Devices {
		if device.Name == name {
			return "", errors.New("device name already exists")
		}
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	token := "sunny_" + base64.RawURLEncoding.EncodeToString(secret)
	digest := sha256.Sum256([]byte(token))
	registry.Devices = append(registry.Devices, Device{Name: name, SHA256: hex.EncodeToString(digest[:]), Enabled: true, CreatedAt: time.Now().Unix()})
	return token, saveRegistry(path, registry)
}

func RevokeDevice(path, name string) error {
	unlock, err := lockRegistry(path)
	if err != nil {
		return err
	}
	defer unlock()
	registry, err := LoadRegistry(path)
	if err != nil {
		return err
	}
	found := false
	for i := range registry.Devices {
		if registry.Devices[i].Name == name {
			registry.Devices[i].Enabled = false
			found = true
		}
	}
	if !found {
		return errors.New("device not found")
	}
	return saveRegistry(path, registry)
}

func DeleteDevice(path, name string) error {
	unlock, err := lockRegistry(path)
	if err != nil {
		return err
	}
	defer unlock()
	registry, err := LoadRegistry(path)
	if err != nil {
		return err
	}
	for i, device := range registry.Devices {
		if device.Name == name {
			if device.Enabled {
				return errors.New("revoke the device before deleting it")
			}
			registry.Devices = append(registry.Devices[:i], registry.Devices[i+1:]...)
			return saveRegistry(path, registry)
		}
	}
	return errors.New("device not found")
}

func lockRegistry(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	return lockRegistryFile(path + ".lock")
}

func saveRegistry(path string, registry Registry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".devices-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := json.NewEncoder(file).Encode(registry); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
