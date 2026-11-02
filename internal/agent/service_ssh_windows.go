package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/sys/windows"
)

// Native Windows OpenSSH rejects ordinary allow ACEs for extra service groups.
// A conditional ACE grants only a token containing the enabled RemoteAgent SID;
// Administrators owns the production key so another LocalService cannot change
// its ACL. No broad grant to the shared LocalService account is necessary.
func serviceSSHFileSDDL(serviceSID string) string {
	return "O:BAG:BAD:P(A;;FA;;;SY)(A;;FA;;;BA)(XA;;FR;;;" + serviceSID + ";(Member_of {SID(" + serviceSID + ")}))"
}

func ConfigureServiceSSH(ctx context.Context, o ServiceSSHOptions, output io.Writer) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if _, err := WindowsServiceStatus(); err != nil {
		return fmt.Errorf("请先安装并配置 Remote Agent Windows Service：%w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if !windows.GetCurrentProcessToken().IsElevated() {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		dir, err := os.MkdirTemp("", "remote-agent-ssh-")
		if err != nil {
			return err
		}
		stdout, stderr := filepath.Join(dir, "stdout.log"), filepath.Join(dir, "stderr.log")
		defer func() { _ = os.Remove(stdout); _ = os.Remove(stderr); _ = os.Remove(dir) }()
		args := []string{"ssh", "configure", "--alias", o.Alias, "--host", o.Host, "--user", o.User, "--port", strconv.Itoa(o.Port), "--identity", o.Identity, "--known-hosts", o.KnownHosts}
		script := serviceElevationScript(exe, args, stdout, stderr)
		cmd := exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		launcher, runErr := cmd.CombinedOutput()
		for _, path := range []string{stdout, stderr} {
			raw, _ := os.ReadFile(path)
			if len(raw) >= 2 && raw[0] == 0xff && raw[1] == 0xfe {
				raw = []byte(decodeUTF16LE(raw[2:]))
			}
			_, _ = output.Write(raw)
		}
		if runErr != nil {
			var exit *exec.ExitError
			if errors.As(runErr, &exit) && exit.ExitCode() == 1223 {
				return errors.New("已取消 SSH 管理员授权，原 SSH 文件保持不变")
			}
			return fmt.Errorf("SSH 配置未完成：%v %s", runErr, strings.TrimSpace(string(launcher)))
		}
		return nil
	}
	root, err := windows.KnownFolderPath(windows.FOLDERID_ProgramData, 0)
	if err != nil {
		return err
	}
	sid, err := gitServiceSID()
	if err != nil {
		return err
	}
	base := filepath.Join(root, "RemoteAgent", "ssh", o.Alias)
	generation := filepath.Join(base, "generations", store.Token())
	key, err := os.ReadFile(o.Identity)
	if err != nil {
		return errors.New("无法读取 SSH 私钥")
	}
	hosts, err := os.ReadFile(o.KnownHosts)
	if err != nil {
		return errors.New("无法读取 known_hosts")
	}
	keyHash, hostsHash := sha256.Sum256(key), sha256.Sum256(hosts)
	script := serviceSSHSetupScript(o, base, generation, sid, hex.EncodeToString(keyHash[:]), hex.EncodeToString(hostsHash[:]))
	cmd := exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if raw, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("SSH 配置失败；旧连接及原文件保留：%v %s", err, strings.TrimSpace(string(raw)))
	}
	_, err = fmt.Fprintf(output, "SSH 连接 %s 已配置。原 SSH 文件保留，无需重启服务。\n后台连接命令：ssh -F \"%s\" %s\n", o.Alias, filepath.Join(base, "config"), o.Alias)
	return err
}

func serviceSSHSetupScript(o ServiceSSHOptions, base, generation, sid, keyHash, hostsHash string) string {
	q := quotePowerShell
	directorySDDL := "O:BAG:BAD:P(A;;FA;;;SY)(A;;FA;;;BA)(XA;;FRFX;;;" + sid + ";(Member_of {SID(" + sid + ")}))"
	return "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue';\n" +
		"$base=" + q(base) + "; $generation=" + q(generation) + "; $directorySDDL=" + q(directorySDDL) + "; $fileSDDL=" + q(serviceSSHFileSDDL(sid)) + ";\n" + `
function Assert-RealPath([string]$Path) {
    $cursor=[IO.Path]::GetFullPath($Path)
    while ($cursor) {
        if ([IO.File]::Exists($cursor) -or [IO.Directory]::Exists($cursor)) {
            if (([IO.File]::GetAttributes($cursor) -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'SSH destination contains a reparse point' }
        }
        $parent=[IO.Path]::GetDirectoryName($cursor)
        if ($parent -eq $cursor) { break }
        $cursor=$parent
    }
}
function Protect-Directory([string]$Path) {
    Assert-RealPath $Path
    [IO.Directory]::CreateDirectory($Path) | Out-Null
    $acl=New-Object System.Security.AccessControl.DirectorySecurity
    $sections=[Security.AccessControl.AccessControlSections]::Access -bor [Security.AccessControl.AccessControlSections]::Owner -bor [Security.AccessControl.AccessControlSections]::Group
    $acl.SetSecurityDescriptorSddlForm($directorySDDL,$sections)
    [IO.Directory]::SetAccessControl($Path,$acl)
}
function Protect-File([string]$Path) {
    Assert-RealPath $Path
    $acl=New-Object System.Security.AccessControl.FileSecurity
    $sections=[Security.AccessControl.AccessControlSections]::Access -bor [Security.AccessControl.AccessControlSections]::Owner -bor [Security.AccessControl.AccessControlSections]::Group
    $acl.SetSecurityDescriptorSddlForm($fileSDDL,$sections)
    [IO.File]::SetAccessControl($Path,$acl)
}
$target=Join-Path $base 'config'
$backup=Join-Path $generation 'previous-config'
$published=$false
$hadPrevious=$false
try {
    Protect-Directory ([IO.Path]::GetDirectoryName($base))
    Protect-Directory $base
    $lock=[IO.File]::Open((Join-Path $base '.configure.lock'),[IO.FileMode]::OpenOrCreate,[IO.FileAccess]::ReadWrite,[IO.FileShare]::None)
    Protect-Directory (Join-Path $base 'generations')
    Protect-Directory $generation
` +
		"    $key=[IO.File]::ReadAllBytes(" + q(o.Identity) + "); $hosts=[IO.File]::ReadAllBytes(" + q(o.KnownHosts) + ");\n" +
		"    $sha=[Security.Cryptography.SHA256]::Create(); if ([BitConverter]::ToString($sha.ComputeHash($key)).Replace('-','').ToLowerInvariant() -ne " + q(keyHash) + " -or [BitConverter]::ToString($sha.ComputeHash($hosts)).Replace('-','').ToLowerInvariant() -ne " + q(hostsHash) + ") { throw 'SSH input changed; rerun configure' };\n" + `
    foreach ($name in @('key','known_hosts','config')) {
        $path=Join-Path $generation $name
        [IO.File]::Create($path).Close()
        Protect-File $path
    }
    [IO.File]::WriteAllBytes((Join-Path $generation 'key'),$key)
    [IO.File]::WriteAllBytes((Join-Path $generation 'known_hosts'),$hosts)
` +
		"    [IO.File]::WriteAllText((Join-Path $generation 'config')," + q(serviceSSHConfig(o, generation)) + ",(New-Object Text.UTF8Encoding($false)));\n" + `
    Assert-RealPath $target
    if ([IO.File]::Exists($target)) {
        $hadPrevious=$true
        Protect-File $target
        [IO.File]::Replace((Join-Path $generation 'config'),$target,$backup)
    } else {
        [IO.File]::Move((Join-Path $generation 'config'),$target)
    }
    $published=$true
    Protect-File $target
    exit 0
} catch {
    $failure=$_.Exception.Message
    if ($published) {
        try {
            Assert-RealPath $target
            Assert-RealPath $backup
            if ($hadPrevious) {
                [IO.File]::Replace($backup,$target,(Join-Path $generation 'failed-config'))
            } else {
                [IO.File]::Move($target,(Join-Path $generation 'failed-config'))
            }
        } catch { $failure+='; rollback failed; retained generations for recovery: '+$_.Exception.Message }
    }
    [Console]::Error.WriteLine($failure)
    exit 1
}
finally { if ($lock) { $lock.Dispose() } }
`
}
