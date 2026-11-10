package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
	bolt "go.etcd.io/bbolt"
)

// ExecutionTopic contains routing metadata only. TaskID is the last accepted turn;
// read_session remains the authority for live status and final output.
type ExecutionTopic struct {
	Topic     string `json:"topic"`
	ClientID  string `json:"clientId"`
	ProjectID string `json:"projectId"`
	SessionID string `json:"sessionId,omitempty"`
	TaskID    string `json:"taskId,omitempty"`
}

type ExecutionTopicPage struct {
	Topics     []ExecutionTopic `json:"topics"`
	NextCursor uint64           `json:"nextCursor,omitempty"`
}

type ExecutionTopicResult struct {
	protocol.TaskStarted
	Mapping   ExecutionTopic `json:"mapping"`
	Recovered bool           `json:"recovered,omitempty"`
}

type executionTopicRecord struct {
	Subscribed bool `json:"subscribed,omitempty"`
	ExecutionTopic
	Owner          string               `json:"owner"`
	Initial        protocol.TaskStart   `json:"initial"`
	CreatedRequest string               `json:"createdRequest"`
	LastRequest    string               `json:"lastRequest"`
	LastPromptHash string               `json:"lastPromptHash"`
	Result         protocol.TaskStarted `json:"result"`
}

func executionTopicStore(c Config, project, session string) (historyStore, string, error) {
	h := historyStore{filepath.Join(filepath.Dir(c.Path), "history.db")}
	manager := false
	for _, service := range c.Services {
		if service.ID == project && isManager(service) {
			manager = true
		}
	}
	if !manager || c.ID == "" || session == "" {
		return h, "", errors.New("主题调度仅用于真实的 CEO 或管家会话；普通执行 session 直接执行任务")
	}
	s, err := h.get(session)
	if err != nil || s.DeletedAt != 0 || s.ServiceID != project {
		return h, "", errors.New("管理会话不属于当前项目")
	}
	return h, store.Hash(c.ID + "\x00" + project + "\x00" + session), nil
}

func executionTopicKey(owner, topic string) (string, error) {
	topic = strings.ToLower(strings.Join(strings.Fields(topic), " "))
	if topic == "" || len(topic) > 320 {
		return "", errors.New("管理角色创建或续办执行会话必须指定 topic（1–320 字节）")
	}
	return store.Hash(owner + "\x00" + topic), nil
}

// ListExecutionTopics restores only this manager session's metadata after a
// restart/context compaction; it never reads another session's transcript.
func ListExecutionTopics(c Config, project, session string, cursor uint64) (ExecutionTopicPage, error) {
	page := ExecutionTopicPage{Topics: []ExecutionTopic{}}
	h, owner, err := executionTopicStore(c, project, session)
	if err != nil {
		return page, err
	}
	rows, err := h.coordinationRows("execution_topics")
	if err != nil {
		return page, err
	}
	all := []ExecutionTopic{}
	for _, data := range rows {
		var row executionTopicRecord
		if err = json.Unmarshal(data, &row); err != nil {
			return page, err
		}
		if row.Owner == owner {
			all = append(all, row.ExecutionTopic)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Topic < all[j].Topic })
	if cursor > uint64(len(all)) {
		return page, errors.New("无效主题分页游标")
	}
	end := min(cursor+50, uint64(len(all)))
	page.Topics = append(page.Topics, all[cursor:end]...)
	if end < uint64(len(all)) {
		page.NextCursor = end
	}
	return page, nil
}

func ordinaryExecutionProject(ctx context.Context, c Config, client, project string) error {
	if client == c.ID {
		for _, service := range c.Services {
			if service.ID == project && service.Scope == "" {
				return nil
			}
		}
	} else {
		data, err := ClientAPI(ctx, c, "GET", "/api/agent/catalog", nil)
		if err != nil {
			return err
		}
		var catalog struct {
			Clients []struct {
				ID       string             `json:"id"`
				Projects []protocol.Service `json:"projects"`
			} `json:"clients"`
		}
		if err = json.Unmarshal(data, &catalog); err != nil {
			return err
		}
		for _, device := range catalog.Clients {
			for _, service := range device.Projects {
				if device.ID == client && service.ID == project && service.Scope == "" {
					return nil
				}
			}
		}
	}
	return errors.New("执行会话必须指向原普通项目，不能指向 CEO 或管家专用项目")
}

func saveExecutionTopic(tx *bolt.Tx, key string, row executionTopicRecord) error {
	b, err := tx.CreateBucketIfNotExists([]byte("execution_topics"))
	if err != nil {
		return err
	}
	if row.SessionID != "" {
		if err := b.ForEach(func(otherKey, data []byte) error {
			if string(otherKey) == key {
				return nil
			}
			var other executionTopicRecord
			if err := json.Unmarshal(data, &other); err != nil {
				return err
			}
			if other.Owner == row.Owner && other.ClientID == row.ClientID && other.SessionID == row.SessionID {
				return errors.New("该执行 session 已属于其他 topic；不同方向须保留独立会话")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	data, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), data)
}

// Claim before network I/O. Concurrent MCP processes and a crash before receipt
// share one immutable target and stable spawn ID; no transaction spans a relay.
func claimExecutionTopic(h historyStore, key string, proposed executionTopicRecord) (row executionTopicRecord, err error) {
	err = h.transact(true, func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("execution_topics"))
		if b != nil && b.Get([]byte(key)) != nil {
			if err := json.Unmarshal(b.Get([]byte(key)), &row); err != nil {
				return err
			}
			if row.Owner != proposed.Owner || row.ClientID != proposed.ClientID || row.ProjectID != proposed.ProjectID || proposed.SessionID != "" && row.SessionID != proposed.SessionID {
				return fmt.Errorf("topic 已绑定 %s/%s/%s；保留原映射，不自动换目标", row.ClientID, row.ProjectID, row.SessionID)
			}
			return nil
		}
		row = proposed
		return saveExecutionTopic(tx, key, row)
	})
	return
}

// StartExecutionTopic creates once or continues the mapped ordinary session.
// Busy/deleted/offline sessions are errors, never a reason to spawn a replacement.
func startExecutionTopic(ctx context.Context, c Config, project, session, request, topic, client string, in protocol.TaskStart) (out ExecutionTopicResult, err error) {
	h, owner, err := executionTopicStore(c, project, session)
	if err != nil {
		return out, err
	}
	key, err := executionTopicKey(owner, topic)
	if err != nil {
		return out, err
	}
	if request == "" || strings.TrimSpace(in.Prompt) == "" || len(in.Prompt) > 32000 {
		return out, errors.New("主题调度需要稳定请求 ID 和 1–32000 字节任务内容")
	}
	if err = ordinaryExecutionProject(ctx, c, client, in.ServiceID); err != nil {
		return out, err
	}
	in.RequestID, in.ParentAgentID, in.ParentSessionID = key, c.ID, session
	// Topic identity is explicit and independent from a turn's display title.
	proposed := executionTopicRecord{ExecutionTopic: ExecutionTopic{Topic: strings.Join(strings.Fields(topic), " "), ClientID: client, ProjectID: in.ServiceID}, Owner: owner, Initial: in, CreatedRequest: request}
	row, err := claimExecutionTopic(h, key, proposed)
	if err != nil {
		return out, err
	}
	out.Mapping = row.ExecutionTopic
	if row.CreatedRequest == request && row.Initial.Prompt != in.Prompt {
		return out, errors.New("同一请求 ID 不能修改任务内容")
	}
	if row.LastRequest == request {
		if row.LastPromptHash != store.Hash(in.Prompt) {
			return out, errors.New("同一请求 ID 不能修改任务内容")
		}
		out.TaskStarted = row.Result
		return out, nil
	}
	var data json.RawMessage
	if row.SessionID == "" {
		// Retry the original prompt after uncertain delivery. A new follow-up is
		// not silently substituted for work that may already be running.
		data, err = ClientTool(ctx, c, client, "sessions.spawn", row.Initial)
		out.Recovered = row.CreatedRequest != request
	} else if row.CreatedRequest == request {
		if row.Initial.Prompt != in.Prompt {
			return out, errors.New("同一请求 ID 不能修改任务内容")
		}
		data, err = ClientTool(ctx, c, client, "sessions.spawn", row.Initial)
	} else {
		data, err = ClientTool(ctx, c, client, "sessions.control", protocol.SessionControl{Operation: "send", CallerSessionID: session, SessionID: row.SessionID, RequestID: request, Prompt: in.Prompt})
	}
	if err != nil {
		return out, fmt.Errorf("topic=%s，目标=%s/%s/%s：%w；保留映射，读取/等待原会话后续办", row.Topic, client, row.ProjectID, row.SessionID, err)
	}
	if err = json.Unmarshal(data, &out.TaskStarted); err != nil {
		return out, err
	}
	if out.Session.ID == "" || out.TaskID == "" || out.Session.ServiceID != row.ProjectID || row.SessionID != "" && out.Session.ID != row.SessionID {
		return out, errors.New("执行器返回无效会话，原映射已保留")
	}
	out.Session = publicSession(out.Session)
	err = h.transact(true, func(tx *bolt.Tx) error {
		var current executionTopicRecord
		if err := json.Unmarshal(tx.Bucket([]byte("execution_topics")).Get([]byte(key)), &current); err != nil {
			return err
		}
		if current.SessionID != "" && row.SessionID == "" {
			out.Mapping = current.ExecutionTopic
			return nil
		}
		current.SessionID, current.TaskID = out.Session.ID, out.TaskID
		current.LastRequest, current.LastPromptHash, current.Result = request, store.Hash(in.Prompt), out.TaskStarted
		// A recovered first dispatch acknowledges its original request, so the
		// follow-up can be submitted explicitly after checking the recovered task.
		if out.Recovered {
			current.LastRequest, current.LastPromptHash = current.CreatedRequest, store.Hash(current.Initial.Prompt)
		}
		out.Mapping = current.ExecutionTopic
		return saveExecutionTopic(tx, key, current)
	})
	return out, err
}

// BindExecutionTopic adopts an existing direct child using metadata pagination.
// It neither reads transcripts nor starts a turn, and cannot overwrite a binding.
func BindExecutionTopic(ctx context.Context, c Config, project, session, topic, client, targetProject, child string) (ExecutionTopic, error) {
	h, owner, err := executionTopicStore(c, project, session)
	if err != nil {
		return ExecutionTopic{}, err
	}
	key, err := executionTopicKey(owner, topic)
	if err != nil {
		return ExecutionTopic{}, err
	}
	if err = ordinaryExecutionProject(ctx, c, client, targetProject); err != nil {
		return ExecutionTopic{}, err
	}
	for cursor := uint64(0); ; {
		data, err := ClientTool(ctx, c, client, "sessions.control", protocol.SessionControl{Operation: "list", CallerSessionID: session, Cursor: cursor})
		if err != nil {
			return ExecutionTopic{}, err
		}
		var page protocol.ConversationPage
		if err = json.Unmarshal(data, &page); err != nil {
			return ExecutionTopic{}, err
		}
		for _, s := range page.Sessions {
			if s.ID == child && s.ServiceID == targetProject && s.DeletedAt == 0 && s.Kind == "web" && s.ParentAgentID == c.ID && s.ParentSessionID == session {
				row, err := claimExecutionTopic(h, key, executionTopicRecord{ExecutionTopic: ExecutionTopic{Topic: strings.Join(strings.Fields(topic), " "), ClientID: client, ProjectID: targetProject, SessionID: child, TaskID: s.TaskID}, Owner: owner})
				return row.ExecutionTopic, err
			}
		}
		if page.NextCursor == 0 {
			break
		}
		if page.NextCursor <= cursor {
			return ExecutionTopic{}, errors.New("无效子会话分页游标")
		}
		cursor = page.NextCursor
	}
	return ExecutionTopic{}, errors.New("找不到属于当前管理会话的原项目直接子会话；未创建或改写映射")
}
