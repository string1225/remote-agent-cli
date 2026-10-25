package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	PublicURL         string
	Downloads         string
	AllowSignup       bool
	GatewayURL        string
	GatewayToken      string
	GatewayAdminToken string
	GatewayModel      string
}
type Server struct {
	Store         *store.Store
	Config        Config
	hub           *hub
	mu            sync.Mutex
	attempts      map[string]limit
	gatewayClient *http.Client
	gatewaySlots  chan struct{}
	gatewayMu     sync.Mutex
	gatewayCalls  map[string]map[*inferenceCall]struct{}
}
type limit struct {
	Count int
	Until time.Time
}

func New(db *store.Store, cfg Config, web fs.FS) (*Server, http.Handler, error) {
	u, err := protocol.ParseServerURL(cfg.PublicURL)
	if err != nil {
		return nil, nil, fmt.Errorf("PUBLIC_URL: %w", err)
	}
	cfg.PublicURL = u.String()
	if cfg.GatewayModel == "" {
		cfg.GatewayModel = "gpt-6.1-sol"
	}
	if err := validateGateway(cfg); err != nil {
		return nil, nil, err
	}
	s := &Server{Store: db, Config: cfg, hub: newHub(), attempts: map[string]limit{}, gatewayClient: newGatewayClient(), gatewaySlots: make(chan struct{}, maxConcurrentInference), gatewayCalls: map[string]map[*inferenceCall]struct{}{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]bool{"ok": true}) })
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]bool{"allowSignup": false})
	})
	mux.HandleFunc("POST /api/register", s.register)
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("GET /api/me", s.me)
	mux.HandleFunc("GET /api/account", s.me)
	mux.HandleFunc("PATCH /api/account", s.changeAccount)
	mux.HandleFunc("GET /api/admin/users", s.adminUsers)
	mux.HandleFunc("POST /api/admin/users", s.adminCreateUser)
	mux.HandleFunc("PATCH /api/admin/users/{id}", s.adminUpdateUser)
	mux.HandleFunc("GET /api/tokens", s.tokens)
	mux.HandleFunc("POST /api/tokens", s.createToken)
	mux.HandleFunc("POST /api/tokens/{id}/reveal", s.revealToken)
	mux.HandleFunc("DELETE /api/tokens/{id}", s.revokeToken)
	mux.HandleFunc("GET /api/admin/codex/accounts", s.codexAccounts)
	mux.HandleFunc("POST /api/admin/codex/{action}", s.codexAccounts)
	mux.HandleFunc("GET /api/admin/tokens", s.adminTokens)
	mux.HandleFunc("DELETE /api/admin/tokens/{id}", s.adminRevokeToken)
	mux.HandleFunc("GET /api/codex", s.codex)
	mux.HandleFunc("POST /api/agent/codex", s.agentCodex)
	mux.HandleFunc("GET /api/token/account", s.tokenAccount)
	mux.HandleFunc("GET /api/agents", s.agents)
	mux.HandleFunc("POST /api/agents", s.enroll)
	mux.HandleFunc("POST /api/agents/{id}/repair", s.repairAgent)
	mux.HandleFunc("POST /api/agents/{id}/rebind", s.rebindAgent)
	mux.HandleFunc("PATCH /api/agents/{id}", s.renameAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", s.revoke)
	mux.HandleFunc("POST /api/enroll", s.redeem)
	mux.HandleFunc("POST /api/agent/channel", s.agentChannel)
	mux.HandleFunc("POST /api/agent/channels/{sync}", s.agentChannel)
	mux.HandleFunc("GET /api/agent/catalog", s.agentCatalog)
	mux.HandleFunc("POST /api/agent/tools", s.agentToolCall)
	mux.HandleFunc("GET /api/agent/connect", s.agentSocket)
	mux.HandleFunc("GET /api/agents/{id}/connect", s.browserSocket)
	mux.HandleFunc("GET /install/{script}", s.installer)
	mux.HandleFunc("GET /install/codex-release", s.codexRelease)
	mux.HandleFunc("GET /downloads/codex/{file}", s.codexDownload)
	mux.HandleFunc("GET /downloads/{file}", s.download)
	mux.Handle("GET /", http.FileServerFS(web))
	var handler http.Handler = s.protect(mux)
	if u.Path != "" {
		mount := http.NewServeMux()
		mount.Handle(u.Path+"/", http.StripPrefix(u.Path, handler))
		mount.HandleFunc("GET "+u.Path, func(w http.ResponseWriter, r *http.Request) {
			target := u.Path + "/"
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusPermanentRedirect)
		})
		handler = mount
	}
	return s, handler, nil
}

func (s *Server) origin() string {
	u, _ := url.Parse(s.Config.PublicURL)
	return u.Scheme + "://" + u.Host
}

func (s *Server) cookiePath() string {
	u, _ := url.Parse(s.Config.PublicURL)
	return strings.TrimSuffix(u.Path, "/") + "/"
}
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	reply(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 128*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, 400, "invalid request body")
		return false
	}
	if d.Decode(&struct{}{}) != io.EOF {
		fail(w, 400, "one JSON object required")
		return false
	}
	return true
}
func (s *Server) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; media-src 'self' blob:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		if strings.HasPrefix(s.Config.PublicURL, "https:") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		if strings.HasPrefix(r.URL.Path, "/codex/v1/") {
			s.gateway(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/install/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/register" {
			s.register(w, r)
			return
		}
		if (r.Method != "GET" && r.Method != "HEAD") && r.Header.Get("Origin") != "" && r.Header.Get("Origin") != s.origin() {
			fail(w, 403, "origin denied")
			return
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			fail(w, 403, "cross-site request denied")
			return
		}
		if r.Method == "POST" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			fail(w, 415, "application/json required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) limited(r *http.Request) bool {
	key, _, _ := net.SplitHostPort(r.RemoteAddr)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.attempts {
		if now.After(v.Until) {
			delete(s.attempts, k)
		}
	}
	v := s.attempts[key]
	if v.Until.IsZero() {
		if len(s.attempts) >= 10000 {
			return true
		}
		v.Until = now.Add(time.Minute)
	}
	v.Count++
	s.attempts[key] = v
	return v.Count > 30
}
func (s *Server) session(r *http.Request) (store.Session, string, error) {
	c, err := r.Cookie("ra_session")
	if err != nil {
		return store.Session{}, "", store.ErrDenied
	}
	v, err := s.Store.Session(c.Value)
	return v, c.Value, err
}
func (s *Server) require(w http.ResponseWriter, r *http.Request) (store.Session, bool) {
	v, _, err := s.session(r)
	if err != nil {
		fail(w, 401, "please sign in")
		return v, false
	}
	return v, true
}
func (s *Server) cookie(w http.ResponseWriter, token string, age int) {
	http.SetCookie(w, &http.Cookie{Name: "ra_session", Value: token, Path: s.cookiePath(), MaxAge: age, HttpOnly: true, Secure: strings.HasPrefix(s.Config.PublicURL, "https:"), SameSite: http.SameSiteStrictMode})
}
func (s *Server) signIn(w http.ResponseWriter, u store.User) {
	token, err := s.Store.NewSessionForUser(u)
	if err != nil {
		fail(w, 500, "cannot create session")
		return
	}
	s.cookie(w, token, 30*24*3600)
	reply(w, 200, map[string]any{"id": u.ID, "username": u.Username, "admin": u.Admin})
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	fail(w, 403, "registration disabled")
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), 12)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if s.limited(r) {
		fail(w, 429, "too many attempts; retry in a minute")
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	u, err := s.Store.User(in.Username)
	hash := u.Password
	if err != nil {
		hash = dummyHash
	}
	check := bcrypt.CompareHashAndPassword(hash, []byte(in.Password))
	if err != nil || check != nil || u.Disabled {
		fail(w, 401, "incorrect username or password")
		return
	}
	s.signIn(w, u)
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_, token, err := s.session(r)
	if err == nil {
		if s.Store.Logout(token) != nil {
			fail(w, 500, "cannot end session")
			return
		}
		s.hub.closeSession(store.Hash(token))
	}
	s.cookie(w, "", -1)
	reply(w, 200, map[string]bool{"ok": true})
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if ok {
		u, err := s.Store.UserByID(v.UserID)
		if err != nil {
			fail(w, 401, "please sign in")
			return
		}
		reply(w, 200, map[string]any{"id": u.ID, "username": u.Username, "admin": u.Admin, "expiresAt": v.ExpiresAt})
	}
}
func (s *Server) agents(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	list, err := s.Store.Agents(v.UserID)
	if err != nil {
		fail(w, 500, "cannot list agents")
		return
	}
	out := []any{}
	for _, a := range list {
		out = append(out, struct {
			store.Agent
			Online bool `json:"online"`
		}{a, s.hub.online(a.ID)})
	}
	reply(w, 200, out)
}
func (s *Server) enroll(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) == 0 || len(in.Name) > 80 {
		fail(w, 400, "device name must be 1–80 bytes")
		return
	}
	list, err := s.Store.Agents(v.UserID)
	if err != nil {
		fail(w, 500, "cannot read agents")
		return
	}
	if len(list) >= 100 {
		fail(w, 409, "100 device limit; revoke unused devices first")
		return
	}
	a, token, err := s.Store.Enroll(v.UserID, in.Name)
	if err != nil {
		fail(w, 500, "cannot create enrollment")
		return
	}
	base := s.Config.PublicURL + "/install/"
	reply(w, 201, map[string]any{
		"agent": a, "expiresAt": time.Now().Add(15 * time.Minute).Unix(),
		"windows":         "'" + token + "' | remote-agent connect --server '" + strings.ReplaceAll(s.Config.PublicURL, "'", "''") + "' --token-stdin",
		"windowsMsi":      s.Config.PublicURL + "/downloads/remote-agent-windows-amd64.msi",
		"windowsArm64Msi": s.Config.PublicURL + "/downloads/remote-agent-windows-arm64.msi",
		"macos":           "curl -fsSL '" + strings.ReplaceAll(base+"macos.sh?token="+token, "'", "'\"'\"'") + "' | bash",
		"linux":           "curl -fsSL '" + strings.ReplaceAll(base+"linux.sh?token="+token, "'", "'\"'\"'") + "' | bash",
	})
}
func (s *Server) redeem(w http.ResponseWriter, r *http.Request) {
	if s.limited(r) {
		fail(w, 429, "too many attempts")
		return
	}
	var in struct {
		Token      string `json:"token"`
		OS         string `json:"os"`
		Arch       string `json:"arch"`
		Rebind     bool   `json:"rebind"`
		ExpectedID string `json:"expectedId"`
		Credential string `json:"credential"`
	}
	if !decode(w, r, &in) {
		return
	}
	if (in.OS != "windows" && in.OS != "darwin" && in.OS != "linux") || (in.Arch != "amd64" && in.Arch != "arm64") {
		fail(w, 400, "unsupported platform")
		return
	}
	var a store.Agent
	var credential string
	var err error
	if in.Rebind {
		a, err = s.Store.RedeemRebind(in.Token, in.ExpectedID, in.OS, in.Arch, in.Credential)
		credential = in.Credential
	} else {
		a, credential, err = s.Store.Redeem(in.Token, in.OS, in.Arch)
	}
	if err != nil {
		fail(w, 401, "invalid, used, or expired enrollment token")
		return
	}
	if in.Rebind {
		s.hub.closeAgent(a.ID)
	}
	reply(w, 200, map[string]any{"id": a.ID, "name": a.Name, "credential": credential, "services": a.Services})
}
func (s *Server) revoke(w http.ResponseWriter, r *http.Request) {
	v, ok := s.require(w, r)
	if !ok {
		return
	}
	if err := s.Store.Revoke(v.UserID, r.PathValue("id")); err != nil {
		fail(w, 404, "agent not found")
		return
	}
	s.hub.closeAgent(r.PathValue("id"))
	reply(w, 200, map[string]bool{"ok": true})
}
func (s *Server) authenticateAgent(r *http.Request) (store.Agent, error) {
	return s.Store.AuthenticateAgent(r.URL.Query().Get("id"), bearer(r))
}
func (s *Server) Close() {
	s.cancelInference("")
	s.gatewayClient.CloseIdleConnections()
	s.hub.closeAll()
}
