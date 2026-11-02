package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestActivationCredentialSurvivesLostResponseAndOriginalConfigIsKept(t *testing.T) {
	var candidate string
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_ = json.NewEncoder(w).Encode(map[string]bool{"proposedCredential": true})
			return
		}
		var in struct{ Credential, Token string }
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Token != strings.Repeat("c", 43) {
			t.Error("binding code lost")
		}
		if candidate == "" {
			candidate = in.Credential
		} else if candidate != in.Credential {
			t.Error("retry generated another credential")
		}
		calls++
		if calls == 1 {
			w.Write([]byte(`{"id":`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "same-device", "name": "测试电脑", "credential": in.Credential})
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	code := strings.Repeat("c", 43)
	if _, err := EnrollActivation(server.URL, code, path); err == nil {
		t.Fatal("lost response accepted")
	}
	if _, err := os.Stat(path + ".activation-pending.json"); err != nil {
		t.Fatal("missing recovery journal", err)
	}
	c, err := EnrollActivation(server.URL, code, path)
	if err != nil || c.ID != "same-device" || c.Credential != candidate {
		t.Fatal("retry failed", err)
	}
	before, _ := os.ReadFile(path)
	if _, err := EnrollActivation(server.URL, code, path); err == nil {
		t.Fatal("existing configuration replaced")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) || calls != 2 {
		t.Fatal("original identity changed")
	}
	if _, err := os.Stat(path + ".activation-pending.json"); !os.IsNotExist(err) {
		t.Fatal("completed journal remained")
	}
}

func TestInvalidActivationCodeCanBeCorrected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write([]byte(`{"proposedCredential":true}`))
			return
		}
		w.WriteHeader(401)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	if _, err := EnrollActivation(server.URL, strings.Repeat("c", 43), path); err == nil {
		t.Fatal("invalid code accepted")
	}
	if _, err := os.Stat(path + ".activation-pending.json"); !os.IsNotExist(err) {
		t.Fatal("a rejected first attempt blocked correction")
	}
}

func TestActivationDoesNotConsumeCodeOnOldServer(t *testing.T) {
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			posts++
		}
		w.WriteHeader(404)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "config.json")
	if _, err := EnrollActivation(server.URL, strings.Repeat("c", 43), path); err == nil || posts != 0 {
		t.Fatal("code sent to incompatible server")
	}
}
