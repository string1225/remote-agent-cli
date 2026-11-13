package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestManagerInstructionsRefreshExistingAndNewRolesWithoutExecutorRecursion(t *testing.T) {
	for _, scope := range []string{"account", "client", "project", ""} {
		t.Run(scope, func(t *testing.T) {
			root := t.TempDir()
			s := protocol.Service{Scope: scope, Workspace: root}
			path := filepath.Join(root, "AGENTS.md")
			original := "保留用户规则\n" + legacyAssistantGuidance + roleStart + "\n即使任务很小，也不要在管理会话亲自执行。\n" + roleEnd + "\n"
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				if err := SyncAssistantInstructions(s); err != nil {
					t.Fatal(err)
				}
			}
			got, _ := os.ReadFile(path)
			backups, _ := filepath.Glob(path + ".role-backup-*")
			if scope == "account" || scope == "client" {
				if !strings.Contains(string(got), "保留用户规则") || !strings.Contains(string(got), "预计5步以内") || strings.Contains(string(got), "即使任务很小") || strings.Contains(string(got), "小事可以直接处理") || strings.Count(string(got), roleStart) != 1 || len(backups) != 1 {
					t.Fatal("migration failed to preserve custom content or was not idempotent")
				}
				backup, _ := os.ReadFile(backups[0])
				if string(backup) != original {
					t.Fatal("migration lost the original file")
				}
			} else if string(got) != original || len(backups) != 0 {
				t.Fatal("shared project rules modified")
			}
			instructions := SessionInstructionsFor(s)
			if scope == "" {
				if instructions != SessionInstructions || AssistantInstructions(s) != "" || strings.Contains(instructions, "超过5步") {
					t.Fatal("executor inherited manager workflow")
				}
			} else if strings.Contains(instructions, "你可以直接完成任务") || !strings.Contains(instructions, "list_execution_topics") || !strings.Contains(instructions, "delegate_work") {
				t.Fatal("manager did not receive mandatory scheduling and layered dispatch")
			}
			if scope != "" && (!strings.Contains(instructions, "默认事件驱动") || !strings.Contains(instructions, "结束管理轮次") || strings.Contains(instructions, "结束汇总前须主动读取/等待")) {
				t.Fatal("manager still holds a polling model")
			}
			if scope != "" && (!strings.Contains(instructions, "预计5步以内") || !strings.Contains(instructions, "超过5步") || strings.Contains(instructions, "即使任务很小") || strings.Contains(instructions, "具体执行必须交给")) {
				t.Fatal("manager lost the five-step exception or kept unconditional delegation")
			}
		})
	}
	root := t.TempDir()
	s := protocol.Service{Scope: "account", Workspace: root}
	if err := SyncAssistantInstructions(s); err != nil {
		t.Fatal("new role", err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.Contains(string(got), managerInstructions) {
		t.Fatal("new role missed persistent instructions")
	}
}

func TestManagerRoleMigrationFailsWithoutOverwritingDamagedContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AGENTS.md")
	original := "custom\n" + roleStart
	os.WriteFile(path, []byte(original), 0600)
	if err := SyncAssistantInstructions(protocol.Service{Scope: "client", Workspace: root}); err == nil {
		t.Fatal("damaged managed block accepted")
	}
	got, _ := os.ReadFile(path)
	if string(got) != original {
		t.Fatal("damaged file overwritten")
	}
}
