package protocol

// Assistant identities are logical roles. ClientID remains the existing device
// identity; ProjectID and SessionID keep their original meanings.
type AssistantRef struct {
	Online    bool     `json:"online"`
	Path      []string `json:"path,omitempty"`
	Scope     string   `json:"scope"`
	ClientID  string   `json:"clientId"`
	ProjectID string   `json:"projectId,omitempty"`
	SessionID string   `json:"sessionId,omitempty"`
	TurnID    string   `json:"turnId,omitempty"`
	Name      string   `json:"name,omitempty"`
}

type CoordinationCommand struct {
	Operation   string       `json:"operation"`
	AccountID   string       `json:"accountId,omitempty"`
	RequestID   string       `json:"requestId,omitempty"`
	WorkID      string       `json:"workId,omitempty"`
	ReplyHandle string       `json:"replyHandle,omitempty"`
	Source      AssistantRef `json:"source"`
	Target      AssistantRef `json:"target"`
	Title       string       `json:"title,omitempty"`
	Prompt      string       `json:"prompt,omitempty"`
	TaskID      string       `json:"taskId,omitempty"`
	Path        []string     `json:"path,omitempty"`
	Report      *WorkReport  `json:"report,omitempty"`
}

type AssistantResolved struct {
	AssistantRef
	Service Service `json:"service"`
}

type WorkReport struct {
	Related     []WorkLink   `json:"related,omitempty"`
	ID          string       `json:"id"`
	WorkID      string       `json:"workId"`
	ReplyHandle string       `json:"replyHandle"`
	Origin      AssistantRef `json:"origin"`
	TaskID      string       `json:"taskId"`
	Status      string       `json:"status"`
	Summary     string       `json:"summary"`
	Captured    bool         `json:"captured"`
	CreatedAt   int64        `json:"createdAt"`
}

type WorkLink struct {
	WorkID      string       `json:"workId"`
	ReplyHandle string       `json:"replyHandle"`
	Origin      AssistantRef `json:"origin"`
	Status      string       `json:"status"`
}

type WorkAccepted struct {
	WorkID      string       `json:"workId"`
	ReplyHandle string       `json:"replyHandle"`
	Target      AssistantRef `json:"target"`
	TaskID      string       `json:"taskId"`
}
