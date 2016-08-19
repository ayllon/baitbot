package main

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
)

var (
	limit uint64
	count uint64
)

func processLink(wg *sync.WaitGroup, parent *url.URL, node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			link := attr.Val
			if parsed, err := url.Parse(link); err == nil {
				resolved := parent.ResolveReference(parsed)
				if resolved.Host != parent.Host {
					logrus.Debug("Ignore link to ", resolved)
				} else if atomic.AddUint64(&count, 1) <= limit {
					wg.Add(1)
					go func() {
						defer wg.Done()
						processPage(wg, resolved)
					}()
				}
			}
		}
	}
}

func extractText(node *html.Node) string {
	buffer := bytes.NewBufferString("")
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			buffer.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			buffer.WriteString(extractText(c))
		}
	}
	return buffer.String()
}

func processParagraph(node *html.Node) {
	text := extractText(node)
	if text != "" {
		processText(text)
	}
}

func processPage(wg *sync.WaitGroup, resource *url.URL) {
	logrus.Info("Processing ", resource)

	response, err := http.Get(resource.String())
	if err != nil {
		logrus.Error(err)
		return
	}
	defer response.Body.Close()

	doc, err := html.Parse(response.Body)
	if err != nil {
		logrus.Error(err)
		return
	}

	var processElement func(*html.Node)
	processElement = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "a":
				processLink(wg, resource, node)
			case "p":
				processParagraph(node)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			processElement(c)
		}
	}
	processElement(doc)
}

var SpiderCmd = &cobra.Command{
	Use: "spider url [url2 [url3]...]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Expecting at least one url")
			cmd.Usage()
			return
		}

		wg := sync.WaitGroup{}
		for _, address := range args {
			wg.Add(1)
			go func(address string) {
				defer wg.Done()
				resource, err := url.Parse(address)
				if err != nil {
					logrus.Error(err)
				} else {
					processPage(&wg, resource)
				}
			}(address)
		}
		wg.Wait()
	},
}

func init() {
	SpiderCmd.PersistentFlags().Uint64Var(&limit, "limit", 100, "Limit")

	MarkovCmd.AddCommand(SpiderCmd)
}
