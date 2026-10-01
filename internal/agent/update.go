package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/buildinfo"
)

const maxUpdateSize = 128 << 20

func updateArtifact() string {
	name := "remote-agent-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func fileDigest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func updateResponse(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Never send enrollment credentials, cookies or model tokens to downloads.
	req.Header.Set("Cache-Control", "no-cache")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach update server: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("update download returned HTTP %d", res.StatusCode)
	}
	return res, nil
}

func updateChecksum(ctx context.Context, client *http.Client, base, name string) (string, error) {
	res, err := updateResponse(ctx, client, base+"/downloads/"+name+".sha256")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 1025))
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(b))
	if len(b) > 1024 || len(fields) != 2 || fields[1] != name || len(fields[0]) != 64 {
		return "", errors.New("invalid update checksum manifest")
	}
	if _, err = hex.DecodeString(fields[0]); err != nil {
		return "", errors.New("invalid update checksum")
	}
	return strings.ToLower(fields[0]), nil
}

func downloadUpdate(ctx context.Context, client *http.Client, url, directory, expected string) (path string, err error) {
	res, err := updateResponse(ctx, client, url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.ContentLength > maxUpdateSize {
		return "", errors.New("update exceeds 128 MiB limit")
	}
	return stageUpdate(res.Body, directory, expected)
}

func stageUpdate(source io.Reader, directory, expected string) (path string, err error) {
	f, err := os.CreateTemp(directory, ".remote-agent-update-*.exe")
	if err != nil {
		return "", err
	}
	path = f.Name()
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(source, maxUpdateSize+1))
	if err != nil {
		return path, err
	}
	if n == 0 || n > maxUpdateSize || hex.EncodeToString(h.Sum(nil)) != expected {
		return path, errors.New("update checksum mismatch or invalid download size; installed binary is unchanged")
	}
	if err = f.Chmod(0700); err != nil {
		return path, err
	}
	if err = f.Sync(); err != nil {
		return path, err
	}
	err = f.Close()
	return path, err
}

// serviceUpdate controls only an existing autostart entry matching this binary
// and configuration. Unmanaged foreground processes need a manual restart.
type serviceUpdate struct {
	managed bool
	stop    func() error
	start   func() error
}

func replaceUpdate(exe, candidate string, service serviceUpdate) error {
	backup := exe + ".previous"
	if err := service.stop(); err != nil {
		return errors.Join(err, service.start())
	}
	// A previous foreground process may still hold this file open on Windows.
	if err := os.Remove(backup); err != nil && !os.IsNotExist(err) {
		return errors.Join(fmt.Errorf("cannot rotate previous binary; stop any older foreground agent: %w", err), service.start())
	}
	if err := os.Rename(exe, backup); err != nil {
		return errors.Join(fmt.Errorf("cannot back up executable: %w", err), service.start())
	}
	if err := os.Rename(candidate, exe); err != nil {
		restore := os.Rename(backup, exe)
		if restore != nil {
			return errors.Join(err, fmt.Errorf("restore %s manually: %w", backup, restore))
		}
		return errors.Join(fmt.Errorf("update replacement failed; restored previous binary: %w", err), service.start())
	}
	if err := service.start(); err != nil {
		// Stop a partially started service before restoring the previous binary.
		stopErr := service.stop()
		if stopErr != nil {
			return errors.Join(err, stopErr, fmt.Errorf("previous binary retained at %s", backup))
		}
		if moveErr := os.Rename(exe, candidate); moveErr != nil {
			return errors.Join(err, moveErr, fmt.Errorf("previous binary retained at %s", backup))
		}
		if restore := os.Rename(backup, exe); restore != nil {
			return errors.Join(err, fmt.Errorf("restore %s manually: %w", backup, restore))
		}
		return errors.Join(fmt.Errorf("service restart failed; restored previous binary: %w", err), service.start())
	}
	return nil
}

// Update obtains the published binary only from the enrolled control server.
// It never reenrolls or writes configuration, history, or SaySo data.
type UpdateOptions struct {
	CheckOnly bool
	Installer bool   // Use this downloaded helper to upgrade the standard installation.
	Server    string // Optional check that the installer belongs to the bound server.
}

func Update(ctx context.Context, configPath string, options UpdateOptions, output io.Writer) error {
	c, err := Load(configPath)
	if err != nil {
		return err
	}
	if options.Server != "" && strings.TrimSuffix(options.Server, "/") != strings.TrimSuffix(c.Server, "/") {
		return errors.New("this device is bound to a different server; use its original installer or remote-agent update")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	helper := exe
	if options.Installer {
		name := "remote-agent"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		exe, err = filepath.EvalSymlinks(filepath.Join(filepath.Dir(configPath), "bin", name))
		if err != nil {
			return fmt.Errorf("cannot find the existing installed CLI: %w", err)
		}
	}
	client := &http.Client{Timeout: 3 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	base, name := strings.TrimSuffix(c.Server, "/"), updateArtifact()
	label := "Current version"
	if options.Installer {
		label = "Installer version"
	}
	fmt.Fprintf(output, "%s: %s (%s/%s)\nChecking %s\n", label, buildinfo.Current(), runtime.GOOS, runtime.GOARCH, base)
	if !options.CheckOnly {
		unlock, err := lockUpdate(exe + ".update.lock")
		if err != nil {
			return err
		}
		defer unlock()
	}
	expected, err := updateChecksum(ctx, client, base, name)
	if err != nil {
		return err
	}
	actual, err := fileDigest(exe)
	if err != nil {
		return err
	}
	if actual == expected {
		if options.Installer && !options.CheckOnly {
			service, err := prepareServiceUpdate(exe, configPath, true)
			if err != nil {
				return err
			}
			if service.managed {
				if err := service.stop(); err != nil {
					return errors.Join(err, service.start())
				}
				if err := service.start(); err != nil {
					return err
				}
				fmt.Fprintln(output, "Autostart agent restarted. Reconnect from the web console.")
			}
		}
		fmt.Fprintln(output, "Already up to date with the server's published build.")
		return nil
	}
	if options.CheckOnly {
		fmt.Fprintln(output, "A different published build is available. Run remote-agent update to install it.")
		return nil
	}
	var candidate string
	if options.Installer {
		fmt.Fprintln(output, "Verifying the downloaded installer against the bound server...")
		f, openErr := os.Open(helper)
		if openErr != nil {
			return openErr
		}
		candidate, err = stageUpdate(f, filepath.Dir(exe), expected)
		f.Close()
	} else {
		fmt.Fprintln(output, "Downloading and verifying the published build...")
		candidate, err = downloadUpdate(ctx, client, base+"/downloads/"+name, filepath.Dir(exe), expected)
	}
	if err != nil {
		return err
	}
	defer os.Remove(candidate)
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	version, err := command(probeCtx, candidate, []string{"version"}).Output()
	if err != nil || !strings.HasPrefix(string(version), "remote-agent ") || len(version) > 256 {
		return errors.New("downloaded binary failed its version check; installed binary is unchanged")
	}
	service, err := prepareServiceUpdate(exe, configPath, options.Installer)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = replaceUpdate(exe, candidate, service); err != nil {
		return err
	}
	fmt.Fprintf(output, "Updated to %sBackup: %s.previous\n", version, exe)
	if service.managed {
		fmt.Fprintln(output, "Autostart agent restarted. Reconnect from the web console.")
	} else {
		fmt.Fprintln(output, "No running matching autostart service was restarted. Restart any foreground agent or custom supervisor to load the update.")
	}
	return nil
}
