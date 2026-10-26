package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	bolt "go.etcd.io/bbolt"
)

// Shared by the fixed IM prompt and MCP instructions, including delegated work.
const IMDeliveryInstructions = "IM 来源任务默认开启完成通知。创建任务后立即简短确认已排队，不要等待执行结束。创建确认和完成消息都不要附网页会话地址，也不能用会话地址代替结果。最终回答须用中文直接给出实质结论、关键结果、验证/提交情况和需用户处理的事项，系统会自动回传保存的来源私聊或群聊，不广播。notify_im_session 通常只发进度；如需主动发送最终结果，使用 final:true，成功入队后系统不再重复发送自动完成结果。其他正常业务链接可保留。"

type taskResult struct {
	Text     string
	Notified bool
	Captured bool
}

func (h historyStore) taskResult(id string) (result taskResult, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte("task_results")); b != nil {
			if data := b.Get([]byte(id)); data != nil {
				return json.Unmarshal(data, &result)
			}
		}
		return nil
	})
	return
}

func (h historyStore) updateTaskResult(id string, change func(*taskResult)) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("task_results"))
		if err != nil {
			return err
		}
		var result taskResult
		if data := b.Get([]byte(id)); data != nil {
			if err := json.Unmarshal(data, &result); err != nil {
				return err
			}
		}
		change(&result)
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), data)
	})
}

func (h historyStore) saveTaskResult(id, text string) error {
	return h.updateTaskResult(id, func(v *taskResult) { v.Text = text; v.Captured = true })
}

// Called only after an explicit final notification to the inherited route was
// accepted by channels.notify. Persist across worker/client restarts.
func RecordFinalNotification(c Config, sessionID string) error {
	h := (&Agent{Config: c}).historyStore()
	s, err := h.get(sessionID)
	if err != nil {
		return err
	}
	if s.TaskID == "" {
		return nil
	}
	return h.updateTaskResult(s.TaskID, func(v *taskResult) { v.Notified = true })
}

func (a *Agent) taskCompletionText(task queuedTask, runErr error) string {
	result, readErr := a.historyStore().taskResult(task.result.TaskID)
	if result.Notified && runErr == nil {
		return ""
	}
	text := strings.TrimSpace(result.Text)
	// Older/plain-text runners have no structured final record. Their assistant
	// output remains a useful fallback; tool/stderr entries are excluded.
	if text == "" && !result.Captured && readErr == nil {
		text = strings.TrimSpace(a.historyStore().turnText(task.result.Session.ID, task.result.TaskID, 1<<20))
	}
	status := "完成"
	if runErr != nil {
		status = "失败"
		if errors.Is(runErr, context.Canceled) {
			status = "取消"
		}
		partial := text
		text = "任务未完成：" + runErr.Error() + "\n请检查已完成内容后再继续，避免重复执行。"
		if partial != "" {
			text += "\n\n已输出的结果（可能不完整）：\n" + partial
		}
	} else if text == "" {
		text = "执行已结束，但未获得可回传的最终结果；请检查任务输出后再确认是否完成。"
	}
	return withoutSessionLinks(fmt.Sprintf("任务%s：%s\n\n%s", status, task.result.Session.Title, text), a.Config.Server)
}

var imMarkdownLink = regexp.MustCompile(`\[([^\]\n]*)\]\((https?://[^\s)]+)\)`)
var imURL = regexp.MustCompile(`https?://[^\s<>）)]+`)

// Scope filtering to this site's workspace session addresses in outbound IM.
// Documentation, repository, download and all other ordinary URLs survive.
func withoutSessionLinks(text, server string) string {
	base, err := url.Parse(server)
	if err != nil || base.Host == "" {
		return text
	}
	isSession := func(raw string) bool {
		u, err := url.Parse(strings.TrimRight(raw, ".,，。"))
		if err != nil || !strings.EqualFold(u.Host, base.Host) || strings.TrimRight(u.Path, "/") != strings.TrimRight(base.Path, "/") {
			return false
		}
		q, err := url.ParseQuery(u.Fragment)
		return err == nil && q.Get("session") != "" && q.Get("client") != "" && q.Get("project") != ""
	}
	text = imMarkdownLink.ReplaceAllStringFunc(text, func(link string) string {
		if isSession(imMarkdownLink.FindStringSubmatch(link)[2]) {
			return "（会话地址已省略）"
		}
		return link
	})
	return imURL.ReplaceAllStringFunc(text, func(link string) string {
		if isSession(link) {
			return "（会话地址已省略）"
		}
		return link
	})
}

// channels.notify accepts at most 16000 bytes. Preserve every part of a long
// result, including UTF-8 boundaries, rather than replacing it with a URL.
func notificationParts(text string) []string {
	var parts []string
	for len(text) > 14000 {
		n := 14000
		for !utf8.RuneStart(text[n]) {
			n--
		}
		parts = append(parts, text[:n])
		text = text[n:]
	}
	if text != "" {
		parts = append(parts, text)
	}
	return parts
}
