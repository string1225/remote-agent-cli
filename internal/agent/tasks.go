package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

type queuedTask struct {
	recorded bool
	ctx      context.Context
	cancel   context.CancelFunc
	start    protocol.TaskStart
	result   protocol.TaskStarted
	complete func(error)
}

func (h historyStore) update(id string, change func(*protocol.Conversation)) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		v, err := sessionFrom(tx, id)
		if err != nil {
			return err
		}
		change(&v)
		return saveSession(tx, v)
	})
}
func (a *Agent) initTasks(ctx context.Context) {
	a.tasksOnce.Do(func() {
		a.taskCtx = ctx
		a.recoverTasks()
		a.initNotifications(ctx)
		a.tasks = make(chan queuedTask, 32)
		a.taskDone = make(chan struct{})
		a.taskCancels = map[string]context.CancelFunc{}
		go a.workTasks(ctx)
	})
}
func (a *Agent) startTask(ctx context.Context, in protocol.TaskStart, complete func(error)) (protocol.TaskStarted, error) {
	if err := a.beginOperation(); err != nil {
		return protocol.TaskStarted{}, err
	}
	defer a.endOperation()
	a.initTasks(ctx)
	var result protocol.TaskStarted
	if in.Model != "" && !soModelPattern.MatchString(in.Model) {
		return result, errors.New("无效模型")
	}
	if strings.TrimSpace(in.Prompt) == "" || len(in.Prompt) > 32000 || len([]rune(in.Title)) > 80 {
		return result, errors.New("任务内容需为 1–32000 字节，标题最多 80 字")
	}
	if in.Source != "web" && in.Source != "im" && in.Source != "sayso" {
		in.Source = "web"
	}
	var service protocol.Service
	a.mu.Lock()
	for _, s := range a.Config.Services {
		if s.ID == in.ServiceID {
			service = s
		}
	}
	provider := a.Providers[service.Provider]
	a.mu.Unlock()
	if service.ID == "" || !provider.Available {
		return result, errors.New("项目不存在或执行器不可用")
	}
	root, err := a.Config.Workspace(service.Workspace)
	if err != nil {
		return result, err
	}
	service.Workspace = root
	if a.taskCtx.Err() != nil {
		return result, errors.New("客户端正在停止")
	}
	if strings.TrimSpace(in.Title) == "" {
		in.Title = soBound(in.Prompt, 80)
	}
	session, err := a.historyStore().create(service, in.Title)
	if err != nil {
		return result, err
	}
	session.Kind = "web"
	session.Source = in.Source
	session.ParentAgentID = in.ParentAgentID
	session.ParentSessionID = in.ParentSessionID
	session.ActionID = in.ActionID
	if in.Notify {
		if in.NotifyAgentID == "" {
			in.NotifyAgentID = a.Config.ID
		}
		if in.NotifyServiceID == "" {
			in.NotifyServiceID = in.ServiceID
		}
		if in.NotifySessionID == "" && in.ParentAgentID == a.Config.ID && in.ParentSessionID != "" {
			if parent, e := a.historyStore().get(in.ParentSessionID); e == nil {
				if parent.Kind == "im" && in.NotifyAgentID == a.Config.ID && parent.ServiceID == in.NotifyServiceID {
					in.NotifySessionID = parent.ID
				} else if parent.NotifyAgentID == in.NotifyAgentID && parent.NotifyServiceID == in.NotifyServiceID {
					in.NotifySessionID = parent.NotifySessionID
				}
			}
		}
	}
	session.NotifyAgentID = in.NotifyAgentID
	session.NotifyServiceID = in.NotifyServiceID
	session.NotifySessionID = in.NotifySessionID
	session.TaskStatus = "queued"
	session.TaskID = store.Token()
	if err = a.historyStore().update(session.ID, func(v *protocol.Conversation) { *v = session }); err != nil {
		return result, err
	}
	result = protocol.TaskStarted{Session: session, TaskID: session.TaskID}
	if _, err = a.historyStore().append(session.ID, protocol.ChatEntry{TurnID: result.TaskID, Role: "user", Text: in.Prompt, Kind: "message"}); err != nil {
		return result, err
	}
	taskCtx, taskCancel := context.WithCancel(a.taskCtx)
	a.mu.Lock()
	a.taskCancels[result.TaskID] = taskCancel
	a.mu.Unlock()
	select {
	case a.tasks <- queuedTask{ctx: taskCtx, cancel: taskCancel, recorded: true, start: in, result: result, complete: complete}:
		return result, nil
	default:
		taskCancel()
		a.mu.Lock()
		delete(a.taskCancels, result.TaskID)
		a.mu.Unlock()
		_ = a.historyStore().update(session.ID, func(v *protocol.Conversation) { v.TaskStatus = "failed" })
		return result, errors.New("客户端任务队列已满，请稍后重试")
	}
}
func (a *Agent) workTasks(ctx context.Context) {
	defer close(a.taskDone)
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-a.tasks:
			err := a.workTask(task.ctx, task)
			task.cancel()
			a.mu.Lock()
			delete(a.taskCancels, task.result.TaskID)
			a.mu.Unlock()
			status := "done"
			if err != nil {
				status = "failed"
				if errors.Is(err, context.Canceled) {
					status = "cancelled"
				}
			}
			_ = a.historyStore().update(task.result.Session.ID, func(v *protocol.Conversation) { v.TaskStatus = status })
			a.notifyTask(task, err)
			if task.complete != nil {
				task.complete(err)
			}
		}
	}
}
func (a *Agent) workTask(ctx context.Context, task queuedTask) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		a.mu.Lock()
		busy := a.busy
		a.mu.Unlock()
		if busy {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				continue
			}
		}
		turnCtx, cancel := context.WithCancel(ctx)
		p := &peer{ctx: turnCtx, cancel: cancel}
		finished := make(chan protocol.Event, 1)
		a.handle(p, protocol.Request{ID: task.result.TaskID, Type: "run", ServiceID: task.start.ServiceID, SessionID: task.result.Session.ID, Prompt: task.start.Prompt, Service: protocol.Service{Model: task.start.Model}, Recorded: task.recorded}, func(e protocol.Event) error {
			if e.Type == "started" {
				return a.historyStore().update(task.result.Session.ID, func(v *protocol.Conversation) { v.TaskStatus = "running"; v.TaskID = task.result.TaskID })
			}
			if e.Type == "done" || e.Type == "error" {
				select {
				case finished <- e:
				default:
				}
			}
			return nil
		})
		select {
		case <-ctx.Done():
			p.close()
			select {
			case <-finished:
			case <-time.After(6 * time.Second):
			}
			return ctx.Err()
		case e := <-finished:
			p.close()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if e.Error == "this device is already running a task" {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(200 * time.Millisecond):
					continue
				}
			}
			if e.Error != "" {
				return errors.New(e.Error)
			}
			return nil
		}
	}
}
func (a *Agent) handleTask(p *peer, req protocol.Request, send func(protocol.Event) error) {
	var in protocol.TaskStart
	if json.Unmarshal(req.Body, &in) != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: "invalid task request"})
		return
	}
	result, err := a.startTask(p.ctx, in, nil)
	if err != nil {
		_ = send(protocol.Event{ID: req.ID, Type: "error", Error: err.Error()})
		return
	}
	_ = send(protocol.Event{ID: req.ID, Type: "result", Data: result})
}

func (a *Agent) cancelTask(id string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	cancel := a.taskCancels[id]
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

// Interrupted work is retained for review and manual continuation, never blindly
// replayed after an upgrade or crash. Read-only inspection leaves unrelated
// existing history bytes untouched (including pre-MSI fixture data).
func (a *Agent) recoverTasks() {
	h := a.historyStore()
	if _, err := os.Stat(h.path); err != nil {
		return
	}
	db, err := bolt.Open(h.path, 0600, &bolt.Options{ReadOnly: true, Timeout: time.Second})
	if err != nil {
		return
	}
	var interrupted []protocol.Conversation
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("sessions"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, data []byte) error {
			var v protocol.Conversation
			if json.Unmarshal(data, &v) == nil && (v.TaskStatus == "queued" || v.TaskStatus == "running") {
				interrupted = append(interrupted, v)
			}
			return nil
		})
	})
	db.Close()
	for _, v := range interrupted {
		_ = h.update(v.ID, func(s *protocol.Conversation) { s.TaskStatus = "failed" })
		_, _ = h.append(v.ID, protocol.ChatEntry{TurnID: v.TaskID, Role: "system", Kind: "error", Text: "客户端重启，之前的任务已中断。请检查已完成内容后再继续。"})
	}
}
