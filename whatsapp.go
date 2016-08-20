package main

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func processWhatsappArchive(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		line = strings.TrimSpace(line)
		first := strings.Split(line, " - ")
		if len(first) > 1 {
			second := strings.Split(first[1], ": ")
			if len(second) > 1 {
				comment := second[1]
				m.ProcessText(comment)
			}
		}
	}
	return err
}

var WhatsappCmd = &cobra.Command{
	Use: "whatsapp",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Need at least one file")
			cmd.Usage()
			return
		}

		for _, path := range args {
			log.Info("Processing ", path)
			if err := processWhatsappArchive(path); err != nil {
				log.Error(err)
			}
		}
	},
}

func init() {
	ImportCmd.AddCommand(WhatsappCmd)
}
