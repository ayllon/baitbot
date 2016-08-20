package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ayllon/baitbot/markov"
	"github.com/spf13/cobra"
	"os"
)

var (
	debug      bool
	dbPath     string
	params     markov.Parameters
	m          *markov.Markov
)

var RootCmd = &cobra.Command{
	Use: "baitbot",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			log.SetLevel(log.DebugLevel)
		}
		log.SetOutput(os.Stderr)
		var err error
		m, err = markov.New(dbPath, params)
		if err != nil {
			log.Fatal(err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		m.Close()
	},
}

var ClearCmd = &cobra.Command{
	Use: "clear",
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Clearing the database")
		m.Clear()
	},
}

var ImportCmd = &cobra.Command{
	Use: "import",
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Increase verbosity")
	RootCmd.PersistentFlags().StringVar(&dbPath, "db", "/tmp/baitbot.db", "Chain database")
	RootCmd.PersistentFlags().BoolVar(&params.DryRun, "dry-run", false, "If set, there will be no insertions")
	RootCmd.PersistentFlags().IntVar(&params.PrefixLen, "prefix-len", 3, "Prefix length")

	RootCmd.AddCommand(ClearCmd)
	RootCmd.AddCommand(ImportCmd)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
