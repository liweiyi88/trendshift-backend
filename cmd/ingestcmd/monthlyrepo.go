package ingestcmd

import (
	"context"
	"errors"
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
	"github.com/liweiyi88/trendshift-backend/logger"
	"github.com/liweiyi88/trendshift-backend/model"
	"github.com/liweiyi88/trendshift-backend/utils/datetime"
	"github.com/spf13/cobra"
)

var verbose bool

func init() {
	ingestMonthlyRepositoryDataCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "prints additional debug information (optional)")
}

var ingestMonthlyRepositoryDataCmd = &cobra.Command{
	Use:   "monthly-repository-data",
	Short: "Fetch, aggregate, and save monthly GitHub repo data",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Init()
		logger.InitSlog(verbose)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		db := database.GetInstance(ctx)

		defer func() {
			if err := db.Close(); err != nil {
				slog.Error("failed to close db", slog.Any("error", err))
				sentry.CaptureException(err)
			}

			stop()
			sentry.Flush(2 * time.Second)
		}()

		tokenPool := github.NewTokenPool(config.GitHubTokens)
		gh := github.NewClient(tokenPool)

		rmr := model.NewRepositoryMonthlyInsightRepo(db)
		ingestor := ingestion.NewMonthlyRepoDataIngestor(rmr, gh)

		for {
			now := time.Now()
			done, err := ingestor.Ingest(ctx, int(now.Month()), now.Year())

			if err != nil {
				if errors.Is(err, github.ErrTokenNotAvailable) {
					earliestResetAt := tokenPool.EarliestReset()

					slog.Warn("no GitHub tokens available, sleeping until earliest reset", slog.Time("reset_at", earliestResetAt))
					sleepDuration := time.Until(earliestResetAt)
					if sleepDuration > 0 {
						slog.Info("sleeping until tokens reset", slog.Duration("sleep", sleepDuration))
						if err := datetime.SleepWithContext(ctx, sleepDuration); err != nil {
							return err
						}
					}
				} else if errors.Is(err, github.ErrTooManyRequests) {
					slog.Warn("fetching repository monthly data with a github token was throttled.")
				} else {
					// Unhandled error, return and let command failed
					return err
				}
			}

			if done {
				sleepDuration := time.Until(datetime.StartOfTomorrow())
				if sleepDuration > 0 {
					slog.Info("Fetch jobs have been done, sleeping until start of tomorrow", slog.Duration("sleep", sleepDuration))
					if err := datetime.SleepWithContext(ctx, sleepDuration); err != nil {
						// Graceful shutdown will cancel the context, lets just return the ctx error.
						return err
					}
				}
			}
		}
	},
}
