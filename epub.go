package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/meskio/epubgo"
	"github.com/spf13/cobra"
)

func processEpub(path string) {
	logrus.Info("Processing ", path)

	epub, err := epubgo.Open(path)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer epub.Close()

	spine, err := epub.Spine()
	if err != nil {
		logrus.Error(err)
	}

	for ; !spine.IsLast(); spine.Next() {
		fd, err := spine.Open()
		if err != nil {
			logrus.Error(err)
			continue
		}
		defer fd.Close()

		if text, err := extractTextFromHtml(fd); err != nil {
			logrus.Error(err)
		} else {
			m.ProcessText(text)
		}
	}
}

var EpubCmd = &cobra.Command{
	Use: "epub",
	Run: func(cmd *cobra.Command, args []string) {
		epubs := args
		if len(epubs) == 0 {
			logrus.Fatal("At least one epub is needed")
		}

		for _, epub := range epubs {
			processEpub(epub)
		}
	},
}

func init() {
	ImportCmd.AddCommand(EpubCmd)
}
