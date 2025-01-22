package main

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// allowedTags is a map of tags that are allowed in the output
var allowedTags = map[string]bool{
	"b": true,
	// "i":          true,
	// "u":          true,
	// "s":          true,
	"a":          true,
	"code":       true,
	"blockquote": true,
}

// filterHTML parses the HTML and removes all tags except the allowed ones
func filterHTML(input string) string {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return input
	}

	// Buffer to store the cleaned HTML
	var buf bytes.Buffer
	w := io.Writer(&buf)

	// Recursively traverse the nodes and write only allowed and closed tags
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if allowedTags[n.Data] {
				// Write the opening tag
				buf.WriteString("<" + n.Data + ">")
				// Process children
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
				// Write the closing tag
				buf.WriteString("</" + n.Data + ">")
			} else {
				// If the tag is not allowed, only process its children
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
			}
		} else if n.Type == html.TextNode {
			_, _ = w.Write([]byte(n.Data))
		}
	}

	// Start traversal from the root node
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		traverse(c)
	}

	return buf.String()
}
