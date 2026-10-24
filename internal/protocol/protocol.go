package protocol

import "encoding/json"

type Service struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	Workspace    string `json:"workspace"`
	ResumeID     string `json:"-"`
	Model        string `json:"-"`
	OutputSchema string `json:"-"`
}

const RelayTransport = "relay-v1"
const MaxRequestBytes = 64 * 1024
const MaxRelayBytes = 256 * 1024

// Signal is a relay envelope. The authenticated server assigns PeerID; requests
// and streaming events travel over the same TLS WebSocket as device heartbeats.
type Signal struct {
	Type      string `json:"type"`
	PeerID    string `json:"peerId,omitempty"`
	Transport string `json:"transport,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	// LeaseSeconds is stamped by the authenticated control server, not the browser.
	// It avoids comparing wall clocks on different machines.
	LeaseSeconds int64     `json:"leaseSeconds,omitempty"`
	Request      *Request  `json:"request,omitempty"`
	Event        *Event    `json:"event,omitempty"`
	Name         string    `json:"name,omitempty"`
	OS           string    `json:"os,omitempty"`
	Arch         string    `json:"arch,omitempty"`
	Services     []Service `json:"services,omitempty"`
	Error        string    `json:"error,omitempty"`
}

type Request struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Service   Service         `json:"service,omitempty"`
	ServiceID string          `json:"serviceId,omitempty"`
	Prompt    string          `json:"prompt,omitempty"`
	RunID     string          `json:"runId,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Title     string          `json:"title,omitempty"`
	Search    string          `json:"search,omitempty"`
	Cursor    uint64          `json:"cursor,omitempty"`
	Method    string          `json:"method,omitempty"`
	Path      string          `json:"path,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

type Event struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
