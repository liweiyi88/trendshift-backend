package ingestcmd

import (
	"context"
	"fmt"

	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/utils/datetime"
	"github.com/spf13/cobra"
)

var ingestMonthlyRepositoryDataCmd = &cobra.Command{
	Use:   "monthly-repository-data",
	Short: "Fetch, aggregate, and save monthly GitHub repo data",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Init()
		client := github.NewClient(config.GitHubToken)

		stargazers := make([]github.Stargazer, 0)
		var cursor *string

		startOfThisMonth := datetime.StartOfThisMonth()
		endOfThisMonth := datetime.EndOfThisMonth()

		for {
			data, nextCursor, err := client.GetRepositoryStars(
				context.Background(),
				"liweiyi88",
				"trendshift",
				cursor,
				&startOfThisMonth,
				&endOfThisMonth)

			if err != nil {
				return fmt.Errorf("failed to get repository stars, error: %v", err)
			}

			stargazers = append(stargazers, data...)
			if nextCursor == nil {
				break
			}

			cursor = nextCursor
		}

		fmt.Printf("Total stargazers fetched: %d\n", len(stargazers))
		return nil
	},
}
