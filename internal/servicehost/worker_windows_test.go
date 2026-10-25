package servicehost

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInheritedWorkerPipesBoundLifecycle(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "core.exe")
	cmd := exec.Command("go", "build", "-o", exe, "../../cmd/agent")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	config := filepath.Join(dir, "config.json")
	data, _ := json.Marshal(map[string]any{"id": "fixture", "credential": "fixture-only", "server": "http://127.0.0.1:1", "allowedRoots": []string{dir}})
	if err := os.WriteFile(config, data, 0600); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(dir, "profile.json")
	data, _ = json.Marshal(map[string]any{"config": config, "environment": map[string]string{"USERPROFILE": dir, "CODEX_HOME": filepath.Join(dir, "codex"), "RA_CODEX_PATH": filepath.Join(dir, "missing-codex"), "RA_QODER_PATH": filepath.Join(dir, "missing-qoder")}})
	if err := os.WriteFile(profile, data, 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(dir, "host.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	w, err := launchWorkerAt(exe, profile, log)
	if err != nil {
		t.Fatal(err)
	}
	defer w.kill()
	for _, kind := range []string{"activate", "pause", "resume", "pause", "resume"} {
		if err := w.call(kind, false); err != nil {
			t.Fatal(kind, err)
		}
	}
	if err := w.stop(); err != nil {
		t.Fatal(err)
	}
}
