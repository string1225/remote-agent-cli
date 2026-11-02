package gitauth

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func currentTestSID(t *testing.T) string {
	t.Helper()
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	return u.User.Sid.String()
}

func TestWindowsStoreRestrictsInheritedAndOwnerAccess(t *testing.T) {
	root := filepath.Join(t.TempDir(), "credentials")
	sid := currentTestSID(t)
	if err := PrepareStore(root, sid); err != nil {
		t.Fatal(err)
	}
	if err := PrepareStore(root, sid); err != nil {
		t.Fatal("existing store was not retained", err)
	}
	file := filepath.Join(root, "fixture")
	if err := os.WriteFile(file, []byte("mock-encrypted-data"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, file} {
		sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		acl, _, err := sd.DACL()
		if err != nil {
			t.Fatal(err)
		}
		ownerRights := false
		for i := uint32(0); i < uint32(acl.AceCount); i++ {
			var ace *windows.ACCESS_ALLOWED_ACE
			if err := windows.GetAce(acl, i, &ace); err != nil {
				t.Fatal(err)
			}
			if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
				continue
			}
			grantee := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
			if grantee == "S-1-3-4" {
				ownerRights = true
				if ace.Mask&windows.WRITE_DAC != 0 {
					t.Fatal("shared owner can rewrite the DACL")
				}
			}
			if grantee != sid && grantee != "S-1-5-18" && grantee != "S-1-5-32-544" && grantee != "S-1-3-4" {
				t.Fatal("unexpected store grantee", grantee)
			}
		}
		if !ownerRights {
			t.Fatal("missing OWNER RIGHTS restriction", path)
		}
	}
}

func TestWindowsStoreRejectsRedirectedDirectory(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "redirect")
	if err := os.Symlink(t.TempDir(), link); err != nil {
		t.Skip("symlink privilege is unavailable")
	}
	if err := PrepareStore(link, currentTestSID(t)); err == nil {
		t.Fatal("redirected credential store accepted")
	}
}

func TestWindowsPipeAuthorizesOwnerAndRejectsFakeService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sid := currentTestSID(t)
	config := filepath.Join(t.TempDir(), "config.json")
	stop, err := Listen(ctx, PipeName(config), sid, sid, fixtureManager(t))
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	conn, err := winio.DialPipeAccessImpLevel(ctx, PipeName(config), windows.GENERIC_READ|windows.GENERIC_WRITE, winio.PipeImpLevelIdentification)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := Exchange(ctx, conn, Request{Version: Version, Action: "status"}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "fixture-user") {
		t.Fatal(output.String())
	}
	if err := Call(ctx, PipeName(config), 0, Request{Version: Version, Action: "status"}, &bytes.Buffer{}); err == nil {
		t.Fatal("pipe accepted without a verified service PID")
	}
	if err := Call(ctx, PipeName(config), uint32(os.Getpid()), Request{Version: Version, Action: "status"}, &bytes.Buffer{}); err != nil {
		t.Fatal("pipe rejected its independently verified server PID", err)
	}
	if _, err := Listen(ctx, PipeName(config), sid, sid, fixtureManager(t)); err == nil {
		t.Fatal("duplicate pipe listener accepted")
	}
}

func TestWindowsIdentityRules(t *testing.T) {
	owner := "S-1-5-21-1-2-3-1000"
	for _, test := range []struct {
		peer    identity
		allowed bool
	}{
		{identity{SID: owner}, true}, {identity{SID: "S-1-5-18"}, true}, {identity{SID: "other", Elevated: true}, true},
		{identity{SID: "S-1-5-19", Service: true}, true}, {identity{SID: "S-1-5-19"}, false}, {identity{SID: "other"}, false},
	} {
		if (allowClient(test.peer, owner) == nil) != test.allowed {
			t.Fatalf("incorrect authorization: %+v", test.peer)
		}
	}
}

func TestInstalledGCMUsesOnlyTemporaryDPAPIStore(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	gcm, err := findGCM(ctx)
	if err != nil {
		t.Skip("GCM is not installed")
	}
	root := t.TempDir()
	store := filepath.Join(root, "dpapi")
	if err := PrepareStore(store, currentTestSID(t)); err != nil {
		t.Fatal(err)
	}
	const secret = "fixture-only-not-a-real-credential"
	for _, action := range []string{"store", "get"} {
		cmd := newCommand(ctx, gcm, action)
		cmd.Dir = root
		cmd.Env = Environment(append(os.Environ(), "GCM_PROVIDER=generic", "GIT_CONFIG_GLOBAL="+filepath.Join(root, "absent-gitconfig"), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_PARAMETERS="), store, false)
		input := "protocol=https\nhost=fixture.invalid\nusername=fixture-user\n"
		if action == "store" {
			input += "password=" + secret + "\n"
		}
		cmd.Stdin = strings.NewReader(input + "\n")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("temporary GCM %s failed: %s (%v)", action, out, err)
		}
		if action == "get" && !strings.Contains(string(out), "password="+secret) {
			t.Fatal("DPAPI credential was not readable after helper restart")
		}
	}
	files := 0
	if err := filepath.WalkDir(store, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files++
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(data, []byte(secret)) {
				t.Fatal("DPAPI store contains plaintext")
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("GCM ignored the configured DPAPI store")
	}
}
