package codexsetup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/protocol"
)

// PersonalConnection identifies the owner of a supplied personal token; device
// or browser account state is deliberately not involved in this exchange.
func PersonalConnection(client *http.Client, server, token string) (c Connection, err error) {
	if _, err = protocol.ParseServerURL(server); err != nil {
		return
	}
	request, err := http.NewRequest("GET", strings.TrimSuffix(server, "/")+"/api/token/account", nil)
	if err != nil {
		return c, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(request)
	if err != nil {
		return c, errors.New("could not verify personal token with gateway")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return c, fmt.Errorf("gateway verification returned HTTP %d; check token and server login", response.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 16384)).Decode(&c); err != nil || c.AccountID == "" || c.AccountUsername == "" {
		return Connection{}, errors.New("invalid token account response")
	}
	if _, err = protocol.ParseServerURL(c.BaseURL); err != nil {
		return Connection{}, errors.New("invalid token account response")
	}
	c.Token = token
	return c, nil
}
