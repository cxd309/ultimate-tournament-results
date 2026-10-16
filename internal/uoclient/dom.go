package uoclient

import (
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// walk visits n and every descendant in document order
// returning false from fn skips that node's children
func walk(n *html.Node, fn func(*html.Node) bool) {
	if !fn(n) {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, fn)
	}
}

// findAll returns every element under n (n included) that match reports true for
func findAll(n *html.Node, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	walk(n, func(c *html.Node) bool {
		if c.Type == html.ElementNode && match(c) {
			found = append(found, c)
		}
		return true
	})
	return found
}

// find returns the first element under n (n included) match reports true for, or nil
func find(n *html.Node, match func(*html.Node) bool) *html.Node {
	var found *html.Node
	walk(n, func(c *html.Node) bool {
		if found != nil {
			return false
		}
		if c.Type == html.ElementNode && match(c) {
			found = c
			return false
		}
		return true
	})
	return found
}

func tag(name string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Data == name }
}

func tagClass(name, class string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Data == name && hasClass(n, class) }
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, class string) bool {
	return slices.Contains(strings.Fields(attr(n, "class")), class)
}

// children returns n's direct element children with the given tag
func children(n *html.Node, name string) []*html.Node {
	var found []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == name {
			found = append(found, c)
		}
	}
	return found
}

// rows returns a table's rows, looking through an implied tbody
func rows(table *html.Node) []*html.Node {
	var found []*html.Node
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		switch {
		case c.Type != html.ElementNode:
		case c.Data == "tr":
			found = append(found, c)
		case c.Data == "tbody" || c.Data == "thead":
			found = append(found, children(c, "tr")...)
		}
	}
	return found
}

// rawText is n's text content exactly as parsed, entities already decoded
func rawText(n *html.Node) string {
	var b strings.Builder
	walk(n, func(c *html.Node) bool {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
		return c.Type != html.ElementNode || (c.Data != "script" && c.Data != "style")
	})
	return b.String()
}

// scriptText is the source of a <script> element, which rawText leaves out
func scriptText(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

var spaceRun = regexp.MustCompile(`\s+`)

// text is n's text content with non-breaking spaces as plain spaces, whitespace
// runs collapsed and the ends trimmed
func text(n *html.Node) string {
	return clean(rawText(n))
}

func clean(s string) string {
	s = strings.ReplaceAll(s, " ", " ")
	return strings.TrimSpace(spaceRun.ReplaceAllString(s, " "))
}

// linkParams parses an UltiOrganizer link's query string, e.g.
// "?view=teamcard&team=57" -> view=teamcard, team=57
func linkParams(href string) url.Values {
	if i := strings.Index(href, "?"); i >= 0 {
		href = href[i+1:]
	}
	v, _ := url.ParseQuery(href)
	return v
}

// links returns every <a> under n whose href is an index.php view of the given name
func links(n *html.Node, view string) []*html.Node {
	return findAll(n, func(c *html.Node) bool {
		return c.Data == "a" && linkParams(attr(c, "href")).Get("view") == view
	})
}

// linkID is the integer value of param in a's href, false if absent or not a number
func linkID(a *html.Node, param string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(linkParams(attr(a, "href")).Get(param)), 10, 64)
	return n, err == nil
}

// firstLinkID is linkID of the first link under n to view, false if there is none
func firstLinkID(n *html.Node, view, param string) (int64, bool) {
	for _, a := range links(n, view) {
		if id, ok := linkID(a, param); ok {
			return id, true
		}
	}
	return 0, false
}

// content is the page's main content div, the part after the left menu
func content(doc *html.Node) *html.Node {
	return find(doc, tagClass("div", "content"))
}

// parseInt parses an optionally signed integer, false for anything else (including "")
func parseInt(s string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n, err == nil
}

// optionalInt is parseInt as a pointer, nil when s isn't a number
func optionalInt(s string) *int64 {
	if n, ok := parseInt(s); ok {
		return &n
	}
	return nil
}

var colorPattern = regexp.MustCompile(`background-color:#([0-9A-Fa-f]{6})`)

// backgroundColor is the first "background-color:#RRGGBB" in n's style, "" if none
func backgroundColor(n *html.Node) string {
	if m := colorPattern.FindStringSubmatch(attr(n, "style")); m != nil {
		return strings.ToUpper(m[1])
	}
	return ""
}

// secondsFromMinutes parses UltiOrganizer's SecToMin "M.SS" game-clock format, where
// M can run past 59, e.g. "84.05" -> 5045
func secondsFromMinutes(s string) (int64, bool) {
	minutes, seconds, ok := strings.Cut(strings.TrimSpace(s), ".")
	if !ok || len(seconds) != 2 {
		return 0, false
	}
	m, err1 := strconv.ParseInt(minutes, 10, 64)
	sec, err2 := strconv.ParseInt(seconds, 10, 64)
	if err1 != nil || err2 != nil || sec > 59 {
		return 0, false
	}
	return m*60 + sec, true
}
