// Package fetcher renders web pages with a shared headless Chrome instance.
package fetcher

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const fetchAttempts = 2

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"

// Result is a single rendered page.
type Result struct {
	HTML     string
	Status   int
	FinalURL string
}

type tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	status atomic.Int32
}

// Fetcher keeps one browser process and hands its tabs out to concurrent
// workers. Close must only be called after all Fetch calls have returned.
type Fetcher struct {
	tabs        chan *tab
	timeout     time.Duration
	allocCancel context.CancelFunc
}

// New starts Chrome and prepares the given number of tabs. An error is
// returned when no browser binary can be found.
func New(workers int, timeout time.Duration) (*Fetcher, error) {
	if workers < 1 {
		workers = 1
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		chromedp.WindowSize(1280, 900),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	probeCtx, probeCancel := chromedp.NewContext(allocCtx)
	err := chromedp.Run(probeCtx, chromedp.Navigate("about:blank"))
	probeCancel()
	if err != nil {
		allocCancel()
		return nil, fmt.Errorf("start chrome: %w", err)
	}

	f := &Fetcher{
		tabs:        make(chan *tab, workers),
		timeout:     timeout,
		allocCancel: allocCancel,
	}
	for i := 0; i < workers; i++ {
		ctx, cancel := chromedp.NewContext(allocCtx)
		f.tabs <- &tab{ctx: ctx, cancel: cancel}
	}
	return f, nil
}

// Fetch loads a page and returns its rendered HTML. A page that fails to
// load is retried once before the error is handed back to the caller.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Result, error) {
	var err error
	for attempt := 0; attempt < fetchAttempts; attempt++ {
		var res *Result
		res, err = f.fetchOnce(ctx, rawURL)
		if err == nil {
			return res, nil
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil, err
}

// Close shuts the browser down.
func (f *Fetcher) Close() {
	for len(f.tabs) > 0 {
		t := <-f.tabs
		t.cancel()
	}
	f.allocCancel()
}

func (f *Fetcher) fetchOnce(ctx context.Context, rawURL string) (*Result, error) {
	t, err := f.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer f.release(t)

	navCtx, cancel := context.WithTimeout(t.ctx, f.timeout)
	defer cancel()

	var html, finalURL string
	err = chromedp.Run(navCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.once.Do(func() {
				chromedp.ListenTarget(t.ctx, func(ev interface{}) {
					resp, ok := ev.(*network.EventResponseReceived)
					if !ok || resp.Response == nil || resp.Type != network.ResourceTypeDocument {
						return
					}
					t.status.Store(int32(resp.Response.Status))
				})
			})
			t.status.Store(0)
			return nil
		}),
		network.Enable(),
		chromedp.Navigate(rawURL),
		waitComplete(),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.Location(&finalURL),
	)
	if err != nil {
		return nil, fmt.Errorf("render %s: %w", rawURL, err)
	}

	return &Result{
		HTML:     html,
		Status:   int(t.status.Load()),
		FinalURL: finalURL,
	}, nil
}

func (f *Fetcher) acquire(ctx context.Context) (*tab, error) {
	select {
	case t := <-f.tabs:
		return t, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (f *Fetcher) release(t *tab) {
	f.tabs <- t
}

// waitComplete blocks until the document has finished loading. Errors are
// ignored while the page is still being set up, the timeout decides when
// a page counts as stuck.
func waitComplete() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for {
			var state string
			err := chromedp.Evaluate("document.readyState", &state).Do(ctx)
			if err == nil && state == "complete" {
				return nil
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("page did not finish loading: %w", ctx.Err())
			case <-time.After(150 * time.Millisecond):
			}
		}
	})
}
