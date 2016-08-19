package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
	"github.com/dustin/gojson"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
)

type (
	Authn struct {
		ConsumerKey       string `yaml:"consumer-key"`
		ConsumerSecret    string `yaml:"consumer-secret"`
		AccessToken       string `yaml:"access-token"`
		AccessTokenSecret string `yaml:"access-token-secret"`
	}
)

const (
	USA = 23424977
	UK  = 23424975
)

var (
	authnFile string
	api       *anaconda.TwitterApi
	authn     Authn

	weoeID     int64
	searchTerm string
)

func readConfig() {
	fd, err := os.Open(authnFile)
	if err != nil {
		logrus.Fatal(err)
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		logrus.Fatal(err)
	}
	err = yaml.Unmarshal(data, &authn)
	if err != nil {
		logrus.Fatal(err)
	}
}

var TwitterCmd = &cobra.Command{
	Use: "twitter",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)

		readConfig()
		anaconda.SetConsumerKey(authn.ConsumerKey)
		anaconda.SetConsumerSecret(authn.ConsumerSecret)
		api = anaconda.NewTwitterApi(authn.AccessToken, authn.AccessTokenSecret)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		api.Close()
	},
}

var TrendingCmd = &cobra.Command{
	Use: "trending",
	Run: func(cmd *cobra.Command, args []string) {
		trends, err := api.GetTrendsByPlace(weoeID, url.Values{})
		if err != nil {
			logrus.Fatal(err)
		}

		for _, trend := range trends.Trends {
			fmt.Println(trend.Name)
		}
	},
}

var GetTweetsCmd = &cobra.Command{
	Use: "get",
	PreRun: func(cmd *cobra.Command, args []string) {
		if searchTerm == "" {
			trends, err := api.GetTrendsByPlace(weoeID, url.Values{})
			if err != nil {
				logrus.Fatal(err)
			}
			searchTerm = trends.Trends[0].Query
			logrus.Info("No search term specified, using ", searchTerm)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		search, err := api.GetSearch(searchTerm+" -RT", url.Values{"count": []string{"100"}})
		if err != nil {
			logrus.Fatal(err)
		}

		for _, tweet := range search.Statuses {
			bytes, err := json.Marshal(&tweet.Text)
			if err != nil {
				logrus.Fatal(err)
			}
			fmt.Println(string(bytes))
		}
	},
}

func init() {
	TwitterCmd.PersistentFlags().StringVar(&authnFile, "authn", "authn.yaml", "Configuration file with access tokens")
	TrendingCmd.PersistentFlags().Int64Var(&weoeID, "weoid", UK, "Where On Earth ID")
	GetTweetsCmd.PersistentFlags().StringVar(&searchTerm, "search", "", "Search term")

	TwitterCmd.AddCommand(TrendingCmd)
	TwitterCmd.AddCommand(GetTweetsCmd)
	RootCmd.AddCommand(TwitterCmd)
}
