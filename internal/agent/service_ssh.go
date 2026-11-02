package agent

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type ServiceSSHOptions struct {
	Alias, Host, User, Identity, KnownHosts string
	Port                                    int
}

var sshAliasPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
var sshUserPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]{0,63}$`)
var sshHostPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,252}$`)

func (o ServiceSSHOptions) Validate() error {
	if !sshAliasPattern.MatchString(o.Alias) || strings.HasSuffix(o.Alias, ".") || !filepath.IsLocal(o.Alias) {
		return errors.New("SSH 别名应为 1–64 个字母、数字、点、下划线或短横线")
	}
	if !sshHostPattern.MatchString(o.Host) && net.ParseIP(o.Host) == nil {
		return errors.New("SSH 主机应为 IP 地址或域名")
	}
	if !sshUserPattern.MatchString(o.User) {
		return errors.New("无效的 SSH 用户名")
	}
	if o.Port < 1 || o.Port > 65535 {
		return errors.New("SSH 端口应在 1–65535 之间")
	}
	for _, path := range []string{o.Identity, o.KnownHosts} {
		if !filepath.IsAbs(path) {
			return errors.New("SSH 密钥和 known_hosts 需要绝对路径")
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return errors.New("SSH 输入必须是已存在的普通文件，且不超过 1 MiB")
		}
	}
	key, err := os.ReadFile(o.Identity)
	if err != nil {
		return errors.New("无法读取指定的 SSH 私钥；请在文件所属用户的终端执行")
	}
	if _, err := ssh.ParsePrivateKey(key); err != nil {
		var encrypted *ssh.PassphraseMissingError
		if errors.As(err, &encrypted) {
			return errors.New("后台 SSH 暂不导入带口令的私钥；请使用单独的无口令部署密钥，原密钥保持不变")
		}
		return errors.New("指定文件不是支持的 SSH 私钥")
	}
	if _, err := knownhosts.New(o.KnownHosts); err != nil {
		return errors.New("known_hosts 格式无效或无法读取")
	}
	return nil
}

func serviceSSHConfig(o ServiceSSHOptions, generation string) string {
	quote := func(path string) string { return `"` + strings.ReplaceAll(filepath.ToSlash(path), `"`, `\"`) + `"` }
	return fmt.Sprintf("Host %s\n    HostName %s\n    User %s\n    Port %s\n    IdentityFile %s\n    UserKnownHostsFile %s\n    GlobalKnownHostsFile none\n    IdentitiesOnly yes\n    IdentityAgent none\n    BatchMode yes\n    StrictHostKeyChecking yes\n    ConnectTimeout 15\n", o.Alias, o.Host, o.User, strconv.Itoa(o.Port), quote(filepath.Join(generation, "key")), quote(filepath.Join(generation, "known_hosts")))
}
