package protocol

// Sessions and their entries travel over the authenticated server relay.
type Conversation struct {
	ChannelProvider string `json:"channelProvider,omitempty"`
	ChannelChatID   string `json:"channelChatId,omitempty"`
	ChannelChatType string `json:"channelChatType,omitempty"`
	ChannelName     string `json:"channelName,omitempty"`
	ChannelPeerName string `json:"channelPeerName,omitempty"`
	NotifyAgentID   string `json:"notifyAgentId,omitempty"`
	NotifyServiceID string `json:"notifyServiceId,omitempty"`
	NotifySessionID string `json:"notifySessionId,omitempty"`
	TaskID          string `json:"taskId,omitempty"`
	ID              string `json:"id"`
	ServiceID       string `json:"serviceId"`
	Title           string `json:"title"`
	Provider        string `json:"provider"`
	Workspace       string `json:"workspace"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
	NativeID        string `json:"nativeId,omitempty"`
	Kind            string `json:"kind,omitempty"`
	Source          string `json:"source,omitempty"`
	ParentAgentID   string `json:"parentAgentId,omitempty"`
	ParentSessionID string `json:"parentSessionId,omitempty"`
	ActionID        string `json:"actionId,omitempty"`
	TaskStatus      string `json:"taskStatus,omitempty"`
}

type ChatEntry struct {
	Attachments []Attachment `json:"attachments,omitempty"`
	Seq         uint64       `json:"seq"`
	TurnID      string       `json:"turnId"`
	Role        string       `json:"role"`
	Text        string       `json:"text"`
	Kind        string       `json:"kind"`
	CreatedAt   int64        `json:"createdAt"`
}

type ConversationPage struct {
	Sessions   []Conversation `json:"sessions"`
	NextCursor uint64         `json:"nextCursor"`
}

type HistoryPage struct {
	Session    Conversation `json:"session"`
	Entries    []ChatEntry  `json:"entries"`
	NextCursor uint64       `json:"nextCursor"`
}
