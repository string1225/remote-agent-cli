package agent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/string1225/remote-agent-cli/internal/buildinfo"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

type MaintenanceOperation struct {
	ID              string `json:"id"`
	AgentID         string `json:"agentId"`
	RequestID       string `json:"requestId"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	Stage           string `json:"stage"`
	Message         string `json:"message"`
	BeforeVersion   string `json:"beforeVersion,omitempty"`
	ExpectedVersion string `json:"expectedVersion,omitempty"`
	AfterVersion    string `json:"afterVersion,omitempty"`
	Downloaded      int64  `json:"downloaded,omitempty"`
	Total           int64  `json:"total,omitempty"`
	StartedAt       int64  `json:"startedAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}
type maintenanceState struct {
	mu            sync.Mutex
	path, agentID string
	items         []MaintenanceOperation
	ctx           context.Context
}
type maintenanceInput struct {
	Kind        string `json:"kind"`
	RequestID   string `json:"requestId"`
	OperationID string `json:"operationId"`
}

var operationIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)

func (a *Agent) maintenanceState() *maintenanceState {
	a.remoteOnce.Do(func() {
		a.mu.Lock()
		path, id := a.Config.Path, a.Config.ID
		a.mu.Unlock()
		m := &maintenanceState{agentID: id, ctx: context.Background()}
		if path != "" {
			m.path = filepath.Join(filepath.Dir(path), "maintenance.json")
			if raw, e := os.ReadFile(m.path); e == nil && len(raw) <= 128<<10 {
				_ = json.Unmarshal(raw, &m.items)
			}
		}
		a.remote = m
	})
	return a.remote
}
func (m *maintenanceState) saveLocked() error {
	if m.path == "" {
		return nil
	}
	raw, e := json.Marshal(m.items)
	if e != nil {
		return e
	}
	// Progress has no credentials. Keep the directory's inherited owner and
	// service grants, so service-written state stays accessible after uninstall.
	return writeMaintenanceFile(m.path, raw)
}

func writeMaintenanceFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".maintenance-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
func (m *maintenanceState) get(id string) (MaintenanceOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, op := range m.items {
		if op.AgentID == m.agentID && (id == "" || op.ID == id) {
			return op, nil
		}
	}
	return MaintenanceOperation{}, nil
}
func (m *maintenanceState) change(id string, edit func(*MaintenanceOperation)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.items {
		if m.items[i].ID == id && m.items[i].AgentID == m.agentID {
			previous := m.items[i]
			edit(&m.items[i])
			m.items[i].UpdatedAt = time.Now().UnixMilli()
			if err := m.saveLocked(); err != nil {
				m.items[i] = previous
				return err
			}
			return nil
		}
	}
	return errors.New("维护操作不存在")
}
func (m *maintenanceState) start(in maintenanceInput) (MaintenanceOperation, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if (in.Kind != "client" && in.Kind != "codex") || !operationIDPattern.MatchString(in.RequestID) {
		return MaintenanceOperation{}, false, errors.New("无效的维护操作")
	}
	for _, op := range m.items {
		if op.AgentID != m.agentID {
			continue
		}
		if op.RequestID == in.RequestID {
			if op.Kind != in.Kind {
				return op, false, errors.New("操作编号已用于其他任务")
			}
			return op, false, nil
		}
		if op.Status == "queued" || op.Status == "running" {
			return op, false, errors.New("这个客户端已有维护操作正在进行")
		}
	}
	op := MaintenanceOperation{ID: store.Token(), AgentID: m.agentID, RequestID: in.RequestID, Kind: in.Kind, Status: "queued", Stage: "queued", Message: "维护请求已保存，正在准备", StartedAt: time.Now().UnixMilli(), UpdatedAt: time.Now().UnixMilli()}
	old := m.items
	m.items = append([]MaintenanceOperation{op}, m.items...)
	if len(m.items) > 20 {
		m.items = m.items[:20]
	}
	if err := m.saveLocked(); err != nil {
		m.items = old
		return op, false, err
	}
	return op, true, nil
}
func (a *Agent) initRemoteMaintenance(ctx context.Context) {
	m := a.maintenanceState()
	m.mu.Lock()
	m.ctx = ctx
	var pending []MaintenanceOperation
	for _, op := range m.items {
		if op.AgentID == m.agentID && (op.Status == "queued" || op.Status == "running") {
			pending = append(pending, op)
		}
	}
	m.mu.Unlock()
	for _, op := range pending {
		if op.Kind != "client" || op.ExpectedVersion == "" {
			_ = m.change(op.ID, func(v *MaintenanceOperation) {
				v.Status = "failed"
				v.Stage = "interrupted"
				v.Message = "客户端已重启，维护操作中断；原配置和安装备份保留，请检查后重试"
			})
			continue
		}
		go a.recoverRemoteUpdate(ctx, op)
	}
}
func (a *Agent) recoverRemoteUpdate(ctx context.Context, op MaintenanceOperation) {
	m := a.maintenanceState()
	until := time.Now().Add(2 * time.Minute)
	for time.Now().Before(until) {
		version, pending, err := remoteRuntimeState(a.Config.Path)
		if err == nil && !pending {
			_ = m.change(op.ID, func(v *MaintenanceOperation) {
				v.AfterVersion = buildinfo.Current()
				if version == op.ExpectedVersion && buildinfo.Current() == op.ExpectedVersion {
					v.Status = "succeeded"
					v.Stage = "complete"
					v.Message = "客户端更新完成，已重新上线"
				} else {
					v.Status = "failed"
					v.Stage = "rollback"
					v.Message = "新版本未完成激活，已保留或恢复上一版本"
				}
			})
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
	_ = m.change(op.ID, func(v *MaintenanceOperation) {
		v.Status = "failed"
		v.Stage = "check_required"
		v.Message = "更新结果暂未确认，请在客户端检查运行状态及更新日志"
	})
}
func (a *Agent) handleMaintenance(req protocol.Request, send func(protocol.Event) error) {
	fail := func(err error) { _ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()}) }
	var in maintenanceInput
	if len(req.Body) > 0 {
		d := json.NewDecoder(strings.NewReader(string(req.Body)))
		d.DisallowUnknownFields()
		if d.Decode(&in) != nil {
			fail(errors.New("无效的维护参数"))
			return
		}
	}
	m := a.maintenanceState()
	if req.Type == "maintenance.status" {
		op, err := m.get(in.OperationID)
		if err != nil {
			fail(err)
		} else {
			_ = send(protocol.Event{ID: req.ID, Type: "result", Data: op})
		}
		return
	}
	if in.Kind == "client" && !remoteClientUpdateSupported() {
		message := "请先在目标电脑更新客户端并以 remote-agent run 启动，之后可远程更新，无需重新绑定"
		if runtime.GOOS == "windows" {
			message = "需要先在这台电脑安装支持远程维护的新版 MSI；完成一次宿主升级后，后续业务更新无需本机授权"
		}
		fail(errors.New(message))
		return
	}
	op, start, err := m.start(in)
	if err != nil {
		fail(err)
		return
	}
	_ = send(protocol.Event{ID: req.ID, Type: "result", Data: op})
	if start {
		m.mu.Lock()
		ctx := m.ctx
		m.mu.Unlock()
		go func() {
			timer := time.NewTimer(200 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				a.runMaintenance(ctx, op)
			}
		}()
	}
}
func (a *Agent) runMaintenance(parent context.Context, op MaintenanceOperation) {
	ctx, cancel := context.WithTimeout(parent, 30*time.Minute)
	defer cancel()
	m := a.maintenanceState()
	set := func(stage, message string) {
		_ = m.change(op.ID, func(v *MaintenanceOperation) { v.Status = "running"; v.Stage = stage; v.Message = message })
	}
	set("preparing", "正在检查客户端环境")
	var err error
	if op.Kind == "client" {
		a.mu.Lock()
		server := a.Config.Server
		a.mu.Unlock()
		release, e := remoteClientRelease(ctx, server)
		err = e
		if err == nil {
			err = m.change(op.ID, func(v *MaintenanceOperation) {
				v.BeforeVersion = buildinfo.Current()
				v.ExpectedVersion = release.Version
				v.Stage = "download"
				v.Message = "正在下载并验证客户端更新"
			})
			if err == nil {
				err = a.applyRemoteClientUpdate(ctx, op, release)
			}
		}
		if parent.Err() != nil {
			return
		}
		if err == nil {
			a.recoverRemoteUpdate(ctx, MaintenanceOperation{ID: op.ID, ExpectedVersion: release.Version})
			return
		}
	} else {
		err = a.installRemoteCodex(ctx, op, set)
	}
	if err != nil {
		_ = m.change(op.ID, func(v *MaintenanceOperation) {
			v.Status = "failed"
			v.Stage = "error"
			v.Message = boundedMaintenanceError(err)
		})
		return
	}
	_ = m.change(op.ID, func(v *MaintenanceOperation) {
		v.Status = "succeeded"
		v.Stage = "complete"
		v.Message = "Codex CLI 已安装或更新，客户端可直接使用，无需重启服务"
	})
}
func boundedMaintenanceError(err error) string {
	value := err.Error()
	if len(value) > 1000 {
		value = value[:1000]
	}
	return value
}

type maintenanceWriter struct{ set func(string, string) }

func (w *maintenanceWriter) Write(p []byte) (int, error) {
	text := string(p)
	switch {
	case strings.Contains(text, "Waiting for local tasks"):
		w.set("checking_tasks", "更新包已验证，正在确认任务已结束")
	case strings.Contains(text, "Starting candidate"):
		w.set("restarting", "正在切换版本并重启客户端，稍后会自动重新连接")
	case strings.Contains(text, "Runtime is up to date"):
		w.set("verifying", "当前已是已发布版本，正在确认运行状态")
	case strings.Contains(text, "restored previous"), strings.Contains(text, "previous runtime remains"):
		w.set("rollback", "新版本未通过检查，正在恢复上一稳定版本")
	}
	return len(p), nil
}
func (a *Agent) installRemoteCodex(ctx context.Context, op MaintenanceOperation, set func(string, string)) error {
	a.mu.Lock()
	if a.maintenance || a.toolInstalling || a.busy || a.operations > 0 || len(a.taskCancels) > 0 || a.recordingActivity > 0 {
		a.mu.Unlock()
		return errors.New("请先结束客户端上的任务和录音，再安装或升级 Codex CLI")
	}
	a.toolInstalling = true
	c := a.Config
	before := a.Providers["codex"].Path
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.toolInstalling = false; a.mu.Unlock() }()
	if before != "" {
		version := providerVersion(ctx, before)
		if err := a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) { v.BeforeVersion = version }); err != nil {
			return err
		}
	}
	bin := filepath.Join(filepath.Dir(c.Path), "bin")
	client := &http.Client{Timeout: 25 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &maintenanceDownloadTransport{base: http.DefaultTransport, report: func(n, total int64) {
		_ = a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) {
			v.Stage = "download"
			v.Message = "正在下载官方 Codex CLI"
			v.Downloaded = n
			v.Total = total
		})
	}, stage: set}}
	set("release", "正在获取官方 Codex CLI 发布信息")
	path, err := codexsetup.Install(ctx, c.Server+"/install/codex-release", bin, runtime.GOOS, runtime.GOARCH, client)
	if err != nil {
		return err
	}
	set("checking_binary", "安装完成，正在检测 Codex CLI")
	provider := probeManagedCodex(path)
	if !provider.Available {
		return errors.New("安装包已保留，但 Codex CLI 未通过启动检查，请查看本机环境")
	}
	c.ManagedCodex = true
	if err = c.Save(); err != nil {
		return err
	}
	a.mu.Lock()
	a.Config.ManagedCodex = true
	if a.Providers == nil {
		a.Providers = make(map[string]Provider)
	}
	a.Providers["codex"] = provider
	a.mu.Unlock()
	version := providerVersion(ctx, path)
	return a.maintenanceState().change(op.ID, func(v *MaintenanceOperation) { v.AfterVersion = version })
}

type maintenanceDownloadTransport struct {
	base   http.RoundTripper
	report func(int64, int64)
	stage  func(string, string)
}

func (t *maintenanceDownloadTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.base.RoundTrip(req)
	if err == nil && res.StatusCode == 200 && strings.Contains(req.URL.Path, "/downloads/") {
		res.Body = &maintenanceDownloadBody{ReadCloser: res.Body, total: res.ContentLength, report: t.report, stage: t.stage}
	}
	return res, err
}

type maintenanceDownloadBody struct {
	io.ReadCloser
	n, total int64
	last     time.Time
	report   func(int64, int64)
	stage    func(string, string)
}

func (b *maintenanceDownloadBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.n += int64(n)
	if time.Since(b.last) > 500*time.Millisecond || err != nil {
		b.last = time.Now()
		b.report(b.n, b.total)
	}
	if err == io.EOF {
		b.stage("installing", "下载完成，正在校验并安装完整 Codex 包")
	}
	return n, err
}
func probeManagedCodex(path string) Provider {
	p := Provider{Name: "codex", Path: path}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	var out boundedBuffer
	cmd := command(ctx, path, []string{"--help"})
	cmd.Stdout = &out
	cmd.Stderr = &out
	p.Available = cmd.Run() == nil && strings.Contains(out.String(), "exec")
	if !p.Available {
		p.Error = "托管 Codex CLI 无法启动，请从客户端菜单重新安装或更新"
	}
	return p
}
func detectConfigured(c Config) map[string]Provider {
	providers := Detect()
	if c.ManagedCodex {
		exe := "codex"
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		providers["codex"] = probeManagedCodex(filepath.Join(filepath.Dir(c.Path), "bin", exe))
	}
	return providers
}
func providerVersion(ctx context.Context, path string) string {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, e := command(ctx, path, []string{"--version"}).Output()
	if e != nil {
		return ""
	}
	return regexp.MustCompile(`[0-9]+(?:\.[0-9]+){1,3}`).FindString(string(out))
}

func (a *Agent) providerAvailable(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Providers[name].Available
}
