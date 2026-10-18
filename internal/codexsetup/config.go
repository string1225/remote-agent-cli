// Package codexsetup installs Codex and configures the platform model provider.
package codexsetup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const Provider = "sunny_agents"

type Connection struct {
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
	Model   string `json:"model"`
}

func Home() (string, error) {
	if dir := os.Getenv("CODEX_HOME"); dir != "" {
		return filepath.Abs(dir)
	}
	home, err := os.UserHomeDir()
	return filepath.Join(home, ".codex"), err
}

// Apply keeps credentials local, in the private TOML file read by both GUI and
// service processes. No shell profile, system proxy or OAuth login is changed.
func Apply(home string, c Connection) (string, error) {
	if _, err := protocol.ParseServerURL(c.BaseURL); err != nil {
		return "", errors.New("invalid HTTPS gateway endpoint")
	}
	if c.Token == "" || strings.ContainsAny(c.Token, "\r\n") {
		return "", errors.New("a personal proxy token is required")
	}
	if err := os.MkdirAll(home, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(home, "config.toml")
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("config.toml is a symbolic link; configure its target explicitly")
	}
	before, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	config := map[string]any{}
	if len(before) > 0 {
		if err := toml.Unmarshal(bytes.TrimPrefix(before, []byte{0xef, 0xbb, 0xbf}), &config); err != nil {
			return "", errors.New("existing config.toml is invalid; it was not changed")
		}
	}
	providers, ok := config["model_providers"].(map[string]any)
	if !ok && config["model_providers"] != nil {
		return "", errors.New("model_providers must be a TOML table")
	}
	if providers == nil {
		providers = map[string]any{}
	}
	providers[Provider] = map[string]any{"name": "Sunny Codex", "base_url": c.BaseURL, "wire_api": "responses", "requires_openai_auth": false, "supports_websockets": false, "http_headers": map[string]string{"Authorization": "Bearer " + c.Token}}
	config["model_providers"] = providers
	config["model_provider"] = Provider
	// A configured default profile overrides root settings in both App and CLI.
	// Switch that profile too, while leaving explicitly chosen other profiles alone.
	if selected, ok := config["profile"].(string); ok && selected != "" {
		if profiles, ok := config["profiles"].(map[string]any); ok {
			if profile, ok := profiles[selected].(map[string]any); ok {
				profile["model_provider"] = Provider
				if profile["model_reasoning_effort"] == "extrahigh" {
					profile["model_reasoning_effort"] = "xhigh"
				}
			}
		}
	}
	if config["model"] == nil || config["model"] == "" {
		config["model"] = c.Model
	}
	if config["model_reasoning_effort"] == "extrahigh" {
		config["model_reasoning_effort"] = "xhigh"
	}
	after, err := toml.Marshal(config)
	if err != nil {
		return "", errors.New("could not serialize Codex configuration")
	}
	backup := ""
	if len(before) > 0 {
		backup = path + ".backup-" + time.Now().UTC().Format("20060102T150405.000000000")
		if err := privateWrite(backup, before, false); err != nil {
			return "", err
		}
	}
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if !bytes.Equal(current, before) {
		return "", errors.New("Codex configuration changed during setup; retry")
	}
	if err := privateWrite(path, after, true); err != nil {
		return "", err
	}
	return backup, nil
}

func privateWrite(path string, data []byte, replace bool) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".codex-setup-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = protect(f.Name()); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if !replace {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			return errors.New("backup already exists")
		}
	}
	return os.Rename(f.Name(), path)
}

// DeviceConnection exchanges an enrolled device credential for its own model
// token. The token is never sent in a URL or logged.
func DeviceConnection(client *http.Client, server, id, credential string) (Connection, error) {
	var c Connection
	if _, err := protocol.ParseServerURL(server); err != nil {
		return c, err
	}
	req, err := http.NewRequest("POST", strings.TrimSuffix(server, "/")+"/api/agent/codex?id="+id, strings.NewReader("{}"))
	if err != nil {
		return c, err
	}
	req.Header.Set("Authorization", "Bearer "+credential)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return c, errors.New("cannot reach the Codex setup service")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return c, fmt.Errorf("Codex setup service returned HTTP %d; check gateway configuration and device account", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&c); err != nil {
		return c, errors.New("invalid Codex setup response")
	}
	return c, nil
}
