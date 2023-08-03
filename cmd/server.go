package cmd

import (
	"github.com/liweiyi88/gti/config"
	"github.com/liweiyi88/gti/web"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the API server",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()
		web.Server()
	},
}
