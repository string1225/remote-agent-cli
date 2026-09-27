package protocol

import "testing"

func TestServerURLMounts(t *testing.T) {
	for _, raw := range []string{"https://codex.sunny-string.cn/agents", "https://host.example/team/agents/", "http://localhost:8080/agents", "https://host.example/"} {
		if _, err := ParseServerURL(raw); err != nil {
			t.Errorf("%s: %v", raw, err)
		}
	}
	for _, raw := range []string{"http://host.example/agents", "https://host.example//agents", "https://host.example/../agents", "https://host.example/%61gents", "https://host.example/agents?token=secret", "https://user:pass@host.example/agents", "https://host.example/agents#fragment"} {
		if _, err := ParseServerURL(raw); err == nil {
			t.Errorf("accepted ambiguous URL: %s", raw)
		}
	}
}
