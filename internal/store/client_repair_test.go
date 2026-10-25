package store

import (
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairClaimRequiresCapabilityPlatformAndConfirmation(t *testing.T) {
	db, e := Open(filepath.Join(t.TempDir(), "store.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	_ = db.CreateUser(User{ID: "alice", Username: "alice"})
	_ = db.CreateUser(User{ID: "bob", Username: "bob"})
	target, ticket, _ := db.Enroll("alice", "Target")
	_, old, _ := db.Redeem(ticket, "windows", "amd64")
	oldMachine := strings.Repeat("a", 64)
	newMachine := strings.Repeat("b", 64)
	_ = db.Heartbeat(target.ID, protocol.Signal{MachineID: oldMachine, Hostname: "old-pc"})
	if _, _, e = db.RebindEnrollment("bob", target.ID); e == nil {
		t.Fatal("non-owner minted repair")
	}
	_, ticket, _ = db.RebindEnrollment("alice", target.ID)
	info, valid, e := db.InspectRepair(ticket, target.ID, old, newMachine)
	if e != nil || !valid || info.MachineID != oldMachine {
		t.Fatal(info, valid, e)
	}
	claim := RepairClaim{Token: ticket, ID: target.ID, Credential: Token(), MachineID: newMachine, Hostname: "new-pc", OS: "windows", Arch: "amd64"}
	if _, _, e = db.ClaimRepair(claim); e == nil {
		t.Fatal("replaced enrolled device without explicit confirmation")
	}
	claim.Force = true
	claim.OS = "linux"
	if _, _, e = db.ClaimRepair(claim); e == nil {
		t.Fatal("force changed immutable OS")
	}
	claim.OS = "windows"
	claim.ID = "wrong"
	if _, _, e = db.ClaimRepair(claim); e == nil {
		t.Fatal("capability escaped its target")
	}
	claim.ID = target.ID
	if _, e = db.AuthenticateAgent(target.ID, old); e != nil {
		t.Fatal("failed claims changed binding")
	}
	claimed, _, e := db.ClaimRepair(claim)
	if e != nil || claimed.ID != target.ID || claimed.MachineID != newMachine {
		t.Fatal(claimed, e)
	}
	if _, e = db.AuthenticateAgent(target.ID, old); e == nil {
		t.Fatal("replaced machine retained access")
	}
	if _, _, e = db.ClaimRepair(claim); e == nil {
		t.Fatal("one-time capability reused")
	}
	if _, valid, e = db.InspectRepair("expired", target.ID, claim.Credential, newMachine); e != nil || !valid {
		t.Fatal("matching local credential should still repair without a fresh capability", e)
	}
	if _, _, e = db.InspectRepair("expired", target.ID, claim.Credential, oldMachine); e == nil {
		t.Fatal("cloned configuration bypassed machine mismatch")
	}
}

func TestRepairMovesSameOwnerProjectsAndRobotReservations(t *testing.T) {
	db, e := Open(filepath.Join(t.TempDir(), "store.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	_ = db.CreateUser(User{ID: "alice", Username: "alice"})
	source, invite, _ := db.Enroll("alice", "Source")
	_, sourceSecret, _ := db.Redeem(invite, "windows", "amd64")
	target, invite, _ := db.Enroll("alice", "Target")
	_, _, _ = db.Redeem(invite, "windows", "amd64")
	_ = db.Heartbeat(source.ID, protocol.Signal{Services: []protocol.Service{{ID: "project", Provider: "codex", Name: "Keep"}}})
	fingerprint := strings.Repeat("d", 64)
	if e = db.BindChannel("alice", ChannelBinding{AgentID: source.ID, ServiceID: "project", Fingerprint: fingerprint}, true); e != nil {
		t.Fatal(e)
	}
	_, ticket, _ := db.RebindEnrollment("alice", target.ID)
	claimed, moved, e := db.ClaimRepair(RepairClaim{Token: ticket, ID: target.ID, Credential: Token(), MachineID: strings.Repeat("a", 64), Hostname: "actual", OS: "windows", Arch: "amd64", Force: true, SourceID: source.ID, SourceCredential: sourceSecret})
	if e != nil || !moved || len(claimed.Services) != 1 || claimed.Services[0].ID != "project" || claimed.RepairSourceID != source.ID {
		t.Fatal(claimed, moved, e)
	}
	if _, e = db.AuthenticateAgent(source.ID, sourceSecret); e == nil {
		t.Fatal("source binding did not become offline")
	}
	if e = db.BindChannel("alice", ChannelBinding{AgentID: target.ID, ServiceID: "project", Fingerprint: fingerprint}, true); e != nil {
		t.Fatal("robot ownership did not follow same-owner replacement", e)
	}
}
