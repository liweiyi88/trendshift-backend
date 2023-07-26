package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var queueUrl string

var rootCmd = &cobra.Command{}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
