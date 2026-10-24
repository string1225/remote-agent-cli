package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// Only explicit account-management operations cross the private loopback control
// boundary. Model/device Tokens never authorize this endpoint.
func (s *Server) codexAccounts(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if s.Config.GatewayURL == "" || s.Config.GatewayAdminToken == "" {
		if r.Method == "GET" {
			reply(w, 200, map[string]any{"enabled": false, "accounts": []any{}})
		} else {
			fail(w, 503, "服务器尚未启用 Codex 账号管理")
		}
		return
	}
	action := "accounts"
	var body []byte
	if r.Method == "POST" {
		action = r.PathValue("action")
		if action != "login" && action != "activate" && action != "cancel" {
			fail(w, 404, "unknown account operation")
			return
		}
		var in struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if !decode(w, r, &in) {
			return
		}
		if len(in.ID) > 100 || len([]rune(in.Name)) > 80 {
			fail(w, 400, "invalid account fields")
			return
		}
		body, _ = json.Marshal(in)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	up, err := http.NewRequestWithContext(ctx, r.Method, strings.TrimSuffix(s.Config.GatewayURL, "/")+"/internal/codex/"+action, bytes.NewReader(body))
	if err != nil {
		fail(w, 503, "账号管理不可用")
		return
	}
	up.Header.Set("Authorization", "Bearer "+s.Config.GatewayAdminToken)
	up.Header.Set("Content-Type", "application/json")
	res, err := s.gatewayClient.Do(up)
	if err != nil {
		fail(w, 503, "无法连接服务器 Codex 账号管理")
		return
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 128*1024+1))
	if err != nil || len(data) > 128*1024 {
		fail(w, 502, "invalid account management response")
		return
	}
	if res.StatusCode != 200 {
		fail(w, 409, "操作未完成：请检查是否有待完成的授权、账号是否存在及设备码登录权限")
		return
	}
	// Decode and encode rather than forwarding arbitrary upstream headers/content.
	type login struct {
		ID              string `json:"id"`
		Target          string `json:"target,omitempty"`
		Name            string `json:"name"`
		State           string `json:"state"`
		VerificationURL string `json:"verificationUrl,omitempty"`
		UserCode        string `json:"userCode,omitempty"`
		ExpiresAt       int64  `json:"expiresAt"`
		Error           string `json:"error,omitempty"`
	}
	var v any
	if action == "accounts" {
		v = &struct {
			Enabled  bool `json:"enabled"`
			Accounts []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Active    bool   `json:"active"`
				State     string `json:"state"`
				AccountID string `json:"accountId,omitempty"`
				ExpiresAt int64  `json:"expiresAt"`
			} `json:"accounts"`
			Login *login `json:"login"`
		}{}
	} else if action == "login" {
		v = &login{}
	} else {
		v = &struct {
			OK bool `json:"ok"`
		}{}
	}
	if json.Unmarshal(data, v) != nil {
		fail(w, 502, "invalid account management response")
		return
	}
	reply(w, 200, v)
}
