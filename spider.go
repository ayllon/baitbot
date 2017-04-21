package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	limit        uint64
	count        uint64
	sleep        time.Duration
	visited      = make(map[string]bool)
	visitedMutex sync.Mutex
)

func visitLink(wg *sync.WaitGroup, link *url.URL) {
	visitedMutex.Lock()
	defer visitedMutex.Unlock()

	if count >= limit {
		log.Warn("Limit reached")
	} else if !visited[link.String()] {
		visited[link.String()] = true
		count++
		wg.Add(1)
		go func() {
			defer wg.Done()
			processPage(wg, link)
		}()
	} else {
		log.Warn("Page already visited: ", link)
	}
}

func processHref(wg *sync.WaitGroup, parent *url.URL, node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			link := attr.Val
			if parsed, err := url.Parse(link); err == nil {
				resolved := parent.ResolveReference(parsed)
				if resolved.Host != parent.Host {
					log.Debug("Ignore link to ", resolved)
				} else {
					visitLink(wg, resolved)
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
	if text := extractText(node); text != "" {
		if err := m.ProcessText(text); err != nil {
			log.Error(err)
		}
	}
}

func processPage(wg *sync.WaitGroup, resource *url.URL) {
	log.Info("Processing ", resource)
	time.Sleep(sleep)

	response, err := http.Get(resource.String())
	if err != nil {
		log.Error(err)
		return
	}
	defer response.Body.Close()

	doc, err := html.Parse(response.Body)
	if err != nil {
		log.Error(err)
		return
	}

	var processElement func(*html.Node)
	processElement = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "a":
				processHref(wg, resource, node)
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
					log.Error(err)
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
	SpiderCmd.PersistentFlags().DurationVar(&sleep, "sleep", 0, "Sleep between requests")

	ImportCmd.AddCommand(SpiderCmd)
}
