package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

var errTodoConflict = errors.New("待办已被修改，请刷新后重试")

func todoText(s string, limit int) bool { return utf8.ValidString(s) && len(s) <= limit }

func (h historyStore) todos(in protocol.TodoStorageCall) (out any, err error) {
	if in.AccountID == "" || len(in.AccountID) > 160 || in.Generation == "" || len(in.Generation) > 160 {
		return nil, errors.New("缺少待办清单身份，请升级服务器")
	}
	err = h.transact(true, func(tx *bolt.Tx) error {
		root, e := tx.CreateBucketIfNotExists([]byte("account_todos"))
		if e != nil {
			return e
		}
		account, e := root.CreateBucketIfNotExists([]byte(in.AccountID))
		if e != nil {
			return e
		}
		b, e := account.CreateBucketIfNotExists([]byte(in.Generation))
		if e != nil {
			return e
		}
		if in.Operation == "probe" {
			out = map[string]bool{"ready": true}
			return nil
		}
		if in.Operation == "freeze" || in.Operation == "unfreeze" {
			if in.MigrationID == "" || len(in.MigrationID) > 100 {
				return errors.New("缺少迁移身份")
			}
			resolved := []byte("m/resolved/" + in.MigrationID)
			if in.Operation == "freeze" {
				if b.Get(resolved) != nil {
					return errors.New("此迁移已回滚")
				}
				if frozen := b.Get([]byte("m/frozen")); frozen != nil && string(frozen) != in.MigrationID {
					return errors.New("清单正在进行其他迁移")
				}
				if e := b.Put([]byte("m/frozen"), []byte(in.MigrationID)); e != nil {
					return e
				}
			} else {
				if string(b.Get([]byte("m/frozen"))) == in.MigrationID {
					if e := b.Delete([]byte("m/frozen")); e != nil {
						return e
					}
				}
				if e := b.Put(resolved, []byte("true")); e != nil {
					return e
				}
			}
			out = map[string]bool{"ok": true}
			return nil
		}
		if in.Operation != "get" && in.Operation != "list" && in.Operation != "export" && b.Get([]byte("m/frozen")) != nil {
			return errors.New("清单正在迁移，请稍后重试")
		}
		if in.Operation == "export" {
			page := protocol.TodoExport{Records: []protocol.TodoRecord{}}
			cursor := b.Cursor()
			size := 0
			for k, v := cursor.Seek([]byte(in.Cursor)); k != nil; k, v = cursor.Next() {
				if !strings.HasPrefix(string(k), "t/") && !strings.HasPrefix(string(k), "r/") {
					continue
				}
				if string(k) == in.Cursor {
					continue
				}
				if len(page.Records) > 0 && (size+len(v) > 24000 || len(page.Records) >= 30) {
					break
				}
				page.Records = append(page.Records, protocol.TodoRecord{Key: string(k), Value: bytes.Clone(v)})
				size += len(v)
				page.NextCursor = string(k)
			}
			if len(page.Records) > 0 {
				c := b.Cursor()
				c.Seek([]byte(page.Records[len(page.Records)-1].Key))
				if k, _ := c.Next(); k != nil {
					page.NextCursor = page.Records[len(page.Records)-1].Key
				} else {
					page.NextCursor = ""
				}
			}
			out = page
			return nil
		}
		if in.Operation == "import" {
			if len(in.Records) > 30 {
				return errors.New("迁移批次过大")
			}
			for _, row := range in.Records {
				if len(row.Key) > 256 || !(strings.HasPrefix(row.Key, "t/") || strings.HasPrefix(row.Key, "r/")) || len(row.Value) > 30000 || !json.Valid(row.Value) {
					return errors.New("无效待办备份")
				}
				if e := b.Put([]byte(row.Key), row.Value); e != nil {
					return e
				}
			}
			out = map[string]bool{"ok": true}
			return nil
		}
		if in.Operation == "list" {
			page := protocol.TodoPage{Todos: []protocol.Todo{}}
			cursor := b.Cursor()
			start := "t/" + in.Cursor
			size := 0
			for k, v := cursor.Seek([]byte(start)); k != nil && strings.HasPrefix(string(k), "t/"); k, v = cursor.Next() {
				if string(k) == start && in.Cursor != "" {
					continue
				}
				var row protocol.Todo
				if e := json.Unmarshal(v, &row); e != nil {
					return e
				}
				if in.Status != "" && row.Status != in.Status {
					continue
				}
				if !strings.Contains(strings.ToLower(row.Title+"\n"+row.Body), strings.ToLower(in.Search)) {
					continue
				}
				if len(page.Todos) > 0 && (size+len(v) > 28000 || len(page.Todos) >= 30) {
					break
				}
				page.Todos = append(page.Todos, row)
				size += len(v)
				page.NextCursor = row.ID
			}
			// A harmless extra empty page is preferable to dropping a matching row.
			if len(page.Todos) > 0 {
				c := b.Cursor()
				c.Seek([]byte("t/" + page.Todos[len(page.Todos)-1].ID))
				if k, _ := c.Next(); k != nil && strings.HasPrefix(string(k), "t/") {
					page.NextCursor = page.Todos[len(page.Todos)-1].ID
				} else {
					page.NextCursor = ""
				}
			}
			out = page
			return nil
		}
		if len(in.ID) > 100 || len(in.RequestID) > 160 {
			return errors.New("无效待办 ID")
		}
		if in.Operation == "create" {
			if in.RequestID == "" || strings.TrimSpace(in.Title) == "" || !todoText(in.Title, 480) || !todoText(in.Body, 8000) {
				return errors.New("待办标题必填且最多 160 个汉字，正文最多 8000 字节")
			}
			in.ID = store.Hash(in.RequestID)
		}
		key := []byte("t/" + in.ID)
		var row protocol.Todo
		raw := b.Get(key)
		if raw != nil {
			if e := json.Unmarshal(raw, &row); e != nil {
				return e
			}
		}
		if in.Operation == "create" {
			if raw != nil {
				if row.ClientID != in.ClientID || row.ProjectID != in.ProjectID || row.SessionID != in.SessionID {
					return errors.New("待办请求 ID 已被占用")
				}
				out = row
				return nil
			}
			now := time.Now().UnixMilli()
			row = protocol.Todo{ID: in.ID, Title: strings.TrimSpace(in.Title), Body: in.Body, Status: "open", ClientID: in.ClientID, ProjectID: in.ProjectID, SessionID: in.SessionID, Version: 1, CreatedAt: now, UpdatedAt: now}
		} else {
			if raw == nil {
				return errors.New("待办不存在")
			}
			if in.Operation == "get" {
				out = row
				return nil
			}
			if in.Operation == "reply" || in.Operation == "ack" {
				if in.RequestID == "" {
					return errors.New("回复缺少请求 ID")
				}
				if prior := b.Get([]byte("r/" + row.ID + "/" + in.RequestID)); prior != nil {
					var reply protocol.TodoReply
					if e := json.Unmarshal(prior, &reply); e != nil {
						return e
					}
					if in.Operation == "reply" {
						if reply.State == "cancelled" {
							return errors.New("此回复已取消，执行状态请在来源会话检查")
						}
						if reply.Text != in.Message {
							return errors.New("回复请求 ID 已被占用")
						}
						row.Reply = &reply
						out = row
						return nil
					}
					if reply.State == "delivered" {
						row.Reply = &reply
						out = row
						return nil
					}
				}
			}
			if in.Operation != "ack" && (in.Version == 0 || row.Version != in.Version) {
				return errTodoConflict
			}
			if row.Reply != nil && row.Reply.State == "pending" && in.Operation != "ack" && in.Operation != "cancel" {
				return errors.New("上一条回复尚未确认送达，请先重试回复")
			}
			switch in.Operation {
			case "update":
				if strings.TrimSpace(in.Title) == "" || !todoText(in.Title, 480) || !todoText(in.Body, 8000) {
					return errors.New("待办标题或正文超出限制")
				}
				row.Title, row.Body = strings.TrimSpace(in.Title), in.Body
			case "complete":
				row.Status = "completed"
			case "cancel":
				row.Status = "cancelled"
				if row.Reply != nil && row.Reply.State == "pending" {
					row.Reply.State = "cancelled"
				}
			case "reopen":
				row.Status = "open"
			case "reply":
				if row.Status != "open" || row.ClientID == "" || row.ProjectID == "" || row.SessionID == "" {
					return errors.New("仅可回复关联会话的未完成待办")
				}
				if strings.TrimSpace(in.Message) == "" || !todoText(in.Message, 8000) {
					return errors.New("回复内容需为 1–8000 字节")
				}
				row.Reply = &protocol.TodoReply{ID: in.RequestID, Text: in.Message, State: "pending"}
			case "ack":
				if row.Reply == nil || row.Reply.ID != in.RequestID {
					return errors.New("待确认回复不存在")
				}
				row.Reply.State = "delivered"
				row.Reply.TaskID = todoReplyTaskID(row.ID, in.RequestID)
			default:
				return errors.New("不支持的待办操作")
			}
			row.Version++
			row.UpdatedAt = time.Now().UnixMilli()
		}
		if row.Reply != nil {
			data, e := json.Marshal(row.Reply)
			if e != nil {
				return e
			}
			if e = b.Put([]byte("r/"+row.ID+"/"+row.Reply.ID), data); e != nil {
				return e
			}
		}
		data, e := json.Marshal(row)
		if e != nil {
			return e
		}
		// Leave room for an acknowledgement's task ID. JSON escaping can make
		// control characters substantially larger than their UTF-8 byte count.
		limit := 27000
		if in.Operation == "ack" {
			limit = 28000
		}
		if len(data) > limit {
			return errors.New("待办内容编码后过长，请缩短正文或回复")
		}
		out = row
		return b.Put(key, data)
	})
	return
}

func (a *Agent) handleTodos(req protocol.Request) (any, error) {
	if req.Type == "todos.validate" {
		var in protocol.TodoCommand
		if json.Unmarshal(req.Body, &in) != nil {
			return nil, errors.New("无效会话引用")
		}
		s, err := a.historyStore().get(in.SessionID)
		if err != nil || s.DeletedAt != 0 || s.ServiceID != in.ProjectID {
			return nil, errors.New("待办来源会话与项目不匹配")
		}
		for _, p := range a.services() {
			if p.ID == s.ServiceID {
				return map[string]bool{"valid": true}, nil
			}
		}
		return nil, errors.New("待办来源项目已移除")
	}
	if req.Type == "todos.deliver" {
		var row protocol.Todo
		if json.Unmarshal(req.Body, &row) != nil || row.Reply == nil || row.Reply.ID == "" || len(row.Reply.ID) > 80 || row.ClientID != a.Config.ID {
			return nil, errors.New("无效待办回复")
		}
		return a.startSessionTask(protocol.Request{ID: todoReplyTaskID(row.ID, row.Reply.ID), ServiceID: row.ProjectID, SessionID: row.SessionID, Prompt: fmt.Sprintf("待办「%s」收到人工回复：\n%s", row.Title, row.Reply.Text), AllowIM: true})
	}
	var in protocol.TodoStorageCall
	if json.Unmarshal(req.Body, &in) != nil {
		return nil, errors.New("无效待办请求")
	}
	return a.historyStore().todos(in)
}

func todoReplyTaskID(todoID, replyID string) string {
	return "todo:" + store.Hash(todoID+"\x00"+replyID)
}
