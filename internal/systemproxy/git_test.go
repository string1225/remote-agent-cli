package systemproxy

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func gitProxyFixture(t *testing.T) GitManager {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git fixture requires git")
	}
	root := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_PARAMETERS", "")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	return GitManager{Git: git, Env: os.Environ(), Path: filepath.Join(root, "proxy-backup.json")}
}

func TestGitProxyTracksPortAndRestoresOriginalValues(t *testing.T) {
	m := gitProxyFixture(t)
	ctx := context.Background()
	original := []string{"http://old-proxy.example:8080", ""}
	if err := m.write(ctx, original); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(m.Git, "config", "--global", "user.name", "Fixture Name")
	cmd.Env = m.Env
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	for _, address := range []string{"127.0.0.1:17891", "127.0.0.1:17891", "127.0.0.1:17892"} {
		if err := m.Enable(ctx, address); err != nil {
			t.Fatal(err)
		}
		values, err := m.read(ctx)
		if err != nil || !slices.Equal(values, []string{"http://" + address}) {
			t.Fatalf("Git did not track active port: %v / %v", values, err)
		}
	}
	if err := m.Disable(ctx, "127.0.0.1:17892"); err != nil {
		t.Fatal(err)
	}
	values, err := m.read(ctx)
	if err != nil || !slices.Equal(values, original) {
		t.Fatalf("original Git values lost: %v / %v", values, err)
	}
	cmd = exec.Command(m.Git, "config", "--global", "user.name")
	cmd.Env = m.Env
	if output, err := cmd.Output(); err != nil || string(output) != "Fixture Name\n" {
		t.Fatal("unrelated Git configuration changed", err)
	}
}

func TestGitProxyOffRestoresUnsetAndPreservesExternalChanges(t *testing.T) {
	for _, external := range []bool{false, true} {
		t.Run(map[bool]string{false: "unset", true: "external"}[external], func(t *testing.T) {
			m := gitProxyFixture(t)
			ctx := context.Background()
			if err := m.Enable(ctx, "127.0.0.1:17891"); err != nil {
				t.Fatal(err)
			}
			var expected []string
			if external {
				expected = []string{"http://changed-by-user.example:9000"}
				if err := m.write(ctx, expected); err != nil {
					t.Fatal(err)
				}
			}
			if err := m.Disable(ctx, "127.0.0.1:17891"); err != nil {
				t.Fatal(err)
			}
			values, err := m.read(ctx)
			if err != nil || !slices.Equal(values, expected) {
				t.Fatalf("wrong Git state after off: %v / %v", values, err)
			}
		})
	}
}

func TestGitProxyRejectsBrokenBackupAndNetworkListener(t *testing.T) {
	m := gitProxyFixture(t)
	ctx := context.Background()
	if err := os.WriteFile(m.Path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Enable(ctx, "127.0.0.1:17891"); err == nil {
		t.Fatal("broken original backup overwritten")
	}
	if err := m.Enable(ctx, "0.0.0.0:17891"); err == nil {
		t.Fatal("non-loopback Git proxy accepted")
	}
	values, err := m.read(ctx)
	if err != nil || len(values) != 0 {
		t.Fatal("failed enable changed Git settings")
	}
}

func TestGitProxyPreservesIncludedProxyAndFailedWrite(t *testing.T) {
	m := gitProxyFixture(t)
	ctx := context.Background()
	config := os.Getenv("GIT_CONFIG_GLOBAL")
	included := filepath.Join(filepath.Dir(config), "included.gitconfig")
	if err := os.WriteFile(included, []byte("[http]\n\tproxy = http://included.example:8080\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(m.Git, "config", "--global", "include.path", included)
	cmd.Env = m.Env
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if err := m.Enable(ctx, "127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	if err := m.Disable(ctx, "127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	values, err := m.read(ctx)
	if err != nil || len(values) != 0 {
		t.Fatal("included proxy was copied into the primary Git config")
	}
	if err = os.WriteFile(config+".lock", nil, 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(config + ".lock")
	if err = m.Enable(ctx, "127.0.0.1:17891"); err == nil {
		t.Fatal("locked Git config was overwritten")
	}
	values, err = m.read(ctx)
	if err != nil || len(values) != 0 {
		t.Fatal("failed write changed original proxy")
	}
}

func TestGitProxyFailedPortChangeKeepsOriginalRecoveryPoint(t *testing.T) {
	m := gitProxyFixture(t)
	ctx := context.Background()
	original := []string{"http://original.example:8080"}
	if err := m.write(ctx, original); err != nil {
		t.Fatal(err)
	}
	if err := m.Enable(ctx, "127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	lock := os.Getenv("GIT_CONFIG_GLOBAL") + ".lock"
	if err := os.WriteFile(lock, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Enable(ctx, "127.0.0.1:17892"); err == nil {
		t.Fatal("locked config unexpectedly changed")
	}
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	if err := m.Disable(ctx, "127.0.0.1:17892"); err != nil {
		t.Fatal(err)
	}
	got, err := m.read(ctx)
	if err != nil || !slices.Equal(got, original) {
		t.Fatal("failed switch lost original proxy recovery", err)
	}
}

func TestGitProxyRejectsEffectiveOverrides(t *testing.T) {
	for _, kind := range []string{"include", "url"} {
		t.Run(kind, func(t *testing.T) {
			m := gitProxyFixture(t)
			ctx := context.Background()
			original := []string{"http://original.example:8080"}
			if err := m.write(ctx, original); err != nil {
				t.Fatal(err)
			}
			var args []string
			if kind == "include" {
				path := filepath.Join(filepath.Dir(m.Path), "late-include.gitconfig")
				if err := os.WriteFile(path, []byte("[http]\nproxy = http://overridden.example:8080\n"), 0600); err != nil {
					t.Fatal(err)
				}
				args = []string{"config", "--global", "include.path", path}
			} else {
				args = []string{"config", "--global", "http.https://github.com/.proxy", "http://overridden.example:8080"}
			}
			cmd := exec.Command(m.Git, args...)
			cmd.Env = m.Env
			if err := cmd.Run(); err != nil {
				t.Fatal(err)
			}
			if err := m.Enable(ctx, "127.0.0.1:17891"); err == nil {
				t.Fatal("ineffective Git proxy reported synchronized")
			}
			status, err := m.Status(ctx, "127.0.0.1:17891")
			if err != nil || status.Managed {
				t.Fatal("Git override reported as managed", err)
			}
			got, err := m.read(ctx)
			if err != nil || !slices.Equal(got, original) {
				t.Fatal("override rollback lost original values", err)
			}
		})
	}
}
