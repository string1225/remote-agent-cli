package protocol

// IM recipient discovery exposes routing metadata, never messages or platform
// credentials. A recipient name may identify more than one session.
type IMSessionQuery struct {
	ServiceID string `json:"serviceId,omitempty"`
	Search    string `json:"search,omitempty"`
	Cursor    uint64 `json:"cursor,omitempty"`
}

type IMSessionRecipient struct {
	SessionID   string `json:"sessionId"`
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Provider    string `json:"provider"`
	ChannelName string `json:"channelName"`
	ChatType    string `json:"chatType"`
	PeerName    string `json:"peerName"`
	NameKnown   bool   `json:"nameKnown"`
	Title       string `json:"title"`
}

type IMSessionPage struct {
	ClientID   string               `json:"clientId"`
	Sessions   []IMSessionRecipient `json:"sessions"`
	NextCursor uint64               `json:"nextCursor"`
}
