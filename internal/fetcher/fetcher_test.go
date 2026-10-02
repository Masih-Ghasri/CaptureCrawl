package fetcher

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFetchRendersDocument(t *testing.T) {
	f, err := New(1, 15*time.Second)
	if err != nil {
		t.Skipf("chrome not available: %v", err)
	}
	defer f.Close()

	const page = "data:text/html,<html><head><title>Demo</title></head><body><h1>hello</h1></body></html>"

	res, err := f.Fetch(context.Background(), page)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(res.HTML, "<h1>hello</h1>") {
		t.Errorf("rendered html = %q, want it to contain the heading", res.HTML)
	}
	if res.FinalURL == "" {
		t.Error("FinalURL is empty")
	}
}

func TestFetchRecoversFromBadURL(t *testing.T) {
	f, err := New(1, 5*time.Second)
	if err != nil {
		t.Skipf("chrome not available: %v", err)
	}
	defer f.Close()

	if _, err := f.Fetch(context.Background(), "http://127.0.0.1:1/"); err == nil {
		t.Error("expected an error for an unreachable page")
	}

	// the tab must still be usable after a failure
	const page = "data:text/html,<html><body><p>ok</p></body></html>"
	res, err := f.Fetch(context.Background(), page)
	if err != nil {
		t.Fatalf("Fetch after failure: %v", err)
	}
	if !strings.Contains(res.HTML, "<p>ok</p>") {
		t.Errorf("rendered html = %q, want it to contain the paragraph", res.HTML)
	}
}
