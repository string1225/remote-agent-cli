package protocol

type TaskStart struct {
	RequestID string `json:"requestId,omitempty"`
	// Legacy notification fields are accepted but do not trigger delivery.
	NotifyAgentID   string `json:"notifyAgentId,omitempty"`
	NotifyServiceID string `json:"notifyServiceId,omitempty"`
	NotifySessionID string `json:"notifySessionId,omitempty"`
	Model           string `json:"model,omitempty"`
	ServiceID       string `json:"serviceId"`
	Title           string `json:"title"`
	Prompt          string `json:"prompt"`
	Source          string `json:"source,omitempty"`
	ParentAgentID   string `json:"parentAgentId,omitempty"`
	ParentSessionID string `json:"parentSessionId,omitempty"`
	ActionID        string `json:"actionId,omitempty"`
	Notify          bool   `json:"notify,omitempty"`
}
type TaskStarted struct {
	Session Conversation `json:"session"`
	TaskID  string       `json:"taskId"`
}
