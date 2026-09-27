package protocol

// Sessions and their entries travel only over the authenticated DataChannel.
type Conversation struct {
	ID        string `json:"id"`
	ServiceID string `json:"serviceId"`
	Title     string `json:"title"`
	Provider  string `json:"provider"`
	Workspace string `json:"workspace"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	NativeID  string `json:"nativeId,omitempty"`
}

type ChatEntry struct {
	Seq       uint64 `json:"seq"`
	TurnID    string `json:"turnId"`
	Role      string `json:"role"`
	Text      string `json:"text"`
	Kind      string `json:"kind"`
	CreatedAt int64  `json:"createdAt"`
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
