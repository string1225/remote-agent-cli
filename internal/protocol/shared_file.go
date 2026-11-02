package protocol

import "time"

const MaxSharedFileBytes int64 = 100 << 20

// SharedFile is a temporary public download capability, never a server path.
type SharedFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	SHA256    string    `json:"sha256"`
	URL       string    `json:"url,omitempty"`
	ExpiresAt time.Time `json:"expiresAt"`
}
