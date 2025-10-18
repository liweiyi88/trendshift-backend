package ingestcmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/liweiyi88/trendshift-backend/config"
	"github.com/liweiyi88/trendshift-backend/database"
	"github.com/liweiyi88/trendshift-backend/github"
	"github.com/liweiyi88/trendshift-backend/ingestion"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/spf13/cobra"
)

var ingestMonthlyRepositoryDataCmd = &cobra.Command{
	Use:   "monthly-repository-data",
	Short: "Fetch, aggregate, and save monthly GitHub repo data",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Init()

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

		rmr := model.NewRepositoryMonthlyInsightRepo(db)
		ingestor := ingestion.NewMonthlyRepoDataIngestor(rmr, gh)

		now := time.Now()
		return ingestor.Ingest(ctx, int(now.Month()), now.Year())
	},
}
