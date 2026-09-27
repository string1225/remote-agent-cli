package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

type saysoStore struct {
	mu      sync.Mutex
	a       *Agent
	dir     string
	state   soState
	uploads map[string]*soUpload
	fatal   error
}

func soNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }
func (a *Agent) saysoStore() (*saysoStore, error) {
	a.saysoOnce.Do(func() {
		path := a.Config.Path
		if path == "" {
			path = DefaultPath()
		}
		s := &saysoStore{a: a, dir: filepath.Join(filepath.Dir(path), "sayso"), uploads: map[string]*soUpload{}}
		a.saysoErr = s.load()
		a.sayso = s
	})
	return a.sayso, a.saysoErr
}
func (s *saysoStore) load() error {
	if err := os.MkdirAll(filepath.Join(s.dir, "recordings"), 0700); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(s.dir, "state.json"))
	if errors.Is(err, os.ErrNotExist) {
		provider := "local"
		if s.a.Providers["codex"].Available {
			provider = "codex"
		}
		s.state = soState{Projects: []*soProject{}, Sessions: []*soSession{}, Actions: []*soAction{}, Runs: []*soRun{}, Settings: soSettings{provider, "default", 2, "codex", "default"}}
		seen := map[string]bool{}
		for _, service := range s.a.services() {
			if seen[service.Workspace] {
				continue
			}
			seen[service.Workspace] = true
			s.state.Projects = append(s.state.Projects, &soProject{store.Token(), service.Name, service.Workspace, "#557461", soNow()})
		}
	} else if err != nil {
		return err
	} else if len(data) > 24<<20 {
		return errors.New("SaySo 数据超过 24 MiB 上限")
	} else if err = json.Unmarshal(data, &s.state); err != nil {
		return fmt.Errorf("SaySo 数据无法读取，原文件已保留: %w", err)
	}
	for _, session := range s.state.Sessions {
		session.Status = "paused"
		session.AnalysisNextAt = ""
		if session.AnalysisStatus == "running" || session.AnalysisStatus == "queued" {
			session.AnalysisStatus = "failed"
			session.AnalysisError = "Agent 重启，分析已中断"
		}
		for _, trace := range session.AnalysisTraces {
			if trace.Status == "running" {
				trace.Status = "failed"
				trace.Error = "Agent 重启，分析已中断"
				trace.FinishedAt = soNow()
			}
		}
	}
	for _, run := range s.state.Runs {
		if run.Status == "running" {
			run.Status = "failed"
			run.Error = "Agent 重启，任务已中断"
			run.FinishedAt = soNow()
		}
	}
	for _, action := range s.state.Actions {
		if action.Status == "running" {
			action.Status = "failed"
			action.UpdatedAt = soNow()
		}
	}
	// Only this directory's generated, unfinished uploads are disposable after restart.
	files, _ := filepath.Glob(filepath.Join(s.dir, "recordings", ".upload-*"))
	for _, path := range files {
		_ = os.Remove(path)
	}
	return s.save()
}
func (s *saysoStore) save() error {
	if s.fatal != nil {
		return s.fatal
	}
	data, err := json.Marshal(s.state)
	if err == nil && len(data) > 24<<20 {
		err = errors.New("SaySo 数据超过 24 MiB 上限，请在目标电脑归档历史后重启 Agent")
	}
	if err == nil {
		var f *os.File
		f, err = os.CreateTemp(s.dir, ".state-*")
		if err == nil {
			name := f.Name()
			defer os.Remove(name)
			if err = f.Chmod(0600); err == nil {
				_, err = f.Write(data)
			}
			if err == nil {
				err = f.Sync()
			}
			closeErr := f.Close()
			if err == nil {
				err = closeErr
			}
			if err == nil {
				err = os.Rename(name, filepath.Join(s.dir, "state.json"))
			}
		}
	}
	if err != nil {
		s.fatal = fmt.Errorf("SaySo 保存失败，请检查目标电脑存储后重启 Agent: %w", err)
	}
	return s.fatal
}

func (a *Agent) handleSayso(p *peer, req protocol.Request, send func(protocol.Event) error) {
	s, err := a.saysoStore()
	var data []byte
	if err == nil {
		data, err = s.call(p, req.Method, req.Path, req.Body)
	}
	if err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()})
		return
	}
	if len(data) <= 24000 {
		_ = send(protocol.Event{ID: req.ID, Type: "result", Data: json.RawMessage(data)})
		return
	}
	// Pull requests stay small; large snapshots are streamed in bounded SCTP frames.
	for len(data) > 0 {
		n := min(16000, len(data))
		if send(protocol.Event{ID: req.ID, Type: "chunk", Data: base64.StdEncoding.EncodeToString(data[:n])}) != nil {
			return
		}
		data = data[n:]
	}
	_ = send(protocol.Event{ID: req.ID, Type: "result", Data: map[string]bool{"chunked": true}})
}

var soModelPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:/-]{0,99}$`)

func soString(body map[string]json.RawMessage, key string, max int, required bool) (string, error) {
	var value string
	if raw, ok := body[key]; ok {
		if json.Unmarshal(raw, &value) != nil {
			return "", fmt.Errorf("%s 必须是文字", key)
		}
	}
	value = strings.TrimSpace(value)
	if len(value) > max || (required && value == "") {
		return "", fmt.Errorf("%s 长度不合法", key)
	}
	return value, nil
}
func (s *saysoStore) session(id string) *soSession {
	for _, v := range s.state.Sessions {
		if v.ID == id {
			return v
		}
	}
	return nil
}
func (s *saysoStore) project(id string) *soProject {
	for _, v := range s.state.Projects {
		if v.ID == id {
			return v
		}
	}
	return nil
}
func (s *saysoStore) action(id string) *soAction {
	for _, v := range s.state.Actions {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func (s *saysoStore) call(p *peer, method, path string, raw json.RawMessage) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ctx.Err() != nil {
		return nil, p.ctx.Err()
	}
	if s.fatal != nil {
		return nil, s.fatal
	}
	if len(raw) > 48<<10 {
		return nil, errors.New("请求过大")
	}
	body := map[string]json.RawMessage{}
	if len(raw) > 0 && json.Unmarshal(raw, &body) != nil {
		return nil, errors.New("无效 JSON")
	}
	if !strings.HasPrefix(path, "/api/") || strings.Contains(path, "..") || strings.ContainsAny(path, "?%\\") {
		return nil, errors.New("不支持的 SaySo 路径")
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/"), "/")
	var value any
	var err error
	switch {
	case path == "/api/state" && method == "GET":
		value = s.state
	case path == "/api/settings" && method == "PATCH":
		settings := s.state.Settings
		for key, target := range map[string]*string{"analysisProvider": &settings.AnalysisProvider, "analysisModel": &settings.AnalysisModel, "executionModel": &settings.ExecutionModel, "executor": &settings.Executor} {
			if _, ok := body[key]; ok {
				v, e := soString(body, key, 100, true)
				if e != nil {
					return nil, e
				}
				*target = v
			}
		}
		if settings.AnalysisProvider != "local" && settings.AnalysisProvider != "codex" {
			return nil, errors.New("不支持的分析方式")
		}
		if settings.Executor != "codex" || !soModelPattern.MatchString(settings.AnalysisModel) || !soModelPattern.MatchString(settings.ExecutionModel) {
			return nil, errors.New("无效模型或执行器")
		}
		if settings.AnalysisProvider == "codex" && !s.a.Providers["codex"].Available {
			return nil, errors.New("目标电脑未安装 Codex，可选择本地规则")
		}
		settings.AnalysisIntervalMinutes = 2
		s.state.Settings = settings
		value = settings
	case parts[0] == "projects" && ((len(parts) == 1 && method == "POST") || (len(parts) == 2 && method == "PATCH")):
		project := &soProject{ID: store.Token(), Color: "#557461", CreatedAt: soNow()}
		if method == "PATCH" {
			old := s.project(parts[1])
			if old == nil {
				return nil, errors.New("项目不存在")
			}
			*project = *old
		}
		if method == "POST" || body["name"] != nil {
			project.Name, err = soString(body, "name", 200, true)
			if err != nil {
				return nil, err
			}
		}
		if method == "POST" || body["rootPath"] != nil {
			root, e := soString(body, "rootPath", 4000, true)
			if e != nil {
				return nil, e
			}
			project.RootPath, e = s.a.Config.Workspace(root)
			if e != nil {
				return nil, e
			}
		}
		if method == "POST" {
			if len(s.state.Projects) >= 100 {
				return nil, errors.New("最多 100 个项目")
			}
			s.state.Projects = append(s.state.Projects, project)
		} else {
			for _, a := range s.state.Actions {
				if a.ProjectID == project.ID && a.Status == "running" {
					return nil, errors.New("任务执行中，暂不能修改项目")
				}
			}
			*s.project(project.ID) = *project
		}
		value = project
	case path == "/api/sessions" && method == "POST":
		id, e := soString(body, "projectId", 100, true)
		if e != nil {
			return nil, e
		}
		if s.project(id) == nil {
			return nil, errors.New("项目不存在")
		}
		if len(s.state.Sessions) >= 500 {
			return nil, errors.New("最多 500 个讨论会话")
		}
		title, e := soString(body, "title", 200, false)
		if e != nil {
			return nil, e
		}
		if title == "" {
			title = time.Now().Format("01月02日 15:04")
		}
		v := &soSession{ID: store.Token(), ProjectID: id, Title: title, Status: "paused", StartedAt: soNow(), Recordings: []soRecording{}, Transcript: []soUtterance{}, AnalysisStatus: "idle", AnalysisTraces: []*soTrace{}}
		s.state.Sessions = append([]*soSession{v}, s.state.Sessions...)
		value = v
	case parts[0] == "sessions" && len(parts) >= 2 && len(parts) <= 3:
		v := s.session(parts[1])
		if v == nil {
			return nil, errors.New("会话不存在")
		}
		switch {
		case len(parts) == 2 && method == "PATCH":
			title := v.Title
			status := v.Status
			if body["title"] != nil {
				title, err = soString(body, "title", 200, true)
				if err != nil {
					return nil, err
				}
			}
			if body["status"] != nil {
				status, err = soString(body, "status", 20, true)
				if err != nil {
					return nil, err
				}
				if status != "paused" && status != "recording" {
					return nil, errors.New("无效录音状态")
				}
			}
			if body["status"] != nil && v.owner != nil && v.owner != p {
				return nil, errors.New("该会话正在另一浏览器录音")
			}
			v.Title = title
			if status == "recording" && v.Status != "recording" {
				s.startRecording(p, v)
			}
			if status == "paused" && v.Status == "recording" {
				s.stopRecording(v)
				s.analyze(p, v, false)
			}
			value = v
		case len(parts) == 2 && method == "DELETE":
			if v.Status == "recording" || v.AnalysisStatus == "running" {
				return nil, errors.New("请先暂停录音并等待分析完成")
			}
			for _, a := range s.state.Actions {
				if a.SessionID == v.ID && a.Status == "running" {
					return nil, errors.New("请等待执行完成")
				}
			}
			s.state.Sessions = slices.DeleteFunc(s.state.Sessions, func(x *soSession) bool { return x.ID == v.ID })
			ids := map[string]bool{}
			for _, a := range s.state.Actions {
				if a.SessionID == v.ID {
					ids[a.ID] = true
				}
			}
			s.state.Actions = slices.DeleteFunc(s.state.Actions, func(x *soAction) bool { return x.SessionID == v.ID })
			s.state.Runs = slices.DeleteFunc(s.state.Runs, func(x *soRun) bool { return ids[x.ActionID] })
			// Retain audio files on disk for local recovery; only indexed files can be fetched.
			value = true
		case len(parts) == 3 && parts[2] == "utterances" && method == "POST":
			text, e := soString(body, "text", 16000, true)
			if e != nil {
				return nil, e
			}
			if len(v.Transcript) >= 10000 {
				return nil, errors.New("讨论已达 10000 条，请新建会话")
			}
			if v.owner != nil && v.owner != p {
				return nil, errors.New("该会话正在另一浏览器录音")
			}
			var at int64
			_ = json.Unmarshal(body["atMs"], &at)
			u := soUtterance{store.Token(), text, max(0, at), soNow()}
			v.Transcript = append(v.Transcript, u)
			value = map[string]any{"utterance": u}
			if v.Status != "recording" {
				s.analyze(p, v, false)
			}
		case len(parts) == 3 && parts[2] == "extract-actions" && method == "POST":
			s.analyze(p, v, true)
			value = map[string]string{"status": v.AnalysisStatus}
		case len(parts) == 3 && parts[2] == "actions" && method == "POST":
			a := &soAction{ID: store.Token(), SessionID: v.ID, ProjectID: v.ProjectID, Status: "draft", Priority: "medium", Source: "手动添加", CreatedAt: soNow(), UpdatedAt: soNow()}
			if err = s.editAction(a, body, true); err != nil {
				return nil, err
			}
			if len(s.state.Actions) >= 3000 {
				return nil, errors.New("最多 3000 条 Action")
			}
			s.state.Actions = append([]*soAction{a}, s.state.Actions...)
			value = a
		default:
			return nil, errors.New("不支持的讨论操作")
		}
	case parts[0] == "actions" && len(parts) >= 2 && len(parts) <= 3:
		a := s.action(parts[1])
		if a == nil {
			return nil, errors.New("Action 不存在")
		}
		if a.Status == "running" {
			return nil, errors.New("Codex 正在执行，暂不能重复操作")
		}
		switch {
		case len(parts) == 2 && method == "PATCH":
			copy := *a
			if err = s.editAction(&copy, body, false); err != nil {
				return nil, err
			}
			*a = copy
			value = a
		case len(parts) == 2 && method == "DELETE":
			s.state.Actions = slices.DeleteFunc(s.state.Actions, func(x *soAction) bool { return x.ID == a.ID })
			s.state.Runs = slices.DeleteFunc(s.state.Runs, func(x *soRun) bool { return x.ActionID == a.ID })
			value = true
		case len(parts) == 3 && parts[2] == "dispatch" && method == "POST":
			value, err = s.dispatch(p, a)
		default:
			return nil, errors.New("不支持的 Action 操作")
		}
	case parts[0] == "runs" && len(parts) == 3 && parts[2] == "cancel" && method == "POST":
		found := false
		for _, run := range s.state.Runs {
			if run.ID == parts[1] {
				found = true
				if run.cancel != nil {
					run.cancel()
				}
				break
			}
		}
		if !found {
			return nil, errors.New("执行记录不存在")
		}
		value = true
	case parts[0] == "audio":
		value, err = s.audio(p, method, parts, body)
	default:
		return nil, errors.New("不支持的 SaySo 操作")
	}
	if err != nil {
		return nil, err
	}
	if method != "GET" && parts[0] != "audio" {
		if err = s.save(); err != nil {
			return nil, err
		}
	}
	return json.Marshal(value)
}
func (s *saysoStore) editAction(a *soAction, b map[string]json.RawMessage, create bool) error {
	for key, target := range map[string]*string{"title": &a.Title, "detail": &a.Detail, "priority": &a.Priority} {
		if b[key] != nil || (create && key == "title") {
			limit := 24000
			if key == "title" {
				limit = 300
			}
			v, e := soString(b, key, limit, key == "title")
			if e != nil {
				return e
			}
			*target = v
		}
	}
	if !slices.Contains([]string{"high", "medium", "low"}, a.Priority) {
		return errors.New("无效优先级")
	}
	if b["status"] != nil {
		status, e := soString(b, "status", 20, true)
		if e != nil {
			return e
		}
		if status != "draft" && status != "ready" {
			return errors.New("只能标记为待完善或 Ready")
		}
		a.Status = status
	} else if !create {
		a.Status = "draft"
	}
	a.UpdatedAt = soNow()
	return nil
}

func (s *saysoStore) stopRecording(v *soSession) {
	if v.timerCancel != nil {
		v.timerCancel()
		v.timerCancel = nil
	}
	v.Status = "paused"
	v.AnalysisNextAt = ""
	v.owner = nil
}
func (s *saysoStore) startRecording(p *peer, v *soSession) {
	for _, other := range s.state.Sessions {
		if other.owner == p {
			s.stopRecording(other)
			s.analyze(p, other, false)
		}
	}
	ctx, cancel := context.WithCancel(p.ctx)
	v.timerCancel = cancel
	v.owner = p
	v.Status = "recording"
	v.AnalysisNextAt = time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339Nano)
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.mu.Lock()
				if v.owner == p && v.timerCancel != nil && p.ctx.Err() != nil {
					s.stopRecording(v)
					_ = s.save()
				}
				s.mu.Unlock()
				return
			case <-ticker.C:
				s.mu.Lock()
				if ctx.Err() == nil {
					s.analyze(p, v, false)
					v.AnalysisNextAt = time.Now().Add(2 * time.Minute).UTC().Format(time.RFC3339Nano)
					_ = s.save()
				}
				s.mu.Unlock()
			}
		}
	}()
}
