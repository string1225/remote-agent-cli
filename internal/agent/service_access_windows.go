package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/string1225/remote-agent-cli/internal/serviceaccess"
	"golang.org/x/sys/windows"
)

func ServiceAccess(ctx context.Context, o serviceaccess.Options, output io.Writer) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if o.Action == "list" || windows.GetCurrentProcessToken().IsElevated() {
		err := serviceaccess.Execute(o, output)
		if o.Action != "list" || !errors.Is(err, os.ErrPermission) || windows.GetCurrentProcessToken().IsElevated() {
			return err
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "remote-agent-access-")
	if err != nil {
		return err
	}
	stdout, stderr := filepath.Join(dir, "stdout.log"), filepath.Join(dir, "stderr.log")
	defer func() { _ = os.Remove(stdout); _ = os.Remove(stderr); _ = os.Remove(dir) }()
	args := []string{"access", o.Action}
	if o.Action == "grant" {
		args = append(args, "--path", o.Path, "--permission", o.Permission)
		if o.Recursive {
			args = append(args, "--recursive")
		}
		if o.GitTrust {
			args = append(args, "--git-trust")
		}
	} else if o.Action == "restore" {
		args = append(args, "--backup", o.Backup)
		if o.Force {
			args = append(args, "--force")
		}
	}
	fmt.Fprintln(output, "正在请求管理员授权；仅修改所选路径的 RemoteAgent 服务 SID 权限，不重启服务。")
	script := serviceElevationScript(exe, args, stdout, stderr)
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
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
			return errors.New("已取消管理员授权，未修改路径权限或 Git 信任")
		}
		return fmt.Errorf("access 操作未完成；如输出已有备份 ID，请用 access list/restore 检查恢复：%v %s", runErr, strings.TrimSpace(string(launcher)))
	}
	return nil
}
