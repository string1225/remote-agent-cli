package server

import "testing"

func TestWindowsMSIDownloadAllowlist(t *testing.T) {
	for _, name := range []string{"remote-agent-windows-amd64.msi", "remote-agent-windows-arm64.msi.sha256", "remote-agent-windows-amd64.exe", "remote-agent-linux-arm64.sha256"} {
		if !artifactRE.MatchString(name) {
			t.Errorf("release artifact rejected: %s", name)
		}
	}
	for _, name := range []string{"../remote-agent-windows-amd64.msi", "remote-agent-linux-amd64.msi", "remote-agent-darwin-arm64.exe", "remote-agent-windows-amd64.msi.exe", "remote-agent-windows-amd64.msi/extra"} {
		if artifactRE.MatchString(name) {
			t.Errorf("invalid artifact accepted: %s", name)
		}
	}
}
