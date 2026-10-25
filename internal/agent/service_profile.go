package agent

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ServiceProfile contains paths and an explicit tool environment, never device
// or model credentials. Business data stays beside the original configuration.
type ServiceProfile struct {
	OwnerSID     string            `json:"ownerSid,omitempty"`
	UpdateServer string            `json:"updateServer,omitempty"`
	Config       string            `json:"config"`
	Environment  map[string]string `json:"environment"`
}

var serviceEnvironmentKeys = map[string]bool{
	"USERPROFILE": true, "HOME": true, "HOMEDRIVE": true, "HOMEPATH": true,
	"APPDATA": true, "LOCALAPPDATA": true, "CODEX_HOME": true, "PATH": true,
	"RA_CODEX_PATH": true, "RA_QODER_PATH": true,
}

func CaptureServiceProfile(config string) (ServiceProfile, error) {
	path, err := filepath.Abs(config)
	if err != nil {
		return ServiceProfile{}, err
	}
	c, err := Load(path)
	if err != nil {
		return ServiceProfile{}, err
	}
	p := ServiceProfile{Config: path, Environment: map[string]string{}}
	p.OwnerSID, err = currentProfileOwner()
	if err != nil {
		return p, err
	}
	p.UpdateServer = c.Server
	for key := range serviceEnvironmentKeys {
		if value := os.Getenv(key); value != "" {
			p.Environment[key] = value
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p, err
	}
	p.Environment["USERPROFILE"] = home
	if p.Environment["CODEX_HOME"] == "" {
		p.Environment["CODEX_HOME"] = filepath.Join(home, ".codex")
	}
	for name, provider := range Detect() {
		if provider.Available {
			if name == "codex" {
				p.Environment["RA_CODEX_PATH"] = provider.Path
			} else {
				p.Environment["RA_QODER_PATH"] = provider.Path
			}
		}
	}
	return p, p.Validate()
}

func (p ServiceProfile) Validate() error {
	if !filepath.IsAbs(p.Config) || !filepath.IsAbs(p.Environment["USERPROFILE"]) || !filepath.IsAbs(p.Environment["CODEX_HOME"]) {
		return errors.New("service config, user profile and Codex home must be absolute paths")
	}
	for key := range p.Environment {
		if !serviceEnvironmentKeys[key] {
			return errors.New("unsupported service environment key")
		}
	}
	return nil
}

func ReadServiceProfile(path string) (ServiceProfile, error) {
	var p ServiceProfile
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	if len(data) > 128<<10 {
		return p, errors.New("service profile is too large")
	}
	if err = json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, p.Validate()
}

func (p ServiceProfile) Save(path string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".service-profile-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
