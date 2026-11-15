package fileshare

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const testR2 = "https://0123456789abcdef0123456789abcdef.r2.cloudflarestorage.com"

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(code int, value any) *http.Response {
	data, _ := json.Marshal(value)
	return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(data))}
}
func testFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func confirmedFile(name string, size int) any {
	return map[string]any{"success": true, "owner_token": "fixture-file-owner", "file": map[string]any{"id": "Abc123def", "url": "https://storage.to/Abc123def", "filename": name, "size": size, "expires_at": time.Now().Add(72 * time.Hour)}}
}

func TestSingleUploadKeepsProofLocalAndDeletesOnlyOwnedCopy(t *testing.T) {
	data := []byte("synthetic generated source only")
	path := testFile(t, "测试代码.txt", data)
	dir := filepath.Join(t.TempDir(), "receipts")
	var visitor string
	puts, deletes := 0, 0
	client := New(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "storage.to" {
			puts++
			body, _ := io.ReadAll(r.Body)
			if !bytes.Equal(body, data) || r.ContentLength != int64(len(data)) || r.Header.Get("Authorization") != "" || r.Header.Get("X-Owner-Token") != "" || r.Header.Get("X-Visitor-Token") != "" {
				t.Error("R2 body or identity isolation failed")
			}
			return response(200, map[string]any{}), nil
		}
		switch r.URL.Path {
		case "/api/upload/init":
			visitor = r.Header.Get("X-Visitor-Token")
			if !visitorPattern.MatchString(visitor) {
				t.Error("missing anonymous identity")
			}
			if r.Header.Get("Authorization") != "" {
				t.Error("platform credential sent")
			}
			return response(200, map[string]any{"success": true, "type": "single", "upload_url": testR2 + "/put?signed=fixture", "r2_key": "key", "headers": map[string][]string{"Host": {r.URL.Host}}}), nil
		case "/api/upload/confirm":
			return response(200, confirmedFile("测试代码.txt", len(data))), nil
		case "/api/file/Abc123def":
			deletes++
			body, _ := io.ReadAll(r.Body)
			if string(body) != "{}" || r.Header.Get("Content-Type") != "application/json" {
				t.Error("management request must have explicit JSON framing")
			}
			if r.Method != "DELETE" || r.Header.Get("Authorization") != "Owner fixture-file-owner" || r.Header.Get("X-Visitor-Token") != visitor {
				t.Error("delete ownership missing")
			}
			return response(200, map[string]any{"success": true}), nil
		}
		return nil, errors.New("unexpected test request")
	})})
	result, err := client.Upload(context.Background(), path, dir)
	if err != nil {
		t.Fatal(err)
	}
	sha := sha256.Sum256(data)
	if result.SHA256 != hex.EncodeToString(sha[:]) || result.Provider != "storage.to" || result.Class != "temporary" || result.LinkType != "page" || result.PreviewURL != "" || puts != 1 {
		t.Fatal("incorrect public metadata")
	}
	output, _ := json.Marshal(result)
	for _, secret := range []string{visitor, "fixture-file-owner", "signed=fixture"} {
		if bytes.Contains(output, []byte(secret)) {
			t.Fatal("proof leaked")
		}
	}
	if err = client.Delete(context.Background(), result.ID, dir); err != nil {
		t.Fatal(err)
	}
	if deletes != 1 {
		t.Fatal("delete not sent")
	}
	if _, err = os.Stat(filepath.Join(dir, result.ID+".json")); !os.IsNotExist(err) {
		t.Fatal("receipt not removed")
	}
	if after, _ := os.ReadFile(path); !bytes.Equal(after, data) {
		t.Fatal("original file changed")
	}
	if err = client.Delete(context.Background(), result.ID, dir); err == nil || deletes != 1 {
		t.Fatal("deleted without local proof")
	}
}

func TestMultipartOrdersPartsAndAbortsFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(strconv.FormatBool(fail), func(t *testing.T) {
			data := bytes.Repeat([]byte("a"), (2<<20)+19)
			path := testFile(t, "code.txt", data)
			var uploaded []byte
			aborted, confirmed, completed := 0, 0, 0
			client := New(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "storage.to" {
					n, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/part/"))
					if fail && n == 2 {
						return response(503, nil), nil
					}
					b, _ := io.ReadAll(r.Body)
					uploaded = append(uploaded, b...)
					if r.Header.Get("X-Owner-Token") != "" || r.Header.Get("X-Visitor-Token") != "" {
						t.Error("API proof sent to R2")
					}
					res := response(200, nil)
					res.Header.Set("ETag", fmt.Sprintf("\"part%d\"", n))
					return res, nil
				}
				switch r.URL.Path {
				case "/api/upload/init":
					return response(200, map[string]any{"success": true, "type": "multipart", "r2_key": "key", "upload_id": "upload", "owner_token": "fixture-upload-owner", "part_size": 1 << 20, "total_parts": 3, "initial_urls": map[string]string{"1": testR2 + "/part/1"}}), nil
				case "/api/upload/parts":
					var in struct {
						Numbers []int `json:"part_numbers"`
					}
					_ = json.NewDecoder(r.Body).Decode(&in)
					if r.Header.Get("Authorization") != "Owner fixture-upload-owner" {
						t.Error("missing upload proof")
					}
					n := in.Numbers[0]
					return response(200, map[string]any{"success": true, "part_urls": []any{map[string]any{"partNumber": n, "url": testR2 + "/part/" + strconv.Itoa(n)}}}), nil
				case "/api/upload/complete-multipart":
					completed++
					var in struct {
						Parts []struct {
							Number int    `json:"partNumber"`
							ETag   string `json:"etag"`
						} `json:"parts"`
					}
					_ = json.NewDecoder(r.Body).Decode(&in)
					if len(in.Parts) != 3 {
						t.Error("missing completion parts")
					} else {
						for i, p := range in.Parts {
							if p.Number != i+1 || p.ETag != fmt.Sprintf("\"part%d\"", i+1) {
								t.Error("part order or ETag lost")
							}
						}
					}
					return response(200, map[string]any{"success": true}), nil
				case "/api/upload/confirm":
					confirmed++
					return response(200, confirmedFile("code.txt", len(data))), nil
				case "/api/upload/abort":
					aborted++
					return response(200, map[string]any{"success": true}), nil
				}
				return nil, errors.New("unexpected test request")
			})})
			result, err := client.Upload(context.Background(), path, filepath.Join(t.TempDir(), "receipts"))
			if fail {
				if err == nil || aborted != 1 || confirmed != 0 || completed != 0 {
					t.Fatal("failed multipart published or not aborted")
				}
				return
			}
			sha := sha256.Sum256(data)
			if err != nil || !bytes.Equal(uploaded, data) || confirmed != 1 || completed != 1 || aborted != 0 || result.SHA256 != hex.EncodeToString(sha[:]) {
				t.Fatal("multipart lost bytes or lifecycle", err)
			}
		})
	}
}

func TestUnsafeURLsAndRedirectsNeverReceiveFileBytes(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1/put", testR2 + ".evil.test/put", "https://user:pass@" + strings.TrimPrefix(testR2, "https://") + "/put", testR2 + ":444/put", "https://evil.test/put"} {
		calls := 0
		client := New(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			calls++
			return response(200, map[string]any{"success": true, "type": "single", "r2_key": "key", "upload_url": raw}), nil
		})})
		if _, err := client.Upload(context.Background(), testFile(t, "code.txt", []byte("safe")), filepath.Join(t.TempDir(), "receipts")); err == nil || calls != 1 {
			t.Fatal("unsafe upload endpoint accepted")
		}
	}
	calls := 0
	client := New(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		res := response(307, map[string]any{"success": false, "error": "redirect"})
		res.Header.Set("Location", "https://evil.test")
		return res, nil
	})})
	if _, err := client.Upload(context.Background(), testFile(t, "code.txt", []byte("safe")), filepath.Join(t.TempDir(), "receipts")); err == nil || calls != 1 {
		t.Fatal("followed redirect")
	}
}

func TestPreflightBlocksObviousSecretsAndRefusedTypesBeforeNetwork(t *testing.T) {
	calls := 0
	client := New(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("network must not be called")
	})})
	for _, name := range []string{".env", ".env.production", "id_ed25519", "auth.json", "secret.pem", "program.exe", "app.apk", "app.ipa"} {
		if _, err := client.Upload(context.Background(), testFile(t, name, []byte("fixture only")), filepath.Join(t.TempDir(), "receipts")); err == nil {
			t.Error("unsafe temporary file accepted", name)
		}
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, _ := writer.Create("nested/.env")
	_, _ = file.Write([]byte("fixture only"))
	_ = writer.Close()
	if _, err := client.Upload(context.Background(), testFile(t, "code.zip", archive.Bytes()), filepath.Join(t.TempDir(), "receipts")); err == nil {
		t.Fatal("archive credential filename accepted")
	}
	if calls != 0 {
		t.Fatal("preflight sent bytes")
	}
}

func TestIdentityIsStableAcrossConcurrentUploadsAndRestart(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "receipts")
	if err := prepareReceipts(dir); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	identities := make(chan string, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := visitorIdentity(dir)
			if err != nil {
				t.Error(err)
			}
			identities <- value
		}()
	}
	wg.Wait()
	close(identities)
	var first string
	for value := range identities {
		if first == "" {
			first = value
		}
		if value == "" || value != first {
			t.Fatal("identity rotated")
		}
	}
	again, err := visitorIdentity(dir)
	if err != nil || again != first {
		t.Fatal("identity not restored")
	}
}

func TestQuotaErrorsAreBoundedAndDoNotLeakProof(t *testing.T) {
	client := New(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return response(429, map[string]any{"success": false, "error": "quota for " + r.Header.Get("X-Visitor-Token")}), nil
	})})
	dir := filepath.Join(t.TempDir(), "receipts")
	_, err := client.Upload(context.Background(), testFile(t, "code.txt", []byte("fixture")), dir)
	identity, _ := visitorIdentity(dir)
	if err == nil || !strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), identity) {
		t.Fatal("quota error missing or proof exposed")
	}
}
