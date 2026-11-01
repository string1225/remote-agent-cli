package agent

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// Only the project registration changes. Historical workspaces, native session
// IDs, channel bindings and SaySo recordings remain available for recovery.
func (a *Agent) updateProject(request protocol.Service) (protocol.Service, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.removingProjects[request.ID] || a.maintenance || a.toolInstalling {
		return protocol.Service{}, errUpdating
	}
	name := strings.TrimSpace(request.Name)
	if name == "" || len([]rune(name)) > 80 {
		return protocol.Service{}, errors.New("项目名称需为 1–80 字")
	}
	workspace, err := a.Config.Workspace(request.Workspace)
	if err != nil {
		return protocol.Service{}, err
	}
	index := -1
	for i, project := range a.Config.Services {
		if project.ID == request.ID {
			index = i
		}
	}
	if index < 0 {
		return protocol.Service{}, errors.New("项目不存在")
	}
	project := a.Config.Services[index]
	if project.Workspace != workspace {
		if a.recordingActivity > 0 {
			return protocol.Service{}, errors.New("请先结束录音或音频上传，再修改工作目录")
		}
		for _, task := range a.taskRecords {
			if task.start.ServiceID == project.ID {
				return protocol.Service{}, errors.New("请先完成或取消此项目的执行和排队消息，再修改工作目录")
			}
		}
		for sessionID := range a.runs {
			session, e := a.historyStore().get(sessionID)
			if e != nil || session.ServiceID == project.ID {
				return protocol.Service{}, errors.New("请先完成此项目的执行，再修改工作目录")
			}
		}
	}
	for _, other := range a.Config.Services {
		if other.ID != project.ID && other.Provider == project.Provider {
			root, e := a.Config.Workspace(other.Workspace)
			if e == nil && root == workspace {
				return protocol.Service{}, errors.New("该目录已有同执行器项目")
			}
		}
	}
	backup, err := os.ReadFile(a.Config.Path)
	if os.IsNotExist(err) {
		backup, err = json.MarshalIndent(a.Config, "", "  ")
	}
	if err != nil {
		return protocol.Service{}, err
	}
	backupPath := a.Config.Path + ".project-settings-backup-" + time.Now().UTC().Format("20060102T150405.000000000")
	if err := os.WriteFile(backupPath, backup, 0600); err != nil {
		return protocol.Service{}, err
	}
	project.Name, project.Workspace = name, workspace
	next := a.Config
	next.Services = append([]protocol.Service{}, a.Config.Services...)
	next.Services[index] = project
	if err := next.Save(); err != nil {
		return protocol.Service{}, err
	}
	a.Config.Services = next.Services
	return project, nil
}
