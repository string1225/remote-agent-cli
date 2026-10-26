package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// Removal unregisters a project, never its directory or local conversation data.
func (a *Agent) removeProject(ctx context.Context, id string, removeChannel bool) error {
	a.mu.Lock()
	if len(a.runs) > 0 {
		a.mu.Unlock()
		return errors.New("请先等待或停止客户端当前任务，再删除项目")
	}
	found, bound := false, false
	for _, p := range a.Config.Services {
		if p.ID == id {
			found = true
			bound = p.Channel != nil
		}
	}
	if !found {
		a.mu.Unlock()
		return errors.New("项目不存在")
	}
	if bound && !removeChannel {
		a.mu.Unlock()
		return errors.New("请确认删除项目并解绑 IM 渠道")
	}
	if a.removingProjects == nil {
		a.removingProjects = map[string]bool{}
	}
	if a.removingProjects[id] {
		a.mu.Unlock()
		return errors.New("项目正在删除")
	}
	a.removingProjects[id] = true
	snapshot, err := json.MarshalIndent(a.Config, "", "  ")
	path := a.Config.Path
	if raw, readErr := os.ReadFile(path); readErr == nil {
		snapshot = raw
	}
	a.mu.Unlock()
	defer func() { a.mu.Lock(); delete(a.removingProjects, id); a.mu.Unlock() }()
	if err != nil {
		return err
	}
	backup := path + ".project-remove-backup-" + time.Now().UTC().Format("20060102T150405.000000000")
	if err := os.WriteFile(backup, snapshot, 0600); err != nil {
		return err
	}
	if bound {
		if err := a.configureChannel(ctx, ChannelConfig{ServiceID: id}); err != nil {
			return err
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	next := a.Config
	next.Services = nil
	for _, p := range a.Config.Services {
		if p.ID != id {
			next.Services = append(next.Services, p)
		}
	}
	if err := next.Save(); err != nil {
		return fmt.Errorf("无法移除项目登记；本机文件和历史已保留，配置备份：%s：%w", backup, err)
	}
	a.Config.Services = next.Services
	return nil
}
