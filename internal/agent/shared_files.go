package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// UploadFile streams a regular file to the enrolled site and rejects redirects
// so device credentials can never be forwarded to another host.
func UploadFile(ctx context.Context, c Config, path string) (protocol.SharedFile, error) {
	var result protocol.SharedFile
	if err := ValidateServer(c.Server); err != nil {
		return result, err
	}
	if !filepath.IsAbs(path) {
		return result, errors.New("上传文件需要绝对路径")
	}
	file, err := os.Open(path)
	if err != nil {
		return result, errors.New("无法打开上传文件")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return result, errors.New("只能上传普通文件")
	}
	if info.Size() > protocol.MaxSharedFileBytes {
		return result, errors.New("上传文件不能超过 100 MiB")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimSuffix(c.Server, "/")+"/api/agent/files?id="+url.QueryEscape(c.ID), io.LimitReader(file, info.Size()))
	if err != nil {
		return result, errors.New("无法创建上传请求")
	}
	req.ContentLength = info.Size()
	req.Header.Set("Authorization", "Bearer "+c.Credential)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-File-Name", url.QueryEscape(filepath.Base(path)))
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return result, errors.New("文件上传失败；请检查站点连接后重试")
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil || len(raw) > 65536 {
		return result, errors.New("上传响应超出限制或连接中断")
	}
	if res.StatusCode != http.StatusOK {
		var failure struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &failure) == nil && failure.Error != "" {
			return result, fmt.Errorf("上传失败：%s", failure.Error)
		}
		return result, fmt.Errorf("上传失败：HTTP %d", res.StatusCode)
	}
	if json.Unmarshal(raw, &result) != nil || result.URL != strings.TrimSuffix(c.Server, "/")+"/files/"+result.ID || result.ID == "" || result.Size != info.Size() || !result.ExpiresAt.After(time.Now()) {
		return protocol.SharedFile{}, errors.New("站点返回无效的下载链接")
	}
	return result, nil
}
