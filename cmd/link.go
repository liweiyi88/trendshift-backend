package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/global"
	"github.com/liweiyi88/trendshift-backend/search"
	"github.com/liweiyi88/trendshift-backend/trending"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link [repository|developer]",
	Short: "Link trending repositories or developers with from GitHub repositories or developers",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		action := args[0]
		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)
		gh := github.NewClient(config.GitHubToken)

		defer func() {
			err := db.Close()

			if err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
				sentry.CaptureException(err)
			}

			stop()
			sentry.Flush(2 * time.Second)
		}()

		appSignal := make(chan os.Signal, 3)
		signal.Notify(appSignal, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-appSignal
			stop()
		}()

		slog.Info("linking repositories...")

		repositories := global.InitRepositories(db)
		search := search.NewSearch()
		githubFetcher := trending.NewGithubFetcher(gh, search, *repositories)

		var err error

		if action == "repository" {
			err = githubFetcher.FetchRepositories(ctx)
		} else if action == "developer" {
			err = githubFetcher.FetchDevelopers(ctx)
		} else {
			slog.Error("invalid action, expected repository or developer")
			return
		}

		if err != nil {
			slog.Error("failed to handle sync action", slog.Any("error", err))
			sentry.CaptureException(err)
		}
	},
}
