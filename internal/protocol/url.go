package protocol

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var basePathRE = regexp.MustCompile(`^(/[A-Za-z0-9_-]+)*$`)

// ParseServerURL accepts an HTTPS origin and optional application mount path.
// HTTP is restricted to loopback, and ambiguous/encoded paths are rejected.
func ParseServerURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawPath != "" || u.Opaque != "" {
		return nil, errors.New("server URL requires a host and optional base path, without credentials, query or fragment")
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	if !basePathRE.MatchString(u.Path) {
		return nil, errors.New("server base path must contain only letters, numbers, underscores, hyphens and single slashes")
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1")) {
		return nil, errors.New("HTTPS is required except on localhost")
	}
	return u, nil
}
