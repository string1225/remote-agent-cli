package main

import (
	"errors"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/agent"
)

// A model token configures local Codex only. It is never redeemed as a device
// code or used to rewrite an existing remote binding, service or system proxy.
func activateEndpoint(_ string, o activationOptions, steps activationSteps, emit func(activationEvent)) error {
	finish := func(err error) error {
		message := "本地 Codex 已配置，可打开新终端运行 codex。未添加或更改远程设备绑定。"
		if err == nil && o.InstallApp {
			message += "已打开 Microsoft Store，请在商店完成 Codex App 安装。"
		}
		if err != nil {
			message = strings.ReplaceAll(err.Error(), o.APIToken, "[已隐藏]")
			if token := strings.TrimSpace(o.APIToken); token != "" {
				message = strings.ReplaceAll(message, token, "[已隐藏]")
			}
		}
		emit(activationEvent{Type: "result", Success: err == nil, Message: message})
		return nil
	}
	if err := agent.ValidateServer(o.Server); err != nil {
		return finish(err)
	}
	if strings.TrimSpace(o.BindingCode) != "" {
		return finish(errors.New("请只填写 Endpoint Token 或设备绑定码中的一种"))
	}
	if !o.ConfigureCodex {
		return finish(errors.New("使用 Endpoint Token 时请勾选「配置 Codex」"))
	}
	emit(activationEvent{Type: "progress", Message: "正在验证 Endpoint Token 的账号…"})
	connection, err := steps.Personal(o.Server, strings.TrimSpace(o.APIToken))
	if err != nil {
		return finish(err)
	}
	c := agent.Config{Server: o.Server}
	if o.InstallCodex {
		emit(activationEvent{Type: "progress", Message: "正在检查或安装 Codex CLI…"})
		if err = steps.Install(c); err != nil {
			return finish(err)
		}
	}
	emit(activationEvent{Type: "progress", Message: "正在备份并配置本地 Codex…"})
	if err = steps.Apply(connection); err != nil {
		return finish(err)
	}
	if o.InstallApp {
		emit(activationEvent{Type: "progress", Message: "正在打开 Codex App 商店页…"})
		if err = steps.App(); err != nil {
			return finish(err)
		}
	}
	return finish(nil)
}
