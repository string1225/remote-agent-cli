package personalgateway

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed adminui/*
var adminFiles embed.FS

const adminPrefix = "/openai/admin"
const adminCookie = "sunny_personal_admin"

type AdminConfig struct {
	PasswordSHA256 string `json:"password_sha256"`
	SessionKey     string `json:"session_key"`
}
type loginWindow struct {
	Start time.Time
	Count int
}
type Admin struct {
	config   AdminConfig
	Origin   string
	mu       sync.Mutex
	attempts map[string]loginWindow
}

func CreateAdmin(path string) (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	password := base64.RawURLEncoding.EncodeToString(secret)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(password))
	config := AdminConfig{PasswordSHA256: hex.EncodeToString(hash[:]), SessionKey: base64.RawURLEncoding.EncodeToString(secret)}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(config); err != nil {
		return "", err
	}
	return password, file.Sync()
}

func LoadAdmin(path, origin string) (*Admin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config AdminConfig
	if json.Unmarshal(data, &config) != nil {
		return nil, errors.New("invalid admin configuration")
	}
	hash, err := hex.DecodeString(config.PasswordSHA256)
	if err != nil || len(hash) != 32 {
		return nil, errors.New("invalid admin password verifier")
	}
	key, err := base64.RawURLEncoding.DecodeString(config.SessionKey)
	if err != nil || len(key) != 32 {
		return nil, errors.New("invalid admin session key")
	}
	return &Admin{config: config, Origin: origin, attempts: map[string]loginWindow{}}, nil
}

func (admin *Admin) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(admin.config.SessionKey))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (admin *Admin) authorized(request *http.Request) bool {
	cookie, err := request.Cookie(adminCookie)
	if err != nil {
		return false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 3 {
		return false
	}
	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || expires < time.Now().Unix() || expires > time.Now().Add(8*time.Hour+time.Minute).Unix() {
		return false
	}
	return hmac.Equal([]byte(parts[2]), []byte(admin.sign(parts[0]+"."+parts[1])))
}

func (admin *Admin) limited(request *http.Request) bool {
	ip := request.Header.Get("X-Real-IP")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(request.RemoteAddr)
	}
	admin.mu.Lock()
	defer admin.mu.Unlock()
	now := time.Now()
	entry := admin.attempts[ip]
	if now.Sub(entry.Start) > 5*time.Minute {
		entry = loginWindow{Start: now}
	}
	if entry.Count >= 10 {
		return true
	}
	if len(admin.attempts) > 1024 {
		for key, item := range admin.attempts {
			if now.Sub(item.Start) > 5*time.Minute {
				delete(admin.attempts, key)
			}
		}
		if len(admin.attempts) > 1024 {
			return true
		}
	}
	entry.Count++
	admin.attempts[ip] = entry
	return false
}

func adminJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

func (admin *Admin) serve(server *Server, w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	path := strings.TrimPrefix(request.URL.Path, adminPrefix)
	if (request.Method == "GET" || request.Method == "HEAD") && (path == "" || path == "/" || path == "/admin.js" || path == "/admin.css") {
		file := "index.html"
		kind := "text/html; charset=utf-8"
		if path == "/admin.js" {
			file = "admin.js"
			kind = "application/javascript"
		}
		if path == "/admin.css" {
			file = "admin.css"
			kind = "text/css"
		}
		data, err := adminFiles.ReadFile("adminui/" + file)
		if err != nil {
			fail(w, 500, "管理页面不可用", "internal_error")
			return
		}
		w.Header().Set("Content-Type", kind)
		if request.Method == "HEAD" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Write(data)
		return
	}
	if request.Method == "POST" {
		if request.Header.Get("Origin") != admin.Origin || request.Header.Get("X-Sunny-Admin") != "1" {
			fail(w, 403, "请从本站管理页面操作。", "permission_error")
			return
		}
		if request.Header.Get("Content-Type") != "application/json" {
			fail(w, 400, "请求格式不正确。", "invalid_request_error")
			return
		}
	}
	if path == "/api/login" && request.Method == "POST" {
		if admin.limited(request) {
			w.Header().Set("Retry-After", "300")
			fail(w, 429, "登录尝试过多，请稍后再试。", "rate_limit_error")
			return
		}
		var body struct {
			Password string `json:"password"`
		}
		if json.NewDecoder(io.LimitReader(request.Body, 4096)).Decode(&body) != nil {
			fail(w, 400, "请输入管理密码。", "invalid_request_error")
			return
		}
		actual := sha256.Sum256([]byte(body.Password))
		expected, _ := hex.DecodeString(admin.config.PasswordSHA256)
		if subtle.ConstantTimeCompare(actual[:], expected) != 1 {
			fail(w, 401, "管理密码不正确。", "authentication_error")
			return
		}
		nonce := make([]byte, 16)
		if _, err := rand.Read(nonce); err != nil {
			fail(w, 500, "无法创建登录会话。", "internal_error")
			return
		}
		payload := strconv.FormatInt(time.Now().Add(8*time.Hour).Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(nonce)
		http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: payload + "." + admin.sign(payload), Path: adminPrefix, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60})
		adminJSON(w, map[string]bool{"ok": true})
		return
	}
	if !admin.authorized(request) {
		fail(w, 401, "请先输入管理密码登录。", "authentication_error")
		return
	}
	if path == "/api/logout" && request.Method == "POST" {
		http.SetCookie(w, &http.Cookie{Name: adminCookie, Path: adminPrefix, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		adminJSON(w, map[string]bool{"ok": true})
		return
	}
	if path == "/api/devices" && request.Method == "GET" {
		registry, err := LoadRegistry(server.RegistryPath)
		if err != nil {
			fail(w, 503, "设备列表暂时不可用。", "gateway_unavailable")
			return
		}
		devices := []map[string]any{}
		for _, device := range registry.Devices {
			devices = append(devices, map[string]any{"name": device.Name, "enabled": device.Enabled, "created_at": device.CreatedAt})
		}
		adminJSON(w, map[string]any{"devices": devices, "endpoint": "https://codex.sunny-string.cn/openai/personal/v1"})
		return
	}
	if path == "/api/status" && request.Method == "GET" {
		_, err := server.Credentials.Get(request.Context(), "")
		adminJSON(w, map[string]any{"logged_in": err == nil, "concurrency": cap(server.Slots)})
		return
	}
	if request.Method == "POST" && (path == "/api/devices" || path == "/api/devices/revoke" || path == "/api/devices/delete") {
		var body struct {
			Name string `json:"name"`
		}
		if json.NewDecoder(io.LimitReader(request.Body, 4096)).Decode(&body) != nil {
			fail(w, 400, "请输入设备名称。", "invalid_request_error")
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		if path == "/api/devices" {
			token, err := AddDevice(server.RegistryPath, body.Name)
			if err != nil {
				fail(w, 409, "设备名称已存在或列表正在更新，请换一个名称或稍后重试。", "conflict_error")
				return
			}
			adminJSON(w, map[string]string{"name": body.Name, "token": token})
			return
		}
		var err error
		if path == "/api/devices/revoke" {
			err = RevokeDevice(server.RegistryPath, body.Name)
		} else {
			err = DeleteDevice(server.RegistryPath, body.Name)
		}
		if err != nil {
			fail(w, 409, "操作未完成。删除前请先撤销 Token，或稍后重试。", "conflict_error")
			return
		}
		adminJSON(w, map[string]bool{"ok": true})
		return
	}
	fail(w, 404, "页面不存在。", "not_found_error")
}
