package access

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Server struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	User        string `json:"user"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fingerprint"`
	HostKey     string `json:"hostKey"`
	PublicKey   string `json:"publicKey"`
	ConfigPath  string `json:"configPath"`
	CreatedAt   int64  `json:"createdAt"`
}

var serverName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
var serverUser = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,63}$`)

func (m *Manager) serverDir(name string) (string, error) {
	if !serverName.MatchString(name) || name == "." || name == ".." {
		return "", errors.New("服务器别名须为 1–64 位字母、数字、点、下划线或横线")
	}
	root := filepath.Join(m.root, "servers")
	if err := secureDir(root); err != nil {
		return "", err
	}
	dir := filepath.Join(root, name)
	if st, err := os.Lstat(dir); err == nil && (!st.IsDir() || st.Mode()&os.ModeSymlink != 0) {
		return "", errors.New("服务器目录必须为真实目录")
	}
	return dir, nil
}
func (m *Manager) ListServers() ([]Server, error) {
	m.sshMu.Lock()
	defer m.sshMu.Unlock()
	return m.listServers()
}
func (m *Manager) listServers() ([]Server, error) {
	root := filepath.Join(m.root, "servers")
	items, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []Server{}, nil
	}
	if err != nil {
		return nil, err
	}
	result := []Server{}
	for _, item := range items {
		if !item.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, item.Name(), "server.json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var s Server
		if json.Unmarshal(raw, &s) != nil || s.Name != item.Name() {
			return nil, errors.New("服务器登记损坏，原文件保留")
		}
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}
func readServer(dir string) (Server, error) {
	var s Server
	raw, err := os.ReadFile(filepath.Join(dir, "server.json"))
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(raw, &s)
	return s, err
}
func sshConnect(ctx context.Context, s Server, auth []ssh.AuthMethod, callback ssh.HostKeyCallback) (*ssh.Client, error) {
	dialer := net.Dialer{Timeout: 12 * time.Second}
	address := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, errors.New("无法连接 SSH 服务器，请检查 IP、端口和网络")
	}
	deadline := time.Now().Add(20 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	config := &ssh.ClientConfig{User: s.User, Auth: auth, HostKeyCallback: callback, Timeout: 12 * time.Second}
	if s.HostKey != "" {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(s.HostKey))
		if err != nil {
			conn.Close()
			return nil, errors.New("主机身份记录损坏")
		}
		config.HostKeyAlgorithms = []string{key.Type()}
	}
	c, ch, req, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		conn.Close()
		return nil, errors.New("SSH 认证失败或主机身份变化，请核对账号、密码和已保存的主机指纹")
	}
	_ = conn.SetDeadline(time.Time{})
	return ssh.NewClient(c, ch, req), nil
}
func (m *Manager) AddServer(ctx context.Context, r Request) (Server, error) {
	m.sshMu.Lock()
	defer m.sshMu.Unlock()
	if net.ParseIP(r.Host) == nil || !serverUser.MatchString(r.User) || len(r.Password) > 4096 {
		return Server{}, errors.New("请提供有效的服务器 IP、SSH 用户名和密码")
	}
	if r.Port == 0 {
		r.Port = 22
	}
	if r.Port < 1 || r.Port > 65535 {
		return Server{}, errors.New("SSH 端口须为 1–65535")
	}
	dir, err := m.serverDir(r.Name)
	if err != nil {
		return Server{}, err
	}
	s, err := readServer(dir)
	if err == nil {
		if s.Host != r.Host || s.User != r.User || s.Port != r.Port {
			return Server{}, errors.New("这个别名已用于其他服务器；请使用新别名，原连接保留")
		}
		return m.testServer(ctx, s)
	}
	if !os.IsNotExist(err) {
		return Server{}, errors.New("服务器记录无法读取，原文件保留")
	}
	if err = secureDir(dir); err != nil {
		return Server{}, err
	}
	s = Server{Name: r.Name, Host: r.Host, User: r.User, Port: r.Port, ConfigPath: filepath.Join(dir, "config"), CreatedAt: time.Now().UnixMilli()}
	if raw, e := os.ReadFile(filepath.Join(dir, "pending.json")); e == nil {
		var pending Server
		if json.Unmarshal(raw, &pending) != nil || pending.Host != s.Host || pending.User != s.User || pending.Port != s.Port {
			return Server{}, errors.New("存在其他目标的未完成配置，原密钥保留；请使用新别名")
		}
		s = pending
	}
	keyPath := filepath.Join(dir, "key")
	raw, err := os.ReadFile(keyPath)
	if os.IsNotExist(err) {
		_, key, e := ed25519.GenerateKey(rand.Reader)
		if e != nil {
			return Server{}, e
		}
		block, e := ssh.MarshalPrivateKey(key, "RemoteAgent "+s.Name)
		if e != nil {
			return Server{}, e
		}
		raw = pem.EncodeToMemory(block)
		if err = writeFile(keyPath, raw); err != nil {
			return Server{}, err
		}
	} else if err != nil {
		return Server{}, errors.New("无法读取待配置密钥")
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		return Server{}, errors.New("SSH 私钥损坏，原文件保留")
	}
	s.PublicKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))) + " remote-agent-" + s.Name
	callback := func(_ string, _ net.Addr, key ssh.PublicKey) error {
		line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
		if s.HostKey != "" && s.HostKey != line {
			return errors.New("host key changed")
		}
		s.HostKey = line
		s.Fingerprint = ssh.FingerprintSHA256(key)
		return nil
	}
	auth := []ssh.AuthMethod{ssh.PublicKeys(signer)}
	if r.Password != "" {
		auth = append(auth, ssh.Password(r.Password))
	}
	ctx, cancel := context.WithTimeout(ctx, 70*time.Second)
	defer cancel()
	client, err := sshConnect(ctx, s, auth, callback)
	r.Password = ""
	if err != nil {
		return Server{}, err
	}
	defer client.Close()
	stop := context.AfterFunc(ctx, func() { client.Close() })
	defer stop()
	// Save the exact key and pinned host before touching authorized_keys. A lost
	// response can then recover with the same identity without duplicating keys.
	if err = writeJSON(filepath.Join(dir, "pending.json"), s); err != nil {
		return Server{}, err
	}
	session, err := client.NewSession()
	if err != nil {
		return Server{}, errors.New("SSH 会话无法建立")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte("no-agent-forwarding,no-port-forwarding,no-X11-forwarding " + s.PublicKey))
	script := `set -eu
umask 077
test -n "$HOME"
dir="$HOME/.ssh"
test ! -L "$dir"
mkdir -p "$dir"
chmod 700 "$dir"
auth="$dir/authorized_keys"
test ! -L "$auth"
key=$(printf '%s' '` + encoded + `' | base64 -d)
if test -f "$auth" && grep -F -- "$key" "$auth" >/dev/null; then exit 0; fi
if test -f "$auth"; then cp -p "$auth" "$auth.remote-agent-backup-` + store.Token() + `"; fi
temp=$(mktemp "$dir/.remote-agent.XXXXXXXX")
trap 'rm -f "$temp"' EXIT HUP INT TERM
if test -f "$auth"; then cat "$auth" > "$temp"; fi
printf '\n%s\n' "$key" >> "$temp"
chmod 600 "$temp"
mv -f "$temp" "$auth"
`
	session.Stdin = strings.NewReader(script)
	err = session.Run("sh -s")
	session.Close()
	if err != nil {
		return Server{}, errors.New("公钥安装未确认；原授权文件及本机待配置密钥保留，可用相同别名重试。目标须支持 POSIX shell")
	}
	if _, err = m.testServer(ctx, s); err != nil {
		return Server{}, errors.New("公钥已保存但密钥登录验证未通过；保留原授权与待配置密钥，请检查服务器公钥认证设置后重试")
	}
	hostLine := knownhosts.Line([]string{net.JoinHostPort(s.Host, strconv.Itoa(s.Port))}, func() ssh.PublicKey { key, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(s.HostKey)); return key }()) + "\n"
	if err = writeFile(filepath.Join(dir, "known_hosts"), []byte(hostLine)); err != nil {
		return Server{}, err
	}
	config := fmt.Sprintf("Host %s\n  HostName %s\n  User %s\n  Port %d\n  IdentityFile \"%s\"\n  UserKnownHostsFile \"%s\"\n  StrictHostKeyChecking yes\n  IdentitiesOnly yes\n  IdentityAgent none\n  BatchMode yes\n", s.Name, s.Host, s.User, s.Port, filepath.ToSlash(keyPath), filepath.ToSlash(filepath.Join(dir, "known_hosts")))
	if err = writeFile(s.ConfigPath, []byte(config)); err != nil {
		return Server{}, err
	}
	if err = writeJSON(filepath.Join(dir, "server.json"), s); err != nil {
		return Server{}, err
	}
	return s, nil
}
func (m *Manager) testServer(ctx context.Context, s Server) (Server, error) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(s.ConfigPath), "key"))
	if err != nil {
		return Server{}, errors.New("服务器私钥不可读")
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		return Server{}, errors.New("服务器私钥无效")
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(s.HostKey))
	if err != nil {
		return Server{}, errors.New("主机身份记录无效")
	}
	client, err := sshConnect(ctx, s, []ssh.AuthMethod{ssh.PublicKeys(signer)}, ssh.FixedHostKey(key))
	if err != nil {
		return Server{}, err
	}
	defer client.Close()
	return s, nil
}
func (m *Manager) TestServer(ctx context.Context, name string) (Server, error) {
	m.sshMu.Lock()
	defer m.sshMu.Unlock()
	dir, err := m.serverDir(name)
	if err != nil {
		return Server{}, err
	}
	s, err := readServer(dir)
	if err != nil {
		return Server{}, errors.New("服务器未登记")
	}
	return m.testServer(ctx, s)
}
func (m *Manager) RemoveServer(name string) (bool, error) {
	m.sshMu.Lock()
	defer m.sshMu.Unlock()
	dir, err := m.serverDir(name)
	if err != nil {
		return false, err
	}
	retired := filepath.Join(m.root, "retired")
	if err = secureDir(retired); err != nil {
		return false, err
	}
	if err = os.Rename(dir, filepath.Join(retired, name+"-"+store.Token())); err != nil {
		return false, errors.New("移除失败，原配置保留")
	}
	return true, nil
}
