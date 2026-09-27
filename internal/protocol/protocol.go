package protocol

import "encoding/json"

type Service struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Workspace string `json:"workspace"`
}

type ICE struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// Signal contains connection metadata only. Prompts and output use DataChannel.
type Signal struct {
	Type      string `json:"type"`
	PeerID    string `json:"peerId,omitempty"`
	Secret    string `json:"secret,omitempty"`
	ExpiresAt int64  `json:"expiresAt,omitempty"`
	// LeaseSeconds is stamped by the authenticated control server, not the browser.
	// It avoids comparing wall clocks on different machines.
	LeaseSeconds int64           `json:"leaseSeconds,omitempty"`
	SDP          json.RawMessage `json:"sdp,omitempty"`
	Candidate    json.RawMessage `json:"candidate,omitempty"`
	ICE          []ICE           `json:"iceServers,omitempty"`
	Name         string          `json:"name,omitempty"`
	OS           string          `json:"os,omitempty"`
	Arch         string          `json:"arch,omitempty"`
	Services     []Service       `json:"services,omitempty"`
	Error        string          `json:"error,omitempty"`
}

type Request struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Secret    string  `json:"secret,omitempty"`
	Service   Service `json:"service,omitempty"`
	ServiceID string  `json:"serviceId,omitempty"`
	Prompt    string  `json:"prompt,omitempty"`
	RunID     string  `json:"runId,omitempty"`
}

type Event struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
