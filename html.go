package main

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
)

func extractTextFromParagraph(node *html.Node, out io.Writer) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			out.Write([]byte(c.Data))
		} else if c.Type == html.ElementNode {
			extractTextFromNode(c, out)
		}
	}
	out.Write([]byte(" "))
}

func extractTextFromNode(node *html.Node, out io.Writer) {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			if c.Data == "p" {
				extractTextFromParagraph(c, out)
			} else {
				extractTextFromNode(c, out)
			}
		}
	}
}

func extractTextFromHtml(r io.Reader) (string, error) {
	node, err := html.Parse(r)
	if err != nil {
		return "", err
	}
	buffer := bytes.NewBufferString("")
	extractTextFromNode(node, buffer)
	return buffer.String(), nil
}
