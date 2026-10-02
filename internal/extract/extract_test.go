package extract

import (
	"reflect"
	"testing"

	"github.com/Masih-Ghasri/CaptureCrawl/internal/model"
)

const fullPage = `<!DOCTYPE html>
<html>
<head>
<title>Docs Home</title>
<style>.x { color: red; }</style>
</head>
<body>
<header><a href="/">Home</a></header>
<nav><a href="/docs">Docs</a></nav>
<main>
<script>alert(1)</script>
<h1>Welcome</h1>
<p>Intro text with <strong>bold</strong> and <a href="/about">a link</a>.</p>
<h2>Features</h2>
<ul>
<li>Fast</li>
<li>Safe
<ul><li>Isolated</li></ul>
</li>
</ul>
<table>
<tr><th>Name</th><th>Value</th></tr>
<tr><td>one</td><td>1</td></tr>
</table>
<pre><code>go build ./...</code></pre>
<p>Footer paragraph</p>
</main>
<footer>All rights reserved</footer>
</body>
</html>`

func TestExtractContent(t *testing.T) {
	page := Extract("https://example.com/", fullPage, 0, "")

	if page.Title != "Docs Home" {
		t.Errorf("Title = %q, want %q", page.Title, "Docs Home")
	}

	want := []model.Block{
		{Kind: model.BlockHeading, Level: 1, Text: "Welcome"},
		{Kind: model.BlockParagraph, Text: "Intro text with bold and a link."},
		{Kind: model.BlockHeading, Level: 2, Text: "Features"},
		{Kind: model.BlockList, Text: "Fast\nSafe\n  Isolated"},
		{Kind: model.BlockTable, Rows: [][]string{{"Name", "Value"}, {"one", "1"}}},
		{Kind: model.BlockCode, Text: "go build ./..."},
		{Kind: model.BlockParagraph, Text: "Footer paragraph"},
	}

	if !reflect.DeepEqual(page.Content, want) {
		t.Errorf("Content = %#v\nwant %#v", page.Content, want)
	}
}

func TestExtractSkipsPageChrome(t *testing.T) {
	page := Extract("https://example.com/", fullPage, 0, "")

	for _, block := range page.Content {
		if block.Text == "All rights reserved" || block.Text == "Docs" {
			t.Errorf("page chrome leaked into content: %q", block.Text)
		}
	}
}

func TestExtractLinks(t *testing.T) {
	const html = `<body>
<a href="mailto:x@y.z">mail</a>
<a href="/page">page</a>
<a href="#top">top</a>
<a href="https://other.com/o">other</a>
<a href="/page#x">dup</a>
</body>`

	page := Extract("https://example.com/start", html, 0, "")

	want := []string{
		"https://example.com/page",
		"https://example.com/start",
		"https://other.com/o",
	}
	if !reflect.DeepEqual(page.Links, want) {
		t.Errorf("Links = %v, want %v", page.Links, want)
	}
}

func TestExtractLooseTextAndInline(t *testing.T) {
	const html = `<body>
<div>Hello <a href="/x">world</a> again<br>next</div>
<span>loose</span>
</body>`

	page := Extract("https://example.com/", html, 2, "https://example.com/")

	want := []model.Block{
		{Kind: model.BlockParagraph, Text: "Hello world again next"},
		{Kind: model.BlockParagraph, Text: "loose"},
	}
	if !reflect.DeepEqual(page.Content, want) {
		t.Errorf("Content = %#v\nwant %#v", page.Content, want)
	}
	if page.Depth != 2 || page.Parent != "https://example.com/" {
		t.Errorf("Depth/Parent = %d/%q, want 2/https://example.com/", page.Depth, page.Parent)
	}
}

func TestExtractTitleFallsBackToHeading(t *testing.T) {
	page := Extract("https://example.com/", `<body><h1>Main Heading</h1></body>`, 0, "")

	if page.Title != "Main Heading" {
		t.Errorf("Title = %q, want %q", page.Title, "Main Heading")
	}
}

func TestExtractEmptyDocument(t *testing.T) {
	page := Extract("https://example.com/", "", 0, "")

	if page.Error != "" {
		t.Errorf("Error = %q, want none", page.Error)
	}
	if len(page.Content) != 0 {
		t.Errorf("Content = %#v, want empty", page.Content)
	}
}
