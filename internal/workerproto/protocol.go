package workerproto

const Version = 1

type Message struct {
	Version  string `json:"version,omitempty"`
	Protocol int    `json:"protocol"`
	Type     string `json:"type"`
	Force    bool   `json:"force,omitempty"`
	Error    string `json:"error,omitempty"`
}
