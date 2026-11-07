package main

import (
	"fmt"
)

func installRemoteAgentSkill() error {
	result, err := syncCurrentUserSkill()
	if err != nil {
		return fmt.Errorf("安装 remote-agent skill 失败，可重新运行 codex setup：%w", err)
	}
	fmt.Printf("remote-agent skill 已安装：%s；版本 %s。\n", result.Path, result.Version)
	return nil
}
