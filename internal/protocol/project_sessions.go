package protocol

// ProjectSessionCreate is shared by the HTTP integration and client relay.
// A message with Execute=false remains paused in the durable task queue.
type ProjectSessionCreate struct {
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
	Execute bool   `json:"execute,omitempty"`
}

// Bytes use JSON base64 on the bounded relay so HTML/control-character escaping
// cannot expand a valid 32000-byte message beyond the 64 KiB request limit.
type ProjectSessionRelay struct {
	Title   string `json:"title,omitempty"`
	Message []byte `json:"message,omitempty"`
	Execute bool   `json:"execute,omitempty"`
}
