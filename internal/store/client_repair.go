package store

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	bolt "go.etcd.io/bbolt"
)

func validMachineID(value string) bool {
	b, e := hex.DecodeString(value)
	return e == nil && len(b) == 32 && strings.ToLower(value) == value
}
func cleanHostname(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, string([]rune(value)[:min(len([]rune(value)), 120)]))
}

func (s *Store) InspectRepair(token, id, credential, machine string) (a Agent, valid bool, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		if get(tx, "agents", id, &a) != nil || a.Revoked || !activeUser(tx, a.UserID) || !validMachineID(machine) {
			return ErrDenied
		}
		valid = credential != "" && subtle.ConstantTimeCompare([]byte(a.TokenHash), []byte(Hash(credential))) == 1
		if valid && (a.MachineID == "" || a.MachineID == machine) {
			return nil
		}
		var invite enrollment
		if get(tx, "enrollments", Hash(token), &invite) != nil || invite.AgentID != id || invite.ExpiresAt <= time.Now().Unix() {
			return ErrDenied
		}
		return nil
	})
	a.TokenHash = ""
	return
}

type RepairClaim struct {
	Token, ID, Credential, MachineID, Hostname, OS, Arch, SourceID, SourceCredential string
	Force                                                                            bool
}

func (s *Store) ClaimRepair(in RepairClaim) (a Agent, channelsMoved bool, err error) {
	credential, e := base64.RawURLEncoding.DecodeString(in.Credential)
	if e != nil || len(credential) != 32 || !validMachineID(in.MachineID) {
		return a, false, ErrDenied
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		var invite enrollment
		if get(tx, "enrollments", Hash(in.Token), &invite) != nil || invite.AgentID != in.ID || invite.ExpiresAt <= time.Now().Unix() || get(tx, "agents", in.ID, &a) != nil || a.Revoked || !activeUser(tx, a.UserID) {
			return ErrDenied
		}
		if a.OS != "" && a.OS != in.OS {
			return ErrDenied
		}
		if a.Enrolled && !in.Force {
			return ErrDenied
		}
		if a.Enrolled && !invite.Recovery {
			return ErrDenied
		}
		a.RepairSourceID = ""
		if err := releaseAgentChannels(tx, a.ID); err != nil {
			return err
		}
		if in.SourceID != "" && in.SourceID != a.ID {
			var source Agent
			if get(tx, "agents", in.SourceID, &source) == nil && source.UserID == a.UserID && !source.Revoked && in.SourceCredential != "" && subtle.ConstantTimeCompare([]byte(source.TokenHash), []byte(Hash(in.SourceCredential))) == 1 {
				var bindings []ChannelBinding
				if err := tx.Bucket([]byte("channel_bots")).ForEach(func(_, v []byte) error {
					var b ChannelBinding
					if err := json.Unmarshal(v, &b); err != nil {
						return err
					}
					if b.AgentID == source.ID {
						bindings = append(bindings, b)
					}
					return nil
				}); err != nil {
					return err
				}
				if err := releaseAgentChannels(tx, source.ID); err != nil {
					return err
				}
				a.Services = source.Services
				// Retain the old record for recovery, but retire its credential.
				source.TokenHash = ""
				source.MachineID = ""
				source.Hostname = ""
				if err := put(tx, "agents", source.ID, source); err != nil {
					return err
				}
				for _, b := range bindings {
					b.AgentID = a.ID
					if err := bindChannel(tx, a.UserID, b, true); err != nil {
						return err
					}
				}
				channelsMoved = true
				a.RepairSourceID = source.ID
			}
		}
		a.Enrolled = true
		a.TokenHash = Hash(in.Credential)
		a.OS = in.OS
		a.Arch = in.Arch
		a.MachineID = in.MachineID
		a.Hostname = cleanHostname(in.Hostname)
		if err := put(tx, "agents", a.ID, a); err != nil {
			return err
		}
		b := tx.Bucket([]byte("enrollments"))
		var remove [][]byte
		if err := b.ForEach(func(k, v []byte) error {
			var x enrollment
			if json.Unmarshal(v, &x) == nil && x.AgentID == a.ID {
				remove = append(remove, append([]byte{}, k...))
			}
			return nil
		}); err != nil {
			return err
		}
		for _, k := range remove {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
	return
}
