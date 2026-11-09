package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// The extension lives in the existing client-owned body so deployed clients
// preserve choices across edits, host migrations and rollbacks. API consumers
// receive structured options and the original Markdown body, without metadata.
const todoOptionsPrefix = "\n<!-- remote-agent-options:v1:"

var legacyTodoOption = regexp.MustCompile(`(?m)^\s*[-*]\s+\*\*([A-L][ ·、.：:)）][^\n*]{1,90})\*\*\s*[：:]\s*([^\n]{0,400})$`)

func encodeTodoOptions(body string, options []protocol.TodoOption) (string, error) {
	if len(options) > 12 {
		return "", errors.New("待办最多提供 12 个选项")
	}
	for _, option := range options {
		if strings.TrimSpace(option.Label) == "" || !utf8.ValidString(option.Label) || !utf8.ValidString(option.Description) || len(option.Label) > 240 || len(option.Description) > 1200 {
			return "", errors.New("选项名称需为 1–240 字节，描述最多 1200 字节")
		}
	}
	if len(options) == 0 {
		return body, nil
	}
	data, err := json.Marshal(options)
	if err != nil {
		return "", err
	}
	body += todoOptionsPrefix + base64.RawURLEncoding.EncodeToString(data) + " -->"
	if len(body) > 8000 {
		return "", errors.New("待办正文与选项合计过长，请缩短描述")
	}
	return body, nil
}

func decodeTodoOptions(row *protocol.Todo) {
	index := strings.LastIndex(row.Body, todoOptionsPrefix)
	if index < 0 || !strings.HasSuffix(row.Body, " -->") {
		// Recognize explicit A/B choices from older tools, never ordinary bullets
		// in a notification. New publishers use the structured options argument.
		matches := legacyTodoOption.FindAllStringSubmatch(row.Body, 13)
		if len(matches) >= 2 && len(matches) <= 12 {
			options := make([]protocol.TodoOption, 0, len(matches))
			for i, match := range matches {
				if match[1][0] != byte('A'+i) {
					return
				}
				options = append(options, protocol.TodoOption{Label: strings.TrimSpace(match[1]), Description: strings.TrimSpace(match[2])})
			}
			if _, err := encodeTodoOptions("", options); err == nil {
				row.Options = options
			}
		}
		return
	}
	data, err := base64.RawURLEncoding.DecodeString(row.Body[index+len(todoOptionsPrefix) : len(row.Body)-4])
	var options []protocol.TodoOption
	if err != nil || json.Unmarshal(data, &options) != nil || len(options) == 0 {
		return
	}
	if _, err := encodeTodoOptions(row.Body[:index], options); err != nil {
		return
	}
	row.Body, row.Options = row.Body[:index], options
}
