package protocol

// ChannelInfo is safe to include in a device heartbeat and account catalog.
// App IDs, secrets and destination identities stay in the client configuration.
type ChannelInfo struct {
	Provider  string `json:"provider"`
	SessionID string `json:"sessionId"`
}
