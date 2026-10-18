package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

//go:embed installers/*
var installers embed.FS
var tokenRE = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var artifactRE = regexp.MustCompile(`^remote-agent-(windows|darwin|linux)-(amd64|arm64)(\.exe)?(\.sha256)?$`)

func (s *Server) installer(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("script")
	platform := strings.TrimPrefix(strings.TrimPrefix(name, "codex-"), "cli-")
	if name != platform && name != "codex-"+platform && name != "cli-"+platform {
		http.NotFound(w, r)
		return
	}
	if platform != "windows.ps1" && platform != "macos.sh" && platform != "linux.sh" {
		http.NotFound(w, r)
		return
	}
	token := r.URL.Query().Get("token")
	if name == platform && !tokenRE.MatchString(token) {
		fail(w, 400, "invalid enrollment token")
		return
	}
	if name != platform {
		token = ""
	}
	data, _ := installers.ReadFile("installers/" + platform)
	server := s.Config.PublicURL
	if strings.HasSuffix(name, ".ps1") {
		server = strings.ReplaceAll(server, "'", "''")
	} else {
		server = strings.ReplaceAll(server, "'", "'\"'\"'")
	}
	body := strings.ReplaceAll(strings.ReplaceAll(string(data), "__SERVER__", server), "__TOKEN__", token)
	mode := "agent"
	if strings.HasPrefix(name, "codex-") {
		mode = "codex"
	}
	if strings.HasPrefix(name, "cli-") {
		mode = "cli"
	}
	body = strings.ReplaceAll(body, "__MODE__", mode)
	body = strings.ReplaceAll(body, "__SETUP_CODEX__", fmt.Sprint(s.Config.GatewayURL != ""))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

var officialRelease struct {
	sync.Mutex
	at      time.Time
	release codexsetup.Release
}

// Includes time spent streaming to clients on slower links (the public server
// may have only a few Mbps of egress). Metadata has its own short timeout.
var releaseHTTP = &http.Client{Timeout: 25 * time.Minute}

func latestCodexRelease() (codexsetup.Release, error) {
	officialRelease.Lock()
	defer officialRelease.Unlock()
	if time.Since(officialRelease.at) < 5*time.Minute {
		return officialRelease.release, nil
	}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/openai/codex/releases/latest", nil)
	if err != nil {
		return codexsetup.Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "remote-agent-codex-installer")
	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return codexsetup.Release{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return codexsetup.Release{}, fmt.Errorf("official release metadata returned HTTP %d", res.StatusCode)
	}
	var release codexsetup.Release
	if err := json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(&release); err != nil {
		return release, err
	}
	valid := codexsetup.Release{}
	for _, asset := range release.Assets {
		for _, goos := range []string{"windows", "darwin", "linux"} {
			for _, arch := range []string{"amd64", "arm64"} {
				name, _ := codexsetup.AssetName(goos, arch)
				if asset.Name == name && strings.HasPrefix(asset.URL, "https://github.com/openai/codex/releases/download/") && strings.HasSuffix(asset.URL, "/"+name) && len(asset.Digest) == 71 && strings.HasPrefix(asset.Digest, "sha256:") {
					valid.Assets = append(valid.Assets, asset)
				}
			}
		}
	}
	if len(valid.Assets) == 0 {
		return valid, fmt.Errorf("no verified official Codex assets found")
	}
	officialRelease.release = valid
	officialRelease.at = time.Now()
	return valid, nil
}

func (s *Server) codexRelease(w http.ResponseWriter, r *http.Request) {
	release, err := latestCodexRelease()
	if err != nil {
		fail(w, 502, "official Codex download is temporarily unavailable")
		return
	}
	rewritten := codexsetup.Release{Assets: make([]codexsetup.Asset, 0, len(release.Assets))}
	for _, asset := range release.Assets {
		asset.URL = s.Config.PublicURL + "/downloads/codex/" + asset.Name
		rewritten.Assets = append(rewritten.Assets, asset)
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	reply(w, 200, rewritten)
}

func (s *Server) codexDownload(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("file")
	valid := false
	for _, goos := range []string{"windows", "darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			asset, _ := codexsetup.AssetName(goos, arch)
			valid = valid || name == asset
		}
	}
	if !valid {
		http.NotFound(w, r)
		return
	}
	release, err := latestCodexRelease()
	if err != nil {
		fail(w, 502, "official Codex download unavailable")
		return
	}
	for _, asset := range release.Assets {
		if asset.Name == name {
			req, err := http.NewRequestWithContext(r.Context(), "GET", asset.URL, nil)
			if err != nil {
				fail(w, 502, "invalid download")
				return
			}
			res, err := releaseHTTP.Do(req)
			if err != nil {
				fail(w, 502, "official Codex download failed")
				return
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				fail(w, 502, "official Codex archive unavailable")
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			_, _ = io.Copy(w, io.LimitReader(res.Body, 256<<20))
			return
		}
	}
	http.NotFound(w, r)
}
func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("file")
	if !artifactRE.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if strings.HasSuffix(name, ".sha256") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	http.ServeFile(w, r, filepath.Join(s.Config.Downloads, name))
}
