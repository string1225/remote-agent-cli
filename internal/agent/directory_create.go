package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type directoryCreated struct {
	directoryOption
	Created bool `json:"created"`
}

// Accept a single child name and anchor the operation to an authorized root.
// os.Root also confines symlinks changed after the parent was validated.
func (a *Agent) createDirectory(ctx context.Context, parent, name string) (directoryCreated, error) {
	var result directoryCreated
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if len(parent) > 4096 || !utf8.ValidString(name) || len(name) > 255 || name == "" || name != strings.TrimSpace(name) || name == "." || name == ".." || strings.HasSuffix(name, ".") || strings.ContainsAny(name, `\/:*?"<>|`) || strings.ContainsFunc(name, unicode.IsControl) {
		return result, errors.New("请输入有效的文件夹名称，不能包含路径、特殊字符或末尾空格、句点")
	}
	stem := strings.ToUpper(strings.TrimRight(strings.SplitN(name, ".", 2)[0], " ."))
	if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || stem == "CONIN$" || stem == "CONOUT$" || len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9' {
		return result, errors.New("该名称是系统保留名称，请更换文件夹名称")
	}
	a.mu.Lock()
	config := a.Config
	config.AllowedRoots = append([]string{}, a.Config.AllowedRoots...)
	a.mu.Unlock()
	base, err := config.Workspace(parent)
	if err != nil {
		return result, errors.New("父目录不存在、无法访问或超出授权范围")
	}
	path := filepath.Join(base, name)
	if len(path) > 4096 {
		return result, errors.New("文件夹路径过长")
	}
	for _, allowed := range config.AllowedRoots {
		canonical, e := canonicalDir(allowed)
		if e != nil || !insideDirectories([]string{canonical}, base) {
			continue
		}
		root, e := os.OpenRoot(canonical)
		if e != nil {
			return result, errors.New("无法访问授权根目录")
		}
		defer root.Close()
		relative, e := filepath.Rel(canonical, base)
		if e != nil {
			return result, e
		}
		directory, e := root.OpenRoot(relative)
		if e != nil {
			return result, errors.New("无法访问父目录，请重新选择")
		}
		defer directory.Close()
		if err := ctx.Err(); err != nil {
			return result, err
		}
		err = directory.Mkdir(name, 0755)
		if err != nil {
			if !errors.Is(err, os.ErrExist) {
				return result, errors.New("无法新建文件夹，请检查父目录权限和名称")
			}
			info, e := directory.Lstat(name)
			if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return result, errors.New("同名文件或链接已存在，请更换名称")
			}
		}
		return directoryCreated{directoryOption: directoryOption{Name: name, Path: path}, Created: err == nil}, nil
	}
	return result, errors.New("父目录超出授权范围")
}
