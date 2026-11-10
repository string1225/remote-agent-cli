package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	bolt "go.etcd.io/bbolt"
)

var errTaskInterrupted = errors.New("任务被新消息插队中断")

type queueRecord struct {
	protocol.QueuedMessage
	SessionID     string   `json:"sessionId"`
	ServiceID     string   `json:"serviceId"`
	AttachmentIDs []string `json:"attachmentIds,omitempty"`
	Order         uint64   `json:"order"`
	PriorityOrder uint64   `json:"priorityOrder,omitempty"`
	Recorded      bool     `json:"recorded,omitempty"`
	Model         string   `json:"model,omitempty"`
}

func queueRecordFrom(tx *bolt.Tx, id string) (row queueRecord, found bool, err error) {
	b := tx.Bucket([]byte("task_queue"))
	if b == nil || b.Get([]byte(id)) == nil {
		return
	}
	found = true
	err = json.Unmarshal(b.Get([]byte(id)), &row)
	return
}
func saveQueueRecord(tx *bolt.Tx, row queueRecord) error {
	b, err := tx.CreateBucketIfNotExists([]byte("task_queue"))
	if err != nil {
		return err
	}
	data, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return b.Put([]byte(row.ID), data)
}
func queuedMessagesFrom(tx *bolt.Tx, sessionID string) ([]protocol.QueuedMessage, error) {
	rows := []queueRecord{}
	if b := tx.Bucket([]byte("task_queue")); b != nil {
		if err := b.ForEach(func(_, value []byte) error {
			var row queueRecord
			if err := json.Unmarshal(value, &row); err != nil {
				return err
			}
			if row.SessionID == sessionID && (row.Status == "queued" || row.Status == "paused" || row.Status == "editing") {
				rows = append(rows, row)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].PriorityOrder != rows[j].PriorityOrder {
			return rows[i].PriorityOrder > rows[j].PriorityOrder
		}
		return rows[i].Order < rows[j].Order
	})
	out := make([]protocol.QueuedMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.QueuedMessage)
	}
	return out, nil
}
func (h historyStore) queueRecord(id string) (row queueRecord, found bool, err error) {
	err = h.transact(false, func(tx *bolt.Tx) error { var e error; row, found, e = queueRecordFrom(tx, id); return e })
	return
}
func (h historyStore) setQueueStatus(id, status string) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		row, found, err := queueRecordFrom(tx, id)
		if err != nil || !found {
			return err
		}
		row.Status = status
		return saveQueueRecord(tx, row)
	})
}

func (a *Agent) persistQueuedTask(task queuedTask) error {
	refs, _, _, err := a.resolveAttachments(task.result.Session.ID, task.attachments)
	if err != nil {
		return err
	}
	return a.historyStore().transact(true, func(tx *bolt.Tx) error {
		if task.newSession {
			// Accept identity, prompt, retry receipt and durable queue together.
			if tx.Bucket([]byte("sessions")).Get([]byte(task.result.Session.ID)) != nil {
				return errors.New("会话 ID 已存在")
			}
			if err := saveSession(tx, task.result.Session); err != nil {
				return err
			}
			if _, err := tx.Bucket([]byte("entries")).CreateBucket([]byte(task.result.Session.ID)); err != nil {
				return err
			}
			if _, err := appendHistory(tx, task.result.Session.ID, protocol.ChatEntry{TurnID: task.result.TaskID, Role: "user", Kind: "message", Text: task.start.Prompt}); err != nil {
				return err
			}
			if task.requestKey != "" {
				b, err := tx.CreateBucketIfNotExists([]byte("task_requests"))
				if err != nil {
					return err
				}
				if b.Get([]byte(task.requestKey)) != nil {
					return errors.New("任务请求已被接收")
				}
				data, err := json.Marshal(task.result)
				if err != nil {
					return err
				}
				if err = b.Put([]byte(task.requestKey), data); err != nil {
					return err
				}
			}
		}
		row, found, err := queueRecordFrom(tx, task.result.TaskID)
		if err != nil {
			return err
		}
		if found && (row.SessionID != task.result.Session.ID || row.ServiceID != task.start.ServiceID) {
			return errors.New("任务 ID 已被其他会话使用")
		}
		if !found {
			b, e := tx.CreateBucketIfNotExists([]byte("task_queue"))
			if e != nil {
				return e
			}
			order, e := b.NextSequence()
			if e != nil {
				return e
			}
			row = queueRecord{QueuedMessage: protocol.QueuedMessage{ID: task.result.TaskID, Prompt: task.start.Prompt, Attachments: refs, Source: task.start.Source, CreatedAt: time.Now().UnixMilli()}, SessionID: task.result.Session.ID, ServiceID: task.start.ServiceID, AttachmentIDs: task.attachments, Order: order}
		}
		row.Status, row.Priority, row.Recorded, row.Model = "queued", task.priority, task.recorded, task.start.Model
		if task.priority {
			row.PriorityOrder, err = tx.Bucket([]byte("task_queue")).NextSequence()
			if err != nil {
				return err
			}
		}
		if err := saveQueueRecord(tx, row); err != nil {
			return err
		}
		if err := attachExecutionTurn(tx, task.result, task.start.ServiceID); err != nil {
			return err
		}
		session, err := sessionFrom(tx, row.SessionID)
		if err != nil {
			return err
		}
		session.ArchivedAt = 0
		if session.TaskStatus != "running" && session.TaskStatus != "queued" {
			session.TaskID, session.TaskStatus, session.TaskError = row.ID, "queued", ""
			session.TaskAcceptedAt, session.TaskStartedAt, session.TaskCompletedAt = row.CreatedAt, 0, 0
		}
		return saveSession(tx, session)
	})
}

func (a *Agent) wakeTasksLocked() {
	select {
	case a.taskWake <- struct{}{}:
	default:
	}
}

func (a *Agent) interruptSessionLocked(sessionID string) {
	if id := a.activeTasks[sessionID]; id != "" {
		if a.interruptedTasks == nil {
			a.interruptedTasks = map[string]bool{}
		}
		a.interruptedTasks[id] = true
		if cancel := a.taskCancels[id]; cancel != nil {
			cancel()
		}
	}
	if cancel := a.runs[sessionID]; cancel != nil {
		cancel()
	}
}

func (a *Agent) dispatchTasks() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for sessionID, ids := range a.pendingTasks {
		if a.activeTasks[sessionID] != "" || len(ids) == 0 {
			continue
		}
		id := ids[0]
		a.pendingTasks[sessionID] = ids[1:]
		task, found := a.taskRecords[id]
		if !found {
			a.wakeTasksLocked()
			continue
		}
		a.activeTasks[sessionID] = id
		a.taskWorkers.Add(1)
		go func() { defer a.taskWorkers.Done(); a.finishTask(task) }()
	}
}

func (a *Agent) changeQueuedTask(req protocol.Request) (protocol.QueuedMessage, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	h := a.historyStore()
	row, found, err := h.queueRecord(req.RunID)
	if err != nil {
		return row.QueuedMessage, err
	}
	if !found || row.SessionID != req.SessionID || row.ServiceID != req.ServiceID {
		return row.QueuedMessage, errors.New("排队消息不存在或不属于当前会话")
	}
	session, sessionErr := h.get(row.SessionID)
	if sessionErr != nil || session.DeletedAt != 0 {
		return row.QueuedMessage, errors.New("会话已删除或不可用")
	}
	if req.Type == "tasks.edit" && (row.Status == "editing" || row.Status == "edited") {
		var body struct {
			Acknowledged bool `json:"acknowledged"`
		}
		if json.Unmarshal(req.Body, &body) == nil && body.Acknowledged && row.Status == "editing" {
			if err := h.setQueueStatus(row.ID, "edited"); err != nil {
				return row.QueuedMessage, err
			}
			row.Status = "edited"
		}
		return row.QueuedMessage, nil
	}
	if a.activeTasks[row.SessionID] == row.ID || (row.Status != "queued" && row.Status != "paused") {
		return row.QueuedMessage, errors.New("这条消息已经开始执行，请刷新后再操作")
	}
	if req.Type == "tasks.edit" {
		if err := h.setQueueStatus(row.ID, "editing"); err != nil {
			return row.QueuedMessage, err
		}
		if task, ok := a.taskRecords[row.ID]; ok {
			task.cancel()
		}
		delete(a.taskCancels, row.ID)
		delete(a.taskRecords, row.ID)
		ids := a.pendingTasks[row.SessionID]
		for i, id := range ids {
			if id == row.ID {
				a.pendingTasks[row.SessionID] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
		_ = h.update(row.SessionID, func(s *protocol.Conversation) {
			if s.TaskID == row.ID {
				s.TaskStatus = "cancelled"
			}
		})
		a.wakeTasksLocked()
		row.Status = "editing"
		return row.QueuedMessage, nil
	}
	if task, ok := a.taskRecords[row.ID]; ok {
		task.priority = true
		if err := a.persistQueuedTask(task); err != nil {
			return row.QueuedMessage, err
		}
		a.taskRecords[row.ID] = task
		ids := a.pendingTasks[row.SessionID]
		next := []string{row.ID}
		for _, id := range ids {
			if id != row.ID {
				next = append(next, id)
			}
		}
		a.pendingTasks[row.SessionID] = next
		a.interruptSessionLocked(row.SessionID)
		a.wakeTasksLocked()
	} else {
		if a.taskCtx == nil || a.taskCtx.Err() != nil {
			return row.QueuedMessage, errors.New("客户端正在停止")
		}
		ctx, cancel := context.WithCancel(a.taskCtx)
		task := queuedTask{ctx: ctx, cancel: cancel, priority: true, recorded: row.Recorded, attachments: row.AttachmentIDs, start: protocol.TaskStart{ServiceID: row.ServiceID, Prompt: row.Prompt, Source: row.Source, Model: row.Model}, result: protocol.TaskStarted{TaskID: row.ID, Session: protocol.Conversation{ID: row.SessionID}}}
		if !a.enqueueTaskLocked(task) {
			cancel()
			return row.QueuedMessage, errors.New("客户端任务队列已满，请稍后重试")
		}
	}
	row.Priority, row.Status = true, "queued"
	return row.QueuedMessage, nil
}

func (h historyStore) recoverQueue() error {
	// A read-only startup must preserve existing database bytes when no queued
	// work needs recovery. Even an empty Bolt write transaction changes metadata.
	db, err := bolt.Open(h.path, 0600, &bolt.Options{ReadOnly: true, Timeout: 5 * time.Second})
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	pending := false
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("task_queue"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, data []byte) error {
			var row queueRecord
			if err := json.Unmarshal(data, &row); err != nil {
				return err
			}
			if row.Status == "queued" || row.Status == "running" {
				pending = true
			}
			return nil
		})
	})
	err = errors.Join(err, db.Close())
	if err != nil || !pending {
		return err
	}
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("task_queue"))
		if b == nil {
			return nil
		}
		var rows []queueRecord
		if err := b.ForEach(func(_, data []byte) error {
			var row queueRecord
			if err := json.Unmarshal(data, &row); err != nil {
				return err
			}
			if row.Status == "queued" {
				row.Status = "paused"
				rows = append(rows, row)
			}
			if row.Status == "running" {
				row.Status = "failed"
				rows = append(rows, row)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, row := range rows {
			if err := saveQueueRecord(tx, row); err != nil {
				return err
			}
		}
		return nil
	})
}
