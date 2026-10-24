package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ClientAPI uses only the enrolled site's HTTPS endpoint. Device credentials are
// never handed to the target client or included in model tool output.
func ClientAPI(ctx context.Context, c Config, method, path string, body any) (json.RawMessage, error) {
	if path != "/api/agent/catalog" && path != "/api/agent/tools" && path != "/api/agent/channel" && path != "/api/agent/channels/sync" {
		return nil, errors.New("unsupported client endpoint")
	}
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(c.Server, "/")+path+"?id="+c.ID, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Credential)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 35 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("无法连接所属站点")
	}
	defer res.Body.Close()
	limit := 2 << 20
	if path == "/api/agent/catalog" {
		limit = 8 << 20
	}
	result, err := io.ReadAll(io.LimitReader(res.Body, int64(limit+1)))
	if err != nil || len(result) > limit {
		return nil, errors.New("站点响应超出限制")
	}
	if res.StatusCode != 200 {
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(result, &failure)
		if failure.Error != "" {
			return nil, errors.New(failure.Error)
		}
		return nil, fmt.Errorf("站点返回 HTTP %d", res.StatusCode)
	}
	if !json.Valid(result) {
		return nil, errors.New("站点返回无效数据")
	}
	return result, nil
}
func ClientTool(ctx context.Context, c Config, target, kind string, body any) (json.RawMessage, error) {
	return ClientAPI(ctx, c, "POST", "/api/agent/tools", map[string]any{"agentId": target, "type": kind, "body": body})
}
