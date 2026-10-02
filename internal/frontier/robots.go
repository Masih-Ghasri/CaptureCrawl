package frontier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt"
)

const UserAgent = "CaptureCrawl"

const maxRobotsSize = 512 * 1024

// Robots holds the crawl rules of a single site. A nil or empty Robots
// allows every path.
type Robots struct {
	data    *robotstxt.RobotsData
	blocked bool
}

// ParseRobots reads robots.txt content. Content that cannot be parsed is
// treated as allowing every path, which matches how browsers behave.
func ParseRobots(content []byte) *Robots {
	data, err := robotstxt.FromBytes(content)
	if err != nil {
		return &Robots{}
	}
	return &Robots{data: data}
}

// Allowed reports whether UserAgent may fetch the given path.
func (r *Robots) Allowed(path string) bool {
	if r == nil || r.blocked {
		return false
	}
	if r.data == nil {
		return true
	}
	return r.data.TestAgent(path, UserAgent)
}

// FetchRobots loads robots.txt for the site behind base. A missing file
// allows everything, an authorization failure blocks the whole site.
func FetchRobots(ctx context.Context, client *http.Client, base *url.URL) (*Robots, error) {
	u := *base
	u.Path = "/robots.txt"
	u.RawQuery = ""
	u.Fragment = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build robots.txt request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		content, err := io.ReadAll(io.LimitReader(resp.Body, maxRobotsSize))
		if err != nil {
			return nil, fmt.Errorf("read robots.txt: %w", err)
		}
		return ParseRobots(content), nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return &Robots{blocked: true}, nil
	default:
		return &Robots{}, nil
	}
}
