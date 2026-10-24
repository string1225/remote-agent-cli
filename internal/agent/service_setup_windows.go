package agent

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// Capture the original user's environment before elevation. Only a private
// profile path crosses UAC; device/model credentials never appear in argv.
func SetupWindowsService(config string, noStart bool) error {
	p, err := CaptureServiceProfile(config)
	if err != nil {
		return err
	}
	exe, err := installedServiceExecutable()
	if err != nil {
		return err
	}
	if windows.GetCurrentProcessToken().IsElevated() {
		return ConfigureWindowsService(p, noStart)
	}
	dir, err := os.MkdirTemp("", "remote-agent-service-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	profile := filepath.Join(dir, "profile.json")
	if err = p.Save(profile); err != nil {
		return err
	}
	args := []string{"service", "configure", "--profile", profile}
	if noStart {
		args = append(args, "--no-start")
	}
	script := serviceElevationScript(exe, args, filepath.Join(dir, "stdout.log"), filepath.Join(dir, "stderr.log"))
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodedPowerShell(script))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		logPath, detail := retainServiceSetupLog(config, dir, strings.TrimSpace(string(out)))
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1223 {
			return fmt.Errorf("administrator approval was cancelled; device credentials are retained; setup log: %s", logPath)
		}
		return fmt.Errorf("Windows Service configuration failed (%v); device credentials are retained; setup log: %s\n%s", err, logPath, detail)
	}
	_, _ = retainServiceSetupLog(config, dir, "Windows Service configured successfully")
	return nil
}

func serviceElevationScript(exe string, args []string, stdout, stderr string) string {
	// RunAs cannot redirect standard streams. Elevate PowerShell instead and
	// redirect the native child inside that elevated process.
	var native []string
	for _, arg := range args {
		native = append(native, windows.EscapeArg(arg))
	}
	inner := "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; try { $p=Start-Process -FilePath " + quotePowerShell(exe) + " -ArgumentList " + quotePowerShell(strings.Join(native, " ")) + " -WindowStyle Hidden -Wait -PassThru -RedirectStandardOutput " + quotePowerShell(stdout) + " -RedirectStandardError " + quotePowerShell(stderr) + "; exit $p.ExitCode } catch { [IO.File]::AppendAllText(" + quotePowerShell(stderr) + ", 'service configuration launcher: ' + $_.Exception.Message + [Environment]::NewLine); exit 1 }"
	nativeArgs := "-NoProfile -NonInteractive -EncodedCommand " + encodedPowerShell(inner)
	shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	return "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; try { $p=Start-Process -FilePath " + quotePowerShell(shell) + " -ArgumentList " + quotePowerShell(nativeArgs) + " -Verb RunAs -WindowStyle Hidden -Wait -PassThru; exit $p.ExitCode } catch { $failure=$_.Exception; while ($failure) { if ($failure.NativeErrorCode -eq 1223) { exit 1223 }; $failure=$failure.InnerException }; [Console]::Error.WriteLine($_.Exception.Message); exit 1 }"
}

func decodeUTF16LE(data []byte) string {
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return string(utf16.Decode(units))
}

func retainServiceSetupLog(config, dir, launcher string) (string, string) {
	var detail strings.Builder
	detail.WriteString(time.Now().Format(time.RFC3339) + " " + launcher + "\n")
	for _, name := range []string{"stdout.log", "stderr.log"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			// Windows PowerShell's native redirection emits UTF-16LE.
			if len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe {
				data = []byte(decodeUTF16LE(data[2:]))
			}
			detail.Write(data)
		}
	}
	path := filepath.Join(filepath.Dir(config), "service-setup.log")
	if err := os.WriteFile(path, []byte(detail.String()), 0600); err != nil {
		return "(unable to save setup log)", detail.String()
	}
	return path, detail.String()
}
