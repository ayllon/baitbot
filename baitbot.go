package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	debug bool
)

var RootCmd = &cobra.Command{
	Use: "baitbot",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		logrus.SetOutput(os.Stderr)
	},
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Increase verbosity")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Error(err)
	}
}
