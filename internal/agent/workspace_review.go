package agent

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

const workspaceReviewLimit = 128 * 1024

var errReviewNotRepository = errors.New("当前目录未启用 Git")

type reviewOutput struct {
	bytes.Buffer
	truncated bool
}

func (w *reviewOutput) Write(data []byte) (int, error) {
	n := len(data)
	remaining := workspaceReviewLimit - w.Len()
	if len(data) > remaining {
		data, w.truncated = data[:remaining], true
	}
	_, _ = w.Buffer.Write(data)
	return n, nil
}

func reviewGit(ctx context.Context, workspace string, args ...string) (string, bool, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return "", false, errors.New("目标电脑未安装 Git")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	// Disable configured helpers and text converters; review must never run a
	// project's hooks or external diff program, refresh its index, or fetch.
	options := []string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "safe.directory=" + workspace, "-C", workspace}
	var env []string
	// GIT_* variables from an agent's parent process must not redirect review
	// into another repository, index, object store or executable.
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		if !strings.HasPrefix(strings.ToUpper(key), "GIT_") {
			env = append(env, variable)
		}
	}
	env = append(env, "GIT_TERMINAL_PROMPT=0", "GIT_CONFIG_NOSYSTEM=1", "GIT_LITERAL_PATHSPECS=1")
	if len(args) > 0 && (args[0] == "diff" || args[0] == "status") {
		// Git can invoke clean/process filters even during status and diff.
		// Enumerate driver names without invoking them, then disable every
		// executable filter for this read; repository configuration is untouched.
		probe := command(ctx, git, append(append([]string{}, options...), "config", "--null", "--name-only", "--get-regexp", `^filter\..*\.(clean|smudge|process|required)$`))
		probe.Env = env
		var filters, probeError reviewOutput
		probe.Stdout, probe.Stderr = &filters, &probeError
		if err := probe.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || probeError.Len() != 0 || ctx.Err() != nil {
				return "", false, errors.New("Git 无法安全读取工作区过滤器配置")
			}
		}
		if filters.truncated {
			return "", false, errors.New("Git 过滤器配置过多，无法预览")
		}
		seen := map[string]bool{}
		for _, key := range strings.Split(filters.String(), "\x00") {
			at := strings.LastIndexByte(key, '.')
			if at < len("filter.") || seen[key[:at]] {
				continue
			}
			driver := key[:at]
			seen[driver] = true
			for _, setting := range []string{"clean=", "smudge=", "process=", "required=false"} {
				options = append(options, "-c", driver+"."+setting)
			}
		}
	}
	cmd := command(ctx, git, append(options, args...))
	var out, stderr reviewOutput
	cmd.Stdout, cmd.Stderr, cmd.Env = &out, &stderr, env
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", false, errors.New("Git 检查超时，请稍后刷新")
		}
		if strings.Contains(stderr.String(), "not a git repository") {
			return "", false, errReviewNotRepository
		}
		return "", false, errors.New("Git 无法读取当前工作区，请在目标电脑检查仓库和目录权限")
	}
	return strings.ToValidUTF8(out.String(), "�"), out.truncated, nil
}

func workspaceChanges(ctx context.Context, workspace string) (protocol.WorkspaceChanges, error) {
	result := protocol.WorkspaceChanges{Files: []protocol.WorkspaceChange{}}
	if _, _, err := reviewGit(ctx, workspace, "rev-parse", "--show-toplevel"); err != nil {
		if !errors.Is(err, errReviewNotRepository) {
			return result, err
		}
		return result, nil
	}
	result.Git = true
	branch, _, _ := reviewGit(ctx, workspace, "symbolic-ref", "--quiet", "--short", "HEAD")
	result.Branch = strings.TrimSpace(branch)
	if result.Branch == "" {
		result.Branch = "detached HEAD"
	}
	status, truncated, err := reviewGit(ctx, workspace, "status", "--porcelain=v1", "-z", "--no-renames", "--untracked-files=all", "--ignore-submodules=all", "--", ".")
	if err != nil {
		return result, err
	}
	result.Truncated = truncated
	prefix, _, err := reviewGit(ctx, workspace, "rev-parse", "--show-prefix")
	if err != nil {
		return result, err
	}
	prefix = strings.TrimSuffix(prefix, "\n")
	rows := strings.Split(status, "\x00")
	for i, row := range rows {
		if len(row) < 4 || (truncated && i == len(rows)-1) {
			continue
		}
		if len(result.Files) >= 200 {
			result.Truncated = true
			break
		}
		path := row[3:]
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		result.Files = append(result.Files, protocol.WorkspaceChange{Path: strings.TrimPrefix(path, prefix), Index: row[:1], Work: row[1:2]})
	}
	return result, nil
}

func workspaceDiff(ctx context.Context, workspace, name string) (protocol.WorkspaceDiff, error) {
	result := protocol.WorkspaceDiff{Path: name}
	if len(name) == 0 || len(name) > 4096 || strings.ContainsAny(name, "\x00\\") || filepath.IsAbs(name) || strings.Contains(name, ":") {
		return result, errors.New("无效的工作区文件路径")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || !insideDirectories([]string{workspace}, filepath.Join(workspace, clean)) {
		return result, errors.New("文件路径超出当前项目")
	}
	changes, err := workspaceChanges(ctx, workspace)
	if err != nil {
		return result, err
	}
	var change *protocol.WorkspaceChange
	for i := range changes.Files {
		if changes.Files[i].Path == name {
			change = &changes.Files[i]
			break
		}
	}
	if change == nil {
		return result, errors.New("文件未在当前变更列表中，请刷新后重试")
	}
	if change.Index == "?" {
		// os.Root follows only links that remain inside the authorized workspace.
		root, err := os.OpenRoot(workspace)
		if err != nil {
			return result, err
		}
		defer root.Close()
		info, err := root.Lstat(clean)
		if err != nil || !info.Mode().IsRegular() {
			return result, errors.New("新增路径不是普通文件，无法预览")
		}
		file, err := root.Open(clean)
		if err != nil {
			return result, err
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, workspaceReviewLimit+1))
		if err != nil {
			return result, err
		}
		result.Truncated = len(data) > workspaceReviewLimit
		if result.Truncated {
			data = data[:workspaceReviewLimit]
			for trim := 0; trim < 3 && !utf8.Valid(data) && len(data) > 0; trim++ {
				data = data[:len(data)-1]
			}
		}
		result.Binary = bytes.ContainsRune(data, 0) || !utf8.Valid(data)
		if !result.Binary {
			result.Untracked = string(data)
		}
		return result, nil
	}
	args := []string{"diff", "--no-ext-diff", "--no-textconv", "--no-color", "--no-renames", "--ignore-submodules=all"}
	result.Unstaged, result.Truncated, err = reviewGit(ctx, workspace, append(args, "--", name)...)
	if err != nil {
		return result, err
	}
	var truncated bool
	stagedArgs := append(append([]string{}, args...), "--cached", "--", name)
	result.Staged, truncated, err = reviewGit(ctx, workspace, stagedArgs...)
	result.Truncated = result.Truncated || truncated
	return result, err
}
