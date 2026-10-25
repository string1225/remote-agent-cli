package clientupdate

import (
	"crypto/ed25519"
	"encoding/hex"
)

// Trust anchors are compiled into the Windows host or Unix client, never
// supplied or replaced by a release manifest.

// ReleasePublicKey can be replaced at build time for isolated acceptance hosts.
// Production builds use this repository's fixed public key.
var ReleasePublicKey = "39d660f6d8471044bf384ba74a8b9bc668dfc2175b8496e1430a122460e475c7"

func TrustKeys() map[string]ed25519.PublicKey {
	key, _ := hex.DecodeString(ReleasePublicKey)
	return map[string]ed25519.PublicKey{"release-2026": ed25519.PublicKey(key)}
}
