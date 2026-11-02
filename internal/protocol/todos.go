package protocol

// Todos belong to an account, independent of projects. Routing is optional for
// personal notes and mandatory for replies that continue an agent session.
type Todo struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Status    string     `json:"status"`
	ClientID  string     `json:"clientId,omitempty"`
	ProjectID string     `json:"projectId,omitempty"`
	SessionID string     `json:"sessionId,omitempty"`
	Version   uint64     `json:"version"`
	CreatedAt int64      `json:"createdAt"`
	UpdatedAt int64      `json:"updatedAt"`
	Reply     *TodoReply `json:"reply,omitempty"`
}

type TodoReply struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	State  string `json:"state"`
	TaskID string `json:"taskId,omitempty"`
}

type TodoCommand struct {
	Operation string `json:"operation"`
	ID        string `json:"id,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Version   uint64 `json:"version,omitempty"`
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	Status    string `json:"status,omitempty"`
	Message   string `json:"message,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Search    string `json:"search,omitempty"`
}

type TodoPage struct {
	Todos      []Todo `json:"todos"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Account and generation are supplied only by the authenticated control site.
type TodoStorageCall struct {
	AccountID   string `json:"accountId"`
	Generation  string `json:"generation"`
	MigrationID string `json:"migrationId,omitempty"`
	TodoCommand
	Records []TodoRecord `json:"records,omitempty"`
}
type TodoRecord struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}
type TodoExport struct {
	Records    []TodoRecord `json:"records"`
	NextCursor string       `json:"nextCursor,omitempty"`
}
