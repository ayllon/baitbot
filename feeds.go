package main

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/SlyMarbo/rss"
	"github.com/dustin/gojson"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

var (
	predefinedFeeds = []string{
		"http://feeds.reuters.com/Reuters/PoliticsNews?format=xml",
		"http://feeds.skynews.com/feeds/rss/world.xml",
		"http://feeds.bbci.co.uk/news/politics/rss.xml",
		"http://rss.cnn.com/rss/edition_world.rss",
	}
)

func stripHtml(in string) string {
	inBuffer := bytes.NewBufferString(in)
	outBuffer := bytes.NewBuffer([]byte{})

	doc, err := html.Parse(inBuffer)
	if err != nil {
		logrus.Warn(err)
		return in
	}
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			outBuffer.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(doc)

	return outBuffer.String()
}

var FeedCmd = &cobra.Command{
	Use: "feed",
	Run: func(cmd *cobra.Command, args []string) {
		feedsUrls := args
		if len(feedsUrls) == 0 {
			feedsUrls = predefinedFeeds
		}

		for _, feedUrl := range feedsUrls {
			feed, err := rss.Fetch(feedUrl)
			if err != nil {
				logrus.Error(err)
				continue
			}

			for _, item := range feed.Items {
				data, err := json.Marshal(item.Title)
				if err != nil {
					logrus.Error(err)
					continue
				}
				fmt.Println(string(data))

				body := item.Content
				if body == "" {
					body = item.Summary
				}

				body = stripHtml(body)
				data, err = json.Marshal(body)
				if err != nil {
					logrus.Error(err)
					continue
				}
				fmt.Println(string(data))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(FeedCmd)
}
