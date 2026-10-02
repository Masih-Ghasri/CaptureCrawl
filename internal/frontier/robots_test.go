package frontier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const robotsContent = `User-agent: *
Disallow: /private/
Allow: /private/open.html

User-agent: Scraper
Disallow: /
`

func TestParseRobots(t *testing.T) {
	r := ParseRobots([]byte(robotsContent))

	tests := []struct {
		path string
		want bool
	}{
		{"/", true},
		{"/docs/guide", true},
		{"/private/", false},
		{"/private/data", false},
		{"/private/open.html", true},
	}

	for _, tt := range tests {
		if got := r.Allowed(tt.path); got != tt.want {
			t.Errorf("Allowed(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestParseRobotsEmpty(t *testing.T) {
	for _, content := range []string{"", "\x00 not really robots"} {
		r := ParseRobots([]byte(content))
		if !r.Allowed("/") {
			t.Error("unusable robots.txt should allow everything")
		}
	}
}

func TestNilRobotsAllowsEverything(t *testing.T) {
	var r *Robots
	if !r.Allowed("/anything") {
		t.Error("nil robots should allow everything")
	}
}

func TestFetchRobots(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprint(w, robotsContent)
	}))
	defer srv.Close()

	r := fetchRobotsFrom(t, srv)

	if r.Allowed("/private/") {
		t.Error("expected /private/ to be disallowed")
	}
	if !r.Allowed("/docs") {
		t.Error("expected /docs to be allowed")
	}
}

func TestFetchRobotsMissing(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	r := fetchRobotsFrom(t, srv)

	if !r.Allowed("/") {
		t.Error("missing robots.txt should allow everything")
	}
}

func TestFetchRobotsForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	r := fetchRobotsFrom(t, srv)

	if r.Allowed("/") {
		t.Error("403 on robots.txt should block the whole site")
	}
}

func fetchRobotsFrom(t *testing.T, srv *httptest.Server) *Robots {
	t.Helper()

	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	r, err := FetchRobots(context.Background(), srv.Client(), base)
	if err != nil {
		t.Fatalf("FetchRobots: %v", err)
	}
	return r
}
