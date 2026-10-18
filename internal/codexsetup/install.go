package codexsetup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Release struct {
	Assets []Asset `json:"assets"`
}

// The compressed release is much smaller than the executable (currently over
// 325 MiB on Windows). Keep separate bounds for transfer and extraction.
const (
	maxBinarySize   = 512 << 20
	maxPackageSize  = 2 << 30
	maxPackageFiles = 10000
)

type Asset struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
	URL    string `json:"browser_download_url"`
}

func AssetName(goos, arch string) (string, error) {
	cpu := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[arch]
	target := map[string]string{"windows": "pc-windows-msvc.exe.zip", "darwin": "apple-darwin.tar.gz", "linux": "unknown-linux-musl.tar.gz"}[goos]
	if cpu == "" || target == "" {
		return "", errors.New("unsupported Codex platform")
	}
	if goos != "windows" {
		return "codex-package-" + cpu + "-" + target, nil
	}
	return "codex-" + cpu + "-" + target, nil
}

// Install accepts release metadata from the configured platform. GitHub's asset
// SHA-256 is mandatory; a partial or mismatched archive never replaces a binary.
func Install(ctx context.Context, releaseURL, bin, goos, arch string, client *http.Client) (string, error) {
	name, err := AssetName(goos, arch)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(req)
	if err != nil {
		return "", errors.New("cannot read official Codex release metadata")
	}
	var release Release
	if response.StatusCode != 200 {
		response.Body.Close()
		return "", fmt.Errorf("Codex release metadata returned HTTP %d", response.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&release)
	response.Body.Close()
	if err != nil {
		return "", errors.New("invalid Codex release metadata")
	}
	var asset Asset
	for _, item := range release.Assets {
		if item.Name == name {
			asset = item
			break
		}
	}
	digest, err := hex.DecodeString(strings.TrimPrefix(asset.Digest, "sha256:"))
	if asset.URL == "" || !strings.HasPrefix(asset.Digest, "sha256:") || err != nil || len(digest) != 32 {
		return "", errors.New("official Codex archive or SHA-256 digest unavailable")
	}
	if err := os.MkdirAll(bin, 0700); err != nil {
		return "", err
	}
	lockPath := filepath.Join(bin, ".codex-install.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("cannot lock Codex installation (if no installer is running, remove %s): %w", lockPath, err)
	}
	defer func() { lock.Close(); os.Remove(lockPath) }()
	req, err = http.NewRequestWithContext(ctx, "GET", asset.URL, nil)
	if err != nil {
		return "", err
	}
	response, err = client.Do(req)
	if err != nil {
		return "", errors.New("could not download Codex")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", fmt.Errorf("Codex download returned HTTP %d", response.StatusCode)
	}
	archive, err := os.CreateTemp(bin, ".codex-archive-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(archive.Name())
	defer archive.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(archive, hash), io.LimitReader(response.Body, (256<<20)+1))
	if err != nil {
		return "", fmt.Errorf("Codex archive download failed: %w", err)
	}
	if size > 256<<20 {
		return "", errors.New("Codex archive exceeded 256 MiB")
	}
	if !bytes.Equal(digest, hash.Sum(nil)) {
		return "", errors.New("Codex archive checksum mismatch")
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	exe := "codex"
	if goos == "windows" {
		exe += ".exe"
	}
	staging, err := os.MkdirTemp(bin, ".codex-stage-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(staging)
	main := strings.TrimSuffix(strings.TrimSuffix(name, ".tar.gz"), ".zip")
	if goos != "windows" {
		main = "bin/codex"
	}
	packageFiles := &packageExtractor{directory: staging, main: main, exe: exe, seen: map[string]bool{}}
	if strings.HasSuffix(name, ".zip") {
		err = extractZIP(archive, size, packageFiles)
	} else {
		err = extractTAR(archive, packageFiles)
	}
	if err != nil {
		return "", err
	}
	if err := packageFiles.validate(goos); err != nil {
		return "", err
	}
	if err := os.Chmod(filepath.Join(staging, main), 0700); err != nil {
		return "", err
	}
	// Keep the architecture entrypoint referenced by codex-package.json. The
	// convenient command is another link to that same binary when supported.
	if err := commandAlias(filepath.Join(staging, main), filepath.Join(staging, exe), goos); err != nil {
		return "", err
	}
	if err := publishPackage(staging, bin, exe); err != nil {
		return "", err
	}
	return filepath.Join(bin, exe), nil
}

func commandAlias(source, destination, goos string) error {
	if goos != "windows" && runtime.GOOS != "windows" {
		target, err := filepath.Rel(filepath.Dir(destination), source)
		if err != nil {
			return err
		}
		// The native binary must resolve to the canonical bin/ directory to find
		// its adjacent code host and the resources in the package parent.
		return os.Symlink(target, destination)
	}
	if err := os.Link(source, destination); err == nil {
		return nil
	}
	// Some home directories use filesystems without hard-link support.
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	if copyErr == nil {
		copyErr = output.Sync()
	}
	return errors.Join(copyErr, output.Close())
}

type packageExtractor struct {
	directory, main, exe string
	seen                 map[string]bool
	total                int64
	entries              int
}

func (p *packageExtractor) relative(name string) (string, error) {
	if strings.ContainsAny(name, `\:`) || strings.HasPrefix(name, "/") {
		return "", errors.New("unsafe Codex archive path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || part != "." && strings.TrimRight(part, ". ") != part || strings.ContainsAny(part, "<>\"|?*\x00") {
			return "", errors.New("unsafe Codex archive path")
		}
	}
	name = path.Clean(name)
	if name == "." {
		return "", nil
	}
	if !filepath.IsLocal(filepath.FromSlash(name)) {
		return "", errors.New("unsafe Codex archive path")
	}
	first := strings.Split(name, "/")[0]
	// The managed bin directory also contains remote-agent and may contain other
	// user tools. A release can publish only Codex package paths within it.
	if first != "bin" && first != "bwrap" && first != "rg" && !strings.HasPrefix(first, "codex-") {
		return "", errors.New("unexpected file in Codex archive")
	}
	if strings.EqualFold(name, p.exe) {
		return "", errors.New("ambiguous Codex entrypoint in archive")
	}
	return filepath.FromSlash(name), nil
}

func (p *packageExtractor) add(name string, mode os.FileMode, size int64, reader io.Reader) error {
	p.entries++
	if p.entries > maxPackageFiles || size < 0 || size > maxBinarySize || p.total+size > maxPackageSize {
		return errors.New("invalid Codex package size or entry count")
	}
	if mode&os.ModeSymlink != 0 || !mode.IsRegular() && !mode.IsDir() {
		return errors.New("links and special files are not supported in Codex archives")
	}
	if mode.IsDir() && size != 0 {
		return errors.New("Codex archive directory contains file data")
	}
	relative, err := p.relative(name)
	if err != nil {
		return err
	}
	if relative == "" {
		if mode.IsDir() {
			return nil
		}
		return errors.New("Codex archive root must be a directory")
	}
	destination := filepath.Join(p.directory, relative)
	if mode.IsDir() {
		return os.MkdirAll(destination, 0700)
	}
	key := strings.ToLower(filepath.ToSlash(relative))
	if p.seen[key] {
		return errors.New("duplicate file or invalid Codex package size")
	}
	p.seen[key], p.total = true, p.total+size
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	permissions := os.FileMode(0600)
	if mode&0111 != 0 || strings.HasSuffix(name, ".exe") {
		permissions = 0700
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, permissions)
	if err != nil {
		return err
	}
	n, copyErr := io.Copy(output, io.LimitReader(reader, size+1))
	if copyErr == nil && n != size {
		copyErr = errors.New("invalid Codex package file size")
	}
	if copyErr == nil {
		copyErr = output.Sync()
	}
	return errors.Join(copyErr, output.Close())
}

func (p *packageExtractor) validate(goos string) error {
	required := [][]string{{p.main}, {"codex-code-mode-host", "bin/codex-code-mode-host"}, {"codex-path/rg"}}
	if goos == "windows" {
		required = [][]string{{p.main}, {"codex-code-mode-host.exe", "bin/codex-code-mode-host.exe"}, {"codex-path/rg.exe"},
			{"codex-command-runner.exe", "codex-resources/codex-command-runner.exe"},
			{"codex-windows-sandbox-setup.exe", "codex-resources/codex-windows-sandbox-setup.exe"}}
	} else if goos == "linux" {
		required = append(required, []string{"bwrap", "codex-resources/bwrap"})
	}
	for _, alternatives := range required {
		found := false
		for _, name := range alternatives {
			if info, err := os.Stat(filepath.Join(p.directory, filepath.FromSlash(name))); err == nil && info.Mode().IsRegular() && info.Size() > 0 {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Codex package is missing required executable %s", alternatives[0])
		}
	}
	return nil
}

func extractTAR(reader io.Reader, output *packageExtractor) error {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return errors.New("invalid Codex gzip archive")
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("invalid Codex tar archive")
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA && header.Typeflag != tar.TypeDir {
			return errors.New("links and special files are not supported in Codex archives")
		}
		if err := output.add(header.Name, header.FileInfo().Mode(), header.Size, archive); err != nil {
			return err
		}
	}
	return nil
}
func extractZIP(reader io.ReaderAt, size int64, output *packageExtractor) error {
	archive, err := zip.NewReader(reader, size)
	if err != nil {
		return errors.New("invalid Codex zip archive")
	}
	for _, file := range archive.File {
		if file.UncompressedSize64 > maxBinarySize {
			return errors.New("invalid Codex package file size")
		}
		input, err := file.Open()
		if err != nil {
			return err
		}
		err = output.add(file.Name, file.Mode(), int64(file.UncompressedSize64), input)
		closeErr := input.Close()
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
	}
	return nil
}

// Publish companions before the entrypoint. Every replacement is an atomic
// rename; a failure restores all earlier entries and retains any failed restore.
func publishPackage(staging, bin, exe string) error {
	backup, err := os.MkdirTemp(bin, ".codex-backup-*")
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(staging)
	if err != nil {
		os.Remove(backup)
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() != exe {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	names = append(names, exe)
	type replacement struct {
		name                string
		backedUp, installed bool
	}
	var replaced []replacement
	rollback := func(cause error) error {
		var restoreErr error
		for i := len(replaced) - 1; i >= 0; i-- {
			record := replaced[i]
			destination := filepath.Join(bin, record.name)
			if record.installed {
				if err := os.Rename(destination, filepath.Join(staging, record.name)); err != nil {
					restoreErr = errors.Join(restoreErr, err)
					continue
				}
			}
			if record.backedUp {
				restoreErr = errors.Join(restoreErr, os.Rename(filepath.Join(backup, record.name), destination))
			}
		}
		if restoreErr != nil {
			return errors.Join(cause, fmt.Errorf("restore previous Codex files from %s: %w", backup, restoreErr))
		}
		os.RemoveAll(backup)
		return fmt.Errorf("Codex installation failed; previous package restored (close active Codex processes before updating): %w", cause)
	}
	for _, name := range names {
		destination := filepath.Join(bin, name)
		replaced = append(replaced, replacement{name: name})
		record := &replaced[len(replaced)-1]
		if _, err := os.Lstat(destination); err == nil {
			if err := os.Rename(destination, filepath.Join(backup, name)); err != nil {
				return rollback(err)
			}
			record.backedUp = true
		} else if !os.IsNotExist(err) {
			return rollback(err)
		}
		if err := os.Rename(filepath.Join(staging, name), destination); err != nil {
			return rollback(err)
		}
		record.installed = true
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("Codex installed; old package cleanup failed at %s: %w", backup, err)
	}
	return nil
}
