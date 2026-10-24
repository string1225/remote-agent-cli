package personalgateway

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type managedAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Home      string `json:"home"`
	AccountID string `json:"accountId,omitempty"`
}
type accountManifest struct {
	Active   string           `json:"active"`
	Accounts []managedAccount `json:"accounts"`
}
type accountLogin struct {
	ID              string `json:"id"`
	Target          string `json:"target,omitempty"`
	Name            string `json:"name"`
	State           string `json:"state"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
	ExpiresAt       int64  `json:"expiresAt"`
	Error           string `json:"error,omitempty"`
	cancel          context.CancelFunc
}
type AccountManager struct {
	mu                         sync.Mutex
	Root, Binary, Proxy, Token string
	manifest                   accountManifest
	sources                    map[string]*CodexSource
	login                      *accountLogin
	launch                     func(context.Context, string) (deviceChallenge, func() (Credential, error), error)
}

func accountID() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
func NewAccountManager(root, binary, proxy, legacyHome, token string) (*AccountManager, error) {
	if len(token) < 32 || strings.ContainsAny(token, "\r\n") || !filepath.IsAbs(root) || !filepath.IsAbs(legacyHome) {
		return nil, errors.New("invalid private account manager configuration")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	m := &AccountManager{Root: root, Binary: binary, Proxy: proxy, Token: token, sources: map[string]*CodexSource{}}
	data, err := os.ReadFile(filepath.Join(root, "accounts.json"))
	if errors.Is(err, os.ErrNotExist) {
		m.manifest = accountManifest{Active: "legacy", Accounts: []managedAccount{{ID: "legacy", Name: "139 原有 Codex 账号", Home: legacyHome}}}
		if err = m.save(m.manifest); err != nil {
			return nil, err
		}
	} else if err != nil || json.Unmarshal(data, &m.manifest) != nil {
		return nil, errors.New("cannot read account registry; original data retained")
	}
	if len(m.manifest.Accounts) == 0 || len(m.manifest.Accounts) > 20 {
		return nil, errors.New("invalid account registry")
	}
	seen := map[string]bool{}
	for _, a := range m.manifest.Accounts {
		if a.ID == "" || seen[a.ID] || !filepath.IsAbs(a.Home) {
			return nil, errors.New("invalid account registry entry")
		}
		seen[a.ID] = true
	}
	if !seen[m.manifest.Active] {
		return nil, errors.New("active account missing")
	}
	m.launch = m.startLogin
	return m, nil
}
func (m *AccountManager) save(v accountManifest) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(m.Root, ".accounts-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	e := f.Close()
	if err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(m.Root, "accounts.json"))
}
func (m *AccountManager) source(home string) *CodexSource {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sources[home] == nil {
		m.sources[home] = &CodexSource{Binary: m.Binary, Proxy: m.Proxy, Home: home}
	}
	return m.sources[home]
}
func (m *AccountManager) ActiveSource() *CodexSource {
	m.mu.Lock()
	home := ""
	for _, a := range m.manifest.Accounts {
		if a.ID == m.manifest.Active {
			home = a.Home
			break
		}
	}
	m.mu.Unlock()
	return m.source(home)
}
func (m *AccountManager) Get(ctx context.Context, previous string) (Credential, error) {
	return m.ActiveSource().Get(ctx, previous)
}
func (m *AccountManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.login != nil && m.login.cancel != nil {
		m.login.cancel()
	}
}
func (m *AccountManager) status(ctx context.Context) map[string]any {
	m.mu.Lock()
	accounts := append([]managedAccount{}, m.manifest.Accounts...)
	active := m.manifest.Active
	var login *accountLogin
	if m.login != nil {
		v := *m.login
		login = &v
	}
	m.mu.Unlock()
	rows := make([]map[string]any, len(accounts))
	var wg sync.WaitGroup
	slots := make(chan struct{}, 4)
	for i, a := range accounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-ctx.Done():
				rows[i] = map[string]any{"id": a.ID, "name": a.Name, "state": "unknown", "active": a.ID == active}
				return
			}
			check, cancel := context.WithTimeout(ctx, 8*time.Second)
			defer cancel()
			credential, err := m.source(a.Home).Get(check, "")
			state := "ready"
			if err != nil {
				state = "login_required"
			}
			rows[i] = map[string]any{"id": a.ID, "name": a.Name, "active": a.ID == active, "state": state, "accountId": credential.AccountID, "expiresAt": credential.Expires.Unix()}
		}()
	}
	wg.Wait()
	return map[string]any{"accounts": rows, "login": login, "enabled": true}
}
func (m *AccountManager) begin(target, name string) (*accountLogin, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 80 {
		return nil, errors.New("请输入 1–80 字的账号名称")
	}
	m.mu.Lock()
	if m.login != nil && m.login.State == "pending" {
		m.mu.Unlock()
		return nil, errors.New("已有授权正在进行，请完成或取消后重试")
	}
	var old managedAccount
	for _, a := range m.manifest.Accounts {
		if a.ID == target {
			old = a
		}
	}
	if target != "" && old.ID == "" || target == "" && len(m.manifest.Accounts) >= 20 {
		m.mu.Unlock()
		return nil, errors.New("账号不存在或超过 20 个账号上限")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	login := &accountLogin{ID: accountID(), Target: target, Name: name, State: "pending", ExpiresAt: time.Now().Add(15 * time.Minute).Unix(), cancel: cancel}
	m.login = login
	m.mu.Unlock()
	home := filepath.Join(m.Root, "credentials", login.ID)
	if err := os.MkdirAll(home, 0700); err != nil {
		cancel()
		m.finish(login, managedAccount{}, err)
		return nil, errors.New("无法建立独立凭据目录")
	}
	// Read the previous identity without changing it; a failed or wrong-account
	// authorization never overwrites an existing working credential directory.
	expected := old.AccountID
	if expected == "" && old.ID != "" {
		check, stop := context.WithTimeout(ctx, 8*time.Second)
		c, err := m.source(old.Home).Get(check, "")
		stop()
		if err == nil {
			expected = c.AccountID
		}
	}
	v, wait, err := m.launch(ctx, home)
	if err != nil {
		cancel()
		m.finish(login, managedAccount{}, err)
		return nil, errors.New("无法开始设备授权；请检查 Codex 设备码登录与服务器网络")
	}
	m.mu.Lock()
	login.VerificationURL = v.VerificationURL
	login.UserCode = v.UserCode
	response := *login
	m.mu.Unlock()
	go func() {
		defer cancel()
		credential, err := wait()
		if err == nil && expected != "" && credential.AccountID != expected {
			err = errors.New("授权的账号与原账号不一致，请使用添加账号")
		}
		id := target
		if id == "" {
			id = accountID()
		}
		m.finish(login, managedAccount{ID: id, Name: name, Home: home, AccountID: credential.AccountID}, err)
	}()
	return &response, nil
}
func (m *AccountManager) finish(login *accountLogin, a managedAccount, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.login != login || login.State != "pending" {
		return
	}
	login.VerificationURL = ""
	login.UserCode = ""
	if err == nil {
		v := m.manifest
		v.Accounts = append([]managedAccount{}, v.Accounts...)
		found := false
		for i, old := range v.Accounts {
			if old.ID == a.ID {
				v.Accounts[i] = a
				found = true
			}
		}
		if !found {
			v.Accounts = append(v.Accounts, a)
		}
		if err = m.save(v); err == nil {
			m.manifest = v
			login.State = "completed"
			return
		}
	}
	login.State = "failed"
	login.Error = "授权未完成、已过期或账号不匹配；原账号和凭据保持可用。"
}
func (m *AccountManager) activate(ctx context.Context, id string) error {
	m.mu.Lock()
	var account managedAccount
	for _, a := range m.manifest.Accounts {
		if a.ID == id {
			account = a
		}
	}
	m.mu.Unlock()
	if account.ID == "" {
		return errors.New("账号不存在")
	}
	if _, err := m.source(account.Home).Get(ctx, ""); err != nil {
		return errors.New("账号需要重新授权，未切换当前账号")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.manifest
	v.Active = id
	if err := m.save(v); err != nil {
		return errors.New("无法保存账号选择")
	}
	m.manifest = v
	return nil
}
func (m *AccountManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Header.Get("Origin") != "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+m.Token)) != 1 {
		http.Error(w, "control access denied", 401)
		return
	}
	if r.URL.Path == "/internal/codex/accounts" && r.Method == "GET" {
		adminJSON(w, m.status(r.Context()))
		return
	}
	var in struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if r.Method != "POST" || json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	switch r.URL.Path {
	case "/internal/codex/login":
		v, err := m.begin(in.ID, in.Name)
		if err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		adminJSON(w, v)
	case "/internal/codex/activate":
		if err := m.activate(r.Context(), in.ID); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		adminJSON(w, map[string]bool{"ok": true})
	case "/internal/codex/cancel":
		m.mu.Lock()
		if m.login != nil && m.login.State == "pending" && m.login.ID == in.ID {
			m.login.cancel()
			m.login.State = "cancelled"
			m.login.UserCode = ""
			m.login.VerificationURL = ""
		}
		m.mu.Unlock()
		adminJSON(w, map[string]bool{"ok": true})
	default:
		http.NotFound(w, r)
	}
}
