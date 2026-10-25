package main

import (
	"fmt"

	"github.com/string1225/remote-agent-cli/internal/codexsetup"
)

func installRemoteAgentSkill() error {
	home, err := codexsetup.Home()
	if err != nil {
		return err
	}
	path, err := codexsetup.InstallRemoteAgentSkill(home)
	if err != nil {
		return fmt.Errorf("安装 remote-agent skill 失败，可重新运行 codex setup：%w", err)
	}
	fmt.Printf("remote-agent skill 已安装：%s；在新 Codex 会话中生效。\n", path)
	return nil
}
