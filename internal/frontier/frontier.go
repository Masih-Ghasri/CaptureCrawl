// Package frontier keeps track of the URLs a crawl still has to visit.
package frontier

import "sync"

type Item struct {
	URL    string
	Depth  int
	Parent string
}

// Frontier is a concurrency safe work queue. URLs are expected to be
// normalized before they are pushed.
type Frontier struct {
	mu       sync.Mutex
	seen     map[string]struct{}
	pending  []Item
	accepted int
	maxPages int
	maxDepth int
}

func New(maxPages, maxDepth int) *Frontier {
	return &Frontier{
		seen:     make(map[string]struct{}),
		maxPages: maxPages,
		maxDepth: maxDepth,
	}
}

// Push queues a URL. It returns false when the URL was already visited,
// when the depth limit is exceeded or when the page budget is used up.
func (f *Frontier) Push(url string, depth int, parent string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.maxDepth > 0 && depth > f.maxDepth {
		return false
	}
	if f.accepted >= f.maxPages {
		return false
	}
	if _, ok := f.seen[url]; ok {
		return false
	}

	f.seen[url] = struct{}{}
	f.pending = append(f.pending, Item{URL: url, Depth: depth, Parent: parent})
	f.accepted++
	return true
}

// Pop returns the next URL in insertion order.
func (f *Frontier) Pop() (Item, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.pending) == 0 {
		return Item{}, false
	}
	item := f.pending[0]
	f.pending = f.pending[1:]
	return item, true
}

func (f *Frontier) Len() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.pending)
}

func (f *Frontier) Seen() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.seen)
}
