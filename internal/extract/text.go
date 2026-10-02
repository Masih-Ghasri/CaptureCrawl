package extract

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func isInline(tag string) bool {
	switch tag {
	case "a", "abbr", "b", "br", "cite", "code", "del", "em", "i", "img",
		"ins", "label", "mark", "q", "s", "small", "span", "strong", "sub",
		"sup", "time", "u", "wbr":
		return true
	}
	return false
}

func isSkipped(tag string) bool {
	switch tag {
	case "script", "style", "noscript", "template", "head", "link", "meta",
		"title", "iframe", "object", "embed", "canvas", "audio", "video",
		"form", "button", "input", "select", "textarea", "svg":
		return true
	}
	return false
}

// collectListItems flattens a list into indented lines. Nested lists are
// appended after their parent item, one indent level deeper.
func collectListItems(list *goquery.Selection, depth int, items *[]string) {
	list.ChildrenFiltered("li").Each(func(_ int, li *goquery.Selection) {
		if text := itemText(li); text != "" {
			*items = append(*items, strings.Repeat("  ", depth)+text)
		}
		li.ChildrenFiltered("ul, ol").Each(func(_ int, nested *goquery.Selection) {
			collectListItems(nested, depth+1, items)
		})
	})
}

// itemText reads the text of a list item while leaving the item's own
// nested lists out, those are collected separately.
func itemText(li *goquery.Selection) string {
	var b strings.Builder

	li.Contents().Each(func(_ int, s *goquery.Selection) {
		node := s.Get(0)
		if node == nil {
			return
		}
		switch node.Type {
		case html.TextNode:
			b.WriteString(node.Data)
		case html.ElementNode:
			tag := strings.ToLower(node.Data)
			switch {
			case tag == "ul" || tag == "ol" || isSkipped(tag):
				return
			case tag == "br":
				b.WriteString(" ")
			default:
				b.WriteString(" ")
				b.WriteString(itemText(s))
				b.WriteString(" ")
			}
		}
	})

	return cleanText(b.String())
}
