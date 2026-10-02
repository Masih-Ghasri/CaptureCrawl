package frontier

import (
	"fmt"
	"sync"
	"testing"
)

func TestFrontierSkipsSeenURLs(t *testing.T) {
	f := New(10, 0)

	if !f.Push("https://example.com/", 0, "") {
		t.Fatal("first push should succeed")
	}
	if f.Push("https://example.com/", 1, "https://example.com/other") {
		t.Error("second push of the same url should be rejected")
	}
	if f.Len() != 1 {
		t.Errorf("Len() = %d, want 1", f.Len())
	}
	if f.Seen() != 1 {
		t.Errorf("Seen() = %d, want 1", f.Seen())
	}
}

func TestFrontierRespectsMaxPages(t *testing.T) {
	f := New(2, 0)

	if !f.Push("https://example.com/1", 0, "") {
		t.Error("first push should succeed")
	}
	if !f.Push("https://example.com/2", 0, "") {
		t.Error("second push should succeed")
	}
	if f.Push("https://example.com/3", 0, "") {
		t.Error("push beyond max-pages should be rejected")
	}
}

func TestFrontierRespectsMaxDepth(t *testing.T) {
	f := New(10, 2)

	if !f.Push("https://example.com/", 2, "") {
		t.Error("push at max depth should succeed")
	}
	if f.Push("https://example.com/deep", 3, "https://example.com/") {
		t.Error("push beyond max depth should be rejected")
	}
}

func TestFrontierPopsInOrder(t *testing.T) {
	f := New(10, 0)

	f.Push("https://example.com/a", 0, "")
	f.Push("https://example.com/b", 1, "https://example.com/a")

	first, ok := f.Pop()
	if !ok {
		t.Fatal("expected an item")
	}
	if first.URL != "https://example.com/a" || first.Depth != 0 {
		t.Errorf("first item = %+v, want url https://example.com/a at depth 0", first)
	}

	second, ok := f.Pop()
	if !ok {
		t.Fatal("expected a second item")
	}
	if second.URL != "https://example.com/b" || second.Parent != "https://example.com/a" {
		t.Errorf("second item = %+v, want url https://example.com/b with parent https://example.com/a", second)
	}

	if _, ok := f.Pop(); ok {
		t.Error("frontier should be empty")
	}
}

func TestFrontierConcurrent(t *testing.T) {
	f := New(1000, 0)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				f.Push(fmt.Sprintf("https://example.com/p/%d", j), 0, "")
			}
		}()
	}
	wg.Wait()

	if f.Seen() != 100 {
		t.Errorf("Seen() = %d, want 100", f.Seen())
	}
	if f.Len() != 100 {
		t.Errorf("Len() = %d, want 100", f.Len())
	}
}
