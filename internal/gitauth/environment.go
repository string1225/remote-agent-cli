// Package gitauth manages local GitHub authorization for Windows service tasks.
package gitauth

import (
	"runtime"
	"strings"

	"github.com/string1225/remote-agent-cli/internal/serviceaccess"
)

// Environment never contains a credential. Git's own -c environment format
// survives provider filtering of *_KEY_* variables, unlike GIT_CONFIG_COUNT.
// Host-scoped overrides leave other hosts and the user's Git files alone.
func Environment(base []string, store string, interactive bool) []string {
	values := map[string]string{
		"GCM_CREDENTIAL_STORE": "dpapi", "GCM_DPAPI_STORE_PATH": store,
		"GCM_INTERACTIVE": "never", "GIT_TERMINAL_PROMPT": "0",
		"GIT_ASKPASS": "", "SSH_ASKPASS": "",
	}
	if runtime.GOOS != "windows" {
		delete(values, "GCM_CREDENTIAL_STORE")
		delete(values, "GCM_DPAPI_STORE_PATH")
	}
	if interactive {
		values["GCM_INTERACTIVE"] = "always"
		// GCM checks this independently of GCM_INTERACTIVE, including when
		// --device --no-ui is used. Only the explicit local login may prompt.
		values["GIT_TERMINAL_PROMPT"] = "1"
	} else {
		// Git executes askpass through its bundled shell. A fixed failing command
		// replaces desktop GUI helpers and provides an actionable service error.
		// The trailing ':' consumes Git's prompt argument without printing it.
		values["GIT_ASKPASS"] = "echo 'Background Git authentication is required. For GitHub, run remote-agent git login in the device owner terminal.' >&2; exit 1; :"
	}
	var parameters string
	env := make([]string, 0, len(base)+5)
	for _, entry := range base {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		key = strings.ToUpper(key)
		// GCM traces can include device codes or secrets when explicitly enabled.
		// Authorization output belongs only to the requesting local terminal.
		if strings.HasPrefix(key, "GCM_TRACE") || strings.HasPrefix(key, "GIT_TRACE") || key == "GIT_CURL_VERBOSE" {
			continue
		}
		if key == "GIT_CONFIG_PARAMETERS" {
			parameters = value
			continue
		}
		if _, replace := values[key]; !replace {
			env = append(env, entry)
		}
	}
	for _, pair := range []string{"credential.https://github.com.helper=", "credential.https://github.com.helper=manager", "core.askPass="} {
		parameters += " '" + pair + "'"
	}
	parameters += serviceaccess.IncludeParameter()
	for _, key := range []string{"GCM_CREDENTIAL_STORE", "GCM_DPAPI_STORE_PATH", "GCM_INTERACTIVE", "GIT_TERMINAL_PROMPT", "GIT_ASKPASS", "SSH_ASKPASS"} {
		if value, ok := values[key]; ok {
			env = append(env, key+"="+value)
		}
	}
	return append(env, "GIT_CONFIG_PARAMETERS="+strings.TrimSpace(parameters))
}
