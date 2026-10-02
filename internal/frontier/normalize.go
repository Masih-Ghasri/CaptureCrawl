package frontier

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

var trackingParams = []string{
	"gclid",
	"fbclid",
	"yclid",
	"mc_cid",
	"mc_eid",
	"utm_id",
}

// Normalize reduces a URL to its canonical form so that the same page
// reached through different links counts as a single visit.
func Normalize(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", raw, err)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q in %q", u.Scheme, raw)
	}
	if u.Host == "" {
		return "", fmt.Errorf("missing host in %q", raw)
	}

	u.Host = strings.ToLower(u.Host)
	switch u.Scheme {
	case "http":
		u.Host = strings.TrimSuffix(u.Host, ":80")
	case "https":
		u.Host = strings.TrimSuffix(u.Host, ":443")
	}

	if u.Path == "" {
		u.Path = "/"
	}
	u.Path = path.Clean(u.Path)
	u.RawPath = ""
	u.Fragment = ""

	q := u.Query()
	for key := range q {
		if isTrackingParam(key) {
			q.Del(key)
		}
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ResolveRef resolves a link found in a page against that page's URL and
// normalizes the result. Non-http links return an error and should be
// skipped by the caller.
func ResolveRef(base, ref string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse base %q: %w", base, err)
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("parse ref %q: %w", ref, err)
	}
	return Normalize(b.ResolveReference(r).String())
}

func isTrackingParam(key string) bool {
	key = strings.ToLower(key)
	if strings.HasPrefix(key, "utm_") {
		return true
	}
	for _, param := range trackingParams {
		if key == param {
			return true
		}
	}
	return false
}
