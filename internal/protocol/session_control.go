package protocol

// CallerAgentID is always replaced by the authenticated relay server.
type SessionControl struct {
	Operation       string `json:"operation"`
	CallerAgentID   string `json:"callerAgentId"`
	CallerSessionID string `json:"callerSessionId"`
	SessionID       string `json:"sessionId,omitempty"`
	TaskID          string `json:"taskId,omitempty"`
	RequestID       string `json:"requestId,omitempty"`
	Prompt          string `json:"prompt,omitempty"`
	Cursor          uint64 `json:"cursor,omitempty"`
	WaitSeconds     int    `json:"waitSeconds,omitempty"`
}

type SessionState struct {
	Session    Conversation `json:"session"`
	TaskID     string       `json:"taskId"`
	Result     string       `json:"result"`
	Complete   bool         `json:"complete"`
	Captured   bool         `json:"captured"`
	NextCursor uint64       `json:"nextCursor,omitempty"`
}
