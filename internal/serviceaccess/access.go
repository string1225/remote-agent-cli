// Package serviceaccess keeps Windows service ACL grants separate from Git trust
// and credentials. No credential material is read or stored here.
package serviceaccess

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	Action, Path, Permission, Backup string
	Recursive, GitTrust, Force       bool
}

var (
	backupID       = regexp.MustCompile(`^[a-f0-9]{32}$`)
	reservedDevice = regexp.MustCompile(`^(CON|PRN|AUX|NUL|CONIN\$|CONOUT\$|COM[1-9¹²³]|LPT[1-9¹²³])$`)
)

func (o Options) Validate() error {
	switch o.Action {
	case "grant":
		if o.Permission != "read" && o.Permission != "modify" {
			return errors.New("请明确指定 --permission read 或 modify")
		}
		if err := validatePath(o.Path); err != nil {
			return err
		}
		if o.Backup != "" || o.Force {
			return errors.New("grant 不接受 restore 参数")
		}
	case "restore":
		if !backupID.MatchString(o.Backup) {
			return errors.New("请用 --backup 指定 grant/list 输出的备份 ID")
		}
		if o.Path != "" || o.Permission != "" || o.Recursive || o.GitTrust {
			return errors.New("restore 不接受 grant 参数")
		}
	case "list":
		if o.Path != "" || o.Permission != "" || o.Backup != "" || o.Recursive || o.GitTrust || o.Force {
			return errors.New("list 不接受授权/恢复参数")
		}
	default:
		return errors.New("请使用 access grant/list/restore")
	}
	return nil
}

func validatePath(path string) error {
	// Restrict to drive-qualified local Windows paths. In particular, reject ADS,
	// device paths, Git's wildcard syntax, and control characters before UAC.
	p := strings.ReplaceAll(path, "/", `\`)
	if strings.IndexFunc(p, func(r rune) bool { return r < 32 }) >= 0 {
		return errors.New("路径不能包含控制字符")
	}
	if len(p) < 4 || p[1] != ':' || p[2] != '\\' || !((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) || strings.ContainsAny(p[2:], ":*?\"<>|\x00\r\n\t") {
		return errors.New("--path 必须是明确的本机绝对路径，例如 C:\\Work\\repo；不支持 UNC、设备路径、ADS 或通配符")
	}
	parts := strings.Split(strings.TrimRight(p[3:], `\`), `\`)
	if len(parts) == 0 || parts[0] == "" {
		return errors.New("不能授权整个驱动器，请选择具体目录或文件")
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part {
			return errors.New("请使用不含 .、..、重复分隔符或尾部空格/点的明确路径")
		}
		stem, _, _ := strings.Cut(strings.ToUpper(part), ".")
		if reservedDevice.MatchString(stem) {
			return errors.New("不能授权 Windows 设备名")
		}
	}
	return nil
}

func gitConfig(paths []string) (string, error) {
	var out strings.Builder
	out.WriteString("# Managed by remote-agent access. Credentials and user Git configuration are separate.\n[safe]\n")
	for _, path := range paths {
		if err := validatePath(path); err != nil {
			return "", err
		}
		p := strings.ReplaceAll(path, `\`, "/")
		p = strings.ReplaceAll(p, `"`, `\"`)
		fmt.Fprintf(&out, "\tdirectory = \"%s\"\n", p)
	}
	return out.String(), nil
}

// IncludeParameter points at a protected Git config read on every Git command,
// so grant/restore need no daemon restart once this version is installed.
func IncludeParameter() string {
	path := trustPath()
	if path == "" {
		return ""
	}
	return " 'include.path=" + strings.ReplaceAll(filepath.ToSlash(path), "'", "'\\''") + "'"
}
