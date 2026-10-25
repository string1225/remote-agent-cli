// Package clientupdate implements signed, credential-free runtime releases.
package clientupdate

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const HostVersion = 1
const WorkerProtocol = 1
const DataVersion = 1
const MaxSize = 128 << 20
const DefaultServer = "https://codex.sunny-string.cn/agents"

var digestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
var versionRE = regexp.MustCompile(`^[A-Za-z0-9._-]{1,80}$`)

type Payload struct {
	SchemaVersion          int    `json:"schemaVersion"`
	Sequence               uint64 `json:"sequence"`
	Version                string `json:"version"`
	OS                     string `json:"os"`
	Arch                   string `json:"arch"`
	MinHostVersion         int    `json:"minHostVersion"`
	WorkerProtocol         int    `json:"workerProtocol"`
	DataVersion            int    `json:"dataVersion"`
	MinRollbackDataVersion int    `json:"minRollbackDataVersion"`
	Size                   int64  `json:"size"`
	SHA256                 string `json:"sha256"`
}

type Release struct {
	Payload
	KeyID     string `json:"keyId"`
	Signature string `json:"signature"`
}

func (p Payload) Artifact() string { return "remote-agent-core-" + p.OS + "-" + p.Arch + ".exe" }

func (r Release) Verify(keys map[string]ed25519.PublicKey, platform, arch string) error {
	p := r.Payload
	if p.SchemaVersion != 1 || p.Sequence == 0 || !versionRE.MatchString(p.Version) || p.OS != platform || p.Arch != arch || p.OS != "windows" || (p.Arch != "amd64" && p.Arch != "arm64") || !digestRE.MatchString(p.SHA256) || p.Size < 1 || p.Size > MaxSize || p.MinHostVersion < 1 || p.WorkerProtocol < 1 || p.DataVersion != DataVersion || p.MinRollbackDataVersion != DataVersion {
		return errors.New("invalid or incompatible runtime release")
	}
	key, ok := keys[r.KeyID]
	sig, err := base64.StdEncoding.DecodeString(r.Signature)
	data, _ := json.Marshal(p)
	// Include key ID in the signed domain; manifests cannot replace trust anchors.
	if !ok || len(key) != ed25519.PublicKeySize || err != nil || !ed25519.Verify(key, append([]byte("remote-agent-update/v1\x00"+r.KeyID+"\x00"), data...), sig) {
		return errors.New("runtime release signature denied")
	}
	return nil
}

func Sign(p Payload, id string, key ed25519.PrivateKey) Release {
	data, _ := json.Marshal(p)
	return Release{Payload: p, KeyID: id, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, append([]byte("remote-agent-update/v1\x00"+id+"\x00"), data...)))}
}

func Compatible(r Release) error {
	if r.MinHostVersion > HostVersion || r.WorkerProtocol != WorkerProtocol {
		return fmt.Errorf("this release requires a newer host MSI (minimum host %d, worker protocol %d); routine runtime updates do not require MSI", r.MinHostVersion, r.WorkerProtocol)
	}
	return nil
}

func HTTPClient() *http.Client {
	return &http.Client{Timeout: 3 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func get(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cache-Control", "no-cache")
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("cannot reach runtime release server")
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("runtime download returned HTTP %d", res.StatusCode)
	}
	return res, nil
}

func Fetch(ctx context.Context, client *http.Client, server, platform, arch string, keys map[string]ed25519.PublicKey) (Release, error) {
	var release Release
	u, err := protocol.ParseServerURL(server)
	if err != nil {
		return release, err
	}
	if platform != "windows" || (arch != "amd64" && arch != "arm64") {
		return release, errors.New("unsupported runtime platform")
	}
	res, err := get(ctx, client, u.String()+"/updates/v1/"+platform+"/"+arch+".json")
	if err != nil {
		return release, err
	}
	defer res.Body.Close()
	d := json.NewDecoder(io.LimitReader(res.Body, 16385))
	d.DisallowUnknownFields()
	if err := d.Decode(&release); err != nil {
		return release, errors.New("invalid runtime manifest")
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return release, errors.New("invalid trailing runtime manifest data")
	}
	if err := release.Verify(keys, platform, arch); err != nil {
		return release, err
	}
	return release, Compatible(release)
}

func Download(ctx context.Context, client *http.Client, server string, release Release, output io.Writer) error {
	u, err := protocol.ParseServerURL(server)
	if err != nil {
		return err
	}
	res, err := get(ctx, client, u.String()+"/downloads/"+release.Artifact())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.ContentLength > release.Size {
		return errors.New("runtime download size mismatch")
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(output, h), io.LimitReader(res.Body, release.Size+1))
	if err != nil {
		return err
	}
	if n != release.Size || hex.EncodeToString(h.Sum(nil)) != release.SHA256 {
		return errors.New("runtime download checksum or size mismatch")
	}
	return nil
}

func NormalizeServer(server string) (string, error) {
	u, err := protocol.ParseServerURL(strings.TrimSuffix(server, "/"))
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
