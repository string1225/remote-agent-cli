package server

import (
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

func TestTodoOptionEncodingPreservesMarkdownAndRejectsOversize(t *testing.T) {
	body := "**背景**\n<!-- ordinary comment -->\n- 保留原文"
	options := []protocol.TodoOption{{Label: "A <script>", Description: "换行\n说明"}}
	packed, err := encodeTodoOptions(body, options)
	if err != nil {
		t.Fatal(err)
	}
	row := protocol.Todo{Body: packed}
	decodeTodoOptions(&row)
	if row.Body != body || len(row.Options) != 1 || row.Options[0] != options[0] {
		t.Fatal("option roundtrip changed content", row)
	}
	for _, invalid := range [][]protocol.TodoOption{{{Label: " "}}, {{Label: "a", Description: strings.Repeat("x", 1201)}}, make([]protocol.TodoOption, 13)} {
		if _, err := encodeTodoOptions(body, invalid); err == nil {
			t.Fatal("invalid choices accepted")
		}
	}
	if _, err := encodeTodoOptions(strings.Repeat("x", 7990), options); err == nil {
		t.Fatal("choices exceeded legacy client body capacity")
	}
	malformed := protocol.Todo{Body: body + todoOptionsPrefix + "invalid -->"}
	before := malformed.Body
	decodeTodoOptions(&malformed)
	if malformed.Body != before || malformed.Options != nil {
		t.Fatal("malformed extension removed user content")
	}
	legacy := protocol.Todo{Body: "请选择：\n- **A · 周五（推荐）**：观察周末\n- **B · 周一**：工作日处理"}
	decodeTodoOptions(&legacy)
	if len(legacy.Options) != 2 || legacy.Options[0].Description != "观察周末" {
		t.Fatal("old publisher's explicit A/B choices were lost", legacy)
	}
	notice := protocol.Todo{Body: "构建完成\n- 代码已提交\n- 验证已通过"}
	decodeTodoOptions(&notice)
	if len(notice.Options) != 0 {
		t.Fatal("notification bullets became invented choices")
	}
}
