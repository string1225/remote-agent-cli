package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/string1225/remote-agent-cli/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func TestOfflineAdminCreateAndGrant(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("DATABASE_PATH", path)
	var output bytes.Buffer
	if err := runAdmin([]string{"create", "--username", " 管理员 ", "--password-stdin"}, strings.NewReader("strong-password\n"), &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "strong-password") {
		t.Fatal("password logged")
	}
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	u, err := db.User("管理员")
	if err != nil || !u.Admin || u.Disabled || bcrypt.CompareHashAndPassword(u.Password, []byte("strong-password")) != nil {
		t.Fatalf("admin not created: %#v %v", u, err)
	}
	if err := db.CreateUser(store.User{ID: "member", Username: "Member"}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if err := runAdmin([]string{"grant", "--username", "member"}, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	u, err = db.User("Member")
	if err != nil || !u.Admin || u.ID != "member" {
		t.Fatalf("grant failed: %#v %v", u, err)
	}
}

func TestOfflineAdminRequiresStdinPasswordAndStoppedServer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	t.Setenv("DATABASE_PATH", path)
	for _, tc := range []struct {
		args     []string
		password string
	}{{[]string{"create", "--username", "admin"}, "secret-password"}, {[]string{"create", "--username", "admin", "--password-stdin"}, "short"}, {[]string{"create", "--username", "admin", "--password", "visible-in-process-list"}, ""}} {
		if err := runAdmin(tc.args, strings.NewReader(tc.password), &bytes.Buffer{}); err == nil {
			t.Fatalf("unsafe command accepted %v", tc.args)
		}
	}
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := runAdmin([]string{"create", "--username", "admin", "--password-stdin"}, strings.NewReader("strong-password"), &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "stop") {
		t.Fatalf("locked DB accepted or unclear error: %v", err)
	}
}

func TestGatewayEnvironmentRequiresPrivateTokenFile(t *testing.T) {
	t.Setenv("CODEX_GATEWAY_URL", "http://127.0.0.1:8088")
	t.Setenv("CODEX_GATEWAY_TOKEN_FILE", "")
	if _, err := configFromEnv(); err == nil {
		t.Fatal("gateway enabled without service token file")
	}
}

func TestGatewayEnvironmentLoadsServiceTokenWithoutExposingIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service-token")
	if err := os.WriteFile(path, []byte("service-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_GATEWAY_URL", "http://127.0.0.1:8088")
	t.Setenv("CODEX_GATEWAY_TOKEN_FILE", path)
	t.Setenv("CODEX_MODEL", "custom-model")
	cfg, err := configFromEnv()
	if err != nil || cfg.GatewayToken != "service-secret" || cfg.GatewayURL != "http://127.0.0.1:8088" || cfg.GatewayModel != "custom-model" {
		t.Fatal("gateway file was not loaded", err)
	}
	if err := os.WriteFile(path, []byte("service-secret\nsecond-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := configFromEnv(); err == nil || strings.Contains(err.Error(), "service-secret") {
		t.Fatalf("multiline service secret accepted or leaked: %v", err)
	}
}

func TestGatewayTokenPermissionsAcceptOnlyPrivateFileOrSystemdCredential(t *testing.T) {
	t.Setenv("CREDENTIALS_DIRECTORY", "/run/credentials/remote-agent.service")
	for _, tc := range []struct {
		path string
		mode os.FileMode
		want bool
	}{
		{"/etc/remote-agent/token", 0600, true},
		{"/etc/remote-agent/token", 0400, true},
		{"/etc/remote-agent/token", 0440, false},
		{"/run/credentials/another.service/token", 0440, false},
		{"/run/credentials/remote-agent.service/token", 0444, false},
		{"/run/credentials/remote-agent.service/token", 0640, false},
	} {
		if got := privateGatewayToken(tc.path, tc.mode); got != tc.want {
			t.Errorf("%s %o: got %t want %t", tc.path, tc.mode, got, tc.want)
		}
	}
	t.Setenv("CREDENTIALS_DIRECTORY", "/tmp/public")
	if privateGatewayToken("/tmp/public/token", 0440) {
		t.Fatal("arbitrary credential directory accepted")
	}
}

func TestSystemdCredentialCanonicalLocation(t *testing.T) {
	// Path policy uses platform-native paths. The special location is Linux-only.
	if os.PathSeparator != '/' {
		t.Skip("Linux credential paths")
	}
	dir := "/run/credentials/remote-agent.service"
	for _, tc := range []struct {
		file, dir string
		want      bool
	}{
		{dir + "/token", dir, true},
		{dir + "/token", dir + "/", false},
		{dir + "/../other.service/token", dir, false},
		{dir + "/sub/token", dir, false},
		{"/run/credentials-evil/unit/token", "/run/credentials-evil/unit", false},
		{"/tmp/public/token", "/tmp/public", false},
		{"/run/credentials/other.service/token", dir, false},
	} {
		if got := systemdCredentialLocation(tc.file, tc.dir); got != tc.want {
			t.Errorf("location %s %s: got %t", tc.file, tc.dir, got)
		}
	}
}

func TestGatewayEnvironmentInheritedSystemdCredential(t *testing.T) {
	if os.Getenv("CREDENTIALS_DIRECTORY") == "" || os.Getenv("CODEX_GATEWAY_TOKEN_FILE") == "" {
		t.Skip("requires a systemd LoadCredential test unit")
	}
	info, err := os.Stat(os.Getenv("CODEX_GATEWAY_TOKEN_FILE"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0440 {
		t.Skip("runtime credential uses ordinary private permissions")
	}
	cfg, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GatewayToken == "" {
		t.Fatal("runtime credential not loaded")
	}
}
