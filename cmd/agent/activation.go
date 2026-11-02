package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/agent"
	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

type activationOptions struct {
	Server         string `json:"server"`
	BindingCode    string `json:"bindingCode,omitempty"`
	AutoStart      bool   `json:"autoStart"`
	WebProxy       bool   `json:"webProxy"`
	ConfigureCodex bool   `json:"configureCodex"`
	InstallCodex   bool   `json:"installCodex"`
}
type activationEvent struct {
	Type    string             `json:"type"`
	Message string             `json:"message,omitempty"`
	Success bool               `json:"success"`
	Bound   bool               `json:"bound"`
	Name    string             `json:"name,omitempty"`
	Options *activationOptions `json:"options,omitempty"`
}
type activationSteps struct {
	Install   func(agent.Config) error
	Configure func(agent.Config) error
	Service   func(agent.Config, bool) error
	Proxy     func(agent.Config, bool) error
	Online    func(agent.Config) error
}

func activationPreferences(path string) string { return path + ".activation-options.json" }
func activationDefaults(path string) activationOptions {
	o := activationOptions{Server: "https://codex.sunny-string.cn/agents", AutoStart: true, ConfigureCodex: true, InstallCodex: true}
	if data, err := os.ReadFile(activationPreferences(path)); err == nil {
		_ = json.Unmarshal(data, &o)
	}
	o.BindingCode = ""
	if cfg, err := agent.Load(path); err == nil {
		o.Server = cfg.Server
	}
	return o
}

func runActivation(args []string, input io.Reader, output io.Writer) error {
	if err := agent.CheckActivationUser(os.Getenv("REMOTE_AGENT_ACTIVATION_OWNER")); err != nil {
		return json.NewEncoder(output).Encode(activationEvent{Type: "result", Message: err.Error()})
	}
	path, err := agent.ActivationConfigPath()
	if err != nil {
		return json.NewEncoder(output).Encode(activationEvent{Type: "result", Message: err.Error()})
	}
	if len(args) == 1 && args[0] == "--inspect" {
		o := activationDefaults(path)
		cfg, err := agent.Load(path)
		return json.NewEncoder(output).Encode(activationEvent{Type: "inspect", Bound: err == nil, Name: cfg.Name, Options: &o})
	}
	if len(args) != 1 || args[0] != "--request-stdin" {
		return errors.New("activation --inspect | --request-stdin")
	}
	var o activationOptions
	if err := json.NewDecoder(io.LimitReader(input, 16384)).Decode(&o); err != nil {
		return errors.New("无效激活请求")
	}
	if runtime.GOOS != "windows" {
		return errors.New("MSI 激活仅适用于 Windows")
	}
	encoder := json.NewEncoder(output)
	return activateDevice(path, o, defaultActivationSteps(), func(e activationEvent) { _ = encoder.Encode(e) })
}

func activateDevice(path string, o activationOptions, steps activationSteps, emit func(activationEvent)) error {
	progress := func(message string) { emit(activationEvent{Type: "progress", Message: message}) }
	finish := func(c agent.Config, err error) error {
		message := "设备「" + c.Name + "」已激活并在线"
		if err != nil {
			message = err.Error()
			for _, secret := range []string{o.BindingCode, c.Credential} {
				if secret != "" {
					message = strings.ReplaceAll(message, secret, "[已隐藏]")
				}
			}
		}
		emit(activationEvent{Type: "result", Success: err == nil, Bound: c.ID != "", Name: c.Name, Message: message})
		return nil // Application failures are shown in the MSI, never uninstall it.
	}
	if err := agent.ValidateServer(o.Server); err != nil {
		return finish(agent.Config{}, err)
	}
	progress("正在绑定设备…")
	c, err := agent.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		c, err = agent.EnrollActivation(o.Server, strings.TrimSpace(o.BindingCode), path)
	}
	if err != nil {
		return finish(c, err)
	}
	if strings.TrimSuffix(o.Server, "/") != c.Server {
		return finish(c, errors.New("本机已绑定其他站点，已保留原账号和设备配置"))
	}
	// Persist only choices, never the binding code or device credential.
	saved := o
	saved.BindingCode = ""
	raw, _ := json.MarshalIndent(saved, "", "  ")
	if err = os.WriteFile(activationPreferences(path), raw, 0600); err != nil {
		return finish(c, err)
	}
	var failures []error
	step := func(label string, fn func() error) bool {
		progress(label)
		if e := fn(); e != nil {
			failures = append(failures, errors.New(label+" "+e.Error()))
			return false
		}
		return true
	}
	if o.InstallCodex {
		step("正在检查或安装 Codex CLI…", func() error { return steps.Install(c) })
	}
	if o.ConfigureCodex {
		step("正在备份并配置 Codex…", func() error { return steps.Configure(c) })
	}
	if step("正在配置并启动 Windows Service，目录授权可能需要数分钟…", func() error { return steps.Service(c, o.AutoStart) }) {
		step("正在应用网页代理选项…", func() error { return steps.Proxy(c, o.WebProxy) })
		step("正在核对设备在线状态…", func() error { return steps.Online(c) })
	}
	return finish(c, errors.Join(failures...))
}

func defaultActivationSteps() activationSteps {
	return activationSteps{
		Install: func(c agent.Config) error {
			if agent.Detect()["codex"].Available {
				return nil
			}
			home, e := os.UserHomeDir()
			if e != nil {
				return e
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			_, e = codexsetup.Install(ctx, c.Server+"/install/codex-release", filepath.Join(home, ".remote-agent", "bin"), runtime.GOOS, runtime.GOARCH, &http.Client{Timeout: 25 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }})
			return e
		},
		Configure: func(c agent.Config) error {
			home, e := codexsetup.Home()
			if e != nil {
				return e
			}
			return applyBoundCodex(&http.Client{Timeout: 30 * time.Second}, c, home, io.Discard)
		},
		Service: func(c agent.Config, automatic bool) error {
			return agent.SetupWindowsServiceWithStartup(c.Path, automatic)
		},
		Proxy:  func(c agent.Config, on bool) error { return applyActivationProxy(c, on) },
		Online: waitActivationOnline,
	}
}

func waitActivationOnline(c agent.Config) error {
	client := &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for end := time.Now().Add(35 * time.Second); time.Now().Before(end); {
		req, _ := http.NewRequest("GET", c.Server+"/api/agent/activation-status?id="+c.ID, nil)
		req.Header.Set("Authorization", "Bearer "+c.Credential)
		res, err := client.Do(req)
		if err == nil {
			var status struct {
				Online bool `json:"online"`
			}
			_ = json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&status)
			res.Body.Close()
			if res.StatusCode == 200 && status.Online {
				return nil
			}
			if res.StatusCode == 401 || res.StatusCode == 403 {
				return errors.New("设备凭据未通过服务器验证，请检查原绑定")
			}
		}
		time.Sleep(time.Second)
	}
	return errors.New("绑定已保存，暂未确认服务在线；请检查网络和本机 agent.log 后重试")
}
