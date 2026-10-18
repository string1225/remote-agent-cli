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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func installFixture(t *testing.T, archive []byte, bin, goos string) (string, error) {
	t.Helper()
	name, err := AssetName(goos, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(archive)
	var fixture *httptest.Server
	fixture = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/release" {
			_ = json.NewEncoder(w).Encode(Release{Assets: []Asset{{Name: name, Digest: "sha256:" + hex.EncodeToString(sum[:]), URL: fixture.URL + "/archive"}}})
			return
		}
		_, _ = w.Write(archive)
	}))
	defer fixture.Close()
	return Install(context.Background(), fixture.URL+"/release", bin, goos, "amd64", fixture.Client())
}

func windowsPackage(t *testing.T, version, omitted string, extras map[string]string) []byte {
	t.Helper()
	var data bytes.Buffer
	w := zip.NewWriter(&data)
	// The official ZIP lists the code-mode host before the architecture CLI.
	entries := []struct{ name, body string }{
		{"codex-code-mode-host.exe", "host-" + version},
		{"codex-command-runner.exe", "runner-" + version},
		{"codex-resources/codex-command-runner.exe", "resource-runner-" + version},
		{"codex-resources/codex-windows-sandbox-setup.exe", "setup-" + version},
		{"codex-path/rg.exe", "search-" + version},
		{"codex-resources/voice/runtime.json", "voice-" + version},
		{"codex-package.json", `{"layoutVersion":1,"entrypoint":"codex-x86_64-pc-windows-msvc.exe","resourcesDir":"codex-resources","pathDir":"codex-path"}`},
		{"codex-x86_64-pc-windows-msvc.exe", "main-" + version},
	}
	for name, body := range extras {
		entries = append(entries, struct{ name, body string }{name, body})
	}
	for _, entry := range entries {
		if entry.name == omitted {
			continue
		}
		f, err := w.Create(entry.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(f, entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestInstallWindowsPreservesCompletePackageAndUpdatesTogether(t *testing.T) {
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "remote-agent.exe"), []byte("keep-agent"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"one", "two"} {
		path, err := installFixture(t, windowsPackage(t, version, "", nil), bin, "windows")
		if err != nil {
			t.Fatal(err)
		}
		for file, expected := range map[string]string{
			"codex.exe": "main-" + version, "codex-code-mode-host.exe": "host-" + version,
			"codex-x86_64-pc-windows-msvc.exe":                "main-" + version,
			"codex-command-runner.exe":                        "runner-" + version,
			"codex-resources/codex-command-runner.exe":        "resource-runner-" + version,
			"codex-resources/codex-windows-sandbox-setup.exe": "setup-" + version,
			"codex-path/rg.exe":                               "search-" + version, "codex-resources/voice/runtime.json": "voice-" + version,
			"remote-agent.exe":   "keep-agent",
			"codex-package.json": `{"layoutVersion":1,"entrypoint":"codex-x86_64-pc-windows-msvc.exe","resourcesDir":"codex-resources","pathDir":"codex-path"}`,
		} {
			got, readErr := os.ReadFile(filepath.Join(bin, filepath.FromSlash(file)))
			if readErr != nil || string(got) != expected {
				t.Fatalf("package file %s = %q, %v; want %q", file, got, readErr, expected)
			}
		}
		if path != filepath.Join(bin, "codex.exe") {
			t.Fatalf("unexpected installed entrypoint %s", path)
		}
	}
}

func TestInstallRejectsIncompleteAndUnsafePackagesBeforeReplacement(t *testing.T) {
	for _, scenario := range []string{"missing-main", "missing-sandbox", "missing-host", "missing-search", "traversal", "backslash", "reserved", "duplicate-main"} {
		t.Run(scenario, func(t *testing.T) {
			bin := t.TempDir()
			old := filepath.Join(bin, "codex.exe")
			if err := os.WriteFile(old, []byte("previous"), 0700); err != nil {
				t.Fatal(err)
			}
			omitted, extras := "", map[string]string{}
			switch scenario {
			case "missing-main":
				omitted = "codex-x86_64-pc-windows-msvc.exe"
			case "missing-sandbox":
				omitted = "codex-resources/codex-windows-sandbox-setup.exe"
			case "missing-host":
				omitted = "codex-code-mode-host.exe"
			case "missing-search":
				omitted = "codex-path/rg.exe"
			case "traversal":
				extras["../escaped.txt"] = "unsafe"
			case "backslash":
				extras[`codex-resources\..\..\escaped.txt`] = "unsafe"
			case "reserved":
				extras["remote-agent.exe"] = "unsafe"
			case "duplicate-main":
				extras["codex.exe"] = "wrong-main"
			}
			if _, err := installFixture(t, windowsPackage(t, "new", omitted, extras), bin, "windows"); err == nil {
				t.Fatal("invalid package was accepted")
			}
			got, err := os.ReadFile(old)
			if err != nil || string(got) != "previous" {
				t.Fatalf("invalid package damaged existing installation: %q, %v", got, err)
			}
		})
	}
}

func TestInstallRejectsArchiveLinksAndOversizedFiles(t *testing.T) {
	for _, scenario := range []string{"zip-link", "tar-link", "oversize"} {
		t.Run(scenario, func(t *testing.T) {
			var data bytes.Buffer
			goos := "windows"
			if scenario == "tar-link" {
				goos = "linux"
				gz := gzip.NewWriter(&data)
				tw := tar.NewWriter(gz)
				if err := tw.WriteHeader(&tar.Header{Name: "codex-resources/escape", Typeflag: tar.TypeSymlink, Linkname: "../../outside"}); err != nil {
					t.Fatal(err)
				}
				if err := tw.Close(); err != nil {
					t.Fatal(err)
				}
				if err := gz.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				zw := zip.NewWriter(&data)
				if scenario == "zip-link" {
					header := &zip.FileHeader{Name: "codex-resources/escape", Method: zip.Store}
					header.SetMode(os.ModeSymlink | 0777)
					w, err := zw.CreateHeader(header)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := io.WriteString(w, "../../outside"); err != nil {
						t.Fatal(err)
					}
				} else {
					if _, err := zw.CreateRaw(&zip.FileHeader{Name: "codex-x86_64-pc-windows-msvc.exe", Method: zip.Store, UncompressedSize64: maxBinarySize + 1}); err != nil {
						t.Fatal(err)
					}
				}
				if err := zw.Close(); err != nil {
					t.Fatal(err)
				}
			}
			bin := t.TempDir()
			if _, err := installFixture(t, data.Bytes(), bin, goos); err == nil {
				t.Fatal("unsafe archive accepted")
			}
			entries, err := os.ReadDir(bin)
			if err != nil || len(entries) != 0 {
				t.Fatalf("failed installation left files behind: %v, %v", entries, err)
			}
		})
	}
}

func TestInstallRollsBackCompanionsWhenRunningWindowsCLIBlocksReplacement(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows denies renaming an executable opened without delete sharing")
	}
	bin := t.TempDir()
	if _, err := installFixture(t, windowsPackage(t, "old", "", nil), bin, "windows"); err != nil {
		t.Fatal(err)
	}
	locked, err := os.Open(filepath.Join(bin, "codex.exe"))
	if err != nil {
		t.Fatal(err)
	}
	defer locked.Close()
	if _, err := installFixture(t, windowsPackage(t, "new", "", nil), bin, "windows"); err == nil {
		t.Fatal("replacing a locked executable unexpectedly succeeded")
	}
	for file, expected := range map[string]string{"codex.exe": "main-old", "codex-code-mode-host.exe": "host-old", "codex-resources/codex-windows-sandbox-setup.exe": "setup-old", "codex-path/rg.exe": "search-old"} {
		got, err := os.ReadFile(filepath.Join(bin, filepath.FromSlash(file)))
		if err != nil || string(got) != expected {
			t.Fatalf("rollback did not restore %s: %q, %v", file, got, err)
		}
	}
	entries, _ := os.ReadDir(bin)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".codex-") {
			t.Fatalf("temporary installation data left behind: %s", entry.Name())
		}
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestExtractSupportsCurrentOfficialWindowsBinarySize(t *testing.T) {
	var data bytes.Buffer
	w := zip.NewWriter(&data)
	f, err := w.Create("codex-x86_64-pc-windows-msvc.exe")
	if err != nil {
		t.Fatal(err)
	}
	const size = 341270320
	if _, err := io.CopyN(f, zeroReader{}, size); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	packageFiles := &packageExtractor{directory: dir, main: "codex-x86_64-pc-windows-msvc.exe", exe: "codex.exe", seen: map[string]bool{}}
	if err := extractZIP(bytes.NewReader(data.Bytes()), int64(data.Len()), packageFiles); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(dir, packageFiles.main)); err != nil || info.Size() != size {
		t.Fatalf("large official executable was not extracted intact: %v, %v", info, err)
	}
}

func TestUnixAssetsUseCompletePackages(t *testing.T) {
	for goos, expected := range map[string]string{"linux": "codex-package-x86_64-unknown-linux-musl.tar.gz", "darwin": "codex-package-x86_64-apple-darwin.tar.gz"} {
		name, err := AssetName(goos, "amd64")
		if err != nil || name != expected {
			t.Fatalf("%s selected %s, %v instead of complete official package", goos, name, err)
		}
	}
}

func TestExtractRejectsHardLinksRootFilesAndExcessDirectories(t *testing.T) {
	for _, scenario := range []string{"hard-link", "root-file", "excess-directories", "directory-data"} {
		t.Run(scenario, func(t *testing.T) {
			packageFiles := &packageExtractor{directory: t.TempDir(), main: "bin/codex", exe: "codex", seen: map[string]bool{}}
			if scenario == "directory-data" {
				if err := packageFiles.add("codex-resources/", os.ModeDir, 1, strings.NewReader("a")); err == nil {
					t.Fatal("directory file data accepted")
				}
				return
			}
			var data bytes.Buffer
			gz := gzip.NewWriter(&data)
			tw := tar.NewWriter(gz)
			headers := []*tar.Header{{Name: "codex-resources/escape", Typeflag: tar.TypeLink, Linkname: "../../outside"}}
			if scenario == "root-file" {
				headers = []*tar.Header{{Name: ".", Typeflag: tar.TypeReg}}
			}
			if scenario == "excess-directories" {
				headers = nil
				for i := 0; i <= maxPackageFiles; i++ {
					headers = append(headers, &tar.Header{Name: "codex-resources/", Typeflag: tar.TypeDir})
				}
			}
			if scenario == "directory-data" {
				headers = []*tar.Header{{Name: "codex-resources/", Typeflag: tar.TypeDir, Size: 1}}
			}
			for _, header := range headers {
				if err := tw.WriteHeader(header); err != nil {
					t.Fatal(err)
				}
			}
			if err := tw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			if err := extractTAR(bytes.NewReader(data.Bytes()), packageFiles); err == nil {
				t.Fatal("malformed or excessive archive accepted")
			}
		})
	}
}

func TestInstallChecksArchiveDigestAndPreservesUnixCompanions(t *testing.T) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	binary := []byte("official-test-binary")
	_ = tw.WriteHeader(&tar.Header{Name: "bin/codex", Mode: 0755, Size: int64(len(binary))})
	_, _ = tw.Write(binary)
	for name, body := range map[string]string{"bin/codex-code-mode-host": "host", "codex-resources/bwrap": "sandbox", "codex-path/rg": "search"} {
		_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(body))})
		_, _ = io.WriteString(tw, body)
	}
	_ = tw.Close()
	_ = gz.Close()
	sum := sha256.Sum256(archive.Bytes())
	digest := "sha256:" + hex.EncodeToString(sum[:])
	var fixture *httptest.Server
	fixture = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/release" {
			_ = json.NewEncoder(w).Encode(map[string]any{"assets": []any{map[string]any{"name": "codex-package-x86_64-unknown-linux-musl.tar.gz", "digest": digest, "browser_download_url": fixture.URL + "/archive"}}})
			return
		}
		_, _ = w.Write(archive.Bytes())
	}))
	defer fixture.Close()
	dir := t.TempDir()
	path, err := Install(context.Background(), fixture.URL+"/release", dir, "linux", "amd64", fixture.Client())
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, binary) {
		t.Fatal("wrong binary extracted")
	}
	for _, name := range []string{"bin/codex", "bin/codex-code-mode-host", "codex-resources/bwrap", "codex-path/rg"} {
		if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err != nil || runtime.GOOS != "windows" && info.Mode()&0100 == 0 {
			t.Fatalf("Unix helper missing or not executable: %s, %v", name, err)
		}
	}
	digest = "sha256:" + string(bytes.Repeat([]byte("0"), 64))
	if _, err := Install(context.Background(), fixture.URL+"/release", dir, "linux", "amd64", fixture.Client()); err == nil {
		t.Fatal("checksum mismatch accepted")
	}
	data, _ = os.ReadFile(filepath.Join(dir, "codex"))
	if !bytes.Equal(data, binary) {
		t.Fatal("failed download damaged existing binary")
	}
}
