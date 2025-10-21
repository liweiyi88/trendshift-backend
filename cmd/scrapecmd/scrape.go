package scrapecmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/global"
	"github.com/liweiyi88/trendshift-backend/scrape"
	"github.com/liweiyi88/trendshift-backend/search"
	"github.com/spf13/cobra"
)

var ScrapeCmd = &cobra.Command{
	Use:   "scrape [repository|developer]",
	Short: "Scrape trending repositories or trending developers form GitHub trending page.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		config.Init()

		search := search.NewSearch()
		ctx, stop := context.WithCancel(context.Background())
		db := database.GetInstance(ctx)
		repositories := global.InitRepositories(db)

		tokenPool := github.NewTokenPool(config.GitHubTokens)
		gh := github.NewClient(tokenPool)
		handler := scrape.NewScrapeHandler(repositories, search, gh)

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

		err := handler.Handle(ctx, action)
		if err != nil {
			slog.Error("failed to handle action", slog.Any("error", err))
			sentry.CaptureException(err)
		}
	},
}
