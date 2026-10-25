package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestPersonalTokenAccountUsesBearerAndDoesNotExposeSecret(t *testing.T) {
	f := setup(t)
	alice, bob := f.user("alice"), f.user("bob")
	data := f.request(alice, "POST", "/api/tokens", map[string]string{"name": "fixture"}, 201)
	var created struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequest("GET", f.url+"/api/token/account", nil)
	request.Header.Set("Authorization", "Bearer "+created.Token)
	response, err := bob.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, _ = io.ReadAll(response.Body)
	if response.StatusCode != 200 || !bytes.Contains(data, []byte(`"accountUsername":"alice"`)) || bytes.Contains(data, []byte(created.Token)) {
		t.Fatalf("token identity not isolated: %d %s", response.StatusCode, data)
	}
}
