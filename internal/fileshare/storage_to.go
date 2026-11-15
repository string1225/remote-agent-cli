// Package fileshare implements anonymous storage.to sharing. Platform/device
// credentials are intentionally absent from its interface.
package fileshare

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const apiBase = "https://storage.to/api"

var fileID = regexp.MustCompile(`^[A-Za-z0-9]{9}$`)
var r2Host = regexp.MustCompile(`^[a-f0-9]{32}\.r2\.cloudflarestorage\.com$`)

type Client struct{ http *http.Client }

func New(client *http.Client) *Client {
	if client == nil {
		client = &http.Client{}
	}
	copy := *client
	copy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{http: &copy}
}

type initialized struct {
	Success     bool                `json:"success"`
	Type        string              `json:"type"`
	URL         string              `json:"upload_url"`
	Key         string              `json:"r2_key"`
	Headers     map[string][]string `json:"headers"`
	UploadID    string              `json:"upload_id"`
	Owner       string              `json:"owner_token"`
	PartSize    int64               `json:"part_size"`
	TotalParts  int                 `json:"total_parts"`
	InitialURLs map[string]string   `json:"initial_urls"`
}

type confirmed struct {
	Success bool   `json:"success"`
	Owner   string `json:"owner_token"`
	File    struct {
		ID        string    `json:"id"`
		URL       string    `json:"url"`
		Name      string    `json:"filename"`
		Size      int64     `json:"size"`
		ExpiresAt time.Time `json:"expires_at"`
	} `json:"file"`
}

func (c *Client) api(ctx context.Context, method, path string, body any, visitor, owner string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	// A JSON body also gives DELETE an explicit content length. Some front
	// proxies otherwise leave a bodyless management request waiting for input.
	if method == "DELETE" && body == nil {
		body = struct{}{}
	}
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return errors.New("无法创建storage.to请求")
		}
		input = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, input)
	if err != nil {
		return errors.New("无法创建storage.to请求")
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if visitor != "" {
		req.Header.Set("X-Visitor-Token", visitor)
	}
	if owner != "" {
		req.Header.Set("Authorization", "Owner "+owner)
	}
	res, err := c.http.Do(req)
	if err != nil {
		reason := "连接失败"
		if errors.Is(err, context.DeadlineExceeded) {
			reason = "请求超时"
		} else if errors.Is(err, context.Canceled) {
			reason = "请求已取消"
		} else if errors.Is(err, io.EOF) {
			reason = "对端提前关闭连接"
		}
		return fmt.Errorf("storage.to %s%s（%T）；未切换到139", path, reason, err)
	}
	defer res.Body.Close()
	if method == "DELETE" && res.StatusCode == http.StatusNoContent {
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil || len(raw) > 65536 {
		return errors.New("storage.to响应中断或超过64 KiB")
	}
	if method == "DELETE" && res.StatusCode == 404 {
		return nil
	}
	var status struct {
		Success *bool  `json:"success"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(raw, &status) != nil {
		return fmt.Errorf("storage.to %s返回非JSON（HTTP %d，可能为安全验证页面）；未切换到139", path, res.StatusCode)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 || status.Success == nil || !*status.Success {
		message := strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, status.Error)
		for _, secret := range []string{visitor, owner} {
			if secret != "" {
				message = strings.ReplaceAll(message, secret, "[hidden]")
			}
		}
		if text := []rune(message); len(text) > 160 {
			message = string(text[:160])
		}
		if res.StatusCode == 429 {
			if seconds, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && seconds >= 0 && seconds <= 86400 {
				message += fmt.Sprintf("；建议%d秒后再试", seconds)
			}
		}
		return fmt.Errorf("storage.to %s失败（HTTP %d）：%s；未切换到139", path, res.StatusCode, message)
	}
	if out != nil && json.Unmarshal(raw, out) != nil {
		return errors.New("storage.to响应格式无效")
	}
	return nil
}

func validateUploadURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || !r2Host.MatchString(u.Hostname()) || u.User != nil || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
		return nil, errors.New("storage.to返回了非预期的R2上传地址，已拒绝上传")
	}
	return u, nil
}

func (c *Client) put(ctx context.Context, raw string, headers map[string][]string, reader io.Reader, size int64) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	target, err := validateUploadURL(raw)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", target.String(), reader)
	if err != nil {
		return "", errors.New("无法创建R2上传请求")
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")
	for name, values := range headers {
		if strings.EqualFold(name, "Host") {
			continue
		}
		if !strings.EqualFold(name, "Content-Type") && !strings.HasPrefix(strings.ToLower(name), "x-amz-") {
			return "", errors.New("R2上传包含非预期请求头，已拒绝")
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	res, err := c.http.Do(req)
	if err != nil {
		return "", errors.New("R2文件上传失败或已取消；原文件保留，未切换到139")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != 200 && res.StatusCode != 204 {
		return "", fmt.Errorf("R2文件上传失败（HTTP %d）；未切换到139", res.StatusCode)
	}
	return res.Header.Get("ETag"), nil
}

func (c *Client) Upload(ctx context.Context, path, receiptDir string) (protocol.SharedFile, error) {
	var result protocol.SharedFile
	if !filepath.IsAbs(path) {
		return result, errors.New("上传文件需要绝对路径")
	}
	f, err := os.Open(path)
	if err != nil {
		return result, errors.New("无法打开上传文件")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return result, errors.New("只能上传普通文件")
	}
	if info.Size() < 1 || info.Size() > protocol.MaxSharedFileBytes {
		return result, errors.New("临时通道接受1字节至100 MiB文件")
	}
	name := filepath.Base(path)
	if len(name) > 240 || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return result, errors.New("文件名过长或含控制字符")
	}
	if err = checkTemporaryFile(f, name, info.Size()); err != nil {
		return result, err
	}
	if err = prepareReceipts(receiptDir); err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	visitor, err := visitorIdentity(receiptDir)
	if err != nil {
		return result, err
	}
	var init initialized
	if err = c.api(ctx, "POST", "/upload/init", map[string]any{"filename": name, "content_type": "application/octet-stream", "size": info.Size()}, visitor, "", &init); err != nil {
		return result, err
	}
	complete := false
	defer func() {
		if !complete && init.Type == "multipart" && init.UploadID != "" {
			cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
			defer stop()
			_ = c.api(cleanup, "POST", "/upload/abort", map[string]string{"upload_id": init.UploadID}, visitor, init.Owner, nil)
		}
	}()
	if init.Key == "" {
		return result, errors.New("storage.to初始化响应缺少存储标识")
	}
	sha := sha256.New()
	crc := crc32.NewIEEE()
	hash := io.MultiWriter(sha, crc)
	if init.Type == "single" {
		_, err = c.put(ctx, init.URL, init.Headers, io.TeeReader(io.NewSectionReader(f, 0, info.Size()), hash), info.Size())
	} else if init.Type == "multipart" {
		err = c.multipart(ctx, f, info.Size(), init, visitor, hash)
	} else {
		err = errors.New("storage.to返回未知上传类型")
	}
	if err != nil {
		return result, err
	}
	after, err := f.Stat()
	if err != nil || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return result, errors.New("上传期间文件发生变化，请重新打包后重试")
	}
	var done confirmed
	err = c.api(ctx, "POST", "/upload/confirm", map[string]any{"filename": name, "size": info.Size(), "content_type": "application/octet-stream", "r2_key": init.Key, "crc32": crc.Sum32()}, visitor, init.Owner, &done)
	if err != nil {
		return result, err
	}
	complete = true
	if !fileID.MatchString(done.File.ID) || done.File.URL != "https://storage.to/"+done.File.ID || done.File.Size != info.Size() || done.File.Name != name || !done.File.ExpiresAt.After(time.Now()) || done.Owner == "" || len(done.Owner) > 4096 || strings.IndexFunc(done.Owner, unicode.IsControl) >= 0 {
		if fileID.MatchString(done.File.ID) && done.Owner != "" {
			cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
			_ = c.api(cleanup, "DELETE", "/file/"+done.File.ID, nil, visitor, done.Owner, nil)
			stop()
		}
		return result, errors.New("storage.to返回无效分享记录，未发布下载链接")
	}
	result = protocol.SharedFile{Provider: "storage.to", Class: "temporary", LinkType: "page", ID: done.File.ID, Name: name, Size: info.Size(), SHA256: hex.EncodeToString(sha.Sum(nil)), URL: done.File.URL, ExpiresAt: done.File.ExpiresAt}
	if err = saveReceipt(receiptDir, receipt{File: result, Visitor: visitor, Owner: done.Owner}); err != nil {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if e := c.api(cleanup, "DELETE", "/file/"+result.ID, nil, visitor, done.Owner, nil); e != nil {
			return protocol.SharedFile{}, fmt.Errorf("上传已完成但管理回执无法保存且清理失败；请检查%s（到期%s），原文件保留", result.URL, result.ExpiresAt.Format(time.RFC3339))
		}
		return protocol.SharedFile{}, errors.New("管理回执无法保存，已删除本次临时副本；原文件保留")
	}
	return result, nil
}

func (c *Client) multipart(ctx context.Context, f *os.File, size int64, init initialized, visitor string, hash io.Writer) error {
	if init.UploadID == "" || init.PartSize < 1<<20 || init.PartSize > protocol.MaxSharedFileBytes || init.TotalParts < 1 || init.TotalParts > 100 || int((size+init.PartSize-1)/init.PartSize) != init.TotalParts {
		return errors.New("storage.to分片参数无效")
	}
	type part struct {
		Number int    `json:"partNumber"`
		ETag   string `json:"etag"`
	}
	var parts []part
	for n := 1; n <= init.TotalParts; n++ {
		uploadURL := init.InitialURLs[strconv.Itoa(n)]
		if uploadURL == "" {
			var response struct {
				URLs []struct {
					Number int    `json:"partNumber"`
					URL    string `json:"url"`
				} `json:"part_urls"`
			}
			if err := c.api(ctx, "POST", "/upload/parts", map[string]any{"upload_id": init.UploadID, "part_numbers": []int{n}}, visitor, init.Owner, &response); err != nil {
				return err
			}
			for _, entry := range response.URLs {
				if entry.Number == n {
					uploadURL = entry.URL
				}
			}
		}
		offset := int64(n-1) * init.PartSize
		length := min(init.PartSize, size-offset)
		etag, err := c.put(ctx, uploadURL, nil, io.TeeReader(io.NewSectionReader(f, offset, length), hash), length)
		if err != nil {
			return err
		}
		if etag == "" || len(etag) > 256 {
			return errors.New("R2分片缺少有效ETag，未发布链接")
		}
		parts = append(parts, part{n, etag})
	}
	return c.api(ctx, "POST", "/upload/complete-multipart", map[string]any{"upload_id": init.UploadID, "parts": parts}, visitor, init.Owner, nil)
}
