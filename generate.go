package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
	"strings"
	"time"
)

var (
	messageLen int
)

var GenerateCmd = &cobra.Command{
	Use: "generate",
	Run: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().Unix())

		var err error
		var seed string

		if len(args) > 0 && len(args) != params.PrefixLen {
			log.Error("Need ", params.PrefixLen, " words as seed")
			return
		} else if len(args) == params.PrefixLen {
			seed = strings.Join(args, " ")
		} else {
			seed, err = m.GetSeed()
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Info("Using as seed ", seed)

		text, err := m.Generate(seed, messageLen)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(text)
	},
}

func init() {
	GenerateCmd.PersistentFlags().IntVar(&messageLen, "len", 200, "Message length")

	RootCmd.AddCommand(GenerateCmd)
}
