package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

func (a *Agent) signalCoordination() {
	a.mu.Lock()
	defer a.mu.Unlock()
	select {
	case a.coordinationWake <- struct{}{}:
	default:
	}
}

func (w coordinationWork) pendingKey() string {
	key := w.Command.WorkID + "/" + w.Accepted.TaskID
	if w.EventKey != "" {
		key += "/" + w.EventKey
	}
	return key
}

func reportWakeID(row coordinationInbox) string {
	if row.WakeAttempt == 0 {
		return store.Hash("report\x00" + row.Report.ID)
	}
	return store.Hash(fmt.Sprintf("report\x00%s\x00%d", row.Report.ID, row.WakeAttempt))
}

func (h historyStore) initializeWorkReport(w *coordinationWork) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("coordination_pending"))
		if b == nil {
			return errors.New("outbox missing")
		}
		var latest coordinationWork
		if err := json.Unmarshal(b.Get([]byte(w.pendingKey())), &latest); err != nil {
			return err
		}
		if latest.Report != nil {
			*w = latest
			return nil
		}
		data, err := json.Marshal(w)
		if err != nil {
			return err
		}
		return b.Put([]byte(w.pendingKey()), data)
	})
}

// Shares the existing work/pending/report/inbox chain. No second delivery queue.
func (a *Agent) watchExecution(in protocol.CoordinationCommand) (any, error) {
	h := a.historyStore()
	s, err := h.get(in.Target.SessionID)
	if err != nil || s.DeletedAt != 0 || s.ServiceID != in.Target.ProjectID || s.ParentAgentID != in.Source.ClientID || s.ParentSessionID != in.Source.SessionID || s.Kind != "web" || in.Target.ClientID != a.Config.ID || len(in.WorkID) != 64 {
		return nil, errors.New("执行订阅不属于该管理会话的直接子会话")
	}
	ordinary := false
	for _, p := range a.services() {
		if p.ID == s.ServiceID && p.Scope == "" {
			ordinary = true
		}
	}
	if !ordinary {
		return nil, errors.New("只能订阅原普通项目执行会话")
	}
	w := coordinationWork{Direct: true, Backfill: s.TaskStatus != "running" && s.TaskStatus != "queued", Command: in, Accepted: protocol.WorkAccepted{WorkID: in.WorkID, ReplyHandle: in.ReplyHandle, Target: in.Target, TaskID: s.TaskID}}
	err = h.transact(true, func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists([]byte("coordination_work"))
		if e != nil {
			return e
		}
		if data := b.Get([]byte(in.WorkID)); data != nil {
			var old coordinationWork
			if json.Unmarshal(data, &old) != nil || !old.Direct || old.Command.AccountID != in.AccountID || old.Command.Source.SessionID != in.Source.SessionID {
				return errors.New("订阅身份冲突")
			}
		}
		data, e := json.Marshal(w)
		if e != nil {
			return e
		}
		if e = b.Put([]byte(in.WorkID), data); e != nil {
			return e
		}
		return saveExecutionPending(tx, w)
	})
	return w.Accepted, err
}

func saveExecutionPending(tx *bolt.Tx, w coordinationWork) error {
	if w.Accepted.TaskID == "" {
		return nil
	}
	b, err := tx.CreateBucketIfNotExists([]byte("coordination_pending"))
	if err != nil {
		return err
	}
	if b.Get([]byte(w.pendingKey())) != nil {
		return nil
	}
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}
	return b.Put([]byte(w.pendingKey()), data)
}

// In the same acceptance transaction as the queue row, so every follow-up turn
// is subscribed even if the dispatching MCP process crashes before its receipt.
func attachExecutionTurn(tx *bolt.Tx, task protocol.TaskStarted, project string) error {
	b := tx.Bucket([]byte("coordination_work"))
	if b == nil {
		return nil
	}
	return b.ForEach(func(_, data []byte) error {
		var w coordinationWork
		if err := json.Unmarshal(data, &w); err != nil {
			return err
		}
		if !w.Direct || w.Accepted.Target.SessionID != task.Session.ID || w.Accepted.Target.ProjectID != project {
			return nil
		}
		w.Accepted.TaskID, w.Report, w.Delivered = task.TaskID, nil, false
		w.Backfill = false
		return saveExecutionPending(tx, w)
	})
}

func (a *Agent) executionCheckpoint(in protocol.CoordinationCommand) (any, error) {
	h := a.historyStore()
	var w coordinationWork
	found, err := h.coordinationGet("coordination_work", in.WorkID, &w)
	if err != nil || !found || !w.Direct || w.Command.AccountID != in.AccountID || w.Accepted.Target.SessionID != in.Target.SessionID {
		return nil, errors.New("执行事件订阅不存在或归属不匹配")
	}
	s, err := h.get(in.Target.SessionID)
	if err != nil || s.TaskID != in.TaskID || s.DeletedAt != 0 || s.TaskStatus != "running" {
		return nil, errors.New("事件轮次已过期或未运行")
	}
	if in.Stage != "plan" && in.Stage != "validation" && in.Stage != "predeploy" {
		return nil, errors.New("检查点仅支持 plan/validation/predeploy")
	}
	if in.Status != "checkpoint" && in.Status != "review_required" && in.Status != "blocked" && in.Status != "needs_input" {
		return nil, errors.New("无效事件状态")
	}
	if strings.TrimSpace(in.Summary) == "" || len(in.Summary) > 1800 || len(in.Evidence) > 4 {
		return nil, errors.New("事件需1–1800字节摘要、最多4项证据")
	}
	for _, e := range in.Evidence {
		if len(e) > 512 {
			return nil, errors.New("单项证据最多512字节")
		}
	}
	if in.Status == "checkpoint" {
		enabled := false
		for _, stage := range w.Command.Checkpoints {
			if stage == in.Stage {
				enabled = true
			}
		}
		if !enabled {
			return map[string]any{"recorded": false, "reason": "checkpoint not enabled", "endTurn": false}, nil
		}
	}
	w.Accepted.TaskID, w.Delivered = in.TaskID, false
	key := in.Stage
	if in.Status != "checkpoint" {
		key = "review"
	} // at most one stopping event per turn
	w.Report = &protocol.WorkReport{TaskStatus: "running", ID: store.Hash(w.Command.WorkID + "\x00" + in.TaskID + "\x00" + key), WorkID: w.Command.WorkID, ReplyHandle: w.Command.ReplyHandle, Origin: w.Accepted.Target, TaskID: in.TaskID, Topic: w.Command.Topic, Event: in.Status, Stage: in.Stage, Status: in.Status, Summary: in.Summary, Evidence: in.Evidence, CreatedAt: time.Now().UnixMilli()}
	if in.Status == "checkpoint" {
		w.EventKey = w.Report.ID
	}
	err = h.transact(true, func(tx *bolt.Tx) error {
		current, e := sessionFrom(tx, in.Target.SessionID)
		if e != nil || current.TaskID != in.TaskID || current.TaskStatus != "running" {
			return errors.New("检查点轮次已结束")
		}
		b, e := tx.CreateBucketIfNotExists([]byte("coordination_pending"))
		if e != nil {
			return e
		}
		if data := b.Get([]byte(w.pendingKey())); data != nil {
			var old coordinationWork
			if e := json.Unmarshal(data, &old); e != nil {
				return e
			}
			if old.Report != nil {
				if old.Report.ID != w.Report.ID || old.Report.Event != w.Report.Event || old.Report.Stage != w.Report.Stage || old.Report.Summary != w.Report.Summary || strings.Join(old.Report.Evidence, "\x00") != strings.Join(w.Report.Evidence, "\x00") {
					return errors.New("检查点已记录，不能用同一事件改写内容")
				}
				return nil
			}
		}
		data, e := json.Marshal(w)
		if e != nil {
			return e
		}
		return b.Put([]byte(w.pendingKey()), data)
	})
	return map[string]any{"eventId": w.Report.ID, "recorded": err == nil, "endTurn": in.Status != "checkpoint", "delivery": "review events are delivered after this execution turn exits"}, err
}

// MCP writes the original topic mapping before it subscribes. Startup/retry can
// reconstruct the manager identity from session metadata, including pre-upgrade
// bind_execution_topic rows, without reading any transcript or inventing IDs.
func (a *Agent) restoreExecutionSubscriptions(ctx context.Context) {
	c := a.configSnapshot()
	h := a.historyStore()
	rows, _ := h.coordinationRows("execution_topics")
	sessions, _ := h.coordinationRows("sessions")
	for _, data := range rows {
		var row executionTopicRecord
		if json.Unmarshal(data, &row) != nil || row.Subscribed || row.SessionID == "" {
			continue
		}
		for _, metadata := range sessions {
			var s protocol.Conversation
			if json.Unmarshal(metadata, &s) != nil || s.DeletedAt != 0 || store.Hash(c.ID+"\x00"+s.ServiceID+"\x00"+s.ID) != row.Owner {
				continue
			}
			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			_ = ensureExecutionSubscription(callCtx, c, s.ServiceID, s.ID, row)
			cancel()
			break
		}
	}
}

func ensureExecutionSubscription(ctx context.Context, c Config, project, session string, row executionTopicRecord) error {
	_, err := ClientAPI(ctx, c, "POST", "/api/agent/coordination", protocol.CoordinationCommand{Operation: "subscribe", Topic: row.Topic, Checkpoints: row.Initial.Checkpoints, Source: protocol.AssistantRef{ClientID: c.ID, ProjectID: project, SessionID: session}, Target: protocol.AssistantRef{ClientID: row.ClientID, ProjectID: row.ProjectID, SessionID: row.SessionID}})
	if err != nil {
		return fmt.Errorf("原执行session已保留，事件订阅待恢复（不要重复创建）：%w", err)
	}
	h, owner, err := executionTopicStore(c, project, session)
	if err != nil {
		return err
	}
	key, err := executionTopicKey(owner, row.Topic)
	if err != nil {
		return err
	}
	return h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("execution_topics"))
		var current executionTopicRecord
		if err := json.Unmarshal(b.Get([]byte(key)), &current); err != nil {
			return err
		}
		current.Subscribed = true
		return saveExecutionTopic(tx, key, current)
	})
}

func StartExecutionTopic(ctx context.Context, c Config, project, session, request, topic, client string, in protocol.TaskStart) (ExecutionTopicResult, error) {
	if len(in.Checkpoints) > 3 {
		return ExecutionTopicResult{}, errors.New("最多3个关键检查点")
	}
	for _, stage := range in.Checkpoints {
		if stage != "plan" && stage != "validation" && stage != "predeploy" {
			return ExecutionTopicResult{}, errors.New("检查点仅支持plan/validation/predeploy")
		}
	}
	out, err := startExecutionTopic(ctx, c, project, session, request, topic, client, in)
	if err != nil {
		return out, err
	}
	h, owner, err := executionTopicStore(c, project, session)
	if err != nil {
		return out, err
	}
	key, _ := executionTopicKey(owner, topic)
	var row executionTopicRecord
	_, err = h.coordinationGet("execution_topics", key, &row)
	if err == nil && in.Checkpoints != nil {
		err = h.transact(true, func(tx *bolt.Tx) error {
			var current executionTopicRecord
			if err := json.Unmarshal(tx.Bucket([]byte("execution_topics")).Get([]byte(key)), &current); err != nil {
				return err
			}
			current.Initial.Checkpoints = append([]string{}, in.Checkpoints...)
			row = current
			return saveExecutionTopic(tx, key, current)
		})
	}
	if err == nil {
		err = ensureExecutionSubscription(ctx, c, project, session, row)
	}
	return out, err
}
