package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

type Config struct {
	Server       string             `json:"server"`
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Credential   string             `json:"credential"`
	AllowedRoots []string           `json:"allowedRoots"`
	AllowWrite   bool               `json:"allowWrite"`
	Services     []protocol.Service `json:"services"`
	Path         string             `json:"-"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".remote-agent", "config.json")
}
func Load(path string) (Config, error) {
	var c Config
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	c.Path = path
	if c.ID == "" || c.Credential == "" {
		return c, errors.New("agent is not enrolled")
	}
	if err = ValidateServer(c.Server); err != nil {
		return c, err
	}
	return c, nil
}
func ValidateServer(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return errors.New("server must be an HTTPS origin")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1") {
		return nil
	}
	return errors.New("HTTPS is required except on localhost")
}
func (c Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(c.Path), ".config-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, c.Path)
}
func Enroll(server, token, path string, roots []string, allowWrite bool) (Config, error) {
	var c Config
	if err := ValidateServer(server); err != nil {
		return c, err
	}
	if _, err := os.Stat(path); err == nil {
		return c, errors.New("configuration already exists; revoke the old device and remove its configuration before reenrolling")
	}
	if len(roots) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return c, err
		}
		roots = []string{home}
	}
	for i, root := range roots {
		canonical, err := canonicalDir(root)
		if err != nil {
			return c, err
		}
		roots[i] = canonical
	}
	// Verify local storage before consuming the one-time token.
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return c, err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".enroll-*")
	if err != nil {
		return c, err
	}
	f.Close()
	os.Remove(f.Name())
	body, _ := json.Marshal(map[string]string{"token": token, "os": runtime.GOOS, "arch": runtime.GOARCH})
	client := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Post(strings.TrimSuffix(server, "/")+"/api/enroll", "application/json", bytes.NewReader(body))
	if err != nil {
		return c, errors.New("cannot reach enrollment server")
	}
	defer res.Body.Close()
	var result struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Credential string `json:"credential"`
		Error      string `json:"error"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 16384)).Decode(&result); err != nil {
		return c, err
	}
	if res.StatusCode != 200 {
		return c, fmt.Errorf("enrollment failed: %s", result.Error)
	}
	c = Config{Server: strings.TrimSuffix(server, "/"), ID: result.ID, Name: result.Name, Credential: result.Credential, AllowedRoots: roots, AllowWrite: allowWrite, Services: []protocol.Service{}, Path: path}
	if err = c.Save(); err != nil {
		return c, fmt.Errorf("device bound but credential could not be saved; revoke this device and enroll again: %w", err)
	}
	return c, nil
}
func canonicalDir(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("workspace must be an absolute path")
	}
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", errors.New("workspace must be a directory")
	}
	return filepath.Clean(path), nil
}
func (c Config) Workspace(path string) (string, error) {
	resolved, err := canonicalDir(path)
	if err != nil {
		return "", err
	}
	for _, root := range c.AllowedRoots {
		canonical, err := canonicalDir(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(canonical, resolved)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return resolved, nil
		}
	}
	return "", errors.New("workspace is outside the locally configured allowedRoots")
}
