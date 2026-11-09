package serviceaccess

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

const fixtureSID = "S-1-5-80-111-222-333-444-555"

func fixture(t *testing.T) (*manager, string) {
	t.Helper()
	root := t.TempDir()
	base := filepath.Join(root, "system", "access")
	repo := filepath.Join(root, "O'Brien # repo")
	for _, p := range []string{base, filepath.Join(repo, ".git"), filepath.Join(repo, "protected")} {
		if err := os.MkdirAll(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, "protected", "mock-credential.txt"), []byte("simulated only"), 0600); err != nil {
		t.Fatal(err)
	}
	m := &manager{base: base, sid: fixtureSID}
	t.Cleanup(m.close)
	obj, err := m.open(filepath.Join(repo, "protected"), true)
	if err != nil {
		t.Fatal(err)
	}
	sddl, err := descriptor(obj)
	if err != nil {
		t.Fatal(err)
	}
	sd, _ := windows.SecurityDescriptorFromString(sddl)
	dacl, _, _ := sd.DACL()
	if err = windows.SetSecurityInfo(obj.handle, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	return m, repo
}

func saved(t *testing.T, m *manager) backup {
	t.Helper()
	records, err := m.records()
	if err != nil || len(records) != 1 {
		t.Fatal(records, err)
	}
	return records[0]
}

func assertRestored(t *testing.T, m *manager, b backup) {
	t.Helper()
	for _, obj := range b.Objects {
		current, err := m.open(obj.Path, false)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := descriptor(current)
		if err != nil || comparableSDDL(actual) != comparableSDDL(obj.Before) {
			t.Fatalf("DACL/owner/inheritance changed at %s:\nbefore %s\nactual %s (%v)", obj.Path, obj.Before, actual, err)
		}
	}
}

func TestAccessRecursiveGrantAndRestorePreserveOwnerAndOriginalACL(t *testing.T) {
	for _, permission := range []string{"read", "modify"} {
		t.Run(permission, func(t *testing.T) {
			m, repo := fixture(t)
			var out bytes.Buffer
			if err := m.run(Options{Action: "grant", Path: repo, Permission: permission, Recursive: true, GitTrust: true}, &out); err != nil {
				t.Fatal(out.String(), err)
			}
			b := saved(t, m)
			for _, obj := range b.Objects {
				if !strings.Contains(obj.After, fixtureSID) {
					t.Fatalf("missing service-only grant: %s", obj.After)
				}
				beforeOwner := strings.Split(obj.Before, "D:")[0]
				afterOwner := strings.Split(obj.After, "D:")[0]
				if beforeOwner != afterOwner {
					t.Fatal("owner/group changed", obj)
				}
				sd, _ := windows.SecurityDescriptorFromString(obj.After)
				acl, _, _ := sd.DACL()
				for i := uint32(0); i < uint32(acl.AceCount); i++ {
					var ace *windows.ACCESS_ALLOWED_ACE
					if err := windows.GetAce(acl, i, &ace); err != nil {
						t.Fatal(err)
					}
					if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String() != fixtureSID {
						continue
					}
					if ace.Mask&(windows.WRITE_DAC|windows.WRITE_OWNER) != 0 || permission == "read" && ace.Mask&(windows.FILE_WRITE_DATA|windows.DELETE) != 0 || permission == "modify" && ace.Mask&(windows.FILE_WRITE_DATA|windows.DELETE) != windows.FILE_WRITE_DATA|windows.DELETE {
						t.Fatalf("incorrect service rights for %s: %x", permission, ace.Mask)
					}
				}
			}
			newFile := filepath.Join(repo, "protected", "new.txt")
			if err := os.WriteFile(newFile, []byte("new"), 0600); err != nil {
				t.Fatal(err)
			}
			current, _ := m.open(newFile, false)
			during, _ := descriptor(current)
			if !strings.Contains(during, fixtureSID) {
				t.Fatal("new contents did not inherit grant", during)
			}
			if err := m.run(Options{Action: "restore", Backup: b.ID}, &out); err != nil {
				t.Fatal(out.String(), err)
			}
			assertRestored(t, m, b)
			after, _ := descriptor(current)
			if strings.Contains(after, fixtureSID) {
				t.Fatal("new contents retained inherited grant", after)
			}
			if err := m.run(Options{Action: "restore", Backup: b.ID}, &out); err != nil {
				t.Fatal("restore must be idempotent", err)
			}
			config, _ := os.ReadFile(filepath.Join(m.base, "trust.gitconfig"))
			if strings.Contains(string(config), "directory =") || saved(t, m).Status != "restored" {
				t.Fatal("trust or journal was not restored", string(config))
			}
		})
	}
}

func TestAccessPreparedJournalRequiresExplicitRecovery(t *testing.T) {
	m, repo := fixture(t)
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "modify", Recursive: true, GitTrust: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b := saved(t, m)
	b.Status = "prepared" // Simulate a process loss during a transaction.
	if err := m.save(b); err != nil {
		t.Fatal(err)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID}, &bytes.Buffer{}); err == nil {
		t.Fatal("incomplete journal restored without explicit --force")
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID, Force: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	assertRestored(t, m, b)
}

func TestAccessSingleFileGrantAndRestoreDoesNotReadContents(t *testing.T) {
	m, repo := fixture(t)
	path := filepath.Join(repo, "protected", "mock-credential.txt")
	var out bytes.Buffer
	if err := m.run(Options{Action: "grant", Path: path, Permission: "read"}, &out); err != nil {
		t.Fatal(err)
	}
	b := saved(t, m)
	if len(b.Objects) != 1 || b.Objects[0].Directory || strings.Contains(out.String(), "simulated only") {
		t.Fatal(b, out.String())
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID}, &out); err != nil {
		t.Fatal(err)
	}
	assertRestored(t, m, b)
	raw, _ := os.ReadFile(path)
	if string(raw) != "simulated only" {
		t.Fatal("file contents changed")
	}
}

func TestAccessRollbackAfterPartialACLOrTrustFailure(t *testing.T) {
	for _, stage := range []string{"grant", "publish", "save"} {
		t.Run(stage, func(t *testing.T) {
			m, repo := fixture(t)
			count := 0
			m.fault = func(s string) error {
				fail := stage == "publish" && s == "write:trust.gitconfig"
				if stage == "grant" && strings.HasPrefix(s, "grant:") || stage == "save" && strings.HasSuffix(s, ".json") {
					count++
					fail = count == 2
				}
				if fail {
					m.fault = nil // Recovery may write normally.
					return errors.New("simulated failure")
				}
				return nil
			}
			if err := m.run(Options{Action: "grant", Path: repo, Permission: "modify", Recursive: true, GitTrust: true}, &bytes.Buffer{}); err == nil {
				t.Fatal("injected fault ignored")
			}
			b := saved(t, m)
			assertRestored(t, m, b)
			if b.Status != "restored" {
				t.Fatal(b.Status)
			}
		})
	}
}

func TestAccessRefusesOverlapDriftAndReplacement(t *testing.T) {
	m, repo := fixture(t)
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "read", Recursive: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b := saved(t, m)
	if err := m.run(Options{Action: "grant", Path: filepath.Join(repo, "protected"), Permission: "modify"}, &bytes.Buffer{}); err == nil {
		t.Fatal("overlapping grant accepted")
	}
	obj, _ := m.open(repo, true)
	if err := grantDACL(obj, fixtureSID, "modify", true); err != nil {
		t.Fatal(err)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatal("ACL drift silently overwritten", err)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID, Force: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	assertRestored(t, m, b)

	// Replacement validation remains enforced with --force. Synthetic identity
	// drift avoids deleting a pinned real file and never touches another tree.
	b.Status = "active"
	b.Objects[0].Identity = "replaced-fixture"
	if err := m.save(b); err != nil {
		t.Fatal(err)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID, Force: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "替换") {
		t.Fatal("replacement identity accepted", err)
	}
}

func TestAccessExactDirectoryGrantDoesNotChangeChildren(t *testing.T) {
	m, repo := fixture(t)
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "read"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b := saved(t, m)
	if len(b.Objects) != 1 {
		t.Fatal("exact grant included descendants")
	}
	child, _ := m.open(filepath.Join(repo, "protected", "mock-credential.txt"), false)
	sddl, _ := descriptor(child)
	if strings.Contains(sddl, fixtureSID) {
		t.Fatal("exact grant escaped into descendants", sddl)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
}

func TestAccessRejectsJunctionBeforeAnyACLChanges(t *testing.T) {
	m, repo := fixture(t)
	outside := t.TempDir()
	junction := filepath.Join(repo, "junction")
	cmd := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", junction, outside)
	if raw, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("junction fixture unavailable: %s %v", raw, err)
	}
	t.Cleanup(func() { _ = os.Remove(junction) })
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "modify", Recursive: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "重解析") {
		t.Fatal("junction accepted", err)
	}
	records, err := m.records()
	if err != nil || len(records) != 0 {
		t.Fatal("preflight changed state", records, err)
	}
}

func TestAccessRestoreRejectsNewJunction(t *testing.T) {
	m, repo := fixture(t)
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "modify", Recursive: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	b := saved(t, m)
	outside := t.TempDir()
	junction := filepath.Join(repo, "new-junction")
	if raw, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", junction, outside).CombinedOutput(); err != nil {
		t.Skipf("junction fixture unavailable: %s %v", raw, err)
	}
	if err := m.run(Options{Action: "restore", Backup: b.ID, Force: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "重解析") {
		t.Fatal("new junction accepted during restore", err)
	}
	if saved(t, m).Status != "active" {
		t.Fatal("restore changed journal before preflight completed")
	}
}

func TestAccessTrustActuallyWorksInGitProtectedScope(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git unavailable")
	}
	m, repo := fixture(t)
	env := append(os.Environ(), "GIT_CONFIG_GLOBAL="+filepath.Join(m.base, "empty-user-config"), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_PARAMETERS=", "GIT_CONFIG_COUNT=0", "GIT_TERMINAL_PROMPT=0", "GIT_TEST_ASSUME_DIFFERENT_OWNER=1")
	gitRun := func(args ...string) ([]byte, error) {
		cmd := exec.Command(git, args...)
		cmd.Env = env
		return cmd.CombinedOutput()
	}
	if raw, err := gitRun("init", "--quiet", repo); err != nil {
		t.Fatal(string(raw), err)
	}
	if raw, err := gitRun("-C", repo, "status", "--porcelain"); err == nil || !strings.Contains(string(raw), "dubious ownership") {
		t.Fatal("fixture did not require exact trust", string(raw), err)
	}
	if err := m.run(Options{Action: "grant", Path: repo, Permission: "read", GitTrust: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	include := "'include.path=" + strings.ReplaceAll(filepath.ToSlash(filepath.Join(m.base, "trust.gitconfig")), "'", "'\\''") + "'"
	env = append(env, "GIT_CONFIG_PARAMETERS="+include)
	if raw, err := gitRun("-C", repo, "status", "--porcelain"); err != nil {
		t.Fatal("command-scope include did not trust exact repo", string(raw), err)
	}
	other := t.TempDir()
	if raw, err := gitRun("init", "--quiet", other); err != nil {
		t.Fatal(string(raw), err)
	}
	if raw, err := gitRun("-C", other, "status", "--porcelain"); err == nil || !strings.Contains(string(raw), "dubious ownership") {
		t.Fatal("unselected repo trusted", string(raw), err)
	}
	b := saved(t, m)
	if err := m.run(Options{Action: "restore", Backup: b.ID}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if raw, err := gitRun("-C", repo, "status", "--porcelain"); err == nil || !strings.Contains(string(raw), "dubious ownership") {
		t.Fatal("restore did not revoke trust", string(raw), err)
	}
	if _, err := os.Stat(filepath.Join(m.base, "empty-user-config")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("user Git config was created/modified", err)
	}
}
