// Package extract turns rendered HTML into structured page data.
package extract

import (
	"strings"

	"github.com/Masih-Ghasri/CaptureCrawl/internal/frontier"
	"github.com/Masih-Ghasri/CaptureCrawl/internal/model"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Extract parses a rendered document and builds a Page from it.
func Extract(pageURL, rawHTML string, depth int, parent string) *model.Page {
	page := &model.Page{
		URL:    pageURL,
		Depth:  depth,
		Parent: parent,
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		page.Error = err.Error()
		return page
	}

	page.Title = titleOf(doc)
	page.Content = contentOf(doc)
	page.Links = linksOf(doc, pageURL)
	return page
}

func titleOf(doc *goquery.Document) string {
	if title := cleanText(doc.Find("title").First().Text()); title != "" {
		return title
	}
	return cleanText(doc.Find("h1").First().Text())
}

// rootOf prefers the page's main content area so that navigation and
// chrome around it stay out of the extracted text.
func rootOf(doc *goquery.Document) *goquery.Selection {
	for _, selector := range []string{"main", "[role=main]", "body"} {
		if sel := doc.Find(selector).First(); sel.Length() > 0 {
			return sel
		}
	}
	return doc.Selection
}

func linksOf(doc *goquery.Document, base string) []string {
	seen := make(map[string]struct{})
	var links []string

	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		abs, err := frontier.ResolveRef(base, href)
		if err != nil {
			return
		}
		if _, dup := seen[abs]; dup {
			return
		}
		seen[abs] = struct{}{}
		links = append(links, abs)
	})

	return links
}

// walk visits the direct content of a node in document order. Loose text
// and inline markup are collected into paragraphs, block level children
// are dispatched on their own.
func walk(sel *goquery.Selection, out *[]model.Block) {
	var buf strings.Builder

	sel.Contents().Each(func(_ int, s *goquery.Selection) {
		node := s.Get(0)
		if node == nil {
			return
		}
		switch node.Type {
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			tag := strings.ToLower(node.Data)
			if isInline(tag) {
				buf.WriteString(" ")
				buf.WriteString(s.Text())
				buf.WriteString(" ")
				return
			}
			flushText(&buf, out)
			processBlock(tag, s, out)
		}
	})

	flushText(&buf, out)
}

func processBlock(tag string, sel *goquery.Selection, out *[]model.Block) {
	switch {
	case isSkipped(tag):
		return

	case tag == "p":
		if text := cleanText(sel.Text()); text != "" {
			*out = append(*out, model.Block{Kind: model.BlockParagraph, Text: text})
		}

	case len(tag) == 2 && tag[0] == 'h' && tag[1] >= '1' && tag[1] <= '6':
		if text := cleanText(sel.Text()); text != "" {
			*out = append(*out, model.Block{
				Kind:  model.BlockHeading,
				Level: int(tag[1] - '0'),
				Text:  text,
			})
		}

	case tag == "ul" || tag == "ol":
		var items []string
		collectListItems(sel, 0, &items)
		if len(items) > 0 {
			*out = append(*out, model.Block{
				Kind: model.BlockList,
				Text: strings.Join(items, "\n"),
			})
		}

	case tag == "table":
		if rows := tableRows(sel); len(rows) > 0 {
			*out = append(*out, model.Block{Kind: model.BlockTable, Rows: rows})
		}

	case tag == "pre":
		if text := strings.TrimSpace(sel.Text()); text != "" {
			*out = append(*out, model.Block{Kind: model.BlockCode, Text: text})
		}

	default:
		walk(sel, out)
	}
}

func flushText(buf *strings.Builder, out *[]model.Block) {
	if text := cleanText(buf.String()); text != "" {
		*out = append(*out, model.Block{Kind: model.BlockParagraph, Text: text})
	}
	buf.Reset()
}

func tableRows(table *goquery.Selection) [][]string {
	var rows [][]string

	table.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		var cells []string
		tr.Children().Each(func(_ int, cell *goquery.Selection) {
			node := cell.Get(0)
			if node == nil || node.Type != html.ElementNode {
				return
			}
			if tag := strings.ToLower(node.Data); tag == "td" || tag == "th" {
				cells = append(cells, cleanText(cell.Text()))
			}
		})
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	})

	return rows
}
