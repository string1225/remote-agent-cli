package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type directoryOption struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
type directoryPage struct {
	Directories []directoryOption `json:"directories"`
	Base        string            `json:"base"`
	Parent      string            `json:"parent"`
	Separator   string            `json:"separator"`
	Recursive   bool              `json:"recursive"`
	Truncated   bool              `json:"truncated"`
}

func directoryKey(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
func insideDirectories(roots []string, path string) bool {
	for _, root := range roots {
		rel, err := filepath.Rel(root, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return true
		}
	}
	return false
}

// Search only metadata within locally authorized roots. Budgets keep large home
// directories and dependency trees from monopolizing the control channel.
func (a *Agent) directories(parent context.Context, base, query string) (directoryPage, error) {
	a.mu.Lock()
	config := a.Config
	config.AllowedRoots = append([]string{}, a.Config.AllowedRoots...)
	config.Services = append(config.Services[:0:0], a.Config.Services...)
	a.mu.Unlock()
	page := directoryPage{Directories: []directoryOption{}, Separator: string(filepath.Separator)}
	if parent.Err() != nil {
		return page, parent.Err()
	}
	if len(base) > 4096 || len(query) > 4096 || strings.ContainsRune(base+query, 0) {
		return page, errors.New("目录搜索内容过长或无效")
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	roots := []string{}
	for _, root := range config.AllowedRoots {
		if path, err := canonicalDir(root); err == nil {
			roots = append(roots, path)
		}
	}
	if len(roots) == 0 {
		return page, errors.New("没有可访问的授权根目录，请在目标电脑检查 Agent 配置")
	}
	seen := map[string]bool{}
	bytes := 0
	add := func(path string) bool {
		if len(path) > 4096 {
			page.Truncated = true
			return true
		}
		key := directoryKey(path)
		if seen[key] {
			return true
		}
		item := directoryOption{Name: filepath.Base(path), Path: path}
		if item.Name == string(filepath.Separator) || item.Name == "." {
			item.Name = path
		}
		data, _ := json.Marshal(item)
		if len(page.Directories) >= 40 || bytes+len(data) > 22000 {
			page.Truncated = true
			return false
		}
		seen[key] = true
		bytes += len(data)
		page.Directories = append(page.Directories, item)
		return true
	}
	query = strings.TrimSpace(query)
	if runtime.GOOS == "windows" {
		query = strings.ReplaceAll(query, "/", `\`)
	}
	if query == "~" || strings.HasPrefix(query, "~"+page.Separator) {
		home, err := os.UserHomeDir()
		if err != nil {
			return page, err
		}
		if query == "~" {
			query = home
		} else {
			query = filepath.Join(home, strings.TrimPrefix(query, "~"+page.Separator))
		}
	}
	known := append([]string{}, roots...)
	for _, service := range config.Services {
		if path, err := canonicalDir(service.Workspace); err == nil && insideDirectories(roots, path) {
			known = append(known, path)
		}
	}
	if base == "" && query == "" {
		for _, path := range known {
			if !add(path) {
				break
			}
		}
		return page, nil
	}
	if base != "" {
		path, err := canonicalDir(base)
		if err != nil || !insideDirectories(roots, path) {
			return page, errors.New("该目录不存在、无法访问或超出授权范围")
		}
		base = path
	}
	pathQuery := filepath.IsAbs(query) || strings.Contains(query, page.Separator) || filepath.VolumeName(query) != ""
	if pathQuery {
		if !filepath.IsAbs(query) {
			if base == "" {
				base = roots[0]
			}
			query = filepath.Join(base, query)
		}
		// Complete partially typed authorized roots without enumerating their parents.
		for _, path := range known {
			if strings.HasPrefix(directoryKey(path), directoryKey(query)) {
				add(path)
			}
		}
		if path, err := canonicalDir(query); err == nil && insideDirectories(roots, path) {
			base = path
			query = ""
			add(path)
		} else {
			dir, leaf := filepath.Split(query)
			path, err := canonicalDir(dir)
			if err != nil || !insideDirectories(roots, path) {
				if len(page.Directories) > 0 {
					return page, nil
				}
				return page, errors.New("路径无法访问；可直接输入目录名搜索，或从授权根目录逐层浏览")
			}
			base = path
			query = leaf
		}
	}
	if base != "" {
		page.Base = base
		up := filepath.Dir(base)
		if up != base && insideDirectories(roots, up) {
			page.Parent = up
		}
	}
	page.Recursive = base == ""
	needle := strings.ToLower(query)
	if page.Recursive {
		for _, path := range known {
			if strings.Contains(strings.ToLower(filepath.Base(path)), needle) {
				add(path)
			}
		}
	}
	type pendingDirectory struct {
		path  string
		depth int
	}
	queue := []pendingDirectory{}
	if base != "" {
		queue = append(queue, pendingDirectory{base, 0})
	} else {
		for _, root := range roots {
			queue = append(queue, pendingDirectory{root, 0})
		}
	}
	visited := map[string]bool{}
	scanned := 0
	for len(queue) > 0 && len(page.Directories) < 40 {
		if ctx.Err() != nil {
			page.Truncated = true
			break
		}
		current := queue[0]
		queue = queue[1:]
		if visited[directoryKey(current.path)] {
			continue
		}
		visited[directoryKey(current.path)] = true
		// Recheck symlinks before opening each directory, including queued entries.
		path, err := canonicalDir(current.path)
		if err != nil || !insideDirectories(roots, path) {
			page.Truncated = true
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			if !page.Recursive {
				return page, errors.New("无法读取该目录，请检查目标电脑的目录权限")
			}
			page.Truncated = true
			continue
		}
		stop := false
		for !stop {
			if ctx.Err() != nil || scanned >= 10000 {
				page.Truncated = true
				stop = true
				break
			}
			entries, readErr := f.ReadDir(128)
			for _, entry := range entries {
				scanned++
				if ctx.Err() != nil || scanned > 10000 {
					page.Truncated = true
					stop = true
					break
				}
				if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
					continue
				}
				child, err := canonicalDir(filepath.Join(path, entry.Name()))
				if err != nil || !insideDirectories(roots, child) {
					continue
				}
				if strings.Contains(strings.ToLower(entry.Name()), needle) {
					if !add(child) {
						stop = true
						break
					}
				}
				if page.Recursive && !skipDirectorySearch(entry.Name()) {
					if current.depth < 4 {
						queue = append(queue, pendingDirectory{child, current.depth + 1})
					} else {
						page.Truncated = true
					}
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					page.Truncated = true
				}
				break
			}
		}
		_ = f.Close()
		if stop {
			break
		}
	}
	if len(queue) > 0 || len(page.Directories) >= 40 {
		page.Truncated = true
	}
	sort.SliceStable(page.Directories, func(i, j int) bool {
		left, right := strings.ToLower(page.Directories[i].Name), strings.ToLower(page.Directories[j].Name)
		lp, rp := strings.HasPrefix(left, needle), strings.HasPrefix(right, needle)
		if lp != rp {
			return lp
		}
		if left == right {
			return page.Directories[i].Path < page.Directories[j].Path
		}
		return left < right
	})
	if parent.Err() != nil {
		return page, parent.Err()
	}
	return page, nil
}
func skipDirectorySearch(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch strings.ToLower(name) {
	case "node_modules", "vendor", "dist", "build", "target", "__pycache__", "appdata", "application data", "library", "$recycle.bin", "system volume information":
		return true
	}
	return false
}
