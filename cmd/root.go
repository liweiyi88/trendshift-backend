package cmd

import (
	"fmt"
	"os"

	"github.com/liweiyi88/trendshift-backend/cmd/githubcmd"
	"github.com/liweiyi88/trendshift-backend/cmd/scrapecmd"
	"github.com/liweiyi88/trendshift-backend/cmd/searchcmd"
	"github.com/liweiyi88/trendshift-backend/cmd/usercmd"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func init() {
	rootCmd.AddCommand(searchcmd.SearchCmd)
	rootCmd.AddCommand(usercmd.UserCmd)
	rootCmd.AddCommand(githubcmd.GihtubSyncCmd)
	rootCmd.AddCommand(scrapecmd.ScrapeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
